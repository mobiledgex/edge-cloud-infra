package gcp

import (
	"context"
	"fmt"

	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/accessapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
)

const gcpVaultPath string = "/secret/data/cloudlet/gcp/credentials"
const gcpAuthKeyPath string = "/secret/data/cloudlet/gcp/auth_key.json"
const gcpAuthKeyName = "auth_key.json"

var gcpProps = map[string]*edgeproto.PropertyInfo{
	"MEX_GCP_PROJECT": {
		Name:        "GCP Project Name",
		Description: "Name of the GCP project",
		Value:       "still-entity-201400",
	},
	"MEX_GCP_ZONE": {
		Name:        "GCP Zone Name",
		Description: "Name of the GCP zone, e.g. us-central1-a",
		Mandatory:   true,
	},
}

func (g *GCPPlatform) GetGcpRegion() string {
	// region is the zone without part after the last hyphen
	zone := g.GetGcpZone()
	zs := strings.Split(zone, "-")
	zs = zs[:len(zs)-1]
	return strings.Join(zs, "-")
}

func (g *GCPPlatform) GetGcpZone() string {
	val, _ := g.properties.GetValue("MEX_GCP_ZONE")
	return val
}

func (g *GCPPlatform) GetGcpProject() string {
	val, _ := g.properties.GetValue("MEX_GCP_PROJECT")
	return val
}

func (a *GCPPlatform) GetProviderSpecificProps(ctx context.Context) (map[string]*edgeproto.PropertyInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetProviderSpecificProps")
	return gcpProps, nil
}

func (m *GCPPlatform) GetAccessData(ctx context.Context, cloudlet *edgeproto.Cloudlet, region string, vaultConfig *vault.Config, dataType string, arg []byte) (map[string]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GCPPlatform GetAccessData", "dataType", dataType)
	switch dataType {
	case accessapi.GetCloudletAccessVars:
		vars, err := infracommon.GetEnvVarsFromVault(ctx, vaultConfig, gcpVaultPath)
		if err != nil {
			return nil, err
		}
		authKeyJSON, err := infracommon.GetVaultDataString(ctx, vaultConfig, gcpAuthKeyPath)
		if err != nil {
			return nil, err
		}
		vars[gcpAuthKeyName] = string(authKeyJSON)
		return vars, nil
	}
	return nil, fmt.Errorf("GCP unhandled GetAccessData type %s", dataType)
}

func (m *GCPPlatform) InitApiAccessProperties(ctx context.Context, accessApi platform.AccessApi, vars map[string]string) error {
	accessVars, err := accessApi.GetCloudletAccessVars(ctx)
	if err != nil {
		return err
	}
	if authKeyJSON, ok := accessVars[gcpAuthKeyName]; ok {
		m.authKeyJSON = authKeyJSON
		delete(accessVars, gcpAuthKeyName)
	}
	m.accessVars = accessVars
	return nil
}
