package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/mobiledgex/edge-cloud-infra/billing"
	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/mctestclient"
	ormtestutil "github.com/mobiledgex/edge-cloud-infra/mc/orm/testutil"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/cloudcommon/nodetest"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var MockESUrl = "http://mock.es"

type OPAttr struct {
	uri         string
	server      *Server
	dc          *grpc.Server
	ds          *testutil.DummyServer
	dcnt        int
	tokenAd     string
	tokenOper   string
	operatorId  string
	countryCode string
	fedAddr     string
	fedId       string
	zones       []ormapi.OPZoneInfo
}

func (o *OPAttr) CleanupOperatorPlatform(ctx context.Context) {
	// nodeMgr is a global object, hence copy the OP specific
	// nodeMgr for cleanup
	o.ds.SetDummyObjs(ctx, testutil.Delete, o.operatorId, o.dcnt)
	o.ds.SetDummyOrgObjs(ctx, testutil.Delete, o.operatorId, o.dcnt)
	o.server.Stop()
	o.dc.Stop()
}

func SetupOperatorPlatform(t *testing.T, ctx context.Context, operatorId, countryCode string) *OPAttr {
	vaultServer, vaultConfig := vault.DummyServer()
	defer vaultServer.Close()

	// any requests that don't have a registered URL will be fetched normally
	mockESUrl := "http://mock.es"
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
	testAlertMgrAddr, err := InitAlertmgrMock()
	require.Nil(t, err)
	de := &nodetest.DummyEventsES{}
	de.InitHttpMock(mockESUrl)

	// run a dummy http server to mimic influxdb
	// this will reply with empty json to everything
	influxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data":[{"Messages": null,"Series": null}]}`)
	}))
	defer influxServer.Close()
	// ==============================

	addr, err := cloudcommon.GetAvailablePort("127.0.0.1:0")
	require.Nil(t, err, "get available port")

	sqlAddr, err := cloudcommon.GetAvailablePort("127.0.0.1:0")
	require.Nil(t, err, "get available port")

	ctrlAddr, err := cloudcommon.GetAvailablePort("127.0.0.1:0")
	require.Nil(t, err, "get available port")

	fedAddr, err := cloudcommon.GetAvailablePort("127.0.0.1:0")
	require.Nil(t, err, "get available port")

	uri := "http://" + addr + "/api/v1"
	config := ServerConfig{
		ServAddr:                addr,
		SqlAddr:                 sqlAddr,
		FederationAddr:          fedAddr,
		RunLocal:                true,
		InitLocal:               true,
		SqlDataDir:              "./.postgres" + operatorId,
		IgnoreEnv:               true,
		SkipVerifyEmail:         true,
		vaultConfig:             vaultConfig,
		AlertMgrAddr:            testAlertMgrAddr,
		AlertmgrResolveTimout:   3 * time.Minute,
		UsageCheckpointInterval: "MONTH",
		BillingPlatform:         billing.BillingTypeFake,
		DeploymentTag:           "local",
		AlertCache:              &edgeproto.AlertCache{},
	}
	unitTestNodeMgrOps = []node.NodeOp{
		node.WithESUrls(MockESUrl),
	}
	defer func() {
		unitTestNodeMgrOps = []node.NodeOp{}
	}()

	server, err := RunServer(&config)
	require.Nil(t, err, "run server")

	Jwks.Init(vaultConfig, "region", "mcorm")
	Jwks.Meta.CurrentVersion = 1
	Jwks.Keys[1] = &vault.JWK{
		Secret:  "12345",
		Refresh: "1s",
	}

	// run dummy controller - this always returns success
	// to all APIs directed to it, and does not actually
	// create or delete objects. We are mocking it out
	// so we can test rbac permissions.
	dc := grpc.NewServer(
		grpc.UnaryInterceptor(testutil.UnaryInterceptor),
		grpc.StreamInterceptor(testutil.StreamInterceptor),
	)
	lis, err := net.Listen("tcp", ctrlAddr)
	require.Nil(t, err)
	ds := testutil.RegisterDummyServer(dc)
	go func() {
		dc.Serve(lis)
	}()

	// wait till mc is ready
	err = server.WaitUntilReady()
	require.Nil(t, err, "server online")

	defaultConfig.DisableRateLimit = true
	enforcer.LogEnforce(true)

	mcClient := mctestclient.NewClient(&ormclient.Client{})

	// Setup Controller, Orgs, Users
	// =============================
	// login as super user
	tokenAd, _, err := mcClient.DoLogin(uri, DefaultSuperuser, DefaultSuperpass, NoOTP, NoApiKeyId, NoApiKey)
	require.Nil(t, err, "login as superuser")

	// create controller
	ctrl := ormapi.Controller{
		Region:   countryCode,
		Address:  ctrlAddr,
		InfluxDB: influxServer.URL,
	}
	status, err := mcClient.CreateController(uri, tokenAd, &ctrl)
	require.Nil(t, err, "create controller")
	require.Equal(t, http.StatusOK, status)

	ctrls, status, err := mcClient.ShowController(uri, tokenAd, ClientNoShowFilter)
	require.Nil(t, err, "show controllers")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(ctrls))
	require.Equal(t, ctrl.Region, ctrls[0].Region)
	require.Equal(t, ctrl.Address, ctrls[0].Address)

	// create an operator
	_, _, tokenOper := testCreateUserOrg(t, mcClient, uri, "oper", "operator", operatorId)

	// admin allow non-edgebox cloudlets on operator org
	setOperatorOrgNoEdgeboxOnly(t, mcClient, uri, tokenAd, operatorId)

	// number of fake objects internally sent back by dummy server
	ds.ShowDummyCount = 0

	// number of dummy objects we add of each type and org
	dcnt := 3
	tag := operatorId
	ds.SetDummyObjs(ctx, testutil.Create, tag, dcnt)
	ds.SetDummyOrgObjs(ctx, testutil.Create, operatorId, dcnt)
	return &OPAttr{
		uri:         uri,
		server:      server,
		dc:          dc,
		ds:          ds,
		operatorId:  operatorId,
		countryCode: countryCode,
		tokenAd:     tokenAd,
		tokenOper:   tokenOper,
		fedAddr:     fedAddr,
	}
}

