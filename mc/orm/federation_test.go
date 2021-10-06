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
	"github.com/mobiledgex/edge-cloud-infra/mc/federation"
	fedcommon "github.com/mobiledgex/edge-cloud-infra/mc/federation/common"
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

type CtrlObj struct {
	addr        string
	dc          *grpc.Server
	ds          *testutil.DummyServer
	dcnt        int
	operatorIds []string
	region      string
}

type OPAttr struct {
	uri    string
	server *Server
	ctrls  []CtrlObj
}

type FederatorAttr struct {
	tokenAd     string
	tokenOper   string
	operatorId  string
	countryCode string
	fedKey      string
	fedAddr     string
	regions     []string
	zones       []federation.ZoneInfo
}

func (o *OPAttr) CleanupOperatorPlatform(ctx context.Context) {
	for _, ctrl := range o.ctrls {
		ctrl.Cleanup(ctx)
	}
	o.server.Stop()
}

func SetupControllerService(t *testing.T, ctx context.Context, operatorIds []string, region string) *CtrlObj {
	ctrlAddr, err := cloudcommon.GetAvailablePort("127.0.0.1:0")
	require.Nil(t, err, "get available port")
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
	// number of fake objects internally sent back by dummy server
	ds.ShowDummyCount = 0

	// number of dummy objects we add of each type and org
	dcnt := 3
	ds.SetDummyObjs(ctx, testutil.Create, "common", dcnt)
	for _, operatorId := range operatorIds {
		ds.SetDummyOrgObjs(ctx, testutil.Create, operatorId, dcnt)
	}
	return &CtrlObj{
		addr:        ctrlAddr,
		ds:          ds,
		dcnt:        dcnt,
		dc:          dc,
		operatorIds: operatorIds,
		region:      region,
	}
}

func (c *CtrlObj) Cleanup(ctx context.Context) {
	c.ds.SetDummyObjs(ctx, testutil.Delete, "common", c.dcnt)
	for _, operatorId := range c.operatorIds {
		c.ds.SetDummyOrgObjs(ctx, testutil.Delete, operatorId, c.dcnt)
	}
	c.dc.Stop()
}

func SetupOperatorPlatform(t *testing.T, ctx context.Context) (*OPAttr, []FederatorAttr) {
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
	// =======================

	addr, err := cloudcommon.GetAvailablePort("127.0.0.1:0")
	require.Nil(t, err, "get available port")

	sqlAddr, err := cloudcommon.GetAvailablePort("127.0.0.1:0")
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

	countryCode := "US"
	operatorIds := []string{"oper1", "oper2"}
	regions := []string{"US-East", "US-West"}

	ctrl1 := SetupControllerService(t, ctx, operatorIds, regions[0])
	ctrl2 := SetupControllerService(t, ctx, operatorIds, regions[1])
	ctrlObjs := []CtrlObj{*ctrl1, *ctrl2}

	opAttr := OPAttr{
		uri:    uri,
		server: server,
		ctrls:  ctrlObjs,
	}

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

	// create controllers
	for _, ctrlObj := range ctrlObjs {
		ctrl := ormapi.Controller{
			Region:   ctrlObj.region,
			Address:  ctrlObj.addr,
			InfluxDB: influxServer.URL,
		}
		status, err := mcClient.CreateController(uri, tokenAd, &ctrl)
		require.Nil(t, err, "create controller")
		require.Equal(t, http.StatusOK, status)
	}

	ctrls, status, err := mcClient.ShowController(uri, tokenAd, ClientNoShowFilter)
	require.Nil(t, err, "show controllers")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(ctrls))

	selfFederators := []FederatorAttr{}
	for _, operatorId := range operatorIds {
		fed := FederatorAttr{}
		// create an operator
		_, _, tokenOper := testCreateUserOrg(t, mcClient, uri, operatorId+"-user", OrgTypeOperator, operatorId)
		fed.tokenOper = tokenOper
		fed.operatorId = operatorId
		fed.countryCode = countryCode
		fed.tokenAd = tokenAd
		fed.regions = regions
		fed.fedAddr = fedAddr

		// admin allow non-edgebox cloudlets on operator org
		setOperatorOrgNoEdgeboxOnly(t, mcClient, uri, tokenAd, operatorId)

		selfFederators = append(selfFederators, fed)
	}

	return &opAttr, selfFederators
}

func getFederationAPI(fedAddr, fedApi string) string {
	return "http://" + fedAddr + fedApi
}

