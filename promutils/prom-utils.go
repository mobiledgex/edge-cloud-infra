package promutils

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
)

var ClusterPrometheusAppLabel = "label_" + cloudcommon.MexAppNameLabel
var ClusterPrometheusAppVersionLabel = "label_" + cloudcommon.MexAppVersionLabel

var PromLabelsAllMobiledgeXApps = `{` + ClusterPrometheusAppLabel + `!=""}`

var PromQCpuClust = `sum(rate(container_cpu_usage_seconds_total{id="/"}[1m]))/sum(machine_cpu_cores)*100`
var PromQMemClust = `sum(container_memory_working_set_bytes{id="/"})/sum(machine_memory_bytes)*100`
var PromQDiskClust = `sum(container_fs_usage_bytes{device=~"^/dev/[sv]d[a-z][1-9]$",id="/"})/sum(container_fs_limit_bytes{device=~"^/dev/[sv]d[a-z][1-9]$",id="/"})*100`
var PromQSentBytesRateClust = `sum(irate(container_network_transmit_bytes_total[1m]))`
var PromQRecvBytesRateClust = `sum(irate(container_network_receive_bytes_total[1m]))`
var PromQTcpConnClust = "node_netstat_Tcp_CurrEstab"
var PromQTcpRetransClust = "node_netstat_Tcp_RetransSegs"
var PromQUdpSentPktsClust = "node_netstat_Udp_OutDatagrams"
var PromQUdpRecvPktsClust = "node_netstat_Udp_InDatagrams"
var PromQUdpRecvErr = "node_netstat_Udp_InErrors"

// This is a template which takes a pod query and adds instance label to it
var PromQAppLabelsWrapperFmt = "max(kube_pod_labels%s)by(label_mexAppName,label_mexAppVersion,pod)*on(pod)group_right(label_mexAppName,label_mexAppVersion)(%s)"

var PromQCpuPod = `sum(rate(container_cpu_usage_seconds_total{image!=""}[1m]))by(pod)`
var PromQMemPod = `sum(container_memory_working_set_bytes{image!=""})by(pod)`
var PromQDiskPod = `sum(container_fs_usage_bytes{image!=""})by(pod)`
var PromQNetRecvRate = `sum(irate(container_network_receive_bytes_total{image!=""}[1m]))by(pod)`
var PromQNetSentRate = `sum(irate(container_network_transmit_bytes_total{image!=""}[1m]))by(pod)`

var PromQAutoScaleCpuTotalU = "stabilized_max_total_worker_node_cpu_utilisation"
var PromQAutoScaleMemTotalU = "stabilized_max_total_worker_node_mem_utilisation"

type PromResp struct {
	Status string   `json:"status,omitempty"`
	Data   PromData `json:"data,omitempty"`
}
type PromData struct {
	ResType string       `json:"resultType,omitempty"`
	Result  []PromMetric `json:"result,omitempty"`
}
type PromMetric struct {
	Labels PromLabels    `json:"metric,omitempty"`
	Values []interface{} `json:"value,omitempty"`
}
type PromLabels struct {
	PodName string `json:"pod,omitempty"`
	App     string `json:"label_mexAppName,omitempty"`
	Version string `json:"label_mexAppVersion,omitempty"`
}
type PromAlert struct {
	Labels      map[string]string
	Annotations map[string]string
	State       string
	ActiveAt    *time.Time `json:"activeAt,omitempty"`
	Value       PromAlertValue
}

// Prometheus Alert Value may be a string or a numeric, depending on
// the version of the prometheus operator used. Handle either.
type PromAlertValue float64

func (s PromAlertValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(s))
}

func (s *PromAlertValue) UnmarshalJSON(b []byte) error {
	var val float64
	err := json.Unmarshal(b, &val)
	if err == nil {
		*s = PromAlertValue(val)
		return nil
	}
	var str string
	err = json.Unmarshal(b, &str)
	if err == nil {
		val, err = strconv.ParseFloat(str, 64)
		if err == nil {
			*s = PromAlertValue(val)
			return nil
		}
		return err
	}
	return err
}

// Takes a float64 representation of a time(in sec) given to use by prometheus
// and turns it into a type.Timestamp format for writing into influxDB
func ParseTime(timeFloat float64) *types.Timestamp {
	sec, dec := math.Modf(timeFloat)
	time := time.Unix(int64(sec), int64(dec*(1e9)))
	ts, _ := types.TimestampProto(time)
	return ts
}

// Returns a prometheus pod-based query joined with k8s labels series
// Function also takes an optional label filter string of form "{label1="val1",label2="val2",..}"
func GetPromQueryWithK8sLabels(labelFilter, podQuery string) string {
	return fmt.Sprintf(PromQAppLabelsWrapperFmt, labelFilter, podQuery)
}
