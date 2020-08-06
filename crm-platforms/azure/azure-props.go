package azure

import (
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
)

var azureProps = map[string]*infracommon.PropertyInfo{
	"MEX_AZURE_LOCATION": {
		Mandatory: true,
	},
	"MEX_AZURE_USER": {
		Mandatory: true,
	},
	"MEX_AZURE_PASS": {
		Secret:    true,
		Mandatory: true,
	},
}

func (a *AzurePlatform) GetK8sProviderSpecificProps() map[string]*infracommon.PropertyInfo {
	return azureProps
}

func (a *AzurePlatform) GetAzureLocation() string {
	if val, ok := a.ManagedK8sProperties.CommonPf.Properties["MEX_AZURE_LOCATION"]; ok {
		return val.Value
	}
	return ""
}

func (a *AzurePlatform) GetAzureUser() string {
	if val, ok := a.ManagedK8sProperties.CommonPf.Properties["MEX_AZURE_USER"]; ok {
		return val.Value
	}
	return ""
}

func (a *AzurePlatform) GetAzurePass() string {
	if val, ok := a.ManagedK8sProperties.CommonPf.Properties["MEX_AZURE_PASS"]; ok {
		return val.Value
	}
	return ""
}
