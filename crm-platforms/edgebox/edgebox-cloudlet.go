package edgebox

import (
	"context"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/fakeinfra"
	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (e *EdgeboxPlatform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, flavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "create cloudlet for edgebox")
	err := e.generic.CreateCloudlet(ctx, cloudlet, pfConfig, flavor, updateCallback)
	if err != nil {
		return err
	}
	return fakeinfra.ShepherdStartup(ctx, cloudlet, pfConfig, updateCallback)
}

func (e *EdgeboxPlatform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "delete cloudlet for edgebox")
	err := e.generic.DeleteCloudlet(ctx, cloudlet, pfConfig, updateCallback)
	if err != nil {
		return err
	}

	updateCallback(edgeproto.UpdateTask, "Stopping Shepherd")
	return intprocess.StopShepherdService(ctx, cloudlet)
}

func (e *EdgeboxPlatform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) (edgeproto.CloudletAction, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "Updating cloudlet", "cloudletName", cloudlet.Key.Name)
	cloudletAction, err := e.generic.UpdateCloudlet(ctx, cloudlet, pfConfig, updateCallback)
	return cloudletAction, err
}

func (e *EdgeboxPlatform) CleanupCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Cleaning up cloudlet", "cloudletName", cloudlet.Key.Name)
	err := e.generic.CleanupCloudlet(ctx, cloudlet, pfConfig, updateCallback)
	return err
}

func (e *EdgeboxPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Saving cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return nil
}

func (e *EdgeboxPlatform) DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Deleting cloudlet access vars", "cloudletName", cloudlet.Key.Name)
	return nil
}
