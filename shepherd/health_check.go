package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	opentracing "github.com/opentracing/opentracing-go"
)

func setupHealthCheckSpan(appInstKey *edgeproto.AppInstKey) (opentracing.Span, context.Context) {
	span := log.StartSpan(log.DebugLevelInfo, "health-check")
	span.SetTag("app", appInstKey.AppKey.Name)
	span.SetTag("cloudlet", appInstKey.ClusterInstKey.CloudletKey.Name)
	span.SetTag("operator", appInstKey.ClusterInstKey.CloudletKey.OperatorKey.Name)
	span.SetTag("cluster", appInstKey.ClusterInstKey.ClusterKey.Name)
	ctx := log.ContextWithSpan(context.Background(), span)
	return span, ctx
}

func HealthCheckDown(appInstKey *edgeproto.AppInstKey) {
	span, ctx := setupHealthCheckSpan(appInstKey)
	defer span.Finish()

	appInst := edgeproto.AppInst{}
	found := AppInstCache.Get(&appInst.Key, &appInst)
	if !found {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find appInst ", "appInst", appInst.Key)
		return
	}
	if appInst.State != edgeproto.TrackedState_READY {
		return
	}
	// TODO - update throguht notify framework(Alerts) that it should be disabled
}

func HealthCheckUp(appInstKey *edgeproto.AppInstKey) {
	span, ctx := setupHealthCheckSpan(appInstKey)
	defer span.Finish()

	appInst := edgeproto.AppInst{}
	found := AppInstCache.Get(&appInst.Key, &appInst)
	if !found {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find appInst ", "appInst", appInst.Key)
		return
	}
	if appInst.State == edgeproto.TrackedState_READY {
		return
	}
	// TODO - update throguht notify framework(Alerts) that it should be enabled
}
