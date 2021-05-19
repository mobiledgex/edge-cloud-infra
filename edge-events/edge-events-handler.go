package edgeevents

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/version"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
	"github.com/mobiledgex/edge-cloud/util"
)

// Implements dmecommon.EdgeEventsHandler interface
type EdgeEventsHandlerPlugin struct {
	mux util.Mutex
	// Hashmap containing Cloudlets mapped to AppInsts mapped to the clients connected to those AppInsts
	CloudletsStruct            *Cloudlets
	EdgeEventsCookieExpiration time.Duration
}

// Map CloudletKey to AppInsts on that cloudlet
type Cloudlets struct {
	CloudletsMap map[edgeproto.CloudletKey]*AppInsts
}

// Map AppInstKey to Clients on appinst
type AppInsts struct {
	AppInstsMap map[edgeproto.AppInstKey]*Clients
}

// Map Client to ClientInfo
type Clients struct {
	ClientsMap map[Client]*ClientInfo
}

// Client uniquely identified by session cookie
type Client struct {
	cookieKey dmecommon.CookieKey
}

// Client info contains the client's specific Send function, last location, and carrier
type ClientInfo struct {
	sendFunc func(event *dme.ServerEdgeEvent)
	lastLoc  *dme.Loc
	carrier  string
}

// Add Client connected to specified AppInst to Map
func (e *EdgeEventsHandlerPlugin) AddClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey, lastLoc *dme.Loc, carrier string, sendFunc func(event *dme.ServerEdgeEvent)) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Initialize CloudletsMap
	if e.CloudletsStruct.CloudletsMap == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "initializing cloudlets map")
		e.CloudletsStruct.CloudletsMap = make(map[edgeproto.CloudletKey]*AppInsts)
	}
	cloudletKey := appInstKey.ClusterInstKey.CloudletKey
	appinsts, ok := e.CloudletsStruct.CloudletsMap[cloudletKey]
	if !ok {
		// add first appinst
		newAppInsts := new(AppInsts)
		newAppInsts.AppInstsMap = make(map[edgeproto.AppInstKey]*Clients)
		appinsts = newAppInsts
		e.CloudletsStruct.CloudletsMap[cloudletKey] = appinsts
	}
	// Get clients on specified appinst
	clients, ok := appinsts.AppInstsMap[appInstKey]
	if !ok {
		// add first client for appinst
		newClients := new(Clients)
		newClients.ClientsMap = make(map[Client]*ClientInfo)
		clients = newClients
		appinsts.AppInstsMap[appInstKey] = clients
	}
	// Initialize client info for new client
	client := Client{cookieKey: cookieKey}
	clients.ClientsMap[client] = &ClientInfo{
		sendFunc: sendFunc,
		lastLoc:  lastLoc,
		carrier:  carrier,
	}
}

// Remove Client connected to specified AppInst from Map
func (e *EdgeEventsHandlerPlugin) RemoveClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Get clients on specified appinst
	clients, err := e.getClients(ctx, appInstKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find appinst to remove", "appInstKey", appInstKey, "error", err)
		return
	}
	// Remove specified client
	client := Client{cookieKey}
	delete(clients.ClientsMap, client)
}

// Update Client's last location
func (e *EdgeEventsHandlerPlugin) UpdateClientLastLocation(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey, lastLoc *dme.Loc) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Update lastLoc field in cliientinfo for specified client
	clientinfo, err := e.getClientInfo(ctx, appInstKey, cookieKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find client to update", "client", cookieKey, "error", err)
		return
	}
	clientinfo.lastLoc = lastLoc
}

func (e *EdgeEventsHandlerPlugin) RemoveCloudletKey(ctx context.Context, cloudletKey edgeproto.CloudletKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.CloudletsStruct.CloudletsMap != nil {
		// Remove cloudlet from map of cloudlets
		delete(e.CloudletsStruct.CloudletsMap, cloudletKey)
	}
}

// Remove AppInst from Map of AppInsts
func (e *EdgeEventsHandlerPlugin) RemoveAppInstKey(ctx context.Context, appInstKey edgeproto.AppInstKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Check to see if appinst exists
	appinsts, err := e.getAppInsts(ctx, appInstKey.ClusterInstKey.CloudletKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to find appinst to remove", "appInstKey", appInstKey, "error", err)
		return
	}
	// Remove appinst from map of appinsts
	delete(appinsts.AppInstsMap, appInstKey)
}

// Handle processing of latency samples and then send back to client
// For now: Avg, Min, Max, StdDev
func (e *EdgeEventsHandlerPlugin) ProcessLatencySamples(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey, samples []*dme.Sample) (*dme.Statistics, error) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Check to see if client is on appinst
	clientinfo, err := e.getClientInfo(ctx, appInstKey, cookieKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find client connected to appinst", "appInstKey", appInstKey, "client", cookieKey, "error", err)
		return nil, fmt.Errorf("cannot find client connected to appinst %v - error is %v", appInstKey, err)
	}
	// Create latencyEdgeEvent with processed stats
	latencyEdgeEvent := new(dme.ServerEdgeEvent)
	latencyEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_LATENCY_PROCESSED
	stats := grpcstats.CalculateStatistics(samples)
	latencyEdgeEvent.Statistics = &stats
	// Send processed stats to client
	clientinfo.sendFunc(latencyEdgeEvent)
	return &stats, nil
}

