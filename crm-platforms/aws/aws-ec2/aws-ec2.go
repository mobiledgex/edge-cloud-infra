package awsec2

import (
	"context"
	"fmt"

	awsgen "github.com/mobiledgex/edge-cloud-infra/crm-platforms/aws/aws-generic"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

type AwsEc2Platform struct {
	awsGenPf        *awsgen.AwsGenericPlatform
	VMProperties    *vmlayer.VMProperties
	BaseImageId     string
	AmiIamAccountId string
	caches          *platform.Caches
	VpcCidr         string
}

func (a *AwsEc2Platform) NameSanitize(name string) string {
	return name
}

// AwsEc2Platform IdSanitize is the same as NameSanitize
func (a *AwsEc2Platform) IdSanitize(name string) string {
	return a.NameSanitize(name)
}

func (a *AwsEc2Platform) SetVMProperties(vmProperties *vmlayer.VMProperties) {
	vmProperties.UseSecgrpForInternalSubnet = true
	vmProperties.RequiresWhitelistOwnIp = true
	a.VMProperties = vmProperties
}

func (a *AwsEc2Platform) GetInternalPortPolicy() vmlayer.InternalPortAttachPolicy {
	return vmlayer.AttachPortAfterCreate
}

func (a *AwsEc2Platform) GetType() string {
	return "awsec2"
}

func (a *AwsEc2Platform) GetProviderSpecificProps(ctx context.Context, pfconfig *platform.PlatformConfig, vaultConfig *vault.Config) (map[string]*edgeproto.PropertyInfo, error) {
	return a.awsGenPf.GetProviderSpecificProps(ctx, pfconfig, vaultConfig)
}

func (a *AwsEc2Platform) InitApiAccessProperties(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config, vars map[string]string) error {
	return a.awsGenPf.GetAwsVaultAccessVars(ctx, key, region, physicalName, vaultConfig)
}

func (a *AwsEc2Platform) InitData(ctx context.Context, caches *platform.Caches) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitData", "AwsEc2Platform", fmt.Sprintf("%+v", a))
	a.caches = caches
	a.awsGenPf = &awsgen.AwsGenericPlatform{Properties: &a.VMProperties.CommonPf.Properties}
}
