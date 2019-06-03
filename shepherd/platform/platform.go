package platform

import (
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// GetType Returns the Cloudlet's stack type, i.e. Openstack, Azure, etc.
	GetType() string
	// Init is called once during shepherd startup.
	Init(key *edgeproto.CloudletKey) error
	// Gets the IP for a cluster
	GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error)
}
