package orm

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/billing"
	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/mctestclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/rbac"
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
	operatorId  string
	countryCode string
	dcnt        int
	tokenAd     string
	tokenOper   string
	nodeMgr     *node.NodeMgr
	database    *gorm.DB
	enforcer    *rbac.Enforcer
	fedAddr     string
}

func (o *OPAttr) SetupGlobals() {
	nodeMgr = o.nodeMgr
	database = o.database
	enforcer = o.enforcer
}

func (o *OPAttr) CleanupOperatorPlatform(ctx context.Context) {
	// nodeMgr is a global object, hence copy the OP specific
	// nodeMgr for cleanup
	o.SetupGlobals()
	o.ds.SetDummyObjs(ctx, testutil.Delete, o.operatorId, o.dcnt)
	o.ds.SetDummyOrgObjs(ctx, testutil.Delete, o.operatorId, o.dcnt)
	o.server.Stop()
	o.dc.Stop()
}

func SetupOperatorPlatform(t *testing.T, ctx context.Context, operatorId, countryCode string) *OPAttr {
	vaultServer, vaultConfig := vault.DummyServer()
	defer vaultServer.Close()

	// mock http to redirect requests
	// ==============================
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
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

	server.echo.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				nodeMgr = config.NodeMgr
				database = server.database
				return next(c)
			}
		})

	server.federationEcho.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				nodeMgr = config.NodeMgr
				database = server.database
				return next(c)
			}
		})

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
		nodeMgr:     nodeMgr,
		database:    server.database,
		enforcer:    enforcer,
		fedAddr:     fedAddr,
	}
}

func TestFederation(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	op1 := SetupOperatorPlatform(t, ctx, "op1", "EU")
	defer op1.CleanupOperatorPlatform(ctx)

	op2 := SetupOperatorPlatform(t, ctx, "op2", "KR")
	defer op2.CleanupOperatorPlatform(ctx)

	//for _, clientRun := range getUnitTestClientRuns() {
	restClient := &ormclient.Client{}
	testFederationInterconnect(t, ctx, restClient, op1, op2)
	//}
}

func testFederationInterconnect(t *testing.T, ctx context.Context, clientRun mctestclient.ClientRun, op1, op2 *OPAttr) {
	mcClient := mctestclient.NewClient(clientRun)
	op1FedReq := &ormapi.OperatorFederation{
		OperatorId:  op1.operatorId,
		CountryCode: op1.countryCode,
		MCC:         "340",
		MNCs:        "120,121,122",
	}
	op1.SetupGlobals()
	op1Resp, status, err := mcClient.CreateSelfFederation(op1.uri, op1.tokenOper, op1FedReq)
	require.Nil(t, err, "create self federation")
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, op1Resp.FederationId)

	op2FedReq := &ormapi.OperatorFederation{
		OperatorId:  op2.operatorId,
		CountryCode: op2.countryCode,
		MCC:         "340",
		MNCs:        "120,121,122",
	}
	op2.SetupGlobals()
	op2Resp, status, err := mcClient.CreateSelfFederation(op2.uri, op2.tokenOper, op2FedReq)
	require.Nil(t, err, "create self federation")
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, op2Resp.FederationId)

	partnerOp2FedReq := &ormapi.OperatorFederation{
		OperatorId:     op2.operatorId,
		CountryCode:    op2.countryCode,
		FederationId:   op2Resp.FederationId,
		FederationAddr: op2.fedAddr,
	}
	op1.SetupGlobals()
	op2SharedZones, status, err := mcClient.CreatePartnerFederation(op1.uri, op1.tokenOper, partnerOp2FedReq)
	require.Nil(t, err, "create self federation")
	require.Equal(t, http.StatusOK, status)

	require.True(t, false)
}
