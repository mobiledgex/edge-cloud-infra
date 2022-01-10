package vmlayer

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// mapStateForSwitchover checks the current state and gives the new state to transition to along. Returns state, generateError, needsCleanup
func (v *VMPlatform) mapStateForSwitchover(ctx context.Context, state edgeproto.TrackedState) (edgeproto.TrackedState, bool, bool) {
	errorState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
	generateError := false
	needsCleanup := false

	switch state {
	case edgeproto.TrackedState_READY:
		return errorState, generateError, needsCleanup
	case edgeproto.TrackedState_CREATE_REQUESTED:
		errorState = edgeproto.TrackedState_CREATE_ERROR
		generateError = true
	case edgeproto.TrackedState_CREATING:
		errorState = edgeproto.TrackedState_CREATE_ERROR
		generateError = true
		needsCleanup = true
	case edgeproto.TrackedState_UPDATE_REQUESTED:
		errorState = edgeproto.TrackedState_UPDATE_ERROR
		generateError = true
	case edgeproto.TrackedState_UPDATING:
		errorState = edgeproto.TrackedState_UPDATE_ERROR
		generateError = true
		needsCleanup = true
	case edgeproto.TrackedState_DELETE_REQUESTED:
		errorState = edgeproto.TrackedState_DELETE_ERROR
		generateError = true
	case edgeproto.TrackedState_DELETING:
		errorState = edgeproto.TrackedState_DELETE_ERROR
		generateError = true
		needsCleanup = true
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "transientStateToErrorState returns", "state", state, "errorState", errorState, "generateError", generateError, "needsCleanup", needsCleanup)
	return errorState, generateError, needsCleanup
}

func (v *VMPlatform) handleTransientClusterInsts(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "handleTransientClusterInsts")

	// fail all pending activities
	clusterInstKeys := []edgeproto.ClusterInstKey{}
	clusterInstsToCleanup := make(map[edgeproto.ClusterInstKey]edgeproto.TrackedState)

	v.Caches.ClusterInstCache.GetAllKeys(ctx, func(k *edgeproto.ClusterInstKey, modRev int64) {
		clusterInstKeys = append(clusterInstKeys, *k)
	})
	for _, k := range clusterInstKeys {
		var clusterInst edgeproto.ClusterInst
		if v.Caches.ClusterInstCache.Get(&k, &clusterInst) {
			errorState, generateError, needsCleanup := v.mapStateForSwitchover(ctx, clusterInst.State)
			immediateErrorState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
			if generateError {
				if needsCleanup {
					// cleanup and then error
					clusterInstsToCleanup[k] = errorState
				} else {
					// send an error right away
					v.Caches.ClusterInstInfoCache.SetError(ctx, &k, immediateErrorState, "CRM switched over while Cluster Instance in transient state")
				}
			}
		}
	}
	ctx, _, err := v.VMProvider.InitOperationContext(ctx, OperationInitStart)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to init context for cleanup", "err", err)
		return
	}
	for k, e := range clusterInstsToCleanup {
		var clusterInst edgeproto.ClusterInst
		if v.Caches.ClusterInstCache.Get(&k, &clusterInst) {
			log.SpanLog(ctx, log.DebugLevelInfra, "cleaning up cluster", "key", k)
			lbName := v.VMProperties.GetRootLBNameForCluster(ctx, &clusterInst)
			err := v.cleanupClusterInst(ctx, lbName, &clusterInst, edgeproto.DummyUpdateCallback)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "error cleaning up cluster", "key", k, "error", err)
				v.Caches.ClusterInstInfoCache.SetError(ctx, &k, e, "CRM switched over while Cluster Instance in transient state, cluster cleanup failed")
			} else {
				v.Caches.ClusterInstInfoCache.SetError(ctx, &k, e, "CRM switched over while Cluster Instance in transient state")
			}
		}
	}

}