func registerFederationAPIs(t *testing.T, partnerFed *FederatorAttr) {
	httpmock.RegisterResponder("POST", getFederationAPI(partnerFed.fedAddr, federation.OperatorPartnerAPI),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			fedReq := federation.OperatorRegistrationRequest{}
			err = json.Unmarshal(body, &fedReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			out := federation.OperatorRegistrationResponse{
				OrigOperatorId:    partnerFed.operatorId,
				OrigFederationId:  partnerFed.fedKey,
				PartnerOperatorId: fedReq.OperatorId,
				DestFederationId:  fedReq.OrigFederationId,
				MCC:               "340",
				MNC:               []string{"120", "121", "122"},
				PartnerZone:       partnerFed.zones,
			}
			return httpmock.NewJsonResponse(200, out)
		},
	)

	httpmock.RegisterResponder("PUT", getFederationAPI(partnerFed.fedAddr, federation.OperatorPartnerAPI),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := federation.UpdateMECNetConf{}
			err = json.Unmarshal(body, &inReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			return httpmock.NewStringResponse(200, "updated successfully"), nil
		},
	)

	httpmock.RegisterResponder("DELETE", getFederationAPI(partnerFed.fedAddr, federation.OperatorPartnerAPI),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := federation.FederationRequest{}
			err = json.Unmarshal(body, &inReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			return httpmock.NewStringResponse(200, "delete partner OP successfully"), nil
		},
	)

	httpmock.RegisterResponder("POST", getFederationAPI(partnerFed.fedAddr, federation.OperatorZoneAPI),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			zoneRegReq := federation.OperatorZoneRegister{}
			err = json.Unmarshal(body, &zoneRegReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			if len(zoneRegReq.Zones) != 1 {
				return httpmock.NewStringResponse(400, "only one zone allowed"), nil
			}

			out := federation.OperatorZoneRegisterResponse{
				LeadOperatorId: partnerFed.operatorId,
				FederationId:   partnerFed.fedKey,
				Zone: federation.ZoneRegisterDetails{
					ZoneId:            zoneRegReq.Zones[0],
					RegistrationToken: zoneRegReq.OrigFederationId,
				},
			}
			return httpmock.NewJsonResponse(200, out)
		},
	)
	httpmock.RegisterResponder("DELETE", getFederationAPI(partnerFed.fedAddr, federation.OperatorZoneAPI),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			zoneDeRegReq := federation.ZoneRequest{}
			err = json.Unmarshal(body, &zoneDeRegReq)
			if err != nil {
				fmt.Printf("failed to unmarshal req data %s: %v\n", body, err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}

			return httpmock.NewStringResponse(200, "successfully deregistered"), nil
		},
	)

	httpmock.RegisterResponder("POST", getFederationAPI(partnerFed.fedAddr, federation.OperatorNotifyZoneAPI),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := federation.NotifyPartnerOperatorZone{}
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
	httpmock.RegisterResponder("DELETE", getFederationAPI(partnerFed.fedAddr, federation.OperatorNotifyZoneAPI),
		func(req *http.Request) (*http.Response, error) {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("failed to read body from request %s: %v\n", req.URL.String(), err)
				return httpmock.NewStringResponse(400, "failed to read body"), nil
			}
			inReq := federation.ZoneRequest{}
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

	// Setup Operator Platform (MC - Self federator)
	op, selfFederators := SetupOperatorPlatform(t, ctx)
	defer op.CleanupOperatorPlatform(ctx)

	// Setup partner federator
	partnerFedId := uuid.New().String()
	partnerFed := &FederatorAttr{
		operatorId:  "partnerOper",
		countryCode: "EU",
		fedKey:      partnerFedId,
		fedAddr:     "111.111.111.111",
	}
	partnerZones := []federation.ZoneInfo{
		federation.ZoneInfo{
			ZoneId:      fmt.Sprintf("%s-testzone0", partnerFed.operatorId),
			GeoLocation: "1.1",
			City:        "New York",
			State:       "New York",
			EdgeCount:   2,
		},
		federation.ZoneInfo{
			ZoneId:      fmt.Sprintf("%s-testzone1", partnerFed.operatorId),
			GeoLocation: "2.2",
			City:        "Nevada",
			State:       "Nevada",
			EdgeCount:   1,
		},
	}
	partnerFed.zones = partnerZones

	// Register mock federation APIs
	registerFederationAPIs(t, partnerFed)

	for _, clientRun := range getUnitTestClientRuns() {
		testFederationInterconnect(t, ctx, clientRun, op, selfFederators, partnerFed)
	}
}

func testPartnerFederationAPIs(t *testing.T, ctx context.Context, mcClient *mctestclient.Client, op *OPAttr, selfFederators []FederatorAttr, partnerFed *FederatorAttr) {
	selfFed1 := selfFederators[0]

	// Verify that selfFed1 has added partnerFed as partner federator (federation planning)
	// ====================================================================================
	fedLookup := &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
	}
	fedInfo, status, err := mcClient.ShowFederator(op.uri, selfFed1.tokenOper, fedLookup)
	require.Nil(t, err, "show federator")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(fedInfo))
	require.Equal(t, fedInfo[0].Type, fedcommon.TypePartner)

	// Partner federator sends federation creation request
	// ==================================================
	opRegReq := federation.OperatorRegistrationRequest{
		OrigFederationId: partnerFed.fedKey,
		DestFederationId: selfFed1.fedKey,
		OperatorId:       partnerFed.operatorId,
		CountryCode:      partnerFed.countryCode,
	}
	opRegRes := federation.OperatorRegistrationResponse{}
	err = sendFederationRequest("POST", selfFed1.fedAddr, federation.OperatorPartnerAPI, &opRegReq, &opRegRes)
	require.Nil(t, err, "partnerFed adds selfFed1 as partner OP")
	// verify federation response
	require.Equal(t, opRegRes.OrigOperatorId, selfFed1.operatorId)
	require.Equal(t, opRegRes.PartnerOperatorId, partnerFed.operatorId)
	require.Equal(t, opRegRes.OrigFederationId, selfFed1.fedKey)
	require.Equal(t, opRegRes.DestFederationId, partnerFed.fedKey)
	require.Equal(t, len(opRegRes.PartnerZone), len(selfFed1.zones), "selfFed1 zones are shared")

	// Verify federation is setup in DB
	federationReq := &ormapi.FederationRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	federations, status, err := mcClient.ShowFederation(op.uri, selfFed1.tokenOper, federationReq)
	require.Nil(t, err, "show federations")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(federations), "federation exists")
	shareZonesWithSelfRoleFound := false
	accessToSelfZonesRoleFound := false
	for _, fed := range federations {
		if fed.PartnerRole == fedcommon.RoleShareZonesWithSelf {
			shareZonesWithSelfRoleFound = true
		}
		if fed.PartnerRole == fedcommon.RoleAccessToSelfZones {
			accessToSelfZonesRoleFound = true
		}
	}
	require.True(t, shareZonesWithSelfRoleFound, "role matches")
	require.True(t, accessToSelfZonesRoleFound, "role matches")

	// partnerFed updates its MCC value and notifies selfFed1 about it
	// ===============================================================
	updateReq := federation.UpdateMECNetConf{
		OrigFederationId: partnerFed.fedKey,
		DestFederationId: selfFed1.fedKey,
		Operator:         partnerFed.operatorId,
		Country:          partnerFed.countryCode,
		MCC:              "999",
	}
	err = sendFederationRequest("PUT", selfFed1.fedAddr, federation.OperatorPartnerAPI, &updateReq, nil)
	require.Nil(t, err, "partnerFed updates its attributes and notifies selfFed1 about it")

	// verify that selfFed1 has successfully updated partnerFed's new MCC value
	fedLookup = &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
	}
	fedInfo, status, err = mcClient.ShowFederator(op.uri, selfFed1.tokenOper, fedLookup)
	require.Nil(t, err, "show federator")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(fedInfo), "federator exists")
	require.Equal(t, fedInfo[0].MCC, updateReq.MCC, "MCC values match")

	// partnerFed sends registration request for selfFed1 zone
	// =======================================================
	for _, sZone := range selfFed1.zones {
		zoneRegReq := federation.OperatorZoneRegister{
			OrigFederationId: partnerFed.fedKey,
			DestFederationId: selfFed1.fedKey,
			Operator:         partnerFed.operatorId,
			Country:          partnerFed.countryCode,
			Zones:            []string{sZone.ZoneId},
		}
		opZoneRes := federation.OperatorZoneRegisterResponse{}
		err = sendFederationRequest("POST", selfFed1.fedAddr, federation.OperatorZoneAPI, &zoneRegReq, &opZoneRes)
		require.Nil(t, err, "partnerFed sends registration request for selfFed1 zone")

		// Verify that registered zones are shown
		zoneLookup := &ormapi.FederatorZoneDetails{
			SelfOperatorId:  selfFed1.operatorId,
			SelfCountryCode: selfFed1.countryCode,
			OperatorId:      selfFed1.operatorId,
			CountryCode:     selfFed1.countryCode,
			ZoneId:          sZone.ZoneId,
		}
		selfFed1Zones, status, err := mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
		require.Nil(t, err, "show self federator zones")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, 1, len(selfFed1Zones))
		zone := selfFed1Zones[0]
		matchStr := fmt.Sprintf("%s/%s", partnerFed.operatorId, partnerFed.countryCode)
		// self zone is shared with partner federator
		require.Equal(t, 1, len(zone.SharedWithFederators))
		require.Equal(t, matchStr, zone.SharedWithFederators[0])
		// self zone is registered by partner federator
		require.Equal(t, 1, len(zone.RegisteredByFederators))
		require.Equal(t, matchStr, zone.RegisteredByFederators[0])
	}

	// partnerFed notifies selfFed1 about a new zone
	// =============================================
	newZone := federation.ZoneInfo{
		ZoneId:      fmt.Sprintf("%s-testzoneX", partnerFed.operatorId),
		GeoLocation: "9.9",
		City:        "Newark",
		State:       "Newark",
		EdgeCount:   2,
	}
	zoneNotifyReq := federation.NotifyPartnerOperatorZone{
		OrigFederationId: partnerFed.fedKey,
		DestFederationId: selfFed1.fedKey,
		Operator:         partnerFed.operatorId,
		Country:          partnerFed.countryCode,
		PartnerZone:      newZone,
	}
	err = sendFederationRequest("POST", selfFed1.fedAddr, federation.OperatorNotifyZoneAPI, &zoneNotifyReq, nil)
	require.Nil(t, err, "partnerFed notifies selfFed1 about a new zone")

	// verify that selfFed1 added this new zone in its db
	zoneLookup := &ormapi.FederatorZoneDetails{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
		ZoneId:          newZone.ZoneId,
	}
	pZones, status, err := mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
	require.Nil(t, err, "show federator zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(pZones))
	require.Equal(t, newZone.ZoneId, pZones[0].ZoneId)
	require.Equal(t, 1, len(pZones[0].SharedWithFederators))
	require.Equal(t, 0, len(pZones[0].RegisteredByFederators))

	// partnerFed notifies selfFed1 about a deleted zone
	// =================================================
	zoneUnshareReq := federation.ZoneRequest{
		OrigFederationId: partnerFed.fedKey,
		DestFederationId: selfFed1.fedKey,
		Operator:         partnerFed.operatorId,
		Country:          partnerFed.countryCode,
		Zone:             newZone.ZoneId,
	}
	err = sendFederationRequest("DELETE", selfFed1.fedAddr, federation.OperatorNotifyZoneAPI, &zoneUnshareReq, nil)
	require.Nil(t, err, "partnerFed notifies selfFed1 about a deleted zone")

	// verify that selfFed1 deleted this zone from its db
	pZones, status, err = mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
	require.Nil(t, err, "show federator zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(pZones))

	// partnerFed sends deregistration request for selfFed1 zone
	// =========================================================
	for _, sZone := range selfFed1.zones {
		zoneDeRegReq := federation.ZoneRequest{
			OrigFederationId: partnerFed.fedKey,
			DestFederationId: selfFed1.fedKey,
			Operator:         partnerFed.operatorId,
			Country:          partnerFed.countryCode,
			Zone:             sZone.ZoneId,
		}
		err = sendFederationRequest("DELETE", selfFed1.fedAddr, federation.OperatorZoneAPI, &zoneDeRegReq, nil)
		require.Nil(t, err, "partnerFed sends deregistration request for selfFed1 zone")

		// Verify that zones are deregistered
		zoneLookup := &ormapi.FederatorZoneDetails{
			SelfOperatorId:  selfFed1.operatorId,
			SelfCountryCode: selfFed1.countryCode,
			OperatorId:      selfFed1.operatorId,
			CountryCode:     selfFed1.countryCode,
			ZoneId:          sZone.ZoneId,
		}
		selfFed1Zones, status, err := mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
		require.Nil(t, err, "show self federator zones")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, 1, len(selfFed1Zones))
		zone := selfFed1Zones[0]
		matchStr := fmt.Sprintf("%s/%s", partnerFed.operatorId, partnerFed.countryCode)
		// self zone is shared with partner federator
		require.Equal(t, 1, len(zone.SharedWithFederators))
		require.Equal(t, matchStr, zone.SharedWithFederators[0])
		// self zone is not registered by partner federator
		require.Equal(t, 0, len(zone.RegisteredByFederators))
	}

	// partnerFed removes selfFed1 as federation partner
	// =================================================
	opFedReq := federation.FederationRequest{
		OrigFederationId: partnerFed.fedKey,
		DestFederationId: selfFed1.fedKey,
		Operator:         partnerFed.operatorId,
		Country:          partnerFed.countryCode,
	}
	err = sendFederationRequest("DELETE", selfFed1.fedAddr, federation.OperatorPartnerAPI, &opFedReq, nil)
	require.Nil(t, err, "partnerFed removes selfFed1 as partner OP")

	// verify that partnerFed has successfully removed federation with selfFed1
	federationReq = &ormapi.FederationRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	federations, status, err = mcClient.ShowFederation(op.uri, selfFed1.tokenOper, federationReq)
	require.Nil(t, err, "show federations")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(federations), "only one federation exist")
	require.Equal(t, fedcommon.RoleShareZonesWithSelf, federations[0].PartnerRole, "role matches")
}

