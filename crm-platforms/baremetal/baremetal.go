package baremetal

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type BareMetalPlatform struct {
	commonPf           infracommon.CommonPlatform
	caches             *platform.Caches
	FlavorList         []*edgeproto.FlavorInfo
	sharedLBName       string
	cloudletKubeConfig string
	externalIps        []string
	internalIps        []string
}

var RootLBFlavor = edgeproto.Flavor{
	Key:   edgeproto.FlavorKey{Name: "rootlb-flavor"},
	Vcpus: uint64(2),
	Ram:   uint64(4096),
	Disk:  uint64(40),
}

func (b *BareMetalPlatform) GetCloudletKubeConfig(cloudletKey *edgeproto.CloudletKey) string {
	return fmt.Sprintf("%s-%s", cloudletKey.Name, "cloudlet-kubeconfig")
}

func (b *BareMetalPlatform) IsCloudletServicesLocal() bool {
	return false
}

func (b *BareMetalPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, caches *platform.Caches, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Init start")
	b.caches = caches
	if err := b.commonPf.InitInfraCommon(ctx, platformConfig, baremetalProps); err != nil {
		return err
	}
	externalIps, err := infracommon.ParseIpRanges(b.GetExternalIpRanges())
	if err != nil {
		return err
	}
	b.externalIps = externalIps
	internalIps, err := infracommon.ParseIpRanges(b.GetInternalIpRanges())
	if err != nil {
		return err
	}
	if len(externalIps) > len(internalIps) {
		log.SpanLog(ctx, log.DebugLevelInfra, "Not enough internal IPs", "numexternal", len(externalIps), "numinternal", len(internalIps))
		return fmt.Errorf("Number of internal IPs defined in BARE_METAL_INTERNAL_IP_RANGES must be b. least b. many b. BARE_METAL_EXTERNAL_IP_RANGES")
	}
	b.internalIps = internalIps
	b.sharedLBName = b.GetSharedLBName(ctx, platformConfig.CloudletKey)
	b.cloudletKubeConfig = b.GetCloudletKubeConfig(platformConfig.CloudletKey)

	if !platformConfig.TestMode {
		err := b.commonPf.InitCloudletSSHKeys(ctx, platformConfig.AccessApi)
		if err != nil {
			return err
		}
		go b.commonPf.RefreshCloudletSSHKeys(platformConfig.AccessApi)
	}

	client, err := b.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: platformConfig.CloudletKey.String(), Type: "baremetalcontrolhost"})
	if err != nil {
		return err
	}
	err = b.SetupLb(ctx, client, b.sharedLBName)
	if err != nil {
		return err
	}
	return nil
}

// TODO: this needs work
func (b *BareMetalPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "GatherCloudletInfo")
	var err error
	info.Flavors, err = b.GetFlavorList(ctx)
	return err
}

func (b *BareMetalPlatform) GetCloudletInfraResources(ctx context.Context) (*edgeproto.InfraResourcesSnapshot, error) {
	var resources edgeproto.InfraResourcesSnapshot
	return &resources, nil
}

// called by controller, make sure it doesn't make b.y calls to infra API
func (b *BareMetalPlatform) GetClusterAdditionalResources(ctx context.Context, cloudlet *edgeproto.Cloudlet, vmResources []edgeproto.VMResource, infraResMap map[string]edgeproto.InfraResource) map[string]edgeproto.InfraResource {
	resInfo := make(map[string]edgeproto.InfraResource)
	return resInfo
}

func (b *BareMetalPlatform) GetClusterAdditionalResourceMetric(ctx context.Context, cloudlet *edgeproto.Cloudlet, resMetric *edgeproto.Metric, resources []edgeproto.VMResource) error {
	externalIpsUsed := uint64(0)
	for _, vmRes := range resources {
		if vmRes.Type == cloudcommon.VMTypeRootLB {
			externalIpsUsed += 1
		}
	}
	resMetric.AddIntVal("externalIpsUsed", externalIpsUsed)
	return nil
}

