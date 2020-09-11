package aws

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const awsVaultPath string = "/secret/data/cloudlet/aws/credentials"

var AWSProps = map[string]*edgeproto.PropertyInfo{
	"AWS_ACCESS_KEY_ID": &edgeproto.PropertyInfo{
		Name:        "AWS Access Key ID",
		Description: "AWS Access Key ID",
		Secret:      true,
		Mandatory:   true,
	},
	"AWS_SECRET_ACCESS_KEY": &edgeproto.PropertyInfo{
		Name:        "AWS Secret Access Key",
		Description: "AWS Secret Access Key",
		Secret:      true,
		Mandatory:   true,
	},

	"AWS_REGION": &edgeproto.PropertyInfo{
		Name:        "AWS Region",
		Description: "AWS Region",
		Mandatory:   true,
	},
}

func (a *AWSPlatform) GetK8sProviderSpecificProps() map[string]*edgeproto.PropertyInfo {
	return AWSProps
}

func (a *AWSPlatform) GetAwsAccessKeyId() string {
	val, _ := a.VMProperties.CommonPf.Properties.GetValue("AWS_ACCESS_KEY_ID")
	return val
}

func (a *AWSPlatform) GetAwsSecretAccessKey() string {
	val, _ := a.VMProperties.CommonPf.Properties.GetValue("AWS_SECRET_ACCESS_KEY")
	return val
}

func (a *AWSPlatform) GetAwsRegion() string {
	val, _ := a.VMProperties.CommonPf.Properties.GetValue("AWS_REGION")
	return val
}

func (a *AWSPlatform) InitApiAccessProperties(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config, vars map[string]string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitApiAccessProperties")
	err := infracommon.InternVaultEnv(ctx, vaultConfig, awsVaultPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to intern vault data for API access", "err", err)
		err = fmt.Errorf("cannot intern vault data from vault %s", err.Error())
		return err
	}
	return nil
}

func (a *AWSPlatform) GetProviderSpecificProps() map[string]*edgeproto.PropertyInfo {
	// for now we use the same as the managed k8s props
	return a.GetK8sProviderSpecificProps()
}
