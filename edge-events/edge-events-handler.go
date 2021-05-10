package edgeevents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/version"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	grpcstats "github.com/mobiledgex/edge-cloud/metrics/grpc"
)

// Implements dmecommon.EdgeEventsHandler interface
type EdgeEventsHandlerPlugin struct {
	// Hashmap containing AppInsts on DME mapped to the clients connected to those AppInsts
	AppInstsStruct             *AppInsts
	EdgeEventsCookieExpiration time.Duration
}

// Map AppInstKey to Clients on appinst
type AppInsts struct {
	AppInstsMap map[edgeproto.AppInstKey]*Clients
	mux         sync.RWMutex
}

// Map Client to ClientInfo
type Clients struct {
	ClientsMap map[Client]*ClientInfo
	mux        sync.RWMutex
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
	mux      sync.RWMutex
}

// Add Client connected to specified AppInst to Map
func (e *EdgeEventsHandlerPlugin) AddClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey, lastLoc *dme.Loc, carrier string, sendFunc func(event *dme.ServerEdgeEvent)) {
	// Get clients on specified appinst
	clients := e.AppInstsStruct.getOrAdd(appInstKey)
	// Initialize client and clientinfo for new client
	client := Client{cookieKey: cookieKey}
	clientinfo := &ClientInfo{
		sendFunc: sendFunc,
		lastLoc:  lastLoc,
		carrier:  carrier,
	}
	clients.add(client, clientinfo)
}

// Remove Client connected to specified AppInst from Map
func (e *EdgeEventsHandlerPlugin) RemoveClientKey(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey) {
	// Get clients on specified appinst
	clients, err := e.AppInstsStruct.get(appInstKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error getting appInstKey from AppInstsMap", "error", err)
		return
	}
	// Remove specified client
	client := Client{cookieKey}
	clients.remove(client)
}

// Update Client's last location
func (e *EdgeEventsHandlerPlugin) UpdateClientLastLocation(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey, lastLoc *dme.Loc) {
	clientinfo, err := e.getClientInfo(appInstKey, cookieKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to get clientinfo", "appInstKey", appInstKey, "client", cookieKey, "error", err)
		return
	}
	// Update lastLoc in clientinfo for specified client
	clientinfo.updateLastLoc(lastLoc)
}

// Remove AppInst from Map of AppInsts
func (e *EdgeEventsHandlerPlugin) RemoveAppInstKey(ctx context.Context, appInstKey edgeproto.AppInstKey) {
	// Remove appInstKey
	e.AppInstsStruct.remove(appInstKey)
}

// Handle processing of latency samples and then send back to client
// For now: Avg, Min, Max, StdDev
func (e *EdgeEventsHandlerPlugin) ProcessLatencySamples(ctx context.Context, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey, samples []*dme.Sample) (*dme.Statistics, error) {
	clientinfo, err := e.getClientInfo(appInstKey, cookieKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to get clientinfo", "appInstKey", appInstKey, "client", cookieKey, "err", err)
		return nil, fmt.Errorf("Unable to get clientinfo for client %v on appInst %v. Error is %s.", cookieKey, appInstKey, err)
	}
	// Create latencyEdgeEvent with processed stats
	latencyEdgeEvent := new(dme.ServerEdgeEvent)
	latencyEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_LATENCY_PROCESSED
	stats := grpcstats.CalculateStatistics(samples)
	latencyEdgeEvent.Statistics = &stats
	// Send processed stats to client
	clientinfo.send(latencyEdgeEvent)
	return &stats, nil
}

// Send a ServerEdgeEvent with Latency Request Event to all clients connected to specified AppInst (and also have an initiated persistent connection)
// When client recieves this event, it will measure latency from itself to appinst and back.
// Client will then send those latency samples back to be processed in the HandleLatencySamples function
// Finally, DME will send the processed latency samples in the form of dme.Latency struct (with calculated avg, min, max, stddev) back to client
func (e *EdgeEventsHandlerPlugin) SendLatencyRequestEdgeEvent(ctx context.Context, appInstKey edgeproto.AppInstKey) {
	// Get clients on specified appinst
	clients, err := e.AppInstsStruct.get(appInstKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find appinst. no clients connected to appinst have edge events connection.", "appInstKey", appInstKey, "err", err)
		return
	}
	// Send latency request to each client on appinst
	clients.mux.RLock()
	defer clients.mux.RUnlock()
	for _, clientinfo := range clients.ClientsMap {
		go func(clientinfo *ClientInfo) {
			latencyRequestEdgeEvent := new(dme.ServerEdgeEvent)
			latencyRequestEdgeEvent.EventType = dme.ServerEdgeEvent_EVENT_LATENCY_REQUEST
			clientinfo.send(latencyRequestEdgeEvent)
		}(clientinfo)
	}
}