func (b *BareMetalPlatform) GetClusterInfraResources(ctx context.Context, clusterKey *edgeproto.ClusterInstKey) (*edgeproto.InfraResources, error) {
	var resources edgeproto.InfraResources
	return &resources, nil
}

func (b *BareMetalPlatform) GetAppInstRuntime(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst) (*edgeproto.AppInstRuntime, error) {
	return &edgeproto.AppInstRuntime{}, nil
}

func (b *BareMetalPlatform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst, clientType string) (ssh.Client, error) {
	return nil, fmt.Errorf("GetClusterPlatformClient TODO")
}

func (b *BareMetalPlatform) GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode, ops ...pc.SSHClientOp) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNodePlatformClient", "node", node)
	if node == nil || node.Name == "" {
		return nil, fmt.Errorf("cannot GetNodePlatformClient, b. node details b.e empty")
	}
	controlIp := b.GetControlAccessIp()
	return b.commonPf.GetSSHClientFromIPAddr(ctx, controlIp, ops...)
}

func (b *BareMetalPlatform) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst) ([]edgeproto.CloudletMgmtNode, error) {
	return []edgeproto.CloudletMgmtNode{}, nil
}

func (b *BareMetalPlatform) GetContainerCommand(ctx context.Context, clusterInst *edgeproto.ClusterInst, app *edgeproto.App, appInst *edgeproto.AppInst, req *edgeproto.ExecRequest) (string, error) {
	return "", fmt.Errorf("GetContainerCommand TODO")
}

func (b *BareMetalPlatform) GetConsoleUrl(ctx context.Context, app *edgeproto.App) (string, error) {
	return "", fmt.Errorf("GetConsoleUrl not supported on BareMetal")
}

func (b *BareMetalPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SaveCloudletAccessVars", "cloudletName", cloudlet.Key.Name)
	return nil
}

func (b *BareMetalPlatform) GetCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config) (map[string]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletAccessVars", "cloudletName", cloudlet.Key.Name)
	return map[string]string{}, nil
}

func (b *BareMetalPlatform) DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCloudletAccessVars", "cloudletName", cloudlet.Key.Name)
	return nil
}

func (b *BareMetalPlatform) SetPowerState(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("SetPowerState not supported on BareMetal")
}

func (b *BareMetalPlatform) runDebug(ctx context.Context, req *edgeproto.DebugRequest) string {
	return "runDebug todo"
}

func (b *BareMetalPlatform) SyncControllerCache(ctx context.Context, caches *platform.Caches, cloudletState dme.CloudletState) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncControllerCache", "state", cloudletState)
	return nil
}

func (b *BareMetalPlatform) GetCloudletManifest(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi platform.AccessApi, flavor *edgeproto.Flavor, caches *platform.Caches) (*edgeproto.CloudletManifest, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Get cloudlet manifest", "cloudletName", cloudlet.Key.Name)
	return &edgeproto.CloudletManifest{Manifest: "fake manifest\n" + pfConfig.CrmAccessPrivateKey}, nil
}

func (b *BareMetalPlatform) VerifyVMs(ctx context.Context, vms []edgeproto.VM) error {
	return nil
}

func (b *BareMetalPlatform) GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error) {
	return nil, fmt.Errorf("GetAccessData TODO")
}

func (b *BareMetalPlatform) GetRestrictedCloudletStatus(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, accessApi platform.AccessApi, updateCallback edgeproto.CacheUpdateCallback) error {
	return nil
}

func (b *BareMetalPlatform) GetRootLBClients(ctx context.Context) (map[string]ssh.Client, error) {
	return nil, nil
}

func (b *BareMetalPlatform) GetVersionProperties() map[string]string {
	return map[string]string{}
}

func (b *BareMetalPlatform) GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error) {
	return &RootLBFlavor, nil
}
