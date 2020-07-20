package generic

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (o *GenericPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SaveCloudletAccessVars not supported")
	return nil
}

func (o *GenericPlatform) ImportDataFromInfra(ctx context.Context, domain vmlayer.VMDomain) error {
	// nothing to do
	return nil
}

func (o *GenericPlatform) GetApiEndpointAddr(ctx context.Context) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetApiEndpointAddr not supported")
	return "", nil
}

func (o *GenericPlatform) GetCloudletManifest(ctx context.Context, name string, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletManifest not supported")
	return "", nil
}
