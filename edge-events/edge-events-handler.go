package edgeevents

import (
	"context"
	"math"
	"net"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

type EdgeEventsHandlerPlugin struct {
	mux util.Mutex
	// Hashmap containing AppInsts on DME mapped to the clients connected to those AppInsts
	AppInstsStruct AppInsts
}

type AppInsts struct {
	AppInstsMap map[edgeproto.AppInstKey]Clients
}

type Clients struct {
	ClientsMap map[net.Addr]*dmecommon.EdgeEventPersistentMgr
}

// Add Client connected to specified AppInst to Map
func (e *EdgeEventsHandlerPlugin) AddClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, addr net.Addr, mgr *dmecommon.EdgeEventPersistentMgr) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.AppInstsStruct.AppInstsMap == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "initializing app insts map")
		e.AppInstsStruct.AppInstsMap = make(map[edgeproto.AppInstKey]Clients)
	}

	clients, ok := e.AppInstsStruct.AppInstsMap[appInstKey]
	if !ok {
		// add first client for appinst
		newClients := new(Clients)
		newClients.ClientsMap = make(map[net.Addr]*dmecommon.EdgeEventPersistentMgr)
		clients = *newClients
	}

	clients.ClientsMap[addr] = mgr
	e.AppInstsStruct.AppInstsMap[appInstKey] = clients
}

// Remove Client connected to specified AppInst from Map
func (e *EdgeEventsHandlerPlugin) RemoveClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, addr net.Addr) {
	e.mux.Lock()
	defer e.mux.Unlock()
	clients, ok := e.AppInstsStruct.AppInstsMap[appInstKey]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find appinst to remove", "appInstKey", appInstKey)
		return
	}

	delete(clients.ClientsMap, addr)
}

// Remove AppInst from Map of AppInsts
func (e *EdgeEventsHandlerPlugin) RemoveAppInstKey(ctx context.Context, appInstKey edgeproto.AppInstKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	_, ok := e.AppInstsStruct.AppInstsMap[appInstKey]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find appinst to remove", "appInstKey", appInstKey)
		return
	}

	delete(e.AppInstsStruct.AppInstsMap, appInstKey)
}

// Send a ServerEdgeEvent with Latency Request Event to all clients connected to specified AppInst (and also have an initiated persistent connection)
// When client recieves this event, it will measure latency from itself to appinst and back.
// Client will then send those latency samples back to be processed in the HandleLatencySamples function
// Finally, DME will send the processed latency samples in the form of dme.Latency struct (with calculated avg, min, max, stddev) back to client
func (e *EdgeEventsHandlerPlugin) SendLatencyRequestEdgeEvent(ctx context.Context, appInst *dmecommon.DmeAppInst, appInstKey edgeproto.AppInstKey) {
	clients, ok := e.AppInstsStruct.AppInstsMap[appInstKey]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find appinst. no clients connected to appinst have edge events connection.", "appInstKey", appInstKey)
		return
	}

	for addr, mgr := range clients.ClientsMap {
		go func(addr net.Addr, mgr *dmecommon.EdgeEventPersistentMgr) {
			latencyRequestEdgeEvent := new(dme.ServerEdgeEvent)
			latencyRequestEdgeEvent.Event = dme.ServerEdgeEvent_EVENT_LATENCY_REQUEST
			log.SpanLog(ctx, log.DebugLevelInfra, "Sending latency request to client", "client addr", addr)
			e.SendEdgeEventToClient(ctx, latencyRequestEdgeEvent, mgr)
		}(addr, mgr)
	}
}

