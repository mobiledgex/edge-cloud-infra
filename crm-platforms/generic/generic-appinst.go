package generic

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (o *GenericPlatform) GetConsoleUrl(ctx context.Context, serverName string) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GenericPlatform not supported")
	return "", nil
}

func (o *GenericPlatform) AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, flavor string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AddAppImageIfNotPresent not supported")
	return nil
}
