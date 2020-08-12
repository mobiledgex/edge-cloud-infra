package gcp

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const gcpVaultPath string = "/secret/data/cloudlet/gcp/credentials"

var gcpProps = map[string]*infracommon.PropertyInfo{
	"MEX_GCP_PROJECT": {
		Value: "still-entity-201400",
	},
	"MEX_GCP_ZONE": {
		Mandatory: true,
	},
	"MEX_GCP_SERVICE_ACCOUNT": {
		Mandatory: true,
		Secret:    true,
	},
	"MEX_GCP_AUTH_KEY_PATH": {
		Value: "/secret/data/cloudlet/gcp/auth_key.json",
	},
}

func (g *GCPPlatform) GetK8sProviderSpecificProps() map[string]*infracommon.PropertyInfo {
	return gcpProps
}

func (g *GCPPlatform) GetGcpAuthKeyUrl() string {
	if val, ok := g.commonPf.Properties["MEX_GCP_AUTH_KEY_PATH"]; ok {
		return val.Value
	}
	return ""
}

func (g *GCPPlatform) GetGcpZone() string {
	if val, ok := g.commonPf.Properties["MEX_GCP_ZONE"]; ok {
		return val.Value
	}
	return ""
}

func (g *GCPPlatform) GetGcpProject() string {
	if val, ok := g.commonPf.Properties["MEX_GCP_PROJECT"]; ok {
		return val.Value
	}
	return ""
}

func (g *GCPPlatform) InitApiAccessProperties(ctx context.Context, region string, vaultConfig *vault.Config, vars map[string]string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitApiAccessProperties")
	err := infracommon.InternVaultEnv(ctx, vaultConfig, gcpVaultPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Failed to intern vault data for API access", "err", err)
		err = fmt.Errorf("cannot intern vault data from vault %s", err.Error())
		return err
	}
	return nil
}
