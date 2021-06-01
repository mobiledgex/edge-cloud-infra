package orm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/lib/pq"
	"github.com/mobiledgex/edge-cloud-infra/billing"
	"github.com/mobiledgex/edge-cloud-infra/billing/chargify"
	"github.com/mobiledgex/edge-cloud-infra/billing/fakebilling"
	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/rbac"
	"github.com/mobiledgex/edge-cloud-infra/version"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/cloudcommon/ratelimit"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	edgetls "github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/nmcclain/ldap"
	gitlab "github.com/xanzy/go-gitlab"
)

// Server struct is just to track sql/db so we can stop them later.
type Server struct {
	config       *ServerConfig
	sql          *intprocess.Sql
	database     *gorm.DB
	echo         *echo.Echo
	vault        *process.Vault
	stopInitData bool
	initDataDone chan error
	initJWKDone  chan struct{}
	notifyServer *notify.ServerMgr
	notifyClient *notify.Client
	sqlListener  *pq.Listener
	ldapServer   *ldap.Server
	done         chan struct{}
}

type ServerConfig struct {
	ServAddr                string
	SqlAddr                 string
	VaultAddr               string
	RunLocal                bool
	InitLocal               bool
	IgnoreEnv               bool
	ApiTlsCertFile          string
	ApiTlsKeyFile           string
	LocalVault              bool
	LDAPAddr                string
	LDAPUsername            string
	LDAPPassword            string
	GitlabAddr              string
	ArtifactoryAddr         string
	PingInterval            time.Duration
	SkipVerifyEmail         bool
	JaegerAddr              string
	vaultConfig             *vault.Config
	SkipOriginCheck         bool
	Hostname                string
	NotifyAddrs             string
	NotifySrvAddr           string
	NodeMgr                 *node.NodeMgr
	BillingPlatform         string
	BillingService          billing.BillingService
	AlertCache              *edgeproto.AlertCache
	AlertMgrAddr            string
	AlertmgrResolveTimout   time.Duration
	UsageCheckpointInterval string
	DomainName              string
	StaticDir               string
	DeploymentTag           string
	RemoveRateLimit         bool
}

var DefaultDBUser = "mcuser"
var DefaultDBName = "mcdb"
var DefaultDBPass = ""
var DefaultSuperuser = "mexadmin"
var DefaultSuperpass = "mexadminfastedgecloudinfra"
var Superuser string

var database *gorm.DB

var enforcer *rbac.Enforcer
var serverConfig *ServerConfig
var gitlabClient *gitlab.Client
var gitlabSync *AppStoreSync
var artifactorySync *AppStoreSync
var nodeMgr *node.NodeMgr
var AlertManagerServer *alertmgr.AlertMgrServer
var allRegionCaches AllRegionCaches

var rateLimitMgr *ratelimit.RateLimitManager

