package openstack

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

type OpenstackPlatform struct {
	openRCVars map[string]string
	commonPf   infracommon.CommonPlatform
}

func (o *OpenstackPlatform) GetType() string {
	return "openstack"
}

func (o *OpenstackPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx,
		log.DebugLevelInfra, "init OpenstackPlatform",
		"physicalName", platformConfig.PhysicalName,
		"vaultAddr", platformConfig.VaultAddr)

	updateCallback(edgeproto.UpdateTask, "Initializing Openstack platform")

	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "vault auth", "type", vaultConfig.Auth.Type())

	updateCallback(edgeproto.UpdateTask, "Fetching Openstack access credentials")
	if err := o.commonPf.InitInfraCommon(ctx, platformConfig, openstackProps, vaultConfig, o); err != nil {
		return err
	}

	if err := o.InitOpenstackProps(ctx, platformConfig.CloudletKey, platformConfig.Region, platformConfig.PhysicalName, vaultConfig, platformConfig.EnvVars); err != nil {
		return err
	}

	o.commonPf.FlavorList, _, _, err = o.GetFlavorInfo(ctx)
	if err != nil {
		return err
	}

	// create rootLB
	sharedRootLbName := o.commonPf.GetRootLBName(platformConfig.CloudletKey)
	updateCallback(edgeproto.UpdateTask, "Creating RootLB")

	crmRootLB, cerr := o.commonPf.NewRootLB(ctx, sharedRootLbName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	o.commonPf.SharedRootLBName = sharedRootLbName
	o.commonPf.SharedRootLB = crmRootLB
	log.SpanLog(ctx, log.DebugLevelInfra, "created shared rootLB", "name", sharedRootLbName)

	vmspec, err := o.commonPf.GetVMSpecForRootLB()
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "calling SetupRootLB")
	updateCallback(edgeproto.UpdateTask, "Setting up RootLB")
	err = o.commonPf.SetupRootLB(ctx, o.commonPf.SharedRootLBName, vmspec, platformConfig.CloudletKey, platformConfig.CloudletVMImagePath, platformConfig.VMImageVersion, edgeproto.DummyUpdateCallback)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "ok, SetupRootLB")

	// set up L7 load balancer
	client, err := o.commonPf.GetSSHClientForServer(ctx, o.commonPf.SharedRootLBName, o.commonPf.GetCloudletExternalNetwork())
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Setting up Proxy")
	err = proxy.InitL7Proxy(ctx, client, proxy.WithDockerNetwork("host"))
	if err != nil {
		return err
	}
	return o.PrepNetwork(ctx)
}

func (o *OpenstackPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	return o.OSGetLimits(ctx, info)
}

// alphanumeric plus -_. first char must be alpha, <= 255 chars.
func (o *OpenstackPlatform) NameSanitize(name string) string {
	r := strings.NewReplacer(
		" ", "",
		"&", "",
		",", "",
		"!", "")
	str := r.Replace(name)
	if str == "" {
		return str
	}
	if !unicode.IsLetter(rune(str[0])) {
		// first character must be alpha
		str = "a" + str
	}
	if len(str) > 255 {
		str = str[:254]
	}
	return str
}

func (o *OpenstackPlatform) GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	return o.commonPf.GetSSHClientForCluster(ctx, clusterInst)
}

func (o *OpenstackPlatform) DeleteResources(ctx context.Context, resourceGroupName string) error {
	return o.HeatDeleteStack(ctx, resourceGroupName)
}

func (o *OpenstackPlatform) CreateAppVM(ctx context.Context, vmAppParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatCreateAppVM(ctx, vmAppParams, updateCallback)
}

func (o *OpenstackPlatform) CreateAppVMWithRootLB(ctx context.Context, vmAppParams, vmLbParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatCreateAppVMWithRootLB(ctx, vmAppParams, vmLbParams, updateCallback)
}

func (o *OpenstackPlatform) CreateRootLBVM(ctx context.Context, serverName, stackName, imgName string, vmspec *vmspec.VMCreationSpec, cloudletKey *edgeproto.CloudletKey, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatCreateRootLBVM(ctx, serverName, stackName, imgName, vmspec, cloudletKey, updateCallback)
}

func (o *OpenstackPlatform) CreateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatCreateCluster(ctx, clusterInst, privacyPolicy, rootLBName, imgName, dedicatedRootLB, updateCallback)
}

func (o *OpenstackPlatform) UpdateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatUpdateCluster(ctx, clusterInst, privacyPolicy, rootLBName, imgName, dedicatedRootLB, updateCallback)
}

func (o *OpenstackPlatform) DeleteClusterResources(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst, rootLBName string, dedicatedRootLB bool) error {
	return o.HeatDeleteCluster(ctx, client, clusterInst, rootLBName, dedicatedRootLB)
}

func (o *OpenstackPlatform) Resync(ctx context.Context) error {
	return fmt.Errorf("Resync not yet implemented")
}
