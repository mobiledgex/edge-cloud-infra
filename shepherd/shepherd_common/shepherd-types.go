package shepherd_common

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Prerequisite - install small edge-cloud utility on the VM running this docker containers
var ResTrackerCmd = "resource-tracker"

// Common interface to deal with AppMetrics
// Pending EDGECLOUD-1183 implementation
type AppStats interface {
	// Returns current resource usage for a app instance
	GetAppStats(ctx context.Context) *AppMetrics
}

// Common interface to deal with ClusterMetrics
type ClusterStats interface {
	// Returns current resource usage for a cluster instance
	GetClusterStats(ctx context.Context) *ClusterMetrics
	GetAppStats(ctx context.Context) map[MetricAppInstKey]*AppMetrics
	GetAlerts(ctx context.Context) []edgeproto.Alert
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

// This structure represents cloudlet utilization stats
// It tracks the Max Available and currently used set of
// resources
type CloudletMetrics struct {
	ComputeTS *types.Timestamp
	// Total number of CPUs
	VCpuMax uint64
	// Current number of CPUs used
	VCpuUsed uint64
	// Total amount of RAM(in MB)
	MemMax uint64
	// Currently used RAM(in MB)
	MemUsed uint64
	// Total amount of Storage(in GB)
	DiskUsed uint64
	// Currently used Storage(in GB)
	DiskMax uint64
	// Total number of Floating IPs available
	FloatingIpsMax uint64
	// Currently used number of Floating IPs
	FloatingIpsUsed uint64
	NetworkTS       *types.Timestamp
	// Total KBytes received
	NetRecv uint64
	// Total KBytes sent
	NetSent   uint64
	IpUsageTS *types.Timestamp
	// Total available IP addresses
	Ipv4Max uint64
	// Currently used IP addrs
	Ipv4Used uint64
}

type ProxyMetrics struct {
	ActiveConn  uint64
	Accepts     uint64
	HandledConn uint64
	Requests    uint64
	Reading     uint64
	Writing     uint64
	Waiting     uint64
	Nginx       bool
	EnvoyStats  map[int32]ConnectionsMetric
	Ts          *types.Timestamp
}

type ConnectionsMetric struct {
	ActiveConn    uint64
	Accepts       uint64
	HandledConn   uint64
	SessionTime   []float64
	AvgBytesSent  float64
	AvgBytesRecvd float64
}

// We keep the name of the pod+ClusterInstKey rather than AppInstKey
// The reson is that we do not have a way to differentiate between different pods in a k8s cluster
// See EDGECLOUD-1183
type MetricAppInstKey struct {
	ClusterInstKey edgeproto.ClusterInstKey
	Pod            string
}
