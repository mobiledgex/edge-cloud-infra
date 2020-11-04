package awsgeneric

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const AwsDefaultVaultPath string = "/secret/data/cloudlet/aws/credentials"
const ArnAccountIdIdx = 4

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
	"AWS_AMI_IAM_OWNER": {
		Name:        "AWS Outpost AMI Owner",
		Description: "IAM Account that owns the base image",
	},
	"AWS_OUTPOST_FLAVORS": {
		Name:        "AWS Outpost Flavors",
		Description: "AWS Outpost Flavors in format flavor1,vcpu,ram,disk;flavor2.. e.g. c5.large,2,4096,40;c5.xlarge,4,8192,40",
	},
	"AWS_USER_ARN": {
		Name:        "AWS User ARN (Amazon Resource Name)",
		Description: "AWS User ARN (Amazon Resource Name)",
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

func (a *AwsGenericPlatform) GetAwsAmiIamOwner() string {
	val, _ := a.Properties.GetValue("AWS_AMI_IAM_OWNER")
	return val
}

func (a *AwsGenericPlatform) GetAwsOutpostVPC() string {
	val, _ := a.Properties.GetValue("AWS_OUTPOST_VPC")
	return val
}

func (a *AwsGenericPlatform) GetAwsOutpostFlavors() string {
	val, _ := a.Properties.GetValue("AWS_OUTPOST_FLAVORS")
	return val
}
func (a *AwsGenericPlatform) GetAwsUserArn() string {
	val, _ := a.Properties.GetValue("AWS_USER_ARN")
	return val
}

func (a *AwsGenericPlatform) GetProviderSpecificProps(ctx context.Context, pfconfig *pf.PlatformConfig, vaultConfig *vault.Config) (map[string]*edgeproto.PropertyInfo, error) {
	vaultPath := AwsDefaultVaultPath
	if pfconfig.CloudletKey.Organization != "aws" {
		// this is not a public cloud aws cloudlet, use the operator specific path
		vaultPath = fmt.Sprintf("/secret/data/%s/cloudlet/%s/%s/%s/%s", pfconfig.Region, "aws", pfconfig.CloudletKey.Organization, pfconfig.PhysicalName, "credentials")
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "GetProviderSpecificProps", "vaultPath", vaultPath)

	// clear existing credentials
	// clear credentials
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")

	err := infracommon.InternVaultEnv(ctx, vaultConfig, vaultPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to intern vault data", "err", err)
		err = fmt.Errorf("cannot intern vault data from vault %s", err.Error())
		return nil, err
	}
	return AWSProps, nil
}

func (a *AwsGenericPlatform) GetUserAccountIdFromArn(ctx context.Context, arn string) (string, error) {
	arns := strings.Split(arn, ":")
	if len(arns) <= ArnAccountIdIdx {
		log.SpanLog(ctx, log.DebugLevelInfra, "Wrong number of fields in ARN", "iamResult.User.Arn", arn)
		return "", fmt.Errorf("Cannot parse IAM ARN: %s", arn)
	}
	return arns[ArnAccountIdIdx], nil
}
