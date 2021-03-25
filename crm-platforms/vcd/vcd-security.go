package vcd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

var VCDClientCtxKey = "VCDClientCtxKey"

var NoVCDClientInContext = "No VCD Client in Context"

var globalVCDClient *govcd.VCDClient
var globalVCDClientLock sync.Mutex
var globalVCDClientLastUpdateTime time.Time

// vcd security related operations

// physicalname (vault key) not needed when  using insure env vars.
func (v *VcdPlatform) PopulateOrgLoginCredsFromEnv(ctx context.Context) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "PopulateOrgLoginCredsFromEnv")

	creds := VcdConfigParams{
		User:      os.Getenv("VCD_USER"),
		Password:  os.Getenv("VCD_PASSWORD"),
		Org:       os.Getenv("VCD_ORG"),
		VcdApiUrl: os.Getenv("VCD_URL"),
		VDC:       os.Getenv("VDC_NAME"),
		Insecure:  true,
	}
	if creds.User == "" {
		return fmt.Errorf("User not defined")
	}
	if creds.Password == "" {
		return fmt.Errorf("Passwd not defined")
	}
	if creds.Org == "" {
		return fmt.Errorf("Org not defined")
	}
	if creds.VcdApiUrl == "" {
		return fmt.Errorf("VcdApiUrl not defined")
	}
	if creds.VDC == "" {
		return fmt.Errorf("missing VDC name")
	}
	v.Creds = &creds
	return nil
}

func (v *VcdPlatform) GetVcdUser() string {
	return v.Creds.User
}
func (v *VcdPlatform) GetVcdPassword() string {
	return v.Creds.Password
}
func (v *VcdPlatform) GetVcdOrgName() string {
	return v.Creds.Org
}
func (v *VcdPlatform) GetVcdVdcName() string {
	return v.Creds.VDC
}

func (v *VcdPlatform) PopulateOrgLoginCredsFromVcdVars(ctx context.Context) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "PopulateOrgLoginCredsFromVault")
	creds := VcdConfigParams{
		User:                  v.GetVCDUser(),
		Password:              v.GetVCDPassword(),
		Org:                   v.GetVCDORG(),
		VcdApiUrl:             v.GetVcdUrl(),
		VDC:                   v.GetVDCName(),
		OauthSgwUrl:           v.GetVcdOauthSgwUrl(),
		OauthAgwUrl:           v.GetVcdOauthAgwUrl(),
		OauthClientId:         v.GetVcdOauthClientId(),
		OauthClientSecret:     v.GetVcdOauthClientSecret(),
		ClientTlsCert:         v.GetVcdClientTlsCert(),
		ClientTlsKey:          v.GetVcdClientTlsKey(),
		ClientRefreshInterval: v.GetVcdClientRefreshInterval(ctx),
		Insecure:              true,
	}
	if creds.OauthSgwUrl != "" {
		if creds.OauthAgwUrl == "" || creds.OauthClientId == "" || creds.OauthClientSecret == "" {
			return fmt.Errorf("OauthAgwUrl is set but other OAUTH related parameter(s) are empty")
		}
	}

	if creds.User == "" {
		return fmt.Errorf("User not defined")
	}
	if creds.Password == "" {
		return fmt.Errorf("Passwd not defined")
	}
	if creds.Org == "" {
		return fmt.Errorf("Org not defined")
	}
	if creds.VcdApiUrl == "" {
		return fmt.Errorf("VCD Href not defined")
	}
	if creds.VDC == "" {
		return fmt.Errorf("missing VDC name")
	}
	v.Creds = &creds

	log.SpanLog(ctx, log.DebugLevelInfra, "client login creds", "user", creds.User, "Org", creds.Org, "Vdc", creds.VDC, "URI", "creds.Href")

	return nil
}

