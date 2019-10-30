//package main
package fakepromexporter

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mobiledgex/edge-cloud/log"
	"gopkg.in/yaml.v2"
)

type ExporterStatsCollector struct {
	Prom Prometheus `yaml:"prometheus"`
}

type Prometheus struct {
	ClustCpu   int `yaml:"clustCpu"`
	TcpConn    int `yaml:"tcpConn"`
	TcpRetrans int `yaml:"tcpRetrans"`
	UdpSent    int `yaml:"udpSent"`
	UdpRecv    int `yaml:"udpRecv"`
	UdpRecvErr int `yaml:"udpRecvErr"`
}

// read in the fake values we want to use
// prolly should move this and the structs to their own file later when i include other stuff besides prom
func GetValuesFromYaml(stats *ExporterStatsCollector, ymlPath string) error {
	yamlFile, err := ioutil.ReadFile(ymlPath)
	if err != nil {
		log.FatalLog("Failed to load exporter yml file", "filename", ymlPath, "error", err)
	}
	err = yaml.Unmarshal(yamlFile, stats)
	if err != nil {
		log.FatalLog("Failed to parse exporter yml file", "filename", ymlPath, "error", err)
	}
	return nil
}

func StartExporter(ctx context.Context) {
	// stats := ExporterStatsCollector{}
	// GetValuesFromYaml(&stats, "/Users/matthewchu/go/src/github.com/mobiledgex/edge-cloud-infra/shepherd/fakePromExporter/fakeStats.yml")

	log.SpanLog(ctx, log.DebugLevelMexos, "Starting fake prometheus exporter...")
	http.HandleFunc("/metrics", exporter)
	http.ListenAndServe(":9100", nil)
}

func exporter(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, exportStr)
}

var exportStr = `# HELP node_netstat_Tcp_CurrEstab mimicking the TcpConns stat
# TYPE node_netstat_Tcp_CurrEstab untyped
node_netstat_Tcp_CurrEstab 100
# HELP node_netstat_Tcp_RetransSegs mimicking the TcpRetrans stat
# TYPE node_netstat_Tcp_RetransSegs untyped
node_netstat_Tcp_RetransSegs 200
# HELP node_netstat_Udp_InDatagrams mimicking the UdpRecv stat
# TYPE node_netstat_Udp_InDatagrams untyped
node_netstat_Udp_InDatagrams 400
# HELP node_netstat_Udp_InErrors mimicking the UdpRecvErr stat
# TYPE node_netstat_Udp_InErrors untyped
node_netstat_Udp_InErrors 500
# HELP node_netstat_Udp_OutDatagrams mimicking the UdpSent stat
# TYPE node_netstat_Udp_OutDatagrams untyped
node_netstat_Udp_OutDatagrams 300
# HELP container_cpu_usage_seconds_total Cumulative cpu time consumed in seconds. For mimicking the cpu metrics
# TYPE container_cpu_usage_seconds_total counter
container_cpu_usage_seconds_total{id="/",image="",pod_name=""} 5598.558165506 
container_cpu_usage_seconds_total{id="idNameThatsNotJustAForwardSlash",image="anythingButAnEmptyString",pod_name="someapplication1"} 142.443278585
# HELP machine_cpu_cores Number of CPU cores on the machine. For mimicking cluster-cpu
# TYPE machine_cpu_cores gauge
machine_cpu_cores 6
# HELP container_memory_working_set_bytes Current working set in bytes. For mimicking the mem metrics
# TYPE container_memory_working_set_bytes gauge
container_memory_working_set_bytes{id="/",image="",pod_name=""} 6.140334592e+09 
container_memory_working_set_bytes{id="idNameThatsNotJustAForwardSlash",image="anythingButAnEmptyString",pod_name="someapplication1"} 3.756834298e+09
# HELP machine_memory_bytes Amount of memory installed on the machine. For mimicking cluster-mem
# TYPE machine_memory_bytes gauge
machine_memory_bytes 8.360390656e+09
# HELP container_fs_usage_bytes Number of bytes that are consumed by the container on this filesystem. For mimicking the disk metrics
# TYPE container_fs_usage_bytes gauge
container_fs_usage_bytes{device="/dev/sda1",id="/",image="",pod_name=""} 1.049339904e+10
container_fs_usage_bytes{device="",id="notAForwardSlash",image="notTheEmptyString",pod_name="someapplication1"} 73728
# HELP container_fs_limit_bytes Number of bytes that can be consumed by the container on this filesystem. For mimicking cluster-disk
# TYPE container_fs_limit_bytes gauge
container_fs_limit_bytes{device="/dev/sda1",id="/"} 6.2725623808e+10
# HELP container_network_transmit_bytes_total Cumulative count of bytes transmitted. For mimicking the network stats
# TYPE container_network_transmit_bytes_total counter
container_network_transmit_bytes_total{image="notTheEmptyString",pod_name="someapplication1"} 5.223803e+06
# HELP container_network_receive_bytes_total Cumulative count of bytes received. For mimicking the network stats
# TYPE container_network_receive_bytes_total counter
container_network_receive_bytes_total{image="notTheEmptyString",pod_name="someapplication1"} 4.506154e+06
`