func getFederationAPI(fedAddr, fedApi string) string {
	return "http://" + fedAddr + fedApi
}

func registerFederationAPIs(t *testing.T, op1, op2 *OPAttr) {
	httpmock.RegisterResponder("POST", getFederationAPI(op2.fedAddr, F_API_OPERATOR_PARTNER),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			fedReq := ormapi.OPRegistrationRequest{}
			err = json.Unmarshal(body, &fedReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			out := ormapi.OPRegistrationResponse{
				OrigOperatorId:    op2.operatorId,
				OrigFederationId:  op2.fedId,
				PartnerOperatorId: fedReq.OperatorId,
				DestFederationId:  fedReq.OrigFederationId,
				MCC:               "340",
				MNC:               []string{"120", "121", "122"},
				PartnerZone:       op2.zones,
			}
			return httpmock.NewJsonResponse(200, out)
		},
	)

	httpmock.RegisterResponder("PUT", getFederationAPI(op2.fedAddr, F_API_OPERATOR_PARTNER),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := ormapi.OPUpdateMECNetConf{}
			err = json.Unmarshal(body, &inReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			return httpmock.NewStringResponse(200, "updated successfully"), nil
		},
	)

	httpmock.RegisterResponder("DELETE", getFederationAPI(op2.fedAddr, F_API_OPERATOR_PARTNER),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := ormapi.OPFederationRequest{}
			err = json.Unmarshal(body, &inReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			return httpmock.NewStringResponse(200, "delete partner OP successfully"), nil
		},
	)

	httpmock.RegisterResponder("POST", getFederationAPI(op2.fedAddr, F_API_OPERATOR_ZONE),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			zoneRegReq := ormapi.OPZoneRegister{}
			err = json.Unmarshal(body, &zoneRegReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			if len(zoneRegReq.Zones) != 1 {
				return httpmock.NewStringResponse(400, "only one zone allowed"), nil
			}

			out := ormapi.OPZoneRegisterResponse{
				LeadOperatorId:    op2.operatorId,
				FederationId:      op2.fedId,
				PartnerOperatorId: op1.operatorId,
				Zone: ormapi.OPZoneRegisterDetails{
					ZoneId:            zoneRegReq.Zones[0],
					RegistrationToken: zoneRegReq.OrigFederationId,
				},
			}
			return httpmock.NewJsonResponse(200, out)
		},
	)
	httpmock.RegisterResponder("DELETE", getFederationAPI(op2.fedAddr, F_API_OPERATOR_ZONE),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			zoneDeRegReq := ormapi.OPZoneRequest{}
			err = json.Unmarshal(body, &zoneDeRegReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			return httpmock.NewStringResponse(200, "successfully deregistered"), nil
		},
	)

	httpmock.RegisterResponder("POST", getFederationAPI(op2.fedAddr, F_API_OPERATOR_NOTIFY_ZONE),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := ormapi.OPZoneNotify{}
			err = json.Unmarshal(body, &inReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			if inReq.PartnerZone.ZoneId == "" {
				return httpmock.NewStringResponse(400, "must specify zone ID"), nil
			}

			return httpmock.NewStringResponse(200, "Added zone successfully"), nil
		},
	)
	httpmock.RegisterResponder("DELETE", getFederationAPI(op2.fedAddr, F_API_OPERATOR_NOTIFY_ZONE),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := ormapi.OPZoneRequest{}
			err = json.Unmarshal(body, &inReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			if inReq.Zone == "" {
				return httpmock.NewStringResponse(400, "must specify zone ID"), nil
			}

			return httpmock.NewStringResponse(200, "Deleted zone successfully"), nil
		},
	)
}