// Handle processing of latency samples and then send back to client
// For now: Avg, Min, Max, StdDev
func (e *EdgeEventsHandlerPlugin) ProcessLatencySamples(ctx context.Context, appInstKey edgeproto.AppInstKey, addr net.Addr, samples []float64) (*dme.Latency, bool) {
	clients, ok := e.AppInstsStruct.AppInstsMap[appInstKey]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find appinst, no clients connected to appinst have edge events connection", "appInstKey", appInstKey)
		return nil, false
	}

	mgr, ok := clients.ClientsMap[addr]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find client connected to appinst", "appInstKey", appInstKey, "client addr", addr)
		return nil, false
	}

	latencyEdgeEvent := new(dme.ServerEdgeEvent)
	latencyEdgeEvent.Event = dme.ServerEdgeEvent_EVENT_LATENCY_PROCESSED
	// Create latency struct
	latency := new(dme.Latency)
	// calculate Min, Max, and Avg
	numSamples := float64(len(samples))
	sum := 0.0
	min := 0.0
	max := 0.0
	for _, sample := range samples {
		sum += sample
		if min == 0 || sample < min {
			min = sample
		}
		if min == 0 || sample > max {
			max = sample
		}
	}
	avg := sum / numSamples
	// calculate StdDev
	diffSquared := 0.0
	for _, sample := range samples {
		diff := sample - avg
		diffSquared += diff * diff
	}
	stddev := math.Sqrt(diffSquared / numSamples)
	// Set latency fields
	latency.Avg = avg
	latency.Min = min
	latency.Max = max
	latency.StdDev = stddev
	latencyEdgeEvent.Latency = latency

	e.SendEdgeEventToClient(ctx, latencyEdgeEvent, mgr)
	return latency, true
}

// Send a ServerEdgeEvent with specified Event to all clients connected to specified AppInst (and also have initiated persistent connection)
func (e *EdgeEventsHandlerPlugin) SendAppInstStateEvent(ctx context.Context, appInst *dmecommon.DmeAppInst, appInstKey edgeproto.AppInstKey, eventType dme.ServerEdgeEvent_ServerEventType) {
	clients, ok := e.AppInstsStruct.AppInstsMap[appInstKey]
	if !ok {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find appinst. no clients connected to appinst have edge events connection.", "appInstKey", appInstKey)
		return
	}

	for client, mgr := range clients.ClientsMap {
		updateServerEdgeEvent := createAppInstStateEvent(ctx, appInst, eventType)
		log.SpanLog(ctx, log.DebugLevelInfra, "Sending update state to client", "client", client)
		e.SendEdgeEventToClient(ctx, updateServerEdgeEvent, mgr)
	}
}

// Send ServerEdgeEvent to specified client via persistent grpc stream
func (e *EdgeEventsHandlerPlugin) SendEdgeEventToClient(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, mgr *dmecommon.EdgeEventPersistentMgr) {
	mgr.Mux.Lock()
	defer mgr.Mux.Unlock()
	svr := mgr.Svr
	log.SpanLog(ctx, log.DebugLevelInfra, "about to send server edge event", "serveredgeevent", serverEdgeEvent, "appInstHealth", serverEdgeEvent.AppinstHealthState, "svr", &mgr.Svr)
	err := (*svr).Send(serverEdgeEvent)
	if err != nil {
		mgr.Err = err
		close(mgr.Terminated)
		log.SpanLog(ctx, log.DebugLevelInfra, "error on send to clients", "error", err)
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "successfully sent update to clients")
	}
}

// Create the ServerEdgeEvent with specified Event
func createAppInstStateEvent(ctx context.Context, appInst *dmecommon.DmeAppInst, eventType dme.ServerEdgeEvent_ServerEventType) *dme.ServerEdgeEvent {
	updateServerEdgeEvent := new(dme.ServerEdgeEvent)
	updateServerEdgeEvent.Event = eventType

	switch eventType {
	case dme.ServerEdgeEvent_EVENT_CLOUDLET_STATE:
		setCloudletState(ctx, appInst, updateServerEdgeEvent)
	case dme.ServerEdgeEvent_EVENT_CLOUDLET_MAINTENANCE:
		setCloudletMaintenanceState(ctx, appInst, updateServerEdgeEvent)
	case dme.ServerEdgeEvent_EVENT_APPINST_HEALTH:
		setAppInstHealthState(ctx, appInst, updateServerEdgeEvent)
	default:
	}

	return updateServerEdgeEvent
}

