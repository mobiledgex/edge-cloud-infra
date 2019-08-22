package main

import (
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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
	cpu       float64
	cpuTS     *types.Timestamp
	mem       uint64
	memTS     *types.Timestamp
	disk      uint64
	diskTS    *types.Timestamp
	netSend   uint64
	netSendTS *types.Timestamp
	netRecv   uint64
	netRecvTS *types.Timestamp
}

type ClusterMetrics struct {
	cpu          float64
	cpuTS        *types.Timestamp
	mem          float64
	memTS        *types.Timestamp
	disk         float64
	diskTS       *types.Timestamp
	netSend      uint64
	netSendTS    *types.Timestamp
	netRecv      uint64
	netRecvTS    *types.Timestamp
	tcpConns     uint64
	tcpConnsTS   *types.Timestamp
	tcpRetrans   uint64
	tcpRetransTS *types.Timestamp
	udpSend      uint64
	udpSendTS    *types.Timestamp
	udpRecv      uint64
	udpRecvTS    *types.Timestamp
	udpRecvErr   uint64
	udpRecvErrTS *types.Timestamp
}

// We keep the name of the pod+ClusterInstKey rather than AppInstKey
// The reson is that we do not have a way to differentiate between different pods in a k8s cluster
// See EDGECLOUD-1183
type MetricAppInstKey struct {
	clusterInstKey edgeproto.ClusterInstKey
	pod            string
}

// K8s Cluster
type K8sClusterStats struct {
	key      edgeproto.ClusterInstKey
	promAddr string
	client   pc.PlatformClient
	ClusterMetrics
}

func (c *K8sClusterStats) GetClusterStats() *ClusterMetrics {
	if c.promAddr == "" {
		return nil
	}
	if err := collectClusterPormetheusMetrics(c); err != nil {
		log.DebugLog(log.DebugLevelMetrics, "Could not collect cluster metrics", "K8s Cluster", c)
		return nil
	}
	return &c.ClusterMetrics
}

// Currently we are collecting stats for all apps in the cluster in one shot
// Implementing  EDGECLOUD-1183 would allow us to query by label and we can have each app be an individual metric
func (c *K8sClusterStats) GetAppStats() map[MetricAppInstKey]*AppMetrics {
	if c.promAddr == "" {
		return nil
	}
	metrics := collectAppPrometheusMetrics(c)
	if metrics == nil {
		log.DebugLog(log.DebugLevelMetrics, "Could not collect app metrics", "K8s Cluster", c)
	}
	return metrics
}
