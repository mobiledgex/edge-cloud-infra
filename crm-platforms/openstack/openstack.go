package openstack

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
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
	sharedRootLbName := o.commonPf.GetRootLBName(platformConfig.CloudletKey)

	log.SpanLog(ctx,
		log.DebugLevelMexos, "init openstack",
		"physicalName", platformConfig.PhysicalName,
		"vaultAddr", platformConfig.VaultAddr)

	updateCallback(edgeproto.UpdateTask, "Initializing Openstack platform")

	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "vault auth", "type", vaultConfig.Auth.Type())

	updateCallback(edgeproto.UpdateTask, "Fetching Openstack access credentials")
	if err := o.commonPf.InitInfraCommon(ctx, platformConfig, openstackProps, vaultConfig, o, nil); err != nil {
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
	updateCallback(edgeproto.UpdateTask, "Creating RootLB")
	crmRootLB, cerr := o.commonPf.NewRootLB(ctx, o.commonPf.SharedRootLBName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	o.commonPf.SharedRootLBName = sharedRootLbName
	o.commonPf.SharedRootLB = crmRootLB
	log.SpanLog(ctx, log.DebugLevelMexos, "created shared rootLB", "name", crmRootLB.Name)

	vmspec, err := o.commonPf.GetVMSpecForRootLB()
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelMexos, "calling SetupRootLB")
	updateCallback(edgeproto.UpdateTask, "Setting up RootLB")
	err = o.commonPf.SetupRootLB(ctx, o.commonPf.SharedRootLBName, vmspec, platformConfig.CloudletKey, platformConfig.CloudletVMImagePath, platformConfig.VMImageVersion, edgeproto.DummyUpdateCallback)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "ok, SetupRootLB")

	// set up L7 load balancer
	client, err := o.commonPf.GetPlatformClientRootLB(ctx, o.commonPf.SharedRootLBName)
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
	rootLBName := o.commonPf.SharedRootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBName = cloudcommon.GetDedicatedLBFQDN(o.commonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey)
	}
	return o.commonPf.GetPlatformClientRootLB(ctx, rootLBName)
}
