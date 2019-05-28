package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

var influxdb = flag.String("influxdb", "0.0.0.0:8086", "InfluxDB address to export to")
var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var tlsCertFile = flag.String("tls", "", "server9 tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var collectInterval = flag.Duration("interval", time.Second*15, "Metrics collection interval")

//do i need these???
var operatorName = flag.String("operator", "local", "Cloudlet Operator Name")
var cloudletName = flag.String("cloudlet", "local", "Cloudlet Name")
var clusterName = flag.String("cluster", "myclust", "Cluster Name")

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
var promQNetRecvRate = "sum(irate(container_network_receive_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"
var promQNetSendRate = "sum(irate(container_network_transmit_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod_name)"

//map keeping track of all the currently running prometheuses
//TODO: figure out exactly what the types need to be
var promMap map[string]*PromStats

var MEXPrometheusAppName = "MEXPrometheusAppName"

var Env = map[string]string{
	"INFLUXDB_USER": "root",
	"INFLUXDB_PASS": "root",
}

var AppInstCache edgeproto.AppInstCache

var InfluxDBName = "clusterstats"
var influxQ *influxq.InfluxQ

var sigChan chan os.Signal

func getIPfromEnv() (string, error) {
	re := regexp.MustCompile(".*PROMETHEUS_PORT_9090_TCP_ADDR=(.*)")
	for _, e := range os.Environ() {
		match := re.FindStringSubmatch(e)
		if len(match) > 1 {
			return match[1], nil
		}
	}
	return "", errors.New("No Prometheus is running")
}

func initEnv() {
	val := os.Getenv("MEX_OPERATOR_NAME")
	if val != "" {
		*operatorName = val
	}
	val = os.Getenv("MEX_CLOUDLET_NAME")
	if val != "" {
		*cloudletName = val
	}
	val = os.Getenv("MEX_CLUSTER_NAME")
	if val != "" {
		*clusterName = val
	}
	val = os.Getenv("MEX_INFLUXDB_ADDR")
	if val != "" {
		*influxdb = val
	}
	val = os.Getenv("MEX_INFLUXDB_USER")
	if val != "" {
		Env["INFLUXDB_USER"] = val
	}
	val = os.Getenv("MEX_INFLUXDB_PASS")
	if val != "" {
		Env["INFLUXDB_PASS"] = val
	}
	val = os.Getenv("MEX_SCRAPE_INTERVAL")
	if val != "" {
		tmp, err := time.ParseDuration(val)
		if err == nil {
			*collectInterval = tmp
		}
	}
}

func appInstCb(key *edgeproto.AppInstKey, old *edgeproto.AppInst) {
	//check for prometheus
	if key.AppKey.Name != MEXPrometheusAppName {
		return
	}
	info := edgeproto.AppInst{}
	found := AppInstCache.Get(key, &info)
	if !found {
		return
	}
	var mapKey = key.ClusterInstKey.ClusterKey.Name
	stats, exists := promMap[mapKey]
	//maybe need to do more than just check for ready
	if info.State == edgeproto.TrackedState_READY {
		fmt.Printf("New Prometheus instance detected in cluster: %s\n", mapKey)
		//get address of prometheus.
		//for now while testing in dind this is ok
		clustIP := "localhost"
		port := info.MappedPorts[0].PublicPort
		promAddress := fmt.Sprintf("%s:%d", clustIP, port)
		//should probably also check to see if it already exists before throwing it in the map
		if !exists {
			stats = NewPromStats(promAddress, *collectInterval, sendMetric)
			promMap[mapKey] = stats
			stats.Start()
		} else {
			fmt.Printf("somethings wrong\n")
		}
	} else { //if its anything other than ready just stop it
		//try to remove it from the prommap
		if exists {
			delete(promMap, mapKey)
			stats.Stop()
		}
	}
}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	initEnv() //leftover from metrics main.go, figure out why this was in there

	fmt.Printf("InfluxDB is at: %s\n", *influxdb)
	fmt.Printf("Metrics collection interval is %s\n", *collectInterval)
	influxQ = influxq.NewInfluxQ(InfluxDBName)
	err := influxQ.Start(*influxdb)
	if err != nil {
		log.FatalLog("Failed to start influx queue",
			"address", *influxdb, "err", err)
	}
	defer influxQ.Stop()

	promMap = make(map[string]*PromStats)

	//register thresher to receive appinst notifications from crm
	edgeproto.InitAppInstCache(&AppInstCache)
	AppInstCache.SetNotifyCb(appInstCb)
	//then init notify, (look at crm/main line 108)
	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient := notify.NewClient(addrs, *tlsCertFile)
	notifyClient.RegisterRecvAppInstCache(&AppInstCache)
	notifyClient.Start()
	defer notifyClient.Stop()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	DebugPrint("Ready\n")

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}

func DebugPrint(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func sendMetric(metric *edgeproto.Metric) {
	influxQ.AddMetric(metric)
}