func RunServer(config *ServerConfig) (retserver *Server, reterr error) {
	server := Server{config: config}
	// keep global pointer to config stored in server for easy access
	serverConfig = server.config
	if config.NodeMgr == nil {
		config.NodeMgr = &node.NodeMgr{}
	}
	nodeMgr = config.NodeMgr
	server.done = make(chan struct{})

	rateLimitMgr = ratelimit.NewRateLimitManager()

	dbuser := os.Getenv("db_username")
	dbpass := os.Getenv("db_password")
	dbname := os.Getenv("db_name")
	Superuser = os.Getenv("superuser")
	superpass := os.Getenv("superpass")
	gitlabToken := os.Getenv("gitlab_token")
	if dbuser == "" || config.IgnoreEnv {
		dbuser = DefaultDBUser
	}
	if dbname == "" || config.IgnoreEnv {
		dbname = DefaultDBName
	}
	if dbpass == "" || config.IgnoreEnv {
		dbpass = DefaultDBPass
	}
	if Superuser == "" || config.IgnoreEnv {
		Superuser = DefaultSuperuser
	}
	if superpass == "" || config.IgnoreEnv {
		superpass = DefaultSuperpass
	}
	if serverConfig.LDAPUsername == "" && !config.IgnoreEnv {
		serverConfig.LDAPUsername = os.Getenv("LDAP_USERNAME")
	}
	if serverConfig.LDAPPassword == "" && !config.IgnoreEnv {
		serverConfig.LDAPPassword = os.Getenv("LDAP_PASSWORD")
	}
	allRegionCaches.init()

	if config.DeploymentTag == "" {
		return nil, fmt.Errorf("Missing deployment tag")
	}

	ctx, span, err := nodeMgr.Init(node.NodeTypeMC, node.CertIssuerGlobal, node.WithName(config.Hostname), node.WithCloudletPoolLookup(&allRegionCaches), node.WithCloudletLookup(&allRegionCaches))
	if err != nil {
		return nil, err
	}
	defer span.Finish()
	defer func() {
		if reterr != nil {
			server.Stop()
		}
	}()
	nodeMgr.UpdateNodeProps(ctx, version.InfraBuildProps("Infra"))

	if config.LocalVault {
		vaultProc := process.Vault{
			Common: process.Common{
				Name: "vault",
			},
			DmeSecret: "123456",
		}
		_, err := vaultProc.StartLocalRoles()
		if err != nil {
			return nil, err
		}
		roles, err := intprocess.SetupVault(&vaultProc)
		if err != nil {
			return nil, err
		}
		roleID := roles.MCRoleID
		secretID := roles.MCSecretID
		config.VaultAddr = process.VaultAddress
		server.vault = &vaultProc
		auth := vault.NewAppRoleAuth(roleID, secretID)
		config.vaultConfig = vault.NewConfig(process.VaultAddress, auth)
	}
	// vaultConfig should only be set by unit tests
	if config.vaultConfig == nil {
		vaultConfig, err := vault.BestConfig(config.VaultAddr)
		if err != nil {
			return nil, err
		}
		config.vaultConfig = vaultConfig
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "vault auth", "type", config.vaultConfig.Auth.Type())
	server.initJWKDone = make(chan struct{}, 1)
	InitVault(config.vaultConfig, server.done, server.initJWKDone)

	switch serverConfig.BillingPlatform {
	case "fake":
		serverConfig.BillingService = &fakebilling.BillingService{}
	case "chargify":
		serverConfig.BillingService = &chargify.BillingService{}
	default:
		return nil, fmt.Errorf("Unable to determine billing platform: %s\n", serverConfig.BillingPlatform)
	}

	err = serverConfig.BillingService.Init(ctx, config.vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to initialize billing services: %v", err)
	}

	if err = checkUsageCheckpointInterval(); err != nil {
		return nil, err
	}

	if gitlabToken == "" {
		log.InfoLog("Note: No gitlab_token env var found")
	}
	gitlabClient = gitlab.NewClient(nil, gitlabToken)
	if err = gitlabClient.SetBaseURL(config.GitlabAddr); err != nil {
		return nil, fmt.Errorf("Gitlab client set base URL to %s, %s",
			config.GitlabAddr, err.Error())
	}

	if config.RunLocal {
		sql := intprocess.Sql{
			Common: process.Common{
				Name: "sql1",
			},
			DataDir:  "./.postgres",
			HttpAddr: config.SqlAddr,
			Username: dbuser,
			Dbname:   dbname,
		}
		_, err := os.Stat(sql.DataDir)
		if config.InitLocal || os.IsNotExist(err) {
			sql.InitDataDir()
		}
		err = sql.StartLocal("")
		if err != nil {
			return nil, fmt.Errorf("local sql start failed, %s",
				err.Error())
		}
		server.sql = &sql
	}

	initdb, err := InitSql(ctx, config.SqlAddr, dbuser, dbpass, dbname)
	if err != nil {
		return nil, fmt.Errorf("sql init failed, %s", err.Error())
	}
	database = initdb
	server.database = database

	enforcer = rbac.NewEnforcer(initdb)
	err = enforcer.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("enforcer init failed, %v", err)
	}

	server.initDataDone = make(chan error, 1)
	go InitData(ctx, Superuser, superpass, config.PingInterval, &server.stopInitData, server.done, server.initDataDone)

	if config.AlertMgrAddr != "" {
		tlsConfig, err := nodeMgr.GetPublicClientTlsConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("Unable to get a client tls config, %s", err.Error())
		}
		AlertManagerServer, err = alertmgr.NewAlertMgrServer(config.AlertMgrAddr, tlsConfig,
			config.AlertCache, config.AlertmgrResolveTimout)
		if err != nil {
			// TODO - this needs to be a fatal failure when we add alertmanager deployment to the ansible scripts
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to start alertmanager server", "error", err)
			err = nil
		}
	}

	e := echo.New()
	e.HideBanner = true
	server.echo = e

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.Use(logger, RateLimit)

	// login route
	root := "api/v1"
	// accessible routes

	// swagger:route POST /login Security Login
	// Login.
	// Login to MC.
	// responses:
	//   200: authToken
	//   400: loginBadRequest
	createMcApi(e, root, "/login", Login, NoAuth, Default)
	// swagger:route POST /usercreate User CreateUser
	// Create User.
	// Creates a new user and allows them to access and manage resources.
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createMcApi(e, root, "/usercreate", CreateUser, Auth, Default)
	createMcApi(e, root, "/passwordresetrequest", PasswordResetRequest, NoAuth, Default)
	// swagger:route POST /publicconfig Config PublicConfig
	// Show Public Configuration.
	// Show Public Configuration for UI
	// responses:
	//   200: success
	//   400: badRequest
	//   404: notFound
	createMcApi(e, root, "/publicconfig", PublicConfig, NoAuth, Default)
	// swagger:route POST /passwordreset Security PasswdReset
	// Reset Login Password.
	// This resets your login password.
	// responses:
	//   200: success
	//   400: badRequest
	createMcApi(e, root, "/passwordreset", PasswordReset, NoAuth, Default)
	createMcApi(e, root, "/verifyemail", VerifyEmail, NoAuth, Default)
	createMcApi(e, root, "/resendverify", ResendVerify, NoAuth, Default)
	// authenticated routes - jwt middleware
	authPrefix := root + "/auth"
	auth := e.Group(authPrefix)
	auth.Use(AuthCookie)
	// refresh auth cookie
	createAuthMcApi(auth, authPrefix, "/refresh", RefreshAuthCookie, Default)

	// swagger:route POST /auth/user/show User ShowUser
	// Show Users.
	// Displays existing users to which you are authorized to access.
	// Security:
	//   Bearer:
	// responses:
	//   200: listUsers
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/user/show", ShowUser, Show)
	createAuthMcApi(auth, authPrefix, "/user/current", CurrentUser, Show)
	// swagger:route POST /auth/user/delete User DeleteUser
	// Delete User.
	// Deletes existing user.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/user/delete", DeleteUser, Delete)
	// swagger:route POST /auth/user/update User UpdateUser
	// Update User.
	// Updates current user.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/user/update", UpdateUser, Update)
	createAuthMcApi(auth, authPrefix, "/user/newpass", NewPassword, Update)
	createAuthMcApi(auth, authPrefix, "/user/create/apikey", CreateUserApiKey, Create)
	createAuthMcApi(auth, authPrefix, "/user/delete/apikey", DeleteUserApiKey, Delete)
	createAuthMcApi(auth, authPrefix, "/user/show/apikey", ShowUserApiKey, Show)
	createAuthMcApi(auth, authPrefix, "/role/assignment/show", ShowRoleAssignment, Show)
	createAuthMcApi(auth, authPrefix, "/role/perms/show", ShowRolePerms, Show)
	createAuthMcApi(auth, authPrefix, "/role/show", ShowRole, Show)
	createAuthMcApi(auth, authPrefix, "/role/adduser", AddUserRole, Default)
	createAuthMcApi(auth, authPrefix, "/role/removeuser", RemoveUserRole, Default)
	createAuthMcApi(auth, authPrefix, "/role/showuser", ShowUserRole, Show)
	// swagger:route POST /auth/org/create Organization CreateOrg
	// Create Organization.
	// Create an Organization to access operator/cloudlet APIs.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/org/create", CreateOrg, Create)
	// swagger:route POST /auth/org/update Organization UpdateOrg
	// Update Organization.
	// API to update an existing Organization.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/org/update", UpdateOrg, Update)
	// swagger:route POST /auth/org/show Organization ShowOrg
	// Show Organizations.
	// Displays existing Organizations in which you are authorized to access.
	// Security:
	//   Bearer:
	// responses:
	//   200: listOrgs
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/org/show", ShowOrg, Show)
	// swagger:route POST /auth/org/delete Organization DeleteOrg
	// Delete Organization.
	// Deletes an existing Organization.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/org/delete", DeleteOrg, Delete)

	createAuthMcApi(auth, authPrefix, "/billingorg/create", CreateBillingOrg, Create)
	// swagger:route POST /auth/billingorg/update BillingOrganization UpdateBillingOrg
	// Update BillingOrganization.
	// API to update an existing BillingOrganization.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/billingorg/update", UpdateBillingOrg, Update)
	// swagger:route POST /auth/billingorg/addchild BillingOrganization AddChildOrg
	// Add Child to BillingOrganization.
	// Adds an Organization to an existing parent BillingOrganization.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/billingorg/addchild", AddChildOrg, Default)
	// swagger:route POST /auth/billingorg/removechild BillingOrganization RemoveChildOrg
	// Remove Child from BillingOrganization.
	// Removes an Organization from an existing parent BillingOrganization.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/billingorg/removechild", RemoveChildOrg, Default)
	// swagger:route POST /auth/billingorg/show BillingOrganization ShowBillingOrg
	// Show BillingOrganizations.
	// Displays existing BillingOrganizations in which you are authorized to access.
	// Security:
	//   Bearer:
	// responses:
	//   200: listBillingOrgs
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/billingorg/show", ShowBillingOrg, Show)
	// swagger:route POST /auth/billingorg/delete BillingOrganization DeleteBillingOrg
	// Delete BillingOrganization.
	// Deletes an existing BillingOrganization.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/billingorg/delete", DeleteBillingOrg, Delete)
	createAuthMcApi(auth, authPrefix, "/billingorg/invoice", GetInvoice, Default)
	createAuthMcApi(auth, authPrefix, "/billingorg/showaccount", ShowAccountInfo, Show)
	createAuthMcApi(auth, authPrefix, "/billingorg/showpaymentprofiles", ShowPaymentInfo, Show)
	createAuthMcApi(auth, authPrefix, "/billingorg/deletepaymentprofile", DeletePaymentInfo, Delete)

	createAuthMcApi(auth, authPrefix, "/controller/create", CreateController, Create)
	createAuthMcApi(auth, authPrefix, "/controller/delete", DeleteController, Delete)
	createAuthMcApi(auth, authPrefix, "/controller/show", ShowController, Show)
	createAuthMcApi(auth, authPrefix, "/gitlab/resync", GitlabResync, Default)
	createAuthMcApi(auth, authPrefix, "/artifactory/resync", ArtifactoryResync, Default)
	createAuthMcApi(auth, authPrefix, "/artifactory/summary", ArtifactorySummary, Default)
	createAuthMcApi(auth, authPrefix, "/config/update", UpdateConfig, Update)
	createAuthMcApi(auth, authPrefix, "/config/reset", ResetConfig, Default)
	createAuthMcApi(auth, authPrefix, "/config/show", ShowConfig, Show)
	createAuthMcApi(auth, authPrefix, "/config/version", ShowVersion, Show)
	createAuthMcApi(auth, authPrefix, "/restricted/user/update", RestrictedUserUpdate, Update)
	createAuthMcApi(auth, authPrefix, "/restricted/org/update", RestrictedUpdateOrg, Update)
	createAuthMcApi(auth, authPrefix, "/audit/showself", ShowAuditSelf, Show)
	createAuthMcApi(auth, authPrefix, "/audit/showorg", ShowAuditOrg, Show)
	createAuthMcApi(auth, authPrefix, "/audit/operations", GetAuditOperations, Default)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccessinvitation/create", CreateCloudletPoolAccessInvitation, Create)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccessinvitation/delete", DeleteCloudletPoolAccessInvitation, Delete)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccessinvitation/show", ShowCloudletPoolAccessInvitation, Show)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccessresponse/create", CreateCloudletPoolAccessResponse, Create)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccessresponse/delete", DeleteCloudletPoolAccessResponse, Delete)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccessresponse/show", ShowCloudletPoolAccessResponse, Show)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccessgranted/show", ShowCloudletPoolAccessGranted, Show)
	createAuthMcApi(auth, authPrefix, "/cloudletpoolaccesspending/show", ShowCloudletPoolAccessPending, Show)
	createAuthMcApi(auth, authPrefix, "/orgcloudlet/show", ShowOrgCloudlet, Show)
	createAuthMcApi(auth, authPrefix, "/orgcloudletinfo/show", ShowOrgCloudletInfo, Show)

	// Support multiple connection types: HTTP(s), Websockets
	addControllerApis("POST", auth, authPrefix)

	// Metrics api route use auth to serve a query to influxDB

	// swagger:route POST /auth/metrics/app DeveloperMetrics AppMetrics
	// App related metrics.
	// Display app related metrics.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/metrics/app", GetMetricsCommon, ShowMetrics)

	// swagger:route POST /auth/metrics/cluster DeveloperMetrics ClusterMetrics
	// Cluster related metrics.
	// Display cluster related metrics.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/metrics/cluster", GetMetricsCommon, ShowMetrics)

	// swagger:route POST /auth/metrics/cloudlet OperatorMetrics CloudletMetrics
	// Cloudlet related metrics.
	// Display cloudlet related metrics.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/metrics/cloudlet", GetMetricsCommon, ShowMetrics)

	// swagger:route POST /auth/metrics/cloudlet/usage OperatorMetrics CloudletUsageMetrics
	// Cloudlet usage related metrics.
	// Display cloudlet usage related metrics.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/metrics/cloudlet/usage", GetMetricsCommon, ShowMetrics)

	// swagger:route POST /auth/metrics/clientapiusage DeveloperMetrics ClientApiUsageMetrics
	// Client api usage related metrics.
	// Display client api usage related metrics.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/metrics/clientapiusage", GetMetricsCommon, ShowMetrics)

	// swagger:route POST /auth/metrics/clientappusage DeveloperMetrics ClientAppUsageMetrics
	// Client app usage related metrics.
	// Display client app usage related metrics.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/metrics/clientappusage", GetMetricsCommon, ShowMetrics)

	// swagger:route POST /auth/metrics/clientcloudletusage DeveloperMetrics ClientCloudletUsageMetrics
	// Client cloudlet usage related metrics.
	// Display client cloudlet usage related metrics.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/metrics/clientcloudletusage", GetMetricsCommon, ShowMetrics)

	createAuthMcApi(auth, authPrefix, "/events/app", GetEventsCommon, Default)
	createAuthMcApi(auth, authPrefix, "/events/cluster", GetEventsCommon, Default)
	createAuthMcApi(auth, authPrefix, "/events/cloudlet", GetEventsCommon, Default)

	// new events/audit apis
	// swagger:route POST /auth/events/show Events SearchEvents
	// Search events
	// Display events based on search filter.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/events/show", ShowEvents, Show)
	// swagger:route POST /auth/events/find Events FindEvents
	// Find events
	// Display events based on find filter.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/events/find", FindEvents, Default)
	// swagger:route POST /auth/events/terms Events TermsEvents
	// Terms Events
	// Display events terms.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/events/terms", EventTerms, Default)

	createAuthMcApi(auth, authPrefix, "/spans/terms", SpanTerms, Default)
	createAuthMcApi(auth, authPrefix, "/spans/show", ShowSpans, Show)
	createAuthMcApi(auth, authPrefix, "/spans/showverbose", ShowSpansVerbose, Show)

	// swagger:route POST /auth/usage/app DeveloperUsage AppUsage
	// App Usage
	// Display app usage.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/usage/app", GetUsageCommon, ShowUsage)
	// swagger:route POST /auth/usage/cluster DeveloperUsage ClusterUsage
	// Cluster Usage
	// Display cluster usage.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/usage/cluster", GetUsageCommon, ShowUsage)
	// swagger:route POST /auth/usage/cloudletpool OperatorUsage CloudletPoolUsage
	// CloudletPool Usage
	// Display cloudletpool usage.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/usage/cloudletpool", GetCloudletPoolUsageCommon, ShowUsage)

	// Alertmanager apis
	// swagger:route POST /auth/alertreceiver/create AlertReceiver CreateAlertReceiver
	// Create Alert Receiver
	// Create alert receiver.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/alertreceiver/create", CreateAlertReceiver, Create)
	// swagger:route POST /auth/alertreceiver/delete AlertReceiver DeleteAlertReceiver
	// Delete Alert Receiver
	// Delete alert receiver.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/alertreceiver/delete", DeleteAlertReceiver, Delete)
	// swagger:route POST /auth/alertreceiver/show AlertReceiver ShowAlertReceiver
	// Show Alert Receiver
	// Show alert receiver.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	createAuthMcApi(auth, authPrefix, "/alertreceiver/show", ShowAlertReceiver, Show)

	createAuthMcApi(auth, authPrefix, "/reporter/create", CreateReporter, Create)
	createAuthMcApi(auth, authPrefix, "/reporter/update", UpdateReporter, Update)
	createAuthMcApi(auth, authPrefix, "/reporter/delete", DeleteReporter, Delete)
	createAuthMcApi(auth, authPrefix, "/reporter/show", ShowReporter, Show)
	createAuthMcApi(auth, authPrefix, "/report/generatedata", GenerateReportData, Default)
	createAuthMcApi(auth, authPrefix, "/report/generate", GenerateReport, Default)
	createAuthMcApi(auth, authPrefix, "/report/show", ShowReport, Show)
	createAuthMcApi(auth, authPrefix, "/report/download", DownloadReport, Default)

	// Generate new short-lived token to authenticate websocket connections
	// Note: Web-client should not store auth token as part of local storage,
	//       instead browser should store it as secure cookies.
	//       For HTTP endpoints, server responds with "set-cookie" to store
	//       cookie in browser. But the same cannot be used for websockets
	//       due to browser limitations.
	//       Hence, authenticated clients can use this API endpoint to
	//       fetch a short-lived token to authenticate websocket endpoints
	auth.POST("/wstoken", GenerateWSAuthToken)

	// Use GET method for websockets as thats the method used
	// in setting up TCP connection by most of the clients
	// Also, authorization is handled as part of websocketUpgrade
	wsPrefix := "ws/" + root + "/auth"
	ws := e.Group(wsPrefix, server.websocketUpgrade)
	addControllerApis("GET", ws, wsPrefix)
	// Metrics api route use ws to serve a query to influxDB
	createMcWebsocketsApi(ws, wsPrefix, "/metrics/app", GetMetricsCommon, Auth, ShowMetrics)
	createMcWebsocketsApi(ws, wsPrefix, "/metrics/cluster", GetMetricsCommon, Auth, ShowMetrics)
	createMcWebsocketsApi(ws, wsPrefix, "/metrics/cloudlet", GetMetricsCommon, Auth, ShowMetrics)
	createMcWebsocketsApi(ws, wsPrefix, "/metrics/cloudlet/usage", GetMetricsCommon, Auth, ShowMetrics)
	createMcWebsocketsApi(ws, wsPrefix, "/metrics/clientapiusage", GetMetricsCommon, Auth, ShowMetrics)
	createMcWebsocketsApi(ws, wsPrefix, "/metrics/clientappusage", GetMetricsCommon, Auth, ShowMetrics)
	createMcWebsocketsApi(ws, wsPrefix, "/metrics/clientcloudletusage", GetMetricsCommon, Auth, ShowMetrics)

	if config.NotifySrvAddr != "" {
		server.notifyServer = &notify.ServerMgr{}
		nodeMgr.RegisterServer(server.notifyServer)

		tlsConfig, err := nodeMgr.InternalPki.GetServerTlsConfig(ctx,
			nodeMgr.CommonName(),
			node.CertIssuerGlobal,
			[]node.MatchCA{node.AnyRegionalMatchCA()})
		if err != nil {
			return nil, err
		}
		edgeproto.InitAlertCache(config.AlertCache)
		// sets the callback to be the alertMgr thread callback
		server.notifyServer.RegisterRecvAlertCache(config.AlertCache)
		if AlertManagerServer != nil {
			config.AlertCache.SetUpdatedCb(AlertManagerServer.UpdateAlert)
		}
		server.notifyServer.Start(nodeMgr.Name(), config.NotifySrvAddr, tlsConfig)
	}
	if config.NotifyAddrs != "" {
		tlsConfig, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
			nodeMgr.CommonName(),
			node.CertIssuerGlobal,
			[]node.MatchCA{node.GlobalMatchCA()})
		if err != nil {
			return nil, err
		}
		addrs := strings.Split(config.NotifyAddrs, ",")
		server.notifyClient = notify.NewClient(nodeMgr.Name(), addrs, edgetls.GetGrpcDialOption(tlsConfig))
		nodeMgr.RegisterClient(server.notifyClient)

		server.notifyClient.Start()
	}

	go func() {
		var err error
		if config.ApiTlsCertFile != "" {
			err = e.StartTLS(config.ServAddr, config.ApiTlsCertFile, config.ApiTlsKeyFile)
		} else {
			err = e.Start(config.ServAddr)
		}
		if err != nil && err != http.ErrServerClosed {
			server.Stop()
			log.FatalLog("Failed to serve", "err", err)
		}
	}()

	ldapServer := ldap.NewServer()
	handler := &ldapHandler{}
	ldapServer.BindFunc("", handler)
	ldapServer.SearchFunc("", handler)
	server.ldapServer = ldapServer
	go func() {
		var err error
		if config.ApiTlsCertFile != "" {
			err = ldapServer.ListenAndServeTLS(config.LDAPAddr, config.ApiTlsCertFile, config.ApiTlsKeyFile)
		} else {
			err = ldapServer.ListenAndServe(config.LDAPAddr)
		}
		if err != nil {
			log.FatalLog("LDAP Server Failed", "err", err)
		}
	}()

	gitlabSync = GitlabNewSync()
	artifactorySync = ArtifactoryNewSync()

	// gitlab/artifactory sync and alertmanager requires data to be initialized
	err = <-server.initDataDone
	if err != nil {
		return nil, err
	}
	gitlabSync.Start(server.done)
	artifactorySync.Start(server.done)
	if AlertManagerServer != nil {
		AlertManagerServer.Start()
	}
	sqlListener, err := initSqlListener(ctx, server.done)
	if err != nil {
		return nil, err
	}
	server.sqlListener = sqlListener
	go func() {
		err := server.sqlListener.Listen(sqlEventsChannel)
		if err != nil && !strings.Contains(err.Error(), "Listener has been closed") {
			log.FatalLog("Failed to listen for sql events", "err", err)
		}
	}()
	if err := allRegionCaches.refreshRegions(ctx); err != nil {
		return nil, err
	}

	return &server, err
}