func (v *VMPlatform) handleTransientAppInsts(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "handleTransientAppInsts")

	// fail all pending activities
	appInstKeys := []edgeproto.AppInstKey{}
	appInstsToCleanup := make(map[edgeproto.AppInstKey]edgeproto.TrackedState)

	v.Caches.AppInstCache.GetAllKeys(ctx, func(k *edgeproto.AppInstKey, modRev int64) {
		appInstKeys = append(appInstKeys, *k)
	})
	for _, k := range appInstKeys {
		var appInst edgeproto.AppInst
		if v.Caches.AppInstCache.Get(&k, &appInst) {
			errorState, generateError, needsCleanup := v.mapStateForSwitchover(ctx, appInst.State)
			immediateErrorState := edgeproto.TrackedState_TRACKED_STATE_UNKNOWN
			if generateError {
				if needsCleanup {
					// cleanup and then error
					appInstsToCleanup[k] = errorState
				} else {
					// send an error right away
					v.Caches.AppInstInfoCache.SetError(ctx, &k, immediateErrorState, "CRM switched over while App Instance in transient state")
				}
			}
		}
	}
	ctx, _, err := v.VMProvider.InitOperationContext(ctx, OperationInitStart)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to init context for cleanup", "err", err)
		return
	}
	for k, e := range appInstsToCleanup {
		app := edgeproto.App{}
		if !v.Caches.AppCache.Get(&k.AppKey, &app) {
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to find app in cache", "appkey", k.AppKey)
			v.Caches.AppInstInfoCache.SetError(ctx, &k, e, "CRM switched over while App Instance in transient state, unable to cleanup")
			continue
		}
		clusterInst := edgeproto.ClusterInst{}
		if cloudcommon.IsClusterInstReqd(&app) {
			clusterInstFound := v.Caches.ClusterInstCache.Get((*edgeproto.ClusterInstKey)(&k.ClusterInstKey), &clusterInst)
			if !clusterInstFound {
				log.SpanLog(ctx, log.DebugLevelInfra, "failed to find clusterinst in cache", "clusterkey", k.ClusterInstKey)
				v.Caches.AppInstInfoCache.SetError(ctx, &k, e, "CRM switched over while App Instance in transient state, unable to cleanup")
			}
		}
		var appInst edgeproto.AppInst
		if v.Caches.AppInstCache.Get(&k, &appInst) {
			log.SpanLog(ctx, log.DebugLevelInfra, "cleaning up appinst", "key", k)
			err := v.cleanupAppInst(ctx, &clusterInst, &app, &appInst, edgeproto.DummyUpdateCallback)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "error cleaning up appinst", "key", k, "error", err)
				v.Caches.AppInstInfoCache.SetError(ctx, &k, e, "CRM switched over while App Instance in transient state, cleanup failed")
			} else {
				v.Caches.AppInstInfoCache.SetError(ctx, &k, e, "CRM switched over while App Instance in transient state")
			}
		}
	}

}

func (v *VMPlatform) ActiveChanged(ctx context.Context, platformActive bool) {
	log.SpanLog(ctx, log.DebugLevelInfra, "ActiveChanged", "platformActive", platformActive)
	if !platformActive {
		return
	}
	var cloudletInternal edgeproto.CloudletInternal
	if !v.Caches.CloudletInternalCache.Get(v.VMProperties.CommonPf.PlatformConfig.CloudletKey, &cloudletInternal) {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error: unable to find cloudlet key in cache", "platformActive", platformActive)
	} else {
		// inform Shepherd via the internal cache of the new active state
		log.SpanLog(ctx, log.DebugLevelInfra, "Updating cloudlet internal cache", "platformActive", platformActive)
		cloudletInternal.Props[CloudletPlatformActive] = fmt.Sprintf("%t", platformActive)
		v.Caches.CloudletInternalCache.Update(ctx, &cloudletInternal, 0)
	}
	// cleanups need to happen in background as ActiveChanged is run via the HA Manager thread and cannot take too much time
	go v.handleTransientClusterInsts(ctx)
	go v.handleTransientAppInsts(ctx)

}