func testFederationInterconnect(t *testing.T, ctx context.Context, clientRun mctestclient.ClientRun, op *OPAttr, selfFederators []FederatorAttr, partnerFed *FederatorAttr) {
	mcClient := mctestclient.NewClient(clientRun)

	// Create self federator objs
	// ==========================
	for ii, selfFed := range selfFederators {
		fedReq := &ormapi.FederatorRequest{
			OperatorId:  selfFed.operatorId,
			CountryCode: selfFed.countryCode,
			MCC:         "340",
			MNCs:        []string{"120", "121", "122"},
		}
		resp, status, err := mcClient.CreateSelfFederator(op.uri, selfFed.tokenOper, fedReq)
		require.Nil(t, err, "create self federator")
		require.Equal(t, http.StatusOK, status)
		require.NotEmpty(t, resp.FederationKey)
		selfFederators[ii].fedKey = resp.FederationKey
	}

	selfFed1 := selfFederators[0]
	selfFed2 := selfFederators[1]

	// selfFed1 creates partner federator obj
	// ======================================
	partnerFedReq := &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
		FederationKey:   partnerFed.fedKey,
		FederationAddr:  partnerFed.fedAddr,
	}
	_, status, err := mcClient.CreatePartnerFederator(op.uri, selfFed1.tokenOper, partnerFedReq)
	require.Nil(t, err, "create partner federator")
	require.Equal(t, http.StatusOK, status)

	// Show federation info
	fedLookup := &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
	}
	fedInfo, status, err := mcClient.ShowFederator(op.uri, selfFed1.tokenOper, fedLookup)
	require.Nil(t, err, "show federators")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(fedInfo))

	// Validate self & partner federation info
	foundSelf := false
	foundPartner := false
	for _, fed := range fedInfo {
		require.Equal(t, selfFed1.operatorId, fed.SelfOperatorId)
		require.Equal(t, selfFed1.countryCode, fed.SelfCountryCode)
		if fed.Type == fedcommon.TypeSelf {
			require.Equal(t, selfFed1.operatorId, fed.OperatorId)
			require.Equal(t, selfFed1.countryCode, fed.CountryCode)
			foundSelf = true
		} else {
			require.Equal(t, partnerFed.operatorId, fed.OperatorId)
			require.Equal(t, partnerFed.countryCode, fed.CountryCode)
			foundPartner = true
		}
	}
	require.True(t, foundSelf, "self federator exists")
	require.True(t, foundPartner, "partner federator exists")

	// selfFed2 should not be able to see selfFed1's partner
	// =====================================================
	fedLookup = &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
	}
	fedInfo, status, err = mcClient.ShowFederator(op.uri, selfFed2.tokenOper, fedLookup)
	require.NotNil(t, err, "do not show partner federator")
	require.Equal(t, http.StatusForbidden, status)
	partnerFedLookup := &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed2.operatorId,
		SelfCountryCode: selfFed2.countryCode,
		Type:            fedcommon.TypePartner,
	}
	fedInfo, status, err = mcClient.ShowFederator(op.uri, selfFed2.tokenOper, partnerFedLookup)
	require.Nil(t, err, "show partner federator")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(fedInfo))

	// Update self federator MCC value
	// ===============================
	updateFed := &cli.MapData{
		Namespace: cli.ArgsNamespace,
		Data:      make(map[string]interface{}),
	}
	updateFed.Data["OperatorId"] = selfFed1.operatorId
	updateFed.Data["CountryCode"] = selfFed1.countryCode
	updateFed.Data["MCC"] = "344"
	_, status, err = mcClient.UpdateSelfFederator(op.uri, selfFed1.tokenOper, updateFed)
	require.Nil(t, err, "update self federation")
	require.Equal(t, http.StatusOK, status)

	// Show federator info
	fedLookup = &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		Type:            fedcommon.TypeSelf,
	}
	fedInfo, status, err = mcClient.ShowFederator(op.uri, selfFed1.tokenOper, fedLookup)
	require.Nil(t, err, "show self federation")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(fedInfo), "one entry")
	require.Equal(t, "344", fedInfo[0].MCC, "matches updated field")

	// Create self federator zones
	// ===========================
	for fid, selfFed := range selfFederators {
		selfFed.zones = []federation.ZoneInfo{}
		for ii, region := range selfFed.regions {
			filter := &edgeproto.Cloudlet{
				Key: edgeproto.CloudletKey{
					Organization: selfFed.operatorId,
				},
			}
			clList, status, err := ormtestutil.TestShowCloudlet(mcClient, op.uri, selfFed.tokenOper, region, filter)
			require.Nil(t, err)
			require.Equal(t, http.StatusOK, status)
			for jj, cl := range clList {
				fedZone := &ormapi.FederatorZoneDetails{
					ZoneId:      fmt.Sprintf("op-testzone-%s-%s-%s", selfFed.operatorId, region, cl.Key.Name),
					OperatorId:  selfFed.operatorId,
					CountryCode: selfFed.countryCode,
					GeoLocation: fmt.Sprintf("%d.%d,%d.%d", ii, jj, ii, jj),
					Region:      region,
					City:        "New York",
					State:       "New York",
					Cloudlets:   []string{cl.Key.Name},
				}
				_, status, err = mcClient.CreateSelfFederatorZone(op.uri, selfFed.tokenOper, fedZone)
				require.Nil(t, err, "create federation zone")
				require.Equal(t, http.StatusOK, status)
				fedZoneInfo := federation.ZoneInfo{
					ZoneId:      fedZone.ZoneId,
					GeoLocation: fedZone.GeoLocation,
					City:        fedZone.City,
					State:       fedZone.State,
					EdgeCount:   len(fedZone.Cloudlets),
				}
				selfFed.zones = append(selfFed.zones, fedZoneInfo)
			}
		}
		selfFederators[fid].zones = selfFed.zones
	}

	// Verify that all zones are created
	// =================================
	for _, selfFed := range selfFederators {
		lookup := &ormapi.FederatorZoneDetails{
			SelfOperatorId:  selfFed.operatorId,
			SelfCountryCode: selfFed.countryCode,
			OperatorId:      selfFed.operatorId,
			CountryCode:     selfFed.countryCode,
		}
		selfFedZones, status, err := mcClient.ShowFederatorZone(op.uri, selfFed.tokenOper, lookup)
		require.Nil(t, err, "show self federator zones")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, len(selfFed.zones), len(selfFedZones), "self federator zones match")
		for _, fedZone := range selfFedZones {
			// these zones are not yet shared/registered
			require.Equal(t, 0, len(fedZone.RegisteredByFederators))
			require.Equal(t, 0, len(fedZone.SharedWithFederators))
		}
	}

	selfFed1Zones := selfFederators[0].zones

	// As part of federation planning, mark zones to be shared with partner federator
	// ==============================================================================
	for _, zone := range selfFed1Zones {
		zoneShReq := &ormapi.FederatorZoneRequest{
			SelfOperatorId:     selfFed1.operatorId,
			SelfCountryCode:    selfFed1.countryCode,
			PartnerOperatorId:  partnerFed.operatorId,
			PartnerCountryCode: partnerFed.countryCode,
			ZoneId:             zone.ZoneId,
		}
		_, status, err := mcClient.ShareSelfFederatorZone(op.uri, selfFed1.tokenOper, zoneShReq)
		require.Nil(t, err, "mark zones to be shared with partner federator")
		require.Equal(t, http.StatusOK, status)
	}

	// All selfFed1 zones are marked to be shared with partner federator
	zoneLookup := &ormapi.FederatorZoneDetails{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      selfFed1.operatorId,
		CountryCode:     selfFed1.countryCode,
	}
	selfFedZones, status, err := mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
	require.Nil(t, err, "show self federator zones")
	require.Equal(t, http.StatusOK, status)
	for _, fedZone := range selfFedZones {
		// these zones are not yet registered
		require.Equal(t, 0, len(fedZone.RegisteredByFederators))
		// these zones are shared with one partner federator
		require.Equal(t, 1, len(fedZone.SharedWithFederators))
		matchStr := fmt.Sprintf("%s/%s", partnerFed.operatorId, partnerFed.countryCode)
		require.Equal(t, matchStr, fedZone.SharedWithFederators[0])
	}

	// No partner zones exist as federation is not yet created
	// =======================================================
	zoneLookup = &ormapi.FederatorZoneDetails{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
	}
	partnerZones, status, err := mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
	require.Nil(t, err, "show partner federator zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(partnerZones))

	// Register partner zone should fail as federation is not yet created
	// ==================================================================
	zoneRegReq := &ormapi.FederatorZoneRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
		ZoneId:             partnerFed.zones[0].ZoneId,
	}
	_, _, err = mcClient.RegisterPartnerFederatorZone(op.uri, selfFed1.tokenOper, zoneRegReq)
	require.NotNil(t, err, "cannot register partner zone as federation does not exist")
	require.Contains(t, err.Error(), "does not exist")

	// Create federation between selfFed1 and partner federator
	// ========================================================
	fedReq := &ormapi.FederationRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	_, status, err = mcClient.CreateFederation(op.uri, selfFed1.tokenOper, fedReq)
	require.Nil(t, err, "create federation")
	require.Equal(t, http.StatusOK, status)

	// Verify federation is created
	federationReq := &ormapi.FederationRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	federations, status, err := mcClient.ShowFederation(op.uri, selfFed1.tokenOper, federationReq)
	require.Nil(t, err, "show federations")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(federations), "federation exists")
	require.Equal(t, fedcommon.RoleShareZonesWithSelf, federations[0].PartnerRole, "role matches")

	// Verify federation does not exist with selfFed2
	federationReq = &ormapi.FederationRequest{
		SelfOperatorId:     selfFed2.operatorId,
		SelfCountryCode:    selfFed2.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	federations, status, err = mcClient.ShowFederation(op.uri, selfFed2.tokenOper, federationReq)
	require.Nil(t, err, "show federations")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(federations), "federation does not exist")

	// Partner zones are shared as part of federation create
	// =====================================================
	zoneLookup = &ormapi.FederatorZoneDetails{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
	}
	partnerZones, status, err = mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
	require.Nil(t, err, "show partner federator zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, len(partnerFed.zones), len(partnerZones))

	// Register all the partner zones to be used
	// =========================================
	for _, pZone := range partnerFed.zones {
		zoneRegReq := &ormapi.FederatorZoneRequest{
			SelfOperatorId:     selfFed1.operatorId,
			SelfCountryCode:    selfFed1.countryCode,
			PartnerOperatorId:  partnerFed.operatorId,
			PartnerCountryCode: partnerFed.countryCode,
			ZoneId:             pZone.ZoneId,
		}
		_, status, err = mcClient.RegisterPartnerFederatorZone(op.uri, selfFed1.tokenOper, zoneRegReq)
		require.Nil(t, err, "register partner federator zone")
		require.Equal(t, http.StatusOK, status)

		// Verify that registered zones are shown
		zoneLookup = &ormapi.FederatorZoneDetails{
			SelfOperatorId:  selfFed1.operatorId,
			SelfCountryCode: selfFed1.countryCode,
			OperatorId:      partnerFed.operatorId,
			CountryCode:     partnerFed.countryCode,
			ZoneId:          pZone.ZoneId,
		}
		partnerZones, status, err = mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
		require.Nil(t, err, "show partner federator zones")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, 1, len(partnerZones))
		zone := partnerZones[0]
		matchStr := fmt.Sprintf("%s/%s", selfFed1.operatorId, selfFed1.countryCode)
		// partner zone is shared with self federator 1
		require.Equal(t, 1, len(zone.SharedWithFederators))
		require.Equal(t, matchStr, zone.SharedWithFederators[0])
		// partner zone is registered by self federator 1
		require.Equal(t, 1, len(zone.RegisteredByFederators))
		require.Equal(t, matchStr, zone.RegisteredByFederators[0])
	}

	// Test federation APIs
	// ====================
	testPartnerFederationAPIs(t, ctx, mcClient, op, selfFederators, partnerFed)

	// --------+
	// Cleanup |
	// --------+

	// Federation deletion between selfFed1 and partner federator should
	// fail if there are partner zones registered
	// =================================================================
	fedReq = &ormapi.FederationRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	_, _, err = mcClient.DeleteFederation(op.uri, selfFed1.tokenOper, fedReq)
	require.NotNil(t, err, "delete federation")
	require.Contains(t, err.Error(), "Please deregister it before removing")

	// Deregister all the partner zones
	// ================================
	for _, pZone := range partnerFed.zones {
		zoneRegReq := &ormapi.FederatorZoneRequest{
			SelfOperatorId:     selfFed1.operatorId,
			SelfCountryCode:    selfFed1.countryCode,
			PartnerOperatorId:  partnerFed.operatorId,
			PartnerCountryCode: partnerFed.countryCode,
			ZoneId:             pZone.ZoneId,
		}
		_, status, err = mcClient.DeRegisterPartnerFederatorZone(op.uri, selfFed1.tokenOper, zoneRegReq)
		require.Nil(t, err, "deregister partner federator zone")
		require.Equal(t, http.StatusOK, status)

		// Verify that zones are deregistered
		zoneLookup = &ormapi.FederatorZoneDetails{
			SelfOperatorId:  selfFed1.operatorId,
			SelfCountryCode: selfFed1.countryCode,
			OperatorId:      partnerFed.operatorId,
			CountryCode:     partnerFed.countryCode,
			ZoneId:          pZone.ZoneId,
		}
		partnerZones, status, err = mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
		require.Nil(t, err, "show partner federator zones")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, 1, len(partnerZones))
		zone := partnerZones[0]
		matchStr := fmt.Sprintf("%s/%s", selfFed1.operatorId, selfFed1.countryCode)
		// partner zone is shared with self federator 1
		require.Equal(t, 1, len(zone.SharedWithFederators))
		require.Equal(t, matchStr, zone.SharedWithFederators[0])
		// partner zone is not registeredy by any self federator
		require.Equal(t, 0, len(zone.RegisteredByFederators))
	}

	// Delete federation between selfFed1 and partner federator
	// ========================================================
	fedReq = &ormapi.FederationRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	_, status, err = mcClient.DeleteFederation(op.uri, selfFed1.tokenOper, fedReq)
	require.Nil(t, err, "delete federation")
	require.Equal(t, http.StatusOK, status)

	// Verify federation is deleted
	federationReq = &ormapi.FederationRequest{
		SelfOperatorId:     selfFed1.operatorId,
		SelfCountryCode:    selfFed1.countryCode,
		PartnerOperatorId:  partnerFed.operatorId,
		PartnerCountryCode: partnerFed.countryCode,
	}
	federations, status, err = mcClient.ShowFederation(op.uri, selfFed1.tokenOper, federationReq)
	require.Nil(t, err, "show federations")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(federations), "no federation exists")

	// No partner zones exist as federation is deleted
	// =======================================================
	zoneLookup = &ormapi.FederatorZoneDetails{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
	}
	partnerZones, status, err = mcClient.ShowFederatorZone(op.uri, selfFed1.tokenOper, zoneLookup)
	require.Nil(t, err, "show partner federator zones")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(partnerZones))

	// Deletion of self federator zone should fail if it is shared with a partner federator
	// ====================================================================================
	zoneReq := &ormapi.FederatorZoneDetails{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      selfFed1.operatorId,
		CountryCode:     selfFed1.countryCode,
		ZoneId:          selfFed1Zones[0].ZoneId,
	}
	_, _, err = mcClient.DeleteSelfFederatorZone(op.uri, selfFed1.tokenOper, zoneReq)
	require.NotNil(t, err, "delete self federator zone should fail as it is shared")

	// Unshare all shared zones
	// ========================
	for _, zone := range selfFed1Zones {
		zoneShReq := &ormapi.FederatorZoneRequest{
			SelfOperatorId:     selfFed1.operatorId,
			SelfCountryCode:    selfFed1.countryCode,
			PartnerOperatorId:  partnerFed.operatorId,
			PartnerCountryCode: partnerFed.countryCode,
			ZoneId:             zone.ZoneId,
		}
		_, status, err := mcClient.UnshareSelfFederatorZone(op.uri, selfFed1.tokenOper, zoneShReq)
		require.Nil(t, err, "mark zones to be unshared with partner federator")
		require.Equal(t, http.StatusOK, status)
	}

	// Delete self federator zones
	// ===========================
	for _, selfFed := range selfFederators {
		for _, zone := range selfFed.zones {
			zoneReq := &ormapi.FederatorZoneDetails{
				SelfOperatorId:  selfFed.operatorId,
				SelfCountryCode: selfFed.countryCode,
				OperatorId:      selfFed.operatorId,
				CountryCode:     selfFed.countryCode,
				ZoneId:          zone.ZoneId,
			}
			_, status, err = mcClient.DeleteSelfFederatorZone(op.uri, selfFed.tokenOper, zoneReq)
			require.Nil(t, err, "delete self federator zone")
			require.Equal(t, http.StatusOK, status)
		}
		// No zones should exist
		zoneReq := &ormapi.FederatorZoneDetails{
			SelfOperatorId:  selfFed.operatorId,
			SelfCountryCode: selfFed.countryCode,
			OperatorId:      selfFed.operatorId,
			CountryCode:     selfFed.countryCode,
		}
		fedZones, status, err := mcClient.ShowFederatorZone(op.uri, selfFed.tokenOper, zoneReq)
		require.Nil(t, err, "show self federator zones")
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, 0, len(fedZones))
	}

	// Deletion of self federator should fail if its associated partner
	// federators still exist
	// ================================================================
	fedDelReq := &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      selfFed1.operatorId,
		CountryCode:     selfFed1.countryCode,
	}
	_, status, err = mcClient.DeleteSelfFederator(op.uri, selfFed1.tokenOper, fedDelReq)
	require.NotNil(t, err, "cannot delete self federator")

	// Delete partner federator obj
	// ============================
	partnerFedReq = &ormapi.FederatorRequest{
		SelfOperatorId:  selfFed1.operatorId,
		SelfCountryCode: selfFed1.countryCode,
		OperatorId:      partnerFed.operatorId,
		CountryCode:     partnerFed.countryCode,
	}
	_, status, err = mcClient.DeletePartnerFederator(op.uri, selfFed1.tokenOper, partnerFedReq)
	require.Nil(t, err, "delete partner federator")
	require.Equal(t, http.StatusOK, status)

	// Delete self federators
	// =======================
	for _, selfFed := range selfFederators {
		fedReq := &ormapi.FederatorRequest{
			SelfOperatorId:  selfFed.operatorId,
			SelfCountryCode: selfFed.countryCode,
			OperatorId:      selfFed.operatorId,
			CountryCode:     selfFed.countryCode,
		}
		_, status, err := mcClient.DeleteSelfFederator(op.uri, selfFed.tokenOper, fedReq)
		require.Nil(t, err, "delete self federator")
		require.Equal(t, http.StatusOK, status)
	}
}
