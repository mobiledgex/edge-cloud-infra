package aws

import (
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
)

var awsProps = map[string]*infracommon.PropertyInfo{
	"MEX_AWS_PROJECT": &infracommon.PropertyInfo{
		Value: "awspoc",
	},
	"MEX_AWS_ZONE":            &infracommon.PropertyInfo{},
	"MEX_AWS_SERVICE_ACCOUNT": &infracommon.PropertyInfo{},
	"MEX_AWS_AUTH_KEY_PATH": &infracommon.PropertyInfo{
		Value: "/secret/data/cloudlet/aws/auth_key.json",
	},
}

func (a *AWSPlatform) GetGcpAuthKeyUrl() string {
	if val, ok := a.commonPf.Properties["MEX_AWS_AUTH_KEY_PATH"]; ok {
		return val.Value
	}
	return ""
}

func (a *AWSPlatform) GetGcpZone() string {
	if val, ok := a.commonPf.Properties["MEX_AWS_ZONE"]; ok {
		return val.Value
	}
	return ""
}

func (a *AWSPlatform) GetGcpProject() string {
	if val, ok := a.commonPf.Properties["MEX_AWS_PROJECT"]; ok {
		return val.Value
	}
	return ""
}
