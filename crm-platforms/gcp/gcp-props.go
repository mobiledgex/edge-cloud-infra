package gcp

import (
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
)

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

func (a *GCPPlatform) GetK8sProviderSpecificProps() map[string]*infracommon.PropertyInfo {
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
