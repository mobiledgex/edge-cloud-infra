package edgeevents

import (
	"fmt"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

/*
 * Utility Functions that safely handle getting, adding, and removing keys in the EdgeEvents maps
 * get, add, and remove key functions for *AppInsts (AppInstMap)
 * get, add, and remove key functions for *Clients (ClientsMap)
 * getLastLoc, updateLastLoc, getCarrier, send functions for *ClientInfo
 */

// Utility function that grabs the Clients on the specified appinst safely
func (a *AppInsts) get(key edgeproto.AppInstKey) (*Clients, error) {
	a.mux.RLock()
	defer a.mux.RUnlock()
	// Get clients on specified appinst
	clients, ok := a.AppInstsMap[key]
	if !ok {
		return nil, fmt.Errorf("Unable to find appinst - appInstKey: %v", key)
	}
	return clients, nil
}

// Utility function that adds a new AppInstKey to AppInstsMap safely
func (a *AppInsts) add(key edgeproto.AppInstKey, clients *Clients) {
	a.mux.Lock()
	defer a.mux.Unlock()
	// add Clients for specified AppInstKey
	a.AppInstsMap[key] = clients
}

// Utility function that gets the value of the specified AppInstKey or adds a new AppInstKey to AppInstsMap safely
func (a *AppInsts) getOrAdd(key edgeproto.AppInstKey) *Clients {
	a.mux.Lock()
	defer a.mux.Unlock()
	// Get clients for specified AppInstKey
	clients, ok := a.AppInstsMap[key]
	if !ok {
		// add first client for appinst
		newClients := new(Clients)
		newClients.ClientsMap = make(map[Client]*ClientInfo)
		clients = newClients
		a.AppInstsMap[key] = clients
	}
	return clients
}

// Utility function that removes an AppInstKey from AppInstsMap safely
func (a *AppInsts) remove(key edgeproto.AppInstKey) {
	a.mux.Lock()
	defer a.mux.Unlock()
	// Remove appinst from map of appinsts
	delete(a.AppInstsMap, key)
}

// Utility function that grabs the ClientInfo for the specified Client safely
func (c *Clients) get(client Client) (*ClientInfo, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	// Get ClientInfo for specified Client
	clientinfo, ok := c.ClientsMap[client]
	if !ok {
		return nil, fmt.Errorf("Unable to find client - Client: %v", client)
	}
	return clientinfo, nil
}

// Utility function that adds a new Client its corresponding ClientInfo to ClientsMap safely
func (c *Clients) add(client Client, clientInfo *ClientInfo) {
	c.mux.Lock()
	defer c.mux.Unlock()
	// add ClientInfo for Client
	c.ClientsMap[client] = clientInfo
}

// Utility function that removes a Client from ClientsMap safely
func (c *Clients) remove(client Client) {
	c.mux.Lock()
	defer c.mux.Unlock()
	// Remove client from map of clients
	delete(c.ClientsMap, client)
}

// Utility function that grabs the last location from ClientInfo safely
func (c *ClientInfo) getLastLoc() *dme.Loc {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.lastLoc
}

// Utility function that updates the last location in ClientInfo safely
func (c *ClientInfo) updateLastLoc(loc *dme.Loc) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.lastLoc = loc
}

// Utility function that grabs the carrier from ClientInfo safely
func (c *ClientInfo) getCarrier() string {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.carrier
}

// Utility function that sends ServerEdgeEvent safely
func (c *ClientInfo) send(event *dme.ServerEdgeEvent) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	c.sendFunc(event)
}
