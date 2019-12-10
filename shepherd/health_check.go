package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
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

func getAlertFromAppInst(appInstKey *edgeproto.AppInstKey) *edgeproto.Alert {
	alert := edgeproto.Alert{}
	alert.Labels["alertname"] = cloudcommon.AlertAppInstDown
	alert.Labels[cloudcommon.AlertLabelDev] = appInstKey.AppKey.DeveloperKey.Name
	alert.Labels[cloudcommon.AlertLabelOperator] = appInstKey.ClusterInstKey.CloudletKey.OperatorKey.Name
	alert.Labels[cloudcommon.AlertLabelCloudlet] = appInstKey.ClusterInstKey.CloudletKey.Name
	alert.Labels[cloudcommon.AlertLabelCluster] = appInstKey.ClusterInstKey.ClusterKey.Name
	alert.Labels[cloudcommon.AlertLabelApp] = appInstKey.AppKey.Name
	alert.Labels[cloudcommon.AlertLabelAppVer] = appInstKey.AppKey.Version
	return &alert
}

func HealthCheckDown(ctx context.Context, appInstKey *edgeproto.AppInstKey) {
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
	// Create and send the Alert
	alert := getAlertFromAppInst(appInstKey)
	AlertCache.UpdateModFunc(ctx, alert.GetKey(), 0, func(old *edgeproto.Alert) (*edgeproto.Alert, bool) {
		if old == nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Update alert", "alert", alert)
			return alert, true
		}
		// don't update if nothing changed
		changed := !alert.Matches(old)
		if changed {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Update alert", "alert", alert)
		}
		return alert, changed
	})
}

func HealthCheckUp(ctx context.Context, appInstKey *edgeproto.AppInstKey) {
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
	// Delete the alert if we can find it
	alert := getAlertFromAppInst(appInstKey)
	AlertCache.Delete(ctx, alert, 0)
}