func (s *Server) WaitUntilReady() error {
	// login won't work until jwt keys are pulled
	<-s.initJWKDone

	// wait until server is online
	for ii := 0; ii < 10; ii++ {
		// if TLS specified, status response will be BadRequest.
		// In any case, as long as the server is responding,
		// then it is ready.
		resp, err := http.Get("http://" + s.config.ServAddr)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for server ready")
}

func (s *Server) Stop() {
	s.stopInitData = true
	close(s.done)
	if s.ldapServer != nil {
		close(s.ldapServer.Quit)
	}
	if s.echo != nil {
		s.echo.Close()
	}
	if s.database != nil {
		s.database.Close()
	}
	if s.sqlListener != nil {
		s.sqlListener.Close()
	}
	if s.sql != nil {
		s.sql.StopLocal()
	}
	if s.vault != nil {
		s.vault.StopLocal()
	}
	if s.notifyServer != nil {
		s.notifyServer.Stop()
	}
	if s.notifyClient != nil {
		s.notifyClient.Stop()
	}
	if AlertManagerServer != nil {
		AlertManagerServer.Stop()
		AlertManagerServer = nil
	}
	nodeMgr.Finish()
}

func ShowVersion(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := GetContext(c)

	if err := authorized(ctx, claims.Username, "", ResourceConfig, ActionView); err != nil {
		return err
	}
	ver := ormapi.Version{
		BuildMaster: version.BuildMaster,
		BuildHead:   version.BuildHead,
		BuildAuthor: version.BuildAuthor,
		Hostname:    cloudcommon.Hostname(),
	}
	return c.JSON(http.StatusOK, ver)
}

func (s *Server) websocketUpgrade(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		upgrader := websocket.Upgrader{}
		if s.config.SkipOriginCheck {
			// Skip origin check restriction.
			// This is to be used for testing purpose only, as it is
			// not safe to allow all origins
			upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return nil
		}
		defer ws.Close()

		// Verify Auth
		// ===========
		// JWT token is received after websocket connection is established, although
		// Websocket server can receive full request header from client before
		// upgrade to websocket

		// Infact most of the golang websocket clients do support that. But the problem
		// is on the UI side. Javascript doesn't support it directly

		// Following are some links describing this issue:
		//  - https://stackoverflow.com/questions/22383089/is-it-possible-to-use-bearer-authentication-for-websocket-upgrade-requests/26123316#26123316
		// The above URL does give another way to send access token, but then it is not
		// safe enough to use

		// Here's another way to solve this, but again complicated and insecure:
		//  - https://devcenter.heroku.com/articles/websocket-security#authentication-authorization

		// In summary, it is not straightforward to implement this from our console UI
		// as we plan to call this directly from React (browser)
		isAuth, err := AuthWSCookie(c, ws)
		if !isAuth {
			return err
		}

		// Set ws on echo context
		SetWs(c, ws)

		// call next handler
		err = next(c)

		// Any handler errors are sent via the websocket here before
		// it is closed. The error is also passed to the caller, but
		// only for being recorded in the audit log.
		if err != nil {
			code, res := getErrorResult(err)
			wsPayload := ormapi.WSStreamPayload{
				Code: code,
				Data: res,
			}
			writeErr := writeWS(c, ws, &wsPayload)
			if writeErr != nil {
				ctx := GetContext(c)
				log.SpanLog(ctx, log.DebugLevelApi, "Failed to write error to websocket stream", "err", err, "writeErr", writeErr)
			}
		}

		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

		return err
	}
}

