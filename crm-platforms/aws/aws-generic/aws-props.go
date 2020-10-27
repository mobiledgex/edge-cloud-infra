package awsgeneric

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const AwsDefaultVaultPath string = "/secret/data/cloudlet/aws/credentials"

var AWSProps = map[string]*edgeproto.PropertyInfo{
	"AWS_ACCESS_KEY_ID": {
		Name:        "AWS Access Key ID",
		Description: "AWS Access Key ID",
		Secret:      true,
		Mandatory:   true,
	},
	"AWS_SECRET_ACCESS_KEY": {
		Name:        "AWS Secret Access Key",
		Description: "AWS Secret Access Key",
		Secret:      true,
		Mandatory:   true,
	},

	"AWS_REGION": {
		Name:        "AWS Region",
		Description: "AWS Region",
		Mandatory:   true,
	},
	// override default for router
	"MEX_ROUTER": {
		Name:        "External Router Type",
		Description: "AWS Router must be " + vmlayer.NoConfigExternalRouter,
		Value:       vmlayer.NoConfigExternalRouter,
	},
	"AWS_OUTPOST_VPC": {
		Name:        "AWS Outpost VPC",
		Description: "Pre-existing VPC for an outpost deployment",
	},
}

func (a *AwsGenericPlatform) GetAwsAccessKeyId() string {
	val, _ := a.Properties.GetValue("AWS_ACCESS_KEY_ID")
	return val
}

func (a *AwsGenericPlatform) GetAwsSecretAccessKey() string {
	val, _ := a.Properties.GetValue("AWS_SECRET_ACCESS_KEY")
	return val
}

func (a *AwsGenericPlatform) GetAwsRegion() string {
	val, _ := a.Properties.GetValue("AWS_REGION")
	return val
}

func (a *AwsGenericPlatform) IsAwsOutpost() bool {
	val, _ := a.Properties.GetValue("AWS_OUTPOST_VPC")
	return val != ""
}

func (a *AwsGenericPlatform) GetOutpostVPC() string {
	val, _ := a.Properties.GetValue("AWS_OUTPOST_VPC")
	return val
}

func (a *AwsGenericPlatform) GetProviderSpecificProps(ctx context.Context, pfconfig *pf.PlatformConfig, vaultConfig *vault.Config) (map[string]*edgeproto.PropertyInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetProviderSpecificProps")
	vaultPath := AwsDefaultVaultPath
	if pfconfig.CloudletKey.Organization != "aws" {
		// this is not a public cloud aws cloudlet, use the operator specific path
		vaultPath = fmt.Sprintf("/secret/data/%s/cloudlet/%s/%s/%s/%s", pfconfig.Region, "aws", pfconfig.CloudletKey.Organization, pfconfig.PhysicalName, "credentials")
	}
	err := infracommon.InternVaultEnv(ctx, vaultConfig, vaultPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to intern vault data", "err", err)
		err = fmt.Errorf("cannot intern vault data from vault %s", err.Error())
		return nil, err
	}
	return AWSProps, nil
}