func (v *VcdPlatform) GetExternalIpNetworkCidr(ctx context.Context, vcdClient *govcd.VCDClient) (string, error) {

	extNet, err := v.GetExtNetwork(ctx, vcdClient)
	if err != nil {
		return "", err
	}

	scope := extNet.OrgVDCNetwork.Configuration.IPScopes.IPScope[0]
	cidr, err := MaskToCidr(scope.Netmask)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalIpNetworkCidr error converting mask to cider", "cidr", cidr, "error", err)
		return "", err
	}
	addr := scope.Gateway + "/" + cidr

	log.SpanLog(ctx, log.DebugLevelInfra, "GetExternalIpNetworkCidr", "addr", addr)

	return addr, nil

}

// same as vsphere (common vmware utils?)
func (v *VcdPlatform) PrepareRootLB(ctx context.Context, client ssh.Client, rootLBName string, secGrpName string, trustPolicy *edgeproto.TrustPolicy, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "PrepareRootLB", "rootLBName", rootLBName)
	iptblStart := time.Now()
	vcdClient := v.GetVcdClientFromContext(ctx)
	if vcdClient == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, NoVCDClientInContext)
		return fmt.Errorf(NoVCDClientInContext)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "PrepareRootLB", "rootLBName", rootLBName)
	// configure iptables based security
	sshCidrsAllowed := []string{}
	externalNet, err := v.GetExternalIpNetworkCidr(ctx, vcdClient)
	if err != nil {
		return err
	}
	sshCidrsAllowed = append(sshCidrsAllowed, externalNet)
	err = v.vmProperties.SetupIptablesRulesForRootLB(ctx, client, sshCidrsAllowed, trustPolicy)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "PrepareRootLB SetupIptableRulesForRootLB failed", "rootLBName", rootLBName, "err", err)
		return err
	}
	if v.Verbose {
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Setup Root LB time %s", cloudcommon.FormatDuration(time.Since(iptblStart), 2)))
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "PrepareRootLB SetupIptableRulesForRootLB complete", "rootLBName", rootLBName, "time", time.Since(iptblStart).String())
	return nil
}

func (v *VcdPlatform) WhitelistSecurityRules(ctx context.Context, client ssh.Client, wlParams *infracommon.WhiteListParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "WhitelistSecurityRules", "wlParams", wlParams)
	// this can be called during LB init so we need to ensure we can reach the server before trying iptables commands
	err := vmlayer.WaitServerReady(ctx, v, client, wlParams.ServerName, vmlayer.MaxRootLBWait)
	if err != nil {
		return err
	}
	return infracommon.AddIngressIptablesRules(ctx, client, wlParams.Label, wlParams.AllowedCIDR, wlParams.DestIP, wlParams.Ports)
}

func (v *VcdPlatform) RemoveWhitelistSecurityRules(ctx context.Context, client ssh.Client, wlParams *infracommon.WhiteListParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RemoveWhitelistSecurityRules", "wlParams", wlParams)
	return infracommon.RemoveIngressIptablesRules(ctx, client, wlParams.Label, wlParams.AllowedCIDR, wlParams.DestIP, wlParams.Ports)
}

func setVer(cli *govcd.VCDClient) error {
	if cli != nil {
		cli.Client.APIVersion = "33.0" // vmwarelab is 10.1
	}
	return nil
}

