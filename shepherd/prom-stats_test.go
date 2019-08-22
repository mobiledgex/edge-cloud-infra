package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/assert"
)

var testMetricSent = 0

var testPayloadData = map[string]string{
	promQCpuClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491286.389,
				"10.01"
			  ]
			}
		  ]
		}
	  }`,
	promQMemClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491347.686,
				"99.99"
			  ]
			}
		  ]
		}
	  }`,
	promQDiskClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491384.455,
				"50.0"
			  ]
			}
		  ]
		}
	  }`,
	promQSendBytesRateClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491412.872,
				"11111"
			  ]
			}
		  ]
		}
	  }`,
	promQRecvBytesRateClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491412.872,
				"22222"
			  ]
			}
		  ]
		}
	  }`,

	promQCpuPod: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {
				"pod_name": "testPod1"
			  },
			  "value": [
				1549491454.802,
				"5.0"
			  ]
			}
			]
		  }
		  }`,
	promQMemPod: `{
		"status": "success",
		"data": {
  		"resultType": "vector",
  		"result": [
			{
	  		"metric": {
				"pod_name": "testPod1"
	  		},
	  		"value": [
				1549484450.932,
				"100000000"
	  		]
			}
  		]
		}
		}`,
	promQDiskPod: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {
				"pod_name": "testPod1"
			},
			"value": [
				1549484450.932,
				"300000000"
			]
			}
		]
		}
		}`,
	promQNetSendRate: `{
		"status": "success",
		"data": {
  		"resultType": "vector",
  		"result": [
			{
	  		"metric": {
				"pod_name": "testPod1"
	  		},
	  		"value": [
				1549484450.932,
				"111111"
	  		]
			}
  		]
		}
		}`,
	promQNetRecvRate: `{
		"status": "success",
		"data": {
  		"resultType": "vector",
  		"result": [
			{
	  		"metric": {
				"pod_name": "testPod1"
	  		},
	  		"value": [
				1549484450.932,
				"222222"
	  		]
			}
  		]
		}
		}`,
}

func testMetricSend(ctx context.Context, metric *edgeproto.Metric) bool {
	testMetricSent = 1
	return true
}

func getTestMetrics(addr string, query string) (*PromResp, error) {
	input := []byte(testPayloadData[query])
	promResp := &PromResp{}
	if err := json.Unmarshal(input, &promResp); err != nil {
		return nil, err
	}
	return promResp, nil
}

func TestPromStats(t *testing.T) {
	log.InitTracer("")
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	testAppKey := MetricAppInstKey{
		clusterInstKey: edgeproto.ClusterInstKey{
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
		Key: testClusterInstKey,
	}

	*platformName = "PLATFORM_TYPE_FAKE"
	testPlatform, _ := getPlatform()

	// Skip this much of the URL
	skiplen := len("/api/v1/query?query=")
	// start up http server to serve Prometheus metrics data
	tsProm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testPayloadData[r.URL.String()[skiplen:]]))
	}))
	defer tsProm.Close()
	// Remove the leading "http://"
	testPromStats, err := NewClusterWorker(tsProm.URL[7:], time.Second*1, testMetricSend, &testClusterInst, testPlatform)
	assert.Nil(t, err, "Get a patform client for fake cloudlet")
	err = collectClusterPormetheusMetrics(testPromStats)
	assert.Nil(t, err, "Fill stats from json")
	testAppKey.pod = "testPod1"
	stat, found := testPromStats.appStatsMap[testAppKey]
	// Check PodStats
	assert.True(t, found, "Pod testPod1 is not found")
	if found {
		assert.Equal(t, float64(5.0), stat.cpu)
		assert.Equal(t, uint64(100000000), stat.mem)
		assert.Equal(t, uint64(300000000), stat.disk)
		assert.Equal(t, uint64(111111), stat.netSend)
		assert.Equal(t, uint64(222222), stat.netRecv)
	}
	// Check ClusterStats
	assert.Equal(t, float64(10.01), testPromStats.clusterStat.cpu)
	assert.Equal(t, float64(99.99), testPromStats.clusterStat.mem)
	assert.Equal(t, float64(50.0), testPromStats.clusterStat.disk)
	assert.Equal(t, uint64(11111), testPromStats.clusterStat.netSend)
	assert.Equal(t, uint64(22222), testPromStats.clusterStat.netRecv)

	// Check callback is called
	assert.Equal(t, int(0), testMetricSent)
	testPromStats.send(ctx, ClusterStatToMetrics(testPromStats)[0])
	assert.Equal(t, int(1), testMetricSent)
}