const StreamAPITag = "StreamAPITag"

func ReadConn(c echo.Context, in interface{}) (bool, error) {
	var err error

	// Mark stream API
	c.Set(StreamAPITag, true)

	if ws := GetWs(c); ws != nil {
		err = ws.ReadJSON(in)
		if err == nil {
			out, err := json.Marshal(in)
			if err == nil {
				LogWsRequest(c, out)
			}
		}
	} else {
		err = c.Bind(in)
	}
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return false, fmt.Errorf("Invalid data")
		}
		// echo returns HTTPError which may include "code: ..., message:", chop code
		if errObj, ok := err.(*echo.HTTPError); ok {
			err = fmt.Errorf("%v", errObj.Message)
		}

		errStr := checkForTimeError(fmt.Sprintf("Invalid data: %v", err))
		return false, fmt.Errorf(errStr)
	}

	return true, nil
}

func WaitForConnClose(c echo.Context, serverClosed chan bool) {
	if ws := GetWs(c); ws != nil {
		clientClosed := make(chan error)
		go func() {
			// Handling close events from client is different here
			// A close message is sent from client, hence just wait
			// on getting a close message
			_, _, err := ws.ReadMessage()
			clientClosed <- err
		}()
		select {
		case <-serverClosed:
			return
		case err := <-clientClosed:
			if _, ok := err.(*websocket.CloseError); !ok {
				ws.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""))
				ws.Close()
			}
		}
	} else {
		if <-serverClosed {
			return
		}
	}
}

