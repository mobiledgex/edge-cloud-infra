package edgebox

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

// edgebox wraps the generic dind implementation with
// mex-specific behavior, such as setting up DNS and
// registry.mobiledgex.net access secrets.

type EdgeboxPlatform struct {
	generic       dind.Platform
	NetworkScheme string
	commonPf      infracommon.CommonPlatform
}

var edgeboxProps = map[string]*infracommon.PropertyInfo{
	"MEX_NETWORK_SCHEME": &infracommon.PropertyInfo{
		Value: cloudcommon.NetworkSchemePrivateIP,
	},
}

func (e *EdgeboxPlatform) GetType() string {
	return "edgebox"
}

func (e *EdgeboxPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	err := e.generic.Init(ctx, platformConfig, updateCallback)
	// Set the test Mode based on what is in PlatformConfig
	infracommon.SetTestMode(platformConfig.TestMode)

	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}

	if err := e.commonPf.InitInfraCommon(ctx, platformConfig, edgeboxProps, vaultConfig, e); err != nil {
		return err
	}

	e.NetworkScheme = e.commonPf.GetCloudletNetworkScheme()
	if e.NetworkScheme != cloudcommon.NetworkSchemePrivateIP &&
		e.NetworkScheme != cloudcommon.NetworkSchemePublicIP {
		return fmt.Errorf("Unsupported network scheme for DIND: %s", e.NetworkScheme)
	}

	fqdn := cloudcommon.GetRootLBFQDN(platformConfig.CloudletKey)
	ipaddr, err := e.GetDINDServiceIP(ctx)
	if err != nil {
		return fmt.Errorf("init cannot get service ip, %s", err.Error())
	}
	if err := e.commonPf.ActivateFQDNA(ctx, fqdn, ipaddr); err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "error in ActivateFQDNA", "err", err)
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "done init edgebox")
	return nil
}

func (e *EdgeboxPlatform) GetCloudletNetworkScheme() string {
	return e.commonPf.Properties["MEX_NETWORK_SCHEME"].Value
}

func (e *EdgeboxPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	return e.generic.GatherCloudletInfo(ctx, info)
}

func (e *EdgeboxPlatform) GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	return e.generic.GetPlatformClient(ctx, clusterInst)
}

func (e *EdgeboxPlatform) NameSanitize(string) string {
	return "not implemented"
}

func (e *EdgeboxPlatform) AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) GetServerDetail(ctx context.Context, serverName string) (*infracommon.ServerDetail, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) GetIPFromServerName(ctx context.Context, networkName, serverName string) (*infracommon.ServerIP, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) GetClusterMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error) {
	return "", "", fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) AttachPortToServer(ctx context.Context, serverName, portName string) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) DetachPortFromServer(ctx context.Context, serverName, portName string) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) CreateAppVM(ctx context.Context, vmAppParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) CreateAppVMWithRootLB(ctx context.Context, vmAppParams, vmLbParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) CreateRootLBVM(ctx context.Context, serverName, stackName, imgName string, vmspec *vmspec.VMCreationSpec, cloudletKey *edgeproto.CloudletKey, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) CreateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) UpdateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) DeleteClusterResources(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst, rootLBName string, dedicatedRootLB bool) error {
	return fmt.Errorf("not implemented")
}

func (o *EdgeboxPlatform) DeleteResources(ctx context.Context, resourceGroupName string) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) NetworkSetupForRootLB(ctx context.Context, client ssh.Client, rootLBName string) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) WhitelistSecurityRules(ctx context.Context, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error {
	return fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) GetVMParams(ctx context.Context, depType infracommon.DeploymentType, serverName, flavorName string, externalVolumeSize uint64, imageName, secGrp string, cloudletKey *edgeproto.CloudletKey, opts ...infracommon.VMParamsOp) (*infracommon.VMParams, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EdgeboxPlatform) Resync(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
