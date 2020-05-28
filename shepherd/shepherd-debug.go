package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var WorkerDebugLevel = log.DebugLevelSampled

func InitDebug(nodeMgr *node.NodeMgr) {
	nodeMgr.Debug.AddDebugFunc("disable-sample-logging", disableSampledLogging)
	nodeMgr.Debug.AddDebugFunc("enable-sample-logging", enableSampledLogging)
}

func disableSampledLogging(ctx context.Context, req *edgeproto.DebugRequest) string {
	WorkerDebugLevel = log.DebugLevelInfo | log.DebugLevelMetrics
	return "set worker debug level to info|metrics"

}

func enableSampledLogging(ctx context.Context, req *edgeproto.DebugRequest) string {
	WorkerDebugLevel = log.DebugLevelSampled
	return "set worker debug level to sampled"
}
