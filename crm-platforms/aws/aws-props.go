package aws

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var AWSProps = map[string]*edgeproto.PropertyInfo{
	"AWS_ACCESS_KEY_ID": &edgeproto.PropertyInfo{
		Name:        "Access Key ID",
		Description: "Access Key ID",
		Secret:      true,
		Mandatory:   true,
	},
	"AWS_SECRET_ACCESS_KEY": &edgeproto.PropertyInfo{
		Name:        "Secret Access Key",
		Description: "Secret Access Key",
		Secret:      true,
		Mandatory:   true,
	},

	"AWS_DEFAULT_REGION": &edgeproto.PropertyInfo{
		Name:        "Default Region",
		Description: "Default Region",
		Value:       "us-west-2",
		Mandatory:   true,
	},
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

func (a *AWSPlatform) GetAwsDefaultRegion() string {
	if val, ok := a.commonPf.Properties["AWS_DEFAULT_REGION"]; ok {
		return val.Value
	}
	return ""
}

func (a *AWSPlatform) GetCloudletProps(ctx context.Context) (*edgeproto.CloudletProps, error) {
	return &edgeproto.CloudletProps{Properties: AWSProps}, nil
}
