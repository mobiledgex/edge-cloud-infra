package main

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
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
// Output is a single container with no app associated with it
// and two containers associated with app testAppInstDocker2
var testDockerResults = []string{`{
	"container": "DockerApp1",
	"memory": {
	  "raw": "11.11MiB / 1.111GiB",
	  "percent": "1.11%"
	},
	"cpu": "1.11%",
	"io": {
	  "network": "111B / 1KB",
	  "block": "1B / 1B"
	}
  }`, `{
	"container": "DockerApp2Container1",
	"memory": {
	  "raw": "21.21MiB / 2.1GiB",
	  "percent": "2.1%"
	},
	"cpu": "2.1%",
	"io": {
	  "network": "21B / 21KB",
	  "block": "21B / 21B"
	}
  }`, `{
	"container": "DockerApp2Container2",
	"memory": {
	  "raw": "22.22MiB / 2.2GiB",
	  "percent": "2.2%"
	},
	"cpu": "2.2%",
	"io": {
	  "network": "22B / 22KB",
	  "block": "2B / 2B"
	}
  }`}

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
				Organization: "testoperator",
				Name:         "testcloudlet",
			},
			Organization: "",
		},
	}

	testOperator := "testoperator"
	testCloudletKey := edgeproto.CloudletKey{
		Organization: testOperator,
		Name:         "testcloudlet",
	}
	testClusterKey := edgeproto.ClusterKey{Name: "testcluster"}
	testClusterInstKey := edgeproto.ClusterInstKey{
		ClusterKey:   testClusterKey,
		CloudletKey:  testCloudletKey,
		Organization: "",
	}
	testClusterInst := edgeproto.ClusterInst{
		Key:        testClusterInstKey,
		Deployment: cloudcommon.AppDeploymentTypeDocker,
	}
	testAppInstDocker2 := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Name:    "DockerApp2",
				Version: "10",
			},
			ClusterInstKey: testClusterInstKey,
		},
		RuntimeInfo: edgeproto.AppInstRuntime{
			ContainerIds: []string{"DockerApp2Container1", "DockerApp2Container2"},
		},
	}

	// Remove all the empty space from docker test data, as that's how it gets returned by docker stats
	for i, s := range testDockerResults {
		testDockerResults[i] = regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), "")
	}
	tmpStr := strings.Join(testDockerResults, "\n")
	buf, err := json.Marshal(testDockerClusterData)
	assert.Nil(t, err, "marshal cluster metrics")
	platform := shepherd_unittest.Platform{
		DockerAppMetrics:     tmpStr,
		DockerClusterMetrics: string(buf),
	}
	edgeproto.InitAppInstCache(&AppInstCache)
	AppInstCache.Update(ctx, &testAppInstDocker2, 0)
	testDockerStats, err := NewClusterWorker(ctx, "", time.Second*1, nil, &testClusterInst, &platform)
	assert.Nil(t, err, "Get a patform client for unit test cloudlet")
	clusterMetrics := testDockerStats.clusterStat.GetClusterStats(ctx)
	appsMetrics := testDockerStats.clusterStat.GetAppStats(ctx)
	assert.NotNil(t, clusterMetrics, "Fill stats from json")
	assert.NotNil(t, appsMetrics, "Fill stats from json")
	testAppKey.Pod = k8smgmt.NormalizeName("DockerApp1")
	testAppKey.App = k8smgmt.NormalizeName("DockerApp1")
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
	testAppKey.Pod = k8smgmt.NormalizeName("DockerApp2")
	testAppKey.App = k8smgmt.NormalizeName("DockerApp2")
	testAppKey.Version = k8smgmt.NormalizeName("10")
	stat, found = appsMetrics[testAppKey]
	// Check PodStats - should be a sum of DockerApp2Container1 and DockerApp2Container2
	assert.True(t, found, "Container DockerApp2 is not found")
	if found {
		assert.Equal(t, float64(2.1)+float64(2.2), stat.Cpu)
		assert.Equal(t, uint64(22240296+23299358), stat.Mem)
		assert.Equal(t, uint64(0), stat.Disk)
		assert.Equal(t, uint64(21504+22528), stat.NetSent)
		assert.Equal(t, uint64(21+22), stat.NetRecv)
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
