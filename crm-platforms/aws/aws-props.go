package aws

import (
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
)

var AWSProps = map[string]*infracommon.PropertyInfo{
	"MEX_AWS_PROJECT": &infracommon.PropertyInfo{
		Value: "awspoc",
	},
	"MEX_AWS_REGION": &infracommon.PropertyInfo{
		Value: "us-west-2",
	},
	"MEX_AWS_ZONE": &infracommon.PropertyInfo{
		Value: "us-west-2a",
	},
	"MEX_AWS_MASTER_ARN": &infracommon.PropertyInfo{
		Value: "arn:aws:kms:us-west-2:177018110765:key/d90efe47-8cd6-407a-8bab-0861d8cd5c0a",
	},
	// "MEX_AWS_AUTH_KEY_PATH": &infracommon.PropertyInfo{
	// 	Value: "/secret/data/cloudlet/aws/auth_key.json",
	// },
}

func (a *AWSPlatform) GetAwsAuthKeyUrl() string {
	if val, ok := a.commonPf.Properties["MEX_AWS_AUTH_KEY_PATH"]; ok {
		return val.Value
	}
	return ""
}

//  Get AWS Zone
func (a *AWSPlatform) GetAWSZone() string {
	if val, ok := a.commonPf.Properties["MEX_AWS_ZONE"]; ok {
		return val.Value
	}
	return ""
}

func (a *AWSPlatform) GetAWSProject() string {
	if val, ok := a.commonPf.Properties["MEX_AWS_PROJECT"]; ok {
		return val.Value
	}
	return ""
}