func writeWS(c echo.Context, ws *websocket.Conn, wsPayload *ormapi.WSStreamPayload) error {
	out, err := json.Marshal(wsPayload)
	if err == nil {
		LogWsResponse(c, string(out))
	}
	return ws.WriteJSON(wsPayload)
}

func WriteStream(c echo.Context, payload *ormapi.StreamPayload) error {
	if ws := GetWs(c); ws != nil {
		wsPayload := ormapi.WSStreamPayload{
			Code: http.StatusOK,
			Data: (*payload).Data,
		}
		return writeWS(c, ws, &wsPayload)
	} else {
		// stream func may return "forbidden", so don't write
		// header until we know it's ok
		if !c.Response().Committed {
			// Write header now that we're streaming back data
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c.Response().WriteHeader(http.StatusOK)
		}
		err := json.NewEncoder(c.Response()).Encode(*payload)
		if err != nil {
			return err
		}
		c.Response().Flush()
	}
	return nil
}

// Create MC API Echo Route and Adds API to ratelimitmgr
func createMcApi(e *echo.Echo, prefix string, path string, h echo.HandlerFunc, apiAuthType ApiAuthType, apiActionType ApiActionType) {
	e.POST(prefix+path, h)
	addApiRateLimit(prefix, path, apiAuthType, Mc, apiActionType)
}

