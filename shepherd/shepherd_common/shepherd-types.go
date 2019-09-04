package shepherd_common

import (
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Common interface to deal with AppMetrics
// Pending EDGECLOUD-1183 implementation
type AppStats interface {
	// Returns current resource usage for a app instance
	GetAppStats() *AppMetrics
}

// Common interface to deal with ClusterMetrics
type ClusterStats interface {
	// Returns current resource usage for a cluster instance
	GetClusterStats() *ClusterMetrics
	GetAppStats() map[MetricAppInstKey]*AppMetrics
}

type AppMetrics struct {
	Cpu       float64
	CpuTS     *types.Timestamp
	Mem       uint64
	MemTS     *types.Timestamp
	Disk      uint64
	DiskTS    *types.Timestamp
	NetSent   uint64
	NetSentTS *types.Timestamp
	NetRecv   uint64
	NetRecvTS *types.Timestamp
}

type ClusterMetrics struct {
	Cpu          float64
	CpuTS        *types.Timestamp
	Mem          float64
	MemTS        *types.Timestamp
	Disk         float64
	DiskTS       *types.Timestamp
	NetSent      uint64
	NetSentTS    *types.Timestamp
	NetRecv      uint64
	NetRecvTS    *types.Timestamp
	TcpConns     uint64
	TcpConnsTS   *types.Timestamp
	TcpRetrans   uint64
	TcpRetransTS *types.Timestamp
	UdpSent      uint64
	UdpSentTS    *types.Timestamp
	UdpRecv      uint64
	UdpRecvTS    *types.Timestamp
	UdpRecvErr   uint64
	UdpRecvErrTS *types.Timestamp
}

// We keep the name of the pod+ClusterInstKey rather than AppInstKey
// The reson is that we do not have a way to differentiate between different pods in a k8s cluster
// See EDGECLOUD-1183
type MetricAppInstKey struct {
	ClusterInstKey edgeproto.ClusterInstKey
	Pod            string
}