func TestFederation(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	unitTest = true

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Setup OP1
	op1 := SetupOperatorPlatform(t, ctx, "op1", "EU")
	defer op1.CleanupOperatorPlatform(ctx)

	// Setup OP2
	op2FedId := uuid.New().String()
	op2 := &OPAttr{
		operatorId:  "op2",
		countryCode: "KR",
		fedId:       op2FedId,
		fedAddr:     "111.111.111.111",
	}
	op2Zones := []ormapi.OPZoneInfo{
		ormapi.OPZoneInfo{
			ZoneId:      fmt.Sprintf("%s-testzone0", op2.operatorId),
			GeoLocation: "1.1",
			City:        "New York",
			State:       "New York",
			EdgeCount:   2,
		},
		ormapi.OPZoneInfo{
			ZoneId:      fmt.Sprintf("%s-testzone1", op2.operatorId),
			GeoLocation: "2.2",
			City:        "Nevada",
			State:       "Nevada",
			EdgeCount:   1,
		},
	}
	op2.zones = op2Zones

	// Register mock federation APIs
	registerFederationAPIs(t, op1, op2)

	for _, clientRun := range getUnitTestClientRuns() {
		testFederationInterconnect(t, ctx, clientRun, op1, op2)
	}
}

func testFederationInterconnect(t *testing.T, ctx context.Context, clientRun mctestclient.ClientRun, op1, op2 *OPAttr) {
	mcClient := mctestclient.NewClient(clientRun)

	// Create federation (OP1)
	op1FedReq := &ormapi.OperatorFederation{
		OperatorId:  op1.operatorId,
		CountryCode: op1.countryCode,
		MCC:         "340",
		MNCs:        "120,121,122",
	}
	op1Resp, status, err := mcClient.CreateFederation(op1.uri, op1.tokenOper, op1FedReq)
	require.Nil(t, err, "create federation")
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, op1Resp.FederationId)

	// Add federation partner (OP2)
	partnerOp2FedReq := &ormapi.OperatorFederation{
		OperatorId:     op2.operatorId,
		CountryCode:    op2.countryCode,
		FederationId:   op2.fedId,
		FederationAddr: op2.fedAddr,
	}
	_, status, err = mcClient.AddFederationPartner(op1.uri, op1.tokenOper, partnerOp2FedReq)
	require.Nil(t, err, "add partner federation")
	require.Equal(t, http.StatusOK, status)

	// Show federation info
	fedLookup := &ormapi.OperatorFederation{}
	fedInfo, status, err := mcClient.ShowFederation(op1.uri, op1.tokenOper, fedLookup)
	require.Nil(t, err, "show federation")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(fedInfo), "all federation OPs")

	// Update federation MCC value
	updateFed := &cli.MapData{
		Namespace: cli.ArgsNamespace,
		Data:      make(map[string]interface{}),
	}
	updateFed.Data["MCC"] = "344"
	_, status, err = mcClient.UpdateFederation(op1.uri, op1.tokenOper, updateFed)
	require.Nil(t, err, "update self federation")
	require.Equal(t, http.StatusOK, status)

	// Show federation info
	fedLookup = &ormapi.OperatorFederation{FederationId: op1Resp.FederationId}
	op1FedInfo, status, err := mcClient.ShowFederation(op1.uri, op1.tokenOper, fedLookup)
	require.Nil(t, err, "show self federation")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(op1FedInfo), "one entry for OP1")
	require.Equal(t, "344", op1FedInfo[0].MCC, "matches updated field")

	// Create OP1 Operator Zones
	clList, status, err := ormtestutil.TestPermShowCloudlet(mcClient, op1.uri, op1.tokenOper, op1.countryCode, op1.operatorId)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	for _, cl := range clList {
		fedZone := &ormapi.OperatorZoneCloudletMap{
			FederationId: op1Resp.FederationId,
			ZoneId:       fmt.Sprintf("op1-testzone%s", cl.Key.Name),
			GeoLocation:  fmt.Sprintf("%s.111", cl.Key.Name),
			City:         "New York",
			State:        "New York",
			Cloudlets:    []string{cl.Key.Name},
		}
		_, status, err = mcClient.CreateFederationZone(op1.uri, op1.tokenOper, fedZone)
		require.Nil(t, err, "create federation zone")
		require.Equal(t, http.StatusOK, status)
	}
	op1ZonesCnt := len(clList)
	op2ZonesCnt := len(op2.zones)

	// Show operator zones, this will include zones shared by federation partner as well
	lookup := &ormapi.OperatorZoneCloudletMap{}
	opZones, status, err := mcClient.ShowFederationZone(op1.uri, op1.tokenOper, lookup)
	require.Nil(t, err, "show federation zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, op1ZonesCnt+op2ZonesCnt, len(opZones), "op1 + op2 zones")

	// Register all the partner zones to be used
	for _, opZone := range opZones {
		zoneRegReq := &ormapi.OperatorZoneCloudletMap{
			ZoneId: opZone.ZoneId,
		}
		_, status, err := mcClient.RegisterFederationZone(op1.uri, op1.tokenOper, zoneRegReq)
		if opZone.FederationId == op1Resp.FederationId {
			require.NotNil(t, err, "cannot register self federation zone")
		} else {
			require.Nil(t, err, "register federation zone")
			require.Equal(t, http.StatusOK, status)
		}
	}

	// Cleanup
	// =======
	// De-register all the partner zones
	for _, opZone := range opZones {
		zoneDeRegReq := &ormapi.OperatorZoneCloudletMap{
			ZoneId: opZone.ZoneId,
		}
		_, status, err := mcClient.DeRegisterFederationZone(op1.uri, op1.tokenOper, zoneDeRegReq)
		if opZone.FederationId == op1Resp.FederationId {
			require.NotNil(t, err, "cannot deregister self federation zone")
		} else {
			require.Nil(t, err, "deregister federation zone")
			require.Equal(t, http.StatusOK, status)
		}
	}

	// Delete OP1 zones
	for _, opZone := range opZones {
		zoneReq := &ormapi.OperatorZoneCloudletMap{
			ZoneId: opZone.ZoneId,
		}
		_, status, err = mcClient.DeleteFederationZone(op1.uri, op1.tokenOper, zoneReq)
		if opZone.FederationId != op1Resp.FederationId {
			require.NotNil(t, err, "cannot delete partner federation zone")
		} else {
			require.Nil(t, err, "delete federation zone")
			require.Equal(t, http.StatusOK, status)
		}
	}

	// No OP1 zones should exist
	zoneLookup := &ormapi.OperatorZoneCloudletMap{
		FederationId: op1Resp.FederationId,
	}
	opZones, status, err = mcClient.ShowFederationZone(op1.uri, op1.tokenOper, zoneLookup)
	require.Nil(t, err, "show federation zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(opZones), "no op1 zones")

	// Remove federation partner (OP2)
	partnerOp2FedReq = &ormapi.OperatorFederation{
		FederationId: op2.fedId,
	}
	_, status, err = mcClient.RemoveFederationPartner(op1.uri, op1.tokenOper, partnerOp2FedReq)
	require.Nil(t, err, "remove partner federation")
	require.Equal(t, http.StatusOK, status)

	// No zones should exist
	zoneLookup = &ormapi.OperatorZoneCloudletMap{}
	opZones, status, err = mcClient.ShowFederationZone(op1.uri, op1.tokenOper, zoneLookup)
	require.Nil(t, err, "show federation zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(opZones), "no zones")

	// Show federation info
	fedLookup = &ormapi.OperatorFederation{
		FederationId: op2.fedId,
	}
	fedInfo, status, err = mcClient.ShowFederation(op1.uri, op1.tokenOper, fedLookup)
	require.Nil(t, err, "show op2 federation")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(fedInfo), "all federation OPs")

	// Delete federation (OP1)
	op1FedReq = &ormapi.OperatorFederation{
		FederationId: op1Resp.FederationId,
	}
	_, status, err = mcClient.DeleteFederation(op1.uri, op1.tokenOper, op1FedReq)
	require.Nil(t, err, "delete federation")
	require.Equal(t, http.StatusOK, status)
}