// Send a ServerEdgeEvent with specified Event to all clients connected to specified AppInst (and also have initiated persistent connection)
func (e *EdgeEventsHandlerPlugin) SendAppInstStateEvent(ctx context.Context, appInst *dmecommon.DmeAppInst, appInstKey edgeproto.AppInstKey, eventType dme.ServerEdgeEvent_ServerEventType) {
	// Get clients on specified appinst
	clients, err := e.AppInstsStruct.get(appInstKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "cannot find appinst. no clients connected to appinst have edge events connection.", "appInstKey", appInstKey, "err", err)
		return
	}
	// Check if appinst is usable. If not do a FindCloudlet for each client
	appInstUsable := dmecommon.IsAppInstUsable(appInst)
	// Send appinst state event to each client on affected appinst
	clients.mux.RLock()
	defer clients.mux.RUnlock()
	for _, clientinfo := range clients.ClientsMap {
		go func(clientinfo *ClientInfo) {
			// Look for a new cloudlet if the appinst is not usable
			newCloudlet := new(dme.FindCloudletReply)
			var err error
			if !appInstUsable {
				err = dmecommon.FindCloudlet(ctx, &appInstKey.AppKey, clientinfo.getCarrier(), clientinfo.getLastLoc(), newCloudlet, e.EdgeEventsCookieExpiration)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "Current appinst is unusable, but error doing FindCloudlet", "err", err)
					err = fmt.Errorf("Current appinst is unusable, but error doing FindCloudlet - error is %s", err.Error())
					newCloudlet = nil
				} else if newCloudlet.Status != dme.FindCloudletReply_FIND_FOUND {
					log.SpanLog(ctx, log.DebugLevelInfra, "Current appinst is unusable, but unable to find any cloudlets", "FindStatus", newCloudlet.Status)
					err = fmt.Errorf("Current appinst is unusable, but unable to find any cloudlets - FindStatus is %s", newCloudlet.Status)
					newCloudlet = nil
				}
			}
			// Send the client the state event
			updateServerEdgeEvent := createAppInstStateEvent(ctx, appInst, eventType, newCloudlet)
			if err != nil {
				updateServerEdgeEvent.ErrorMsg = err.Error()
			}
			clientinfo.send(updateServerEdgeEvent)
		}(clientinfo)
	}
}

// Send ServerEdgeEvent to specified client via persistent grpc stream
func (e *EdgeEventsHandlerPlugin) SendEdgeEventToClient(ctx context.Context, serverEdgeEvent *dme.ServerEdgeEvent, appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey) {
	clientinfo, err := e.getClientInfo(appInstKey, cookieKey)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "unable to get clientinfo", "appInstKey", appInstKey, "client", cookieKey, "err", err)
		return
	}
	clientinfo.send(serverEdgeEvent)
}

// Create the ServerEdgeEvent with specified Event
func createAppInstStateEvent(ctx context.Context, appInst *dmecommon.DmeAppInst, eventType dme.ServerEdgeEvent_ServerEventType, newCloudlet *dme.FindCloudletReply) *dme.ServerEdgeEvent {
	updateServerEdgeEvent := new(dme.ServerEdgeEvent)
	updateServerEdgeEvent.EventType = eventType
	updateServerEdgeEvent.NewCloudlet = newCloudlet
	// Populate the corresponding ServerEdgeEvent field based on eventType
	switch eventType {
	case dme.ServerEdgeEvent_EVENT_CLOUDLET_STATE:
		updateServerEdgeEvent.CloudletState = appInst.CloudletState
	case dme.ServerEdgeEvent_EVENT_CLOUDLET_MAINTENANCE:
		updateServerEdgeEvent.MaintenanceState = appInst.MaintenanceState
	case dme.ServerEdgeEvent_EVENT_APPINST_HEALTH:
		updateServerEdgeEvent.HealthCheck = appInst.AppInstHealth
	default:
	}
	return updateServerEdgeEvent
}

// Helper function that gets the clientinfo of the specified client on the specified appinst
func (e *EdgeEventsHandlerPlugin) getClientInfo(appInstKey edgeproto.AppInstKey, cookieKey dmecommon.CookieKey) (*ClientInfo, error) {
	// Get clients on specified appinst
	clients, err := e.AppInstsStruct.get(appInstKey)
	if err != nil {
		return nil, fmt.Errorf("error getting appInstKey from AppInstsMap - %v", err)
	}
	// Get clientinfo for specified client
	client := Client{cookieKey}
	clientinfo, err := clients.get(client)
	if err != nil {
		return nil, fmt.Errorf("error getting client from ClientsMap - %v", err)
	}
	return clientinfo, nil
}

func (e *EdgeEventsHandlerPlugin) GetVersionProperties() map[string]string {
	return version.InfraBuildProps("EdgeEvents")
}
