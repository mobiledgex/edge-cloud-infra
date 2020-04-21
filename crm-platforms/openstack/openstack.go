package openstack

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

type OpenstackPlatform struct {
	openRCVars map[string]string
	vmPlatform vmlayer.VMPlatform
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
	if err := o.InitOpenstackProps(ctx, platformConfig.CloudletKey, platformConfig.Region, platformConfig.PhysicalName, vaultConfig, platformConfig.EnvVars); err != nil {
		return err
	}
	o.vmPlatform.FlavorList, _, _, err = o.GetFlavorInfo(ctx)
	if err != nil {
		return err
	}
	err = o.PrepNetwork(ctx)
	if err != nil {
		return err
	}
	if err := o.vmPlatform.CommonPf.InitInfraCommon(ctx, platformConfig, vmlayer.VMProviderProps, vaultConfig, o); err != nil {
		return err
	}
	return o.vmPlatform.InitVMProvider(ctx, o, updateCallback)

	// create rootLB
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
	return o.vmPlatform.GetSSHClientForCluster(ctx, clusterInst)
}

func (o *OpenstackPlatform) DeleteResources(ctx context.Context, resourceGroupName string) error {
	return o.HeatDeleteStack(ctx, resourceGroupName)
}

func (o *OpenstackPlatform) GetResourceID(ctx context.Context, resourceType vmlayer.ResourceType, resourceName string) (string, error) {
	switch resourceType {
	case vmlayer.ResourceTypeSecurityGroup:
		return o.GetSecurityGroupIDForName(ctx, resourceName)
		// TODO other types as needed
	}
	return "", fmt.Errorf("GetResourceID not implemented for resource type: %s ", resourceType)
}

func (o *OpenstackPlatform) Resync(ctx context.Context) error {
	return fmt.Errorf("Resync not yet implemented")
}