// Translate edgeproto CloudletState to dme CloudletState
func setCloudletState(ctx context.Context, appInst *dmecommon.DmeAppInst, serverEdgeEvent *dme.ServerEdgeEvent) {
	switch appInst.CloudletState {
	case edgeproto.CloudletState_CLOUDLET_STATE_UNKNOWN:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_STATE_UNKNOWN
	case edgeproto.CloudletState_CLOUDLET_STATE_ERRORS:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_STATE_ERRORS
	case edgeproto.CloudletState_CLOUDLET_STATE_READY:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_STATE_READY
	case edgeproto.CloudletState_CLOUDLET_STATE_OFFLINE:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_OFFLINE
	case edgeproto.CloudletState_CLOUDLET_STATE_NOT_PRESENT:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_OFFLINE
	case edgeproto.CloudletState_CLOUDLET_STATE_INIT:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_STATE_INIT
	case edgeproto.CloudletState_CLOUDLET_STATE_UPGRADE:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_UPGRADE
	case edgeproto.CloudletState_CLOUDLET_STATE_NEED_SYNC:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_STATE_UNKNOWN
	default:
		serverEdgeEvent.CloudletState = dme.ServerEdgeEvent_CLOUDLET_STATE_UNKNOWN
	}
}

// Translate edgeproto MaintenanceState to dme CloudletMaintenanceState
func setCloudletMaintenanceState(ctx context.Context, appInst *dmecommon.DmeAppInst, serverEdgeEvent *dme.ServerEdgeEvent) {
	switch appInst.MaintenanceState {
	case edgeproto.MaintenanceState_NORMAL_OPERATION:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_NORMAL
	case edgeproto.MaintenanceState_MAINTENANCE_START:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_UNDER_MAINTENANCE
	case edgeproto.MaintenanceState_FAILOVER_REQUESTED:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_FAILING_OVER
	case edgeproto.MaintenanceState_FAILOVER_DONE:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_FAILOVER_DONE
	case edgeproto.MaintenanceState_FAILOVER_ERROR:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_FAILOVER_ERROR
	case edgeproto.MaintenanceState_MAINTENANCE_START_NO_FAILOVER:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_UNDER_MAINTENANCE
	case edgeproto.MaintenanceState_CRM_REQUESTED:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_UNDER_MAINTENANCE
	case edgeproto.MaintenanceState_CRM_UNDER_MAINTENANCE:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_UNDER_MAINTENANCE
	case edgeproto.MaintenanceState_CRM_ERROR:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_MAINTENANCE_FAILED
	case edgeproto.MaintenanceState_UNDER_MAINTENANCE:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_UNDER_MAINTENANCE
	default:
		serverEdgeEvent.CloudletMaintenanceState = dme.ServerEdgeEvent_MAINTENANCE_STATE_UNKNOWN
	}
}

// Translate edgeproto HealthCheck to dme AppinstHealthState
func setAppInstHealthState(ctx context.Context, appInst *dmecommon.DmeAppInst, serverEdgeEvent *dme.ServerEdgeEvent) {
	switch appInst.AppInstHealth {
	case edgeproto.HealthCheck_HEALTH_CHECK_UNKNOWN:
		serverEdgeEvent.AppinstHealthState = dme.ServerEdgeEvent_HEALTH_CHECK_UNKNOWN
	case edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE:
		serverEdgeEvent.AppinstHealthState = dme.ServerEdgeEvent_HEALTH_CHECK_FAIL
	case edgeproto.HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL:
		serverEdgeEvent.AppinstHealthState = dme.ServerEdgeEvent_APPINST_DOWN
	case edgeproto.HealthCheck_HEALTH_CHECK_OK:
		serverEdgeEvent.AppinstHealthState = dme.ServerEdgeEvent_HEALTH_CHECK_OK
	default:
		serverEdgeEvent.AppinstHealthState = dme.ServerEdgeEvent_HEALTH_CHECK_UNKNOWN
	}
}
