package main

import (
	"context"
	"time"

	edgeevents "github.com/mobiledgex/edge-cloud-infra/edge-events"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func GetEdgeEventsHandler(ctx context.Context, edgeEventsCookieExpiration time.Duration) (dmecommon.EdgeEventsHandler, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetEdgeEventHandler")

	edgeEventsHandlerPlugin := new(edgeevents.EdgeEventsHandlerPlugin)
	log.SpanLog(ctx, log.DebugLevelInfra, "initializing app insts struct")
	appInstsStruct := new(edgeevents.AppInsts)
	log.SpanLog(ctx, log.DebugLevelInfra, "initializing app insts map")
	appInstsStruct.AppInstsMap = make(map[edgeproto.AppInstKey]*edgeevents.Clients)
	edgeEventsHandlerPlugin.AppInstsStruct = appInstsStruct
	edgeEventsHandlerPlugin.EdgeEventsCookieExpiration = edgeEventsCookieExpiration
	return edgeEventsHandlerPlugin, nil
}

func main() {}
