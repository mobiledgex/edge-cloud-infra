package aws

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func (o *AWSPlatform) GetConsoleUrl(ctx context.Context, serverName string) (string, error) {
	return "", fmt.Errorf("GetConsoleUrl not implemented")
}

func (o *AWSPlatform) AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, flavor string, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("AddAppImageIfNotPresent not implemented")
}
