package aws

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const awsVaultPath string = "/secret/data/cloudlet/aws/credentials"

var AWSProps = map[string]*infracommon.PropertyInfo{
	"AWS_ACCESS_KEY_ID": {
		Secret:    true,
		Mandatory: true,
	},
	"AWS_SECRET_ACCESS_KEY": {
		Secret:    true,
		Mandatory: true,
	},
	"AWS_REGION": {
		Mandatory: true,
	},
}

func (a *AWSPlatform) GetK8sProviderSpecificProps() map[string]*infracommon.PropertyInfo {
	return AWSProps
}

func (a *AWSPlatform) GetAwsAccessKeyId() string {
	if val, ok := a.commonPf.Properties["AWS_ACCESS_KEY_ID"]; ok {
		return val.Value
	}
	return ""
}

func (a *AWSPlatform) GetAwsSecretAccessKey() string {
	if val, ok := a.commonPf.Properties["AWS_SECRET_ACCESS_KEY"]; ok {
		return val.Value
	}
	return ""
}

func (a *AWSPlatform) GetAwsRegion() string {
	if val, ok := a.commonPf.Properties["AWS_REGION"]; ok {
		return val.Value
	}
	return ""
}

func (a *AWSPlatform) InitApiAccessProperties(ctx context.Context, region string, vaultConfig *vault.Config, vars map[string]string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitApiAccessProperties")
	err := infracommon.InternVaultEnv(ctx, vaultConfig, awsVaultPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to intern vault data for API access", "err", err)
		err = fmt.Errorf("cannot intern vault data from vault %s", err.Error())
		return err
	}
	return nil
}
