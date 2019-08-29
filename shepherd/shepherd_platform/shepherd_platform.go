package shepherd_platform

import (
	"context"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Platform abstracts the underlying cloudlet platform.
type Platform interface {
	// Set context for Span logging
	SetContext(ctx context.Context)
	// GetType Returns the Cloudlet's stack type, i.e. Openstack, Azure, etc.
	GetType() string
	// Init is called once during shepherd startup.
	Init(key *edgeproto.CloudletKey, physicalName, vaultAddr string) error
	// Gets the IP for a cluster
	GetClusterIP(clusterInst *edgeproto.ClusterInst) (string, error)
	// Gets a platform client to be able to runn commands against (mainly for curling the prometheuses)
	GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error)
}
