package main

import (
	"fmt"
	// 	"io/ioutil"
	"net/http"
	// 	"gopkg.in/yaml.v2"
	// 	"github.com/mobiledgex/edge-cloud/log"
	// 	"github.com/prometheus/client_golang/prometheus"
	// 	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ExporterStatsCollector struct {
	Prom Prometheus `yaml:"prometheus"`
}

type Prometheus struct {
	TcpConn    int `yaml:"tcpConn"`
	TcpRetrans int `yaml:"tcpRetrans"`
	UdpSent    int `yaml:"udpSent"`
	UdpRecv    int `yaml:"udpRecv"`
	UdpRecvErr int `yaml:"udpRecvErr"`
}

// func (es ExporterStatsCollector) Describe(ch chan<- *prometheus.Desc) {
// 	prometheus.DescribeByCollect(es, ch)
// }

// func (es ExporterStatsCollector) Collect(ch chan<- prometheus.Metric) {
// 	ch <- prometheus.MustNewConstMetric(tcpConnDesc, prometheus.UntypedValue, float64(es.Prom.TcpConn))
// 	ch <- prometheus.MustNewConstMetric(tcpRetrans, prometheus.UntypedValue, float64(es.Prom.TcpRetrans))
// 	ch <- prometheus.MustNewConstMetric(udpSent, prometheus.UntypedValue, float64(es.Prom.UdpSent))
// 	ch <- prometheus.MustNewConstMetric(udpRecv, prometheus.UntypedValue, float64(es.Prom.UdpRecv))
// 	ch <- prometheus.MustNewConstMetric(udpRecvErr, prometheus.UntypedValue, float64(es.Prom.UdpRecvErr))
// }

// var (
// 	tcpConnDesc = prometheus.NewDesc("node_netstat_Tcp_CurrEstab", "mimicking the TcpConns stat", nil, nil)
// 	tcpRetrans  = prometheus.NewDesc("node_netstat_Tcp_RetransSegs", "mimicking the TcpRetrans stat", nil, nil)
// 	udpSent     = prometheus.NewDesc("node_netstat_Udp_OutDatagrams", "mimicking the UdpSent stat", nil, nil)
// 	udpRecv     = prometheus.NewDesc("node_netstat_Udp_InDatagrams", "mimicking the UdpRecv stat", nil, nil)
// 	udpRecvErr  = prometheus.NewDesc("node_netstat_Udp_InErrors", "mimicking the UdpRecvErr stat", nil, nil)
// )

// //read in the fake values we want to use
// //prolly should move this and the structs to their own file later when i include other stuff besides prom
// func GetValuesFromYaml(stats *ExporterStatsCollector, ymlPath string) error {
// 	yamlFile, err := ioutil.ReadFile(ymlPath)
// 	if err != nil {
// 		log.FatalLog("Failed to load exporter yml file", "filename", ymlPath)
// 	}
// 	err = yaml.Unmarshal(yamlFile, stats)
// 	if err != nil {
// 		log.FatalLog("Failed to parse exporter yml file", "filename", ymlPath)
// 	}
// 	return nil
// }

func main() {
	// reg := prometheus.NewPedanticRegistry()
	// stats := ExporterStatsCollector{}
	// GetValuesFromYaml(&stats, "/Users/matthewchu/go/src/github.com/mobiledgex/edge-cloud-infra/shepherd/fakePromEndpoint/fakeStats.yml")
	// prometheus.WrapRegistererWith(nil, reg).MustRegister(stats)

	// http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/metrics", tempExporter)
	http.ListenAndServe(":9100", nil)
}

// temporary for now until i figure out how to export netstat stuff without a trailing ".0"
func tempExporter(w http.ResponseWriter, r *http.Request) {
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
`
