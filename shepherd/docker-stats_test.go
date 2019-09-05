package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_unittest"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
)

// Sample output of formatted 'docker stats' cmd
var testDockerResult = `{
	"App": "DockerApp1",
	"memory": {
	  "raw": "11.11MiB / 1.111GiB",
	  "percent": "1.11%"
	},
	"cpu": "1.11%",
	"io": {
	  "network": "111B / 1KB",
	  "block": "1B / 1B"
	}
  }`

// Example output of resource-tracker
var testDockerClusterData = shepherd_common.ClusterMetrics{
	Cpu:        10.10101010,
	Mem:        11.111111,
	Disk:       12.12121212,
	NetSent:    1313131313,
	NetRecv:    1414141414,
	TcpConns:   1515,
	TcpRetrans: 16,
	UdpSent:    1717,
	UdpRecv:    1818,
	UdpRecvErr: 19,
}

func TestDockerStats(t *testing.T) {
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	testAppKey := shepherd_common.MetricAppInstKey{
		ClusterInstKey: edgeproto.ClusterInstKey{
			ClusterKey: edgeproto.ClusterKey{
				Name: "testcluster",
			},
			CloudletKey: edgeproto.CloudletKey{
				OperatorKey: edgeproto.OperatorKey{
					Name: "testoper",
				},
				Name: "testcloudlet",
			},
			Developer: "",
		},
	}

	testOperatorKey := edgeproto.OperatorKey{Name: "testoper"}
	testCloudletKey := edgeproto.CloudletKey{
		OperatorKey: testOperatorKey,
		Name:        "testcloudlet",
	}
	testClusterKey := edgeproto.ClusterKey{Name: "testcluster"}
	testClusterInstKey := edgeproto.ClusterInstKey{
		ClusterKey:  testClusterKey,
		CloudletKey: testCloudletKey,
		Developer:   "",
	}
	testClusterInst := edgeproto.ClusterInst{
		Key:        testClusterInstKey,
		Deployment: cloudcommon.AppDeploymentTypeDocker,
	}

	buf, err := json.Marshal(testDockerClusterData)
	assert.Nil(t, err, "marshal cluster metrics")
	platform := shepherd_unittest.Platform{
		DockerAppMetrics:     testDockerResult,
		DockerClusterMetrics: string(buf),
	}

	testPromStats, err := NewClusterWorker(ctx, "", time.Second*1, nil, &testClusterInst, &platform)
	assert.Nil(t, err, "Get a patform client for unit test cloudlet")
	clusterMetrics := testPromStats.clusterStat.GetClusterStats(ctx)
	appsMetrics := testPromStats.clusterStat.GetAppStats(ctx)
	assert.NotNil(t, clusterMetrics, "Fill stats from json")
	assert.NotNil(t, appsMetrics, "Fill stats from json")
	testAppKey.Pod = k8smgmt.NormalizeName("DockerApp1")
	stat, found := appsMetrics[testAppKey]
	// Check PodStats
	assert.True(t, found, "Container DockerApp1 is not found")
	if found {
		assert.Equal(t, float64(1.11), stat.Cpu)
		assert.Equal(t, uint64(11649679), stat.Mem)
		assert.Equal(t, uint64(0), stat.Disk)
		assert.Equal(t, uint64(1024), stat.NetSent)
		assert.Equal(t, uint64(111), stat.NetRecv)
		assert.NotNil(t, stat.CpuTS, "CPU timestamp")
	}

	// Check ClusterStats
	assert.Equal(t, testDockerClusterData.Cpu, clusterMetrics.Cpu)
	assert.NotNil(t, clusterMetrics.CpuTS, "CPU timestamp for cluster")
	assert.Equal(t, testDockerClusterData.Mem, clusterMetrics.Mem)
	assert.Equal(t, testDockerClusterData.Disk, clusterMetrics.Disk)
	assert.Equal(t, testDockerClusterData.NetSent, clusterMetrics.NetSent)
	assert.Equal(t, testDockerClusterData.NetRecv, clusterMetrics.NetRecv)
	assert.Equal(t, testDockerClusterData.TcpConns, clusterMetrics.TcpConns)
	assert.Equal(t, testDockerClusterData.TcpRetrans, clusterMetrics.TcpRetrans)
	assert.Equal(t, testDockerClusterData.UdpSent, clusterMetrics.UdpSent)
	assert.Equal(t, testDockerClusterData.UdpRecv, clusterMetrics.UdpRecv)
	assert.Equal(t, testDockerClusterData.UdpRecvErr, clusterMetrics.UdpRecvErr)
}