// GetClient gets a new client object.  Copies are made of the global client object, which instantiates a new
// http client but shares the access token
func (v *VcdPlatform) GetClient(ctx context.Context, creds *VcdConfigParams) (client *govcd.VCDClient, err error) {
	apiUrl := creds.VcdApiUrl
	if creds.OauthAgwUrl != "" {
		apiUrl = creds.OauthAgwUrl
	}
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse request to org %s at %s err: %s", creds.Org, creds.VcdApiUrl, err)
	}
	if v.TestMode {
		err := v.PopulateOrgLoginCredsFromEnv(ctx)
		if err != nil {
			return nil, err
		}
	}
	if creds == nil {
		// usually indicates we called GetClient before InitProvider
		return nil, fmt.Errorf("nil creds passed to GetClient")
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "GetClient", "user", creds.User, "OauthSgwUrl", creds.OauthSgwUrl)
	globalVCDClientLock.Lock()
	defer globalVCDClientLock.Unlock()

	tokenDuration := time.Since(globalVCDClientLastUpdateTime)
	expired := (uint64(tokenDuration.Seconds()) >= creds.ClientRefreshInterval)
	log.SpanLog(ctx, log.DebugLevelInfra, "Check for token expired", "tokenDuration", tokenDuration, "ClientRefreshInterval", creds.ClientRefreshInterval, "exipred", expired)
	if globalVCDClient == nil || expired {
		log.SpanLog(ctx, log.DebugLevelInfra, "Need to refresh global token")

		globalVCDClient = govcd.NewVCDClient(*u, creds.Insecure, govcd.WithOauthUrl(creds.OauthSgwUrl), govcd.WithClientTlsCerts(creds.ClientTlsCert, creds.ClientTlsKey))
		user := creds.User
		pass := creds.Password

		if creds.OauthSgwUrl != "" {
			log.SpanLog(ctx, log.DebugLevelInfra, "Using OAUTH client credentials", "OauthSgwUrl", creds.OauthSgwUrl)
			user = creds.OauthClientId
			pass = creds.OauthClientSecret
		}
		_, err := globalVCDClient.GetAuthResponse(user, pass, creds.Org)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Unable to login to org", "org", creds.Org, "err", err)
			globalVCDClient = nil
			return nil, fmt.Errorf("Unable to login to org %s at %s err: %s", creds.Org, creds.VcdApiUrl, err)
		}
		globalVCDClientLastUpdateTime = time.Now()
	}
	vcdClient, err := globalVCDClient.CopyClient()
	if err != nil {
		return nil, fmt.Errorf("CopyClient failed - %v", err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetClient connected", "API Version", vcdClient.Client.APIVersion)
	return vcdClient, nil
}

// New
func (v *VcdPlatform) ConfigureCloudletSecurityRules(ctx context.Context, egressRestricted bool, TrustPolicy *edgeproto.TrustPolicy, action vmlayer.ActionType, updateCallback edgeproto.CacheUpdateCallback) error {

	log.SpanLog(ctx, log.DebugLevelInfra, "ConfigureCloudletSecurityRules tbi")
	return nil
}

// GetVcdClientFromContext returns a client object if one exists, otherwise nil
func (v *VcdPlatform) GetVcdClientFromContext(ctx context.Context) *govcd.VCDClient {
	vcdClient, found := ctx.Value(VCDClientCtxKey).(*govcd.VCDClient)
	if !found {
		return nil
	}
	return vcdClient
}

func (v *VcdPlatform) InitOperationContext(ctx context.Context, operationStage vmlayer.OperationInitStage) (context.Context, vmlayer.OperationInitResult, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitOperationContext", "operationStage", operationStage)

	if operationStage == vmlayer.OperationInitStart {
		// getClient will setup the client within the context.  First ensure it is not already set, which
		// indicates an error because we don't want it to be setup twice as it may get cleaned up erroneously.
		// So we look for the client and expect a NoVCDClientInContext error
		vcdClient := v.GetVcdClientFromContext(ctx)
		if vcdClient != nil {
			// This indicates we called InitOperationContext with OperationInitStart twice before OperationInitComplete
			// which is unavoidable in some flows
			log.SpanLog(ctx, log.DebugLevelInfra, "InitOperationContext VCDClient is already in context")
			return ctx, vmlayer.OperationAlreadyInitialized, nil
		}
		// now get a new client
		var err error
		vcdClient, err = v.GetClient(ctx, v.Creds)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Failed to initialize vcdClient", "err", err)
			return ctx, vmlayer.OperationInitFailed, err
		} else {
			ctx = context.WithValue(ctx, VCDClientCtxKey, vcdClient)
			log.SpanLog(ctx, log.DebugLevelInfra, "Updated context with client", "APIVersion", vcdClient.Client.APIVersion, "key", VCDClientCtxKey)
			return ctx, vmlayer.OperationNewlyInitialized, nil
		}
	} else {
		// because we re-use copies of the context, we do not try to disconnect the client.
		// Disconnect generally does not work in VCD anyway
		return ctx, vmlayer.OperationNewlyInitialized, nil
	}
}
