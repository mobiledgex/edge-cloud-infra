package main

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/azure"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/gcp"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/mexdind"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/openstack"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
)

func GetPlatform(ctx context.Context, plat string) (platform.Platform, error) {
	var outPlatform platform.Platform
	switch plat {
	case "PLATFORM_TYPE_OPENSTACK":
		outPlatform = &openstack.Platform{}
	case "PLATFORM_TYPE_AZURE":
		outPlatform = &azure.Platform{}
	case "PLATFORM_TYPE_GCP":
		outPlatform = &gcp.Platform{}
	case "PLATFORM_TYPE_MEXDIND":
		outPlatform = &mexdind.Platform{}
	default:
		return nil, fmt.Errorf("unknown platform %s", plat)
	}
	outPlatform.SetContext(ctx)
	return outPlatform, nil
}

func main() {}