// Send a ServerEdgeEvent with Latency Request Event to all clients connected to specified AppInst (and also have an initiated persistent connection)
// When client recieves this event, it will measure latency from itself to appinst and back.
// Client will then send those latency samples back to be processed in the HandleLatencySamples function
// Finally, DME will send the processed latency samples in the form of dme.Latency struct (with calculated avg, min, max, stddev) back to client
func (e *EdgeEventsHandlerPlugin) SendLatencyRequestEdgeEvent(ctx context.Context, appInstKey edgeproto.AppInstKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Get clients on specified appinst
	clients, err := e.getClients(ctx, appInstKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find appinst. no clients connected to appinst have edge events connection.", "appInstKey", appInstKey, "error", err)
		return
	}
	// Send latency request to each client on appinst
	go func() {
		e.mux.Lock()
		defer e.mux.Unlock()
		for _, clientinfo := range clients.ClientsMap {
			latencyRequestEdgeEvent := new(dme.ServerEdgeEvent)
			latencyRequestEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_LATENCY_REQUEST
			clientinfo.sendFunc(latencyRequestEdgeEvent)
		}
	}()
}

// Send a AppInstState EdgeEvent with specified Event to all clients connected to specified AppInst (and also have initiated persistent connection)
func (e *EdgeEventsHandlerPlugin) SendAppInstStateEdgeEvent(ctx context.Context, appInst *dmecommon.DmeAppInst, appInstKey edgeproto.AppInstKey, eventType dme.ServerEdgeEvent_ServerEventType) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Get clients on specified appinst
	clients, err := e.getClients(ctx, appInstKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find appinst. no clients connected to appinst have edge events connection.", "appInstKey", appInstKey, "error", err)
		return
	}
	// Check if appinst is usable. If not do a FindCloudlet for each client
	usability := getUsability(appInst.MaintenanceState, appInst.CloudletState, appInst.AppInstHealth)
	// Send appinst state event to each client on affected appinst
	go func() {
		e.mux.Lock()
		defer e.mux.Unlock()
		for _, clientinfo := range clients.ClientsMap {
			appInstStateEdgeEvent := e.createAppInstStateEdgeEvent(ctx, appInst, appInstKey, clientinfo, eventType, usability)
			clientinfo.sendFunc(appInstStateEdgeEvent)
		}
	}()
}

func (e *EdgeEventsHandlerPlugin) SendCloudletStateEdgeEvent(ctx context.Context, cloudlet *dmecommon.DmeCloudlet, cloudletKey edgeproto.CloudletKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Get appinsts on specified cloudlet
	appinsts, err := e.getAppInsts(ctx, cloudletKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find cloudlet. no appinsts on cloudlet", "cloudlet", cloudletKey, "error", err)
		return
	}
	// Check if cloudlet is usable. If not do a FindCloudlet for each client (only use cloudlet.State)
	usability := getUsability(dme.MaintenanceState_NORMAL_OPERATION, cloudlet.State, dme.HealthCheck_HEALTH_CHECK_OK)
	// Send cloudlet state event to each client on each appinst on the affected cloudlet
	go func() {
		e.mux.Lock()
		e.mux.Unlock()
		for key, clients := range appinsts.AppInstsMap {
			for _, clientinfo := range clients.ClientsMap {
				cloudletStateEdgeEvent := e.createCloudletStateEdgeEvent(ctx, cloudlet, key, clientinfo, usability)
				clientinfo.sendFunc(cloudletStateEdgeEvent)
			}
		}
	}()
}

func (e *EdgeEventsHandlerPlugin) SendCloudletMaintenanceStateEdgeEvent(ctx context.Context, cloudlet *dmecommon.DmeCloudlet, cloudletKey edgeproto.CloudletKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Get appinsts on specified cloudlet
	appinsts, err := e.getAppInsts(ctx, cloudletKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find cloudlet. no appinsts on cloudlet", "cloudlet", cloudletKey, "error", err)
		return
	}
	// Check if cloudlet is usable. If not do a FindCloudlet for each client (only use cloudlet.MaintenanceState)
	usability := getUsability(cloudlet.MaintenanceState, dme.CloudletState_CLOUDLET_STATE_READY, dme.HealthCheck_HEALTH_CHECK_OK)
	// Send cloudlet state event to each client on each appinst on the affected cloudlet
	go func() {
		e.mux.Lock()
		e.mux.Unlock()
		for key, clients := range appinsts.AppInstsMap {
			for _, clientinfo := range clients.ClientsMap {
				cloudletStateEdgeEvent := e.createCloudletMaintenanceStateEdgeEvent(ctx, cloudlet, key, clientinfo, usability)
				clientinfo.sendFunc(cloudletStateEdgeEvent)
			}
		}
	}()
}

// Send ServerEdgeEvent to specified client via persistent grpc stream
func (e *EdgeEventsHandlerPlugin) SendEdgeEventToClient(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey) {
	e.mux.Lock()
	defer e.mux.Unlock()
	// Check to see if client is on appinst
	clientinfo, err := e.getClientInfo(ctx, appInstKey, cookieKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find client connected to appinst", "appInstKey", appInstKey, "client", cookieKey, "error", err)
		return
	}
	clientinfo.sendFunc(serverEdgeEvent)
}

func (e *EdgeEventsHandlerPlugin) GetVersionProperties() map[string]string {
	return version.InfraBuildProps("EdgeEvents")
}