// Create MC API in the AUTH group and Adds API to ratelimitmgr
func createAuthMcApi(auth *echo.Group, prefix string, path string, h echo.HandlerFunc, apiActionType ApiActionType) {
	auth.POST(path, h)
	addApiRateLimit(prefix, path, Auth, Mc, apiActionType)
}

// Create MC API that uses websockets route and Adds API to ratelimitmgr
func createMcWebsocketsApi(ws *echo.Group, prefix string, path string, h echo.HandlerFunc, apiAuthType ApiAuthType, apiActionType ApiActionType) {
	ws.GET(path, h)
	addApiRateLimit(prefix, path, apiAuthType, Mc, apiActionType)
}

// add api to ratelimitmgr
func addApiRateLimit(prefix string, path string, apiAuthType ApiAuthType, apiType ApiType, apiActionType ApiActionType) {
	// If RemoveRateLimit, do not add to ratelimitmgr
	if serverConfig.RemoveRateLimit {
		return
	}
	methodName := "/" + prefix + path
	var fullEpRateLimitSettings *edgeproto.RateLimitSettings
	var perIpRateLimitSettings *edgeproto.RateLimitSettings
	if apiType == Controller {
		// If controller API, use LeakyBucket to "leak" requests to controller and let controller accept or reject
		fullEpRateLimitSettings = McControllerApiFullEndpointRateLimitSettings
		perIpRateLimitSettings = McControllerApiPerIpRateLimitSettings
	} else {
		if apiAuthType == Auth {
			// Switch through different MC API action types
			switch apiActionType {
			case Create:
				fullEpRateLimitSettings = McCreateApiFullEndpointRateLimitSettings
				perIpRateLimitSettings = McCreateApiPerIpRateLimitSettings
			case Delete:
				fullEpRateLimitSettings = McDeleteApiFullEndpointRateLimitSettings
				perIpRateLimitSettings = McDeleteApiPerIpRateLimitSettings
			case Show:
				fullEpRateLimitSettings = McShowApiFullEndpointRateLimitSettings
				perIpRateLimitSettings = McShowApiPerIpRateLimitSettings
			case Update:
				fullEpRateLimitSettings = McUpdateApiFullEndpointRateLimitSettings
				perIpRateLimitSettings = McUpdateApiPerIpRateLimitSettings
			case ShowMetrics:
				fullEpRateLimitSettings = McShowMetricsApiFullEndpointRateLimitSettings
				perIpRateLimitSettings = McShowMetricsApiPerIpRateLimitSettings
			case ShowUsage:
				fullEpRateLimitSettings = McShowUsageApiFullEndpointRateLimitSettings
				perIpRateLimitSettings = McShowUsageApiPerIpRateLimitSettings
			case Default:
				fallthrough
			default:
				fullEpRateLimitSettings = McDefaultApiFullEndpointRateLimitSettings
				perIpRateLimitSettings = McDefaultApiPerIpRateLimitSettings
			}
		} else {
			// No auth MC APIs
			fullEpRateLimitSettings = NoAuthMcApiFullEndpointRateLimitSettings
			perIpRateLimitSettings = NoAuthMcApiPerIpRateLimitSettings
		}
	}
	rateLimitMgr.AddApiEndpointLimiter(methodName, fullEpRateLimitSettings, perIpRateLimitSettings, nil, nil)
}
