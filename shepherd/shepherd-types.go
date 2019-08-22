package main

import (
	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// Common interface to deal with AppMetrics
type AppStats interface {
	// Returns current resource usage for a app instance
	GetAppStats() *AppMetrics
}

// Common interface to deal with ClusterMetrics
type ClusterStats interface {
	GetClusterStats() *ClusterMetrics
	// Returns current resource usage for a cluster instance
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

// TODO - why not edgeproto.AppInstKey? how to account for the mobiledgeX infra pods(i.e. prometheus, weave, etc.)
type MetricAppInstKey struct {
	operator  string
	cloudlet  string
	cluster   string
	pod       string
	developer string
}

// K8s App. We use it to deal with a kuberneter app metrics
type K8sAppStats struct {
	edgeproto.AppKey
	client pc.PlatformClient
}

func (app *K8sAppStats) GetAppStats() *AppMetrics {
	//TODO - nill is error
	return nil
}

// Docker App. TODO: NEEDS IMPLEMENTING
type DockerAppStats struct {
	edgeproto.AppKey
	client pc.PlatformClient
}

func (app *DockerAppStats) GetAppStats() *AppMetrics {
	//TODO - nill is error
	return nil
}

// OpenstackVM App. TODO: NEEDS IMPLEMENTING
type OpenStackVmAppStats struct {
	edgeproto.AppKey
	client pc.PlatformClient
}

func (app *OpenStackVmAppStats) GetAppStats() *AppMetrics {
	//TODO - nill is error
	return nil
}
