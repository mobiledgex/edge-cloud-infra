package main

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/assert"
)

func TestDockerStats(t *testing.T) {

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

	*platformName = "PLATFORM_TYPE_UNITTEST"
	testPlatform, _ := getPlatform()

	testPromStats, err := NewClusterWorker("", time.Second*1, nil, &testClusterInst, testPlatform)
	assert.Nil(t, err, "Get a patform client for unit test cloudlet")
	clusterMetrics := testPromStats.clusterStat.GetClusterStats()
	appsMetrics := testPromStats.clusterStat.GetAppStats()
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
	assert.Equal(t, float64(10.10101010), clusterMetrics.Cpu)
	assert.NotNil(t, clusterMetrics.CpuTS, "CPU timestamp for cluster")
	assert.Equal(t, float64(11.111111), clusterMetrics.Mem)
	assert.Equal(t, float64(12.12121212), clusterMetrics.Disk)
	assert.Equal(t, uint64(1313131313), clusterMetrics.NetSent)
	assert.Equal(t, uint64(1414141414), clusterMetrics.NetRecv)
}
