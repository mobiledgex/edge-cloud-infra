package main

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/aws"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/azure"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/edgebox"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/fakeinfra"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/gcp"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/openstack"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vmpool"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vsphere"
	"github.com/mobiledgex/edge-cloud-infra/managedk8s"
	"github.com/mobiledgex/edge-cloud-infra/plugin/common"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
)

func GetPlatform(plat string) (platform.Platform, error) {
	var outPlatform platform.Platform
	switch plat {
	case "PLATFORM_TYPE_OPENSTACK":
		openstackProvider := openstack.OpenstackPlatform{}
		outPlatform = &vmlayer.VMPlatform{
			Type:       vmlayer.VMProviderOpenstack,
			VMProvider: &openstackProvider,
		}
	case "PLATFORM_TYPE_VSPHERE":
		vsphereProvider := vsphere.VSpherePlatform{}
		outPlatform = &vmlayer.VMPlatform{
			Type:       vmlayer.VMProviderVSphere,
			VMProvider: &vsphereProvider,
		}
	case "PLATFORM_TYPE_VM_POOL":
		vmpoolProvider := vmpool.VMPoolPlatform{}
		outPlatform = &vmlayer.VMPlatform{
			Type:       vmlayer.VMProviderVMPool,
			VMProvider: &vmpoolProvider,
		}
	case "PLATFORM_TYPE_AZURE":
		azureProvider := &azure.AzurePlatform{}
		outPlatform = &managedk8s.ManagedK8sPlatform{
			Type:     managedk8s.ManagedK8sProviderAzure,
			Provider: azureProvider,
		}
	case "PLATFORM_TYPE_GCP":
		gcpProvider := &gcp.GCPPlatform{}
		outPlatform = &managedk8s.ManagedK8sPlatform{
			Type:     managedk8s.ManagedK8sProviderGCP,
			Provider: gcpProvider,
		}
	case "PLATFORM_TYPE_AWS":
		awsProvider := &aws.AWSPlatform{}
		outPlatform = &managedk8s.ManagedK8sPlatform{
			Type:     managedk8s.ManagedK8sProviderAWS,
			Provider: awsProvider,
		}
	case "PLATFORM_TYPE_EDGEBOX":
		outPlatform = &edgebox.EdgeboxPlatform{}
	case "PLATFORM_TYPE_FAKEINFRA":
		outPlatform = &fakeinfra.Platform{}
	default:
		return nil, fmt.Errorf("unknown platform %s", plat)
	}
	return outPlatform, nil
}

func GetClusterSvc() (platform.ClusterSvc, error) {
	return &common.ClusterSvc{}, nil
}

func main() {}
