package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelNotify | log.DebugLevelApi | log.DebugLevelMetrics)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	retry := newRetryTracker()
	key := testutil.AppInstData[0].Key

	// no error should not register a retry
	retry.registerDeployResult(ctx, key, nil)
	require.Equal(t, 0, len(retry.allFailures))

	// already exists error should not register a retry
	retry.registerDeployResult(ctx, key, key.ExistsError())
	require.Equal(t, 0, len(retry.allFailures))

	// error should register a retry
	retry.registerDeployResult(ctx, key, fmt.Errorf("failure"))
	require.Equal(t, 1, len(retry.allFailures))

	// retryTracker should return failure
	failure := retry.hasFailure(ctx, key.AppKey, key.ClusterInstKey.CloudletKey)
	require.True(t, failure)

	cacheData.init()
	minmax := newMinMaxChecker(&cacheData)
	runCount := 0
	minmax.workers.Init("test-retry", func(ctx context.Context, k interface{}) {
		appkey, ok := k.(edgeproto.AppKey)
		require.True(t, ok)
		require.Equal(t, key.AppKey, appkey)
		runCount++
	})
	// do retry should queue recheck and clear failure
	retry.doRetry(ctx, minmax)
	require.Equal(t, 0, len(retry.allFailures))
	minmax.workers.WaitIdle()
	require.Equal(t, 1, runCount)
}
