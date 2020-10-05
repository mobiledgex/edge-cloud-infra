package main

import (
	"context"

	edgeevents "github.com/mobiledgex/edge-cloud-infra/edge-events"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/log"
)

func GetEdgeEventsHandler(ctx context.Context) (dmecommon.EdgeEventsHandler, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetEdgeEventHandler")

	return &edgeevents.EdgeEventsHandlerPlugin{}, nil
}

func main() {}
