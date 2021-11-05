package main

import (
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/crmutil"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/notify"
)

// ControllerData contains cache data for controller
type ControllerData struct {
	*crmutil.ControllerData
}

// NewControllerData creates a new instance to track data from the controller
func NewControllerData(pf platform.Platform, nodeMgr *node.NodeMgr) *ControllerData {
	cd := &ControllerData{}
	cd.ControllerData = crmutil.NewControllerData(pf, &edgeproto.CloudletKey{}, nodeMgr)
	return cd
}

func InitClientNotify(client *notify.Client, nodeMgr *node.NodeMgr, cd *ControllerData) {
	crmutil.InitClientNotify(client, nodeMgr, cd.ControllerData)
}
