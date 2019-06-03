package main

import (
	//"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	//"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/platform"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/platform/dind"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/platform/openstack"
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
var platformName = flag.String("platform", "", "Platform type of Cloudlet")

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

var Env = map[string]string{
	"INFLUXDB_USER": "root",
	"INFLUXDB_PASS": "root",
}

//map keeping track of all the currently running prometheuses
var promMap map[string]*PromStats
var MEXPrometheusAppName = "MEXPrometheusAppName"
var AppInstCache edgeproto.AppInstCache

//is there a better way to get clusterinsts than having this cache
var ClusterInstCache edgeproto.ClusterInstCache
var pf platform.Platform

var InfluxDBName = "clusterstats"
var influxQ *influxq.InfluxQ

var sigChan chan os.Signal

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
		clusterInst := edgeproto.ClusterInst{}
		found := ClusterInstCache.Get(&key.ClusterInstKey, &clusterInst)
		if !found {
			fmt.Printf("Unable to find clusterInst for prometheus\n")
			return
		}
		clustIP, _ := pf.GetClusterIP(&clusterInst)
		port := info.MappedPorts[0].PublicPort
		promAddress := fmt.Sprintf("%s:%d", clustIP, port)
		fmt.Printf("prometheus found at: %s\n", promAddress)
		if !exists {
			stats = NewPromStats(promAddress, *collectInterval, sendMetric, key.ClusterInstKey)
			promMap[mapKey] = stats
			stats.Start()
		} else { //somehow this cluster's prometheus was already registered
			fmt.Printf("Error, Prometheus app already registered for this cluster\n")
		}
	} else { //if its anything other than ready just stop it
		//try to remove it from the prommap
		if exists {
			delete(promMap, mapKey)
			stats.Stop()
		}
	}
}

func getPlatform() (platform.Platform, error) {
	var plat platform.Platform
	var err error
	switch *platformName {
	case "dind":
		plat = &dind.Platform{}
	case "openstack":
		plat = &openstack.Platform{}
	default:
		err = fmt.Errorf("No recognizable platform provided. Supported platforms are dind, openstack, and azure\n")
	}
	return plat, err
}

//openstack appinst go line 44
func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)

	fmt.Printf("InfluxDB is at: %s\n", *influxdb)
	fmt.Printf("Metrics collection interval is %s\n", *collectInterval)
	var err error
	pf, err = getPlatform()
	if err != nil {
		log.FatalLog("Failed to get platform", "platformName", platformName, "err", err)
	}
	influxQ = influxq.NewInfluxQ(InfluxDBName)
	err = influxQ.Start(*influxdb)
	if err != nil {
		log.FatalLog("Failed to start influx queue",
			"address", *influxdb, "err", err)
	}
	defer influxQ.Stop()

	promMap = make(map[string]*PromStats)

	//register thresher to receive appinst notifications from crm
	edgeproto.InitAppInstCache(&AppInstCache)
	AppInstCache.SetNotifyCb(appInstCb)
	edgeproto.InitClusterInstCache(&ClusterInstCache)
	//then init notify, (look at crm/main line 108)
	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient := notify.NewClient(addrs, *tlsCertFile)
	notifyClient.RegisterRecvAppInstCache(&AppInstCache)
	notifyClient.RegisterRecvClusterInstCache(&ClusterInstCache)
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
