package shepherd_fake

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
)

var promQCpuClust = "sum(rate(container_cpu_usage_seconds_total%7Bid%3D%22%2F%22%7D%5B1m%5D))%2Fsum(machine_cpu_cores)*100"
var promQMemClust = "sum(container_memory_working_set_bytes%7Bid%3D%22%2F%22%7D)%2Fsum(machine_memory_bytes)*100"
var promQDiskClust = "sum(container_fs_usage_bytes%7Bdevice%3D~%22%5E%2Fdev%2F%5Bsv%5Dd%5Ba-z%5D%5B1-9%5D%24%22%2Cid%3D%22%2F%22%7D)%2Fsum(container_fs_limit_bytes%7Bdevice%3D~%22%5E%2Fdev%2F%5Bsv%5Dd%5Ba-z%5D%5B1-9%5D%24%22%2Cid%3D%22%2F%22%7D)*100"
var promQSendBytesRateClust = "sum(irate(container_network_transmit_bytes_total%5B1m%5D))"
var promQRecvBytesRateClust = "sum(irate(container_network_receive_bytes_total%5B1m%5D))"
var promQTcpConnClust = "node_netstat_Tcp_CurrEstab"
var promQTcpRetransClust = "node_netstat_Tcp_RetransSegs"
var promQUdpSendPktsClust = "node_netstat_Udp_OutDatagrams"
var promQUdpRecvPktsClust = "node_netstat_Udp_InDatagrams"
var promQUdpRecvErr = "node_netstat_Udp_InErrors"

var promQCpuPod = "sum(rate(container_cpu_usage_seconds_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"
var promQMemPod = "sum(container_memory_working_set_bytes%7Bimage!%3D%22%22%7D)by(pod_name)"
var promQDiskPod = "sum(container_fs_usage_bytes%7Bimage!%3D%22%22%7D)by(pod_name)"
var promQNetRecvRate = "sum(irate(container_network_receive_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"
var promQNetSendRate = "sum(irate(container_network_transmit_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"

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
	l, err := net.Listen("tcp", "127.0.0.1:9090")
	//only one fakeProm will be up at a time, so basically its whichever shepherd creates it first is the one that will get tested
	if err != nil {
		//if it fails, do nothing and return
		fmt.Printf("failed to initialize listener on port 9090, assuming there is already a fake prometheus server up")
		return
	}
	fakeProm := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(e2eTestData[r.URL.String()[skiplen:]]))
	}))
	//assign the listener for port 9090
	fakeProm.Listener.Close()
	fakeProm.Listener = l
	fakeProm.Start()
	defer fakeProm.Close()

	//loop forever
	for true {
	}
}

func checkInflux() (bool, error) {
	//start an influx client so we can read from it
	influxAddr := "127.0.0.1:8086" //maybe somehow pull this from "local_multi.yml" so its not hardcoded
	influxAuth := &cloudcommon.InfluxCreds{}
	influxQ := influxq.NewInfluxQ(cloudcommon.DeveloperMetricsDbName, influxAuth.User, influxAuth.Pass)
	err := influxQ.Start(influxAddr, "")
	if err != nil {
		return false, fmt.Errorf("Failed to start influx queue address %s, %v", influxAddr, err)
	}

	//check all the measurements
	var result []client.Result
	result, err = influxQ.QueryDB("select \"cpu\" from \"cluster-cpu\" LIMIT 1")
	if err != nil {
		return false, err
	}
	fmt.Printf("asdf %v\n", result[0].Series[0].Values[0])

	result, err = influxQ.QueryDB("select \"mem\" from \"cluster-mem\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"disk\" from \"cluster-disk\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"sendBytes\",\"recvBytes\" from \"cluster-network\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"tcpConns\",\"tcpRetrans\" from \"cluster-tcp\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"udpSend\",\"udpRecv\",\"udpRecvErr\" from \"cluster-udp\" LIMIT 1")

	result, err = influxQ.QueryDB("select \"cpu\" from \"appinst-cpu\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"mem\" from \"appinst-mem\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"disk\" from \"appinst-disk\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"sendBytes\",\"recvBytes\" from \"appinst-network\" LIMIT 1")

	return true, nil
}
