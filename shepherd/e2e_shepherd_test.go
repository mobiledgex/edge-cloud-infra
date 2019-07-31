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

//spawn a fake prometheus with fake data numbers for the following queries

var fakeProm *httptest.Server

var e2eTestData = map[string]string{
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
	promQTcpConnClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491978.657,
				"33333"
			  ]
			}
		  ]
		}
	  }`,
	promQTcpRetransClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491007.677,
				"44444"
			  ]
			}
		  ]
		}
	  }`,
	promQUdpSendPktsClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491432.234,
				"55555"
			  ]
			}
		  ]
		}
	  }`,
	promQUdpRecvPktsClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491473.683,
				"66666"
			  ]
			}
		  ]
		}
	  }`,
	promQUdpRecvErr: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491345.543,
				"77777"
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
				1549485795.932,
				"200000000"
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
				1549484450.424,
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

func SetupFakeProm() {
	// Skip this much of the URL
	skiplen := len("/api/v1/query?query=")
	// start up http server to serve Prometheus metrics data
	fakeProm = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(e2eTestData[r.URL.String()[skiplen:]]))
	}))
}

func GetFakePromAddr() string {
	return fakeProm.URL[7:]
}

//delete fake prometheus 30 seconds from now?
func DeleteFakeProm() {
	fakeProm.Close()
}

func checkInflux() {

}
