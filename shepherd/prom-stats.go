package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

var promQCpuClust = "sum(rate(container_cpu_usage_seconds_total%7Bid%3D%22%2F%22%7D%5B1m%5D))%2Fsum(machine_cpu_cores)*100"
var promQMemClust = "sum(container_memory_working_set_bytes%7Bid%3D%22%2F%22%7D)%2Fsum(machine_memory_bytes)*100"
var promQDiskClust = "sum(container_fs_usage_bytes%7Bdevice%3D~%22%5E%2Fdev%2F%5Bsv%5Dd%5Ba-z%5D%5B1-9%5D%24%22%2Cid%3D%22%2F%22%7D)%2Fsum(container_fs_limit_bytes%7Bdevice%3D~%22%5E%2Fdev%2F%5Bsv%5Dd%5Ba-z%5D%5B1-9%5D%24%22%2Cid%3D%22%2F%22%7D)*100"
var promQSentBytesRateClust = "sum(irate(container_network_transmit_bytes_total%5B1m%5D))"
var promQRecvBytesRateClust = "sum(irate(container_network_receive_bytes_total%5B1m%5D))"
var promQTcpConnClust = "node_netstat_Tcp_CurrEstab"
var promQTcpRetransClust = "node_netstat_Tcp_RetransSegs"
var promQUdpSentPktsClust = "node_netstat_Udp_OutDatagrams"
var promQUdpRecvPktsClust = "node_netstat_Udp_InDatagrams"
var promQUdpRecvErr = "node_netstat_Udp_InErrors"

var promQCpuPod = "sum(rate(container_cpu_usage_seconds_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod)"
var promQMemPod = "sum(container_memory_working_set_bytes%7Bimage!%3D%22%22%7D)by(pod)"
var promQDiskPod = "sum(container_fs_usage_bytes%7Bimage!%3D%22%22%7D)by(pod)"
var promQNetRecvRate = "sum(irate(container_network_receive_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod)"
var promQNetSentRate = "sum(irate(container_network_transmit_bytes_total%7Bimage!%3D%22%22%7D%5B1m%5D))by(pod)"

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
}
type PromAlert struct {
	Labels      map[string]string
	Annotations map[string]string
	State       string
	ActiveAt    *time.Time `json:"activeAt,omitempty"`
	Value       PromValue
}

const platformClientHeaderSize = 3

//trims the output from the Client.PlatformClient.Output request so that to get rid of the header stuff tacked on by it
func outputTrim(output string) string {
	lines := strings.SplitN(output, "\n", platformClientHeaderSize+1)
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

func getPromMetrics(ctx context.Context, addr string, query string, client ssh.Client) (*PromResp, error) {
	reqURI := "'http://" + addr + "/api/v1/query?query=" + query + "'"
	resp, err := client.Output("curl " + reqURI)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to get prom metrics", "reqURI", reqURI, "err", err)
		return nil, err
	}
	trimmedResp := outputTrim(resp)
	promResp := &PromResp{}
	if err = json.Unmarshal([]byte(trimmedResp), promResp); err != nil {
		return nil, err
	}
	return promResp, nil
}

func getPromAlerts(ctx context.Context, addr string, client ssh.Client) ([]edgeproto.Alert, error) {
	reqURI := "'http://" + addr + "/api/v1/alerts'"
	out, err := client.Output("curl " + reqURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to run <%s>, %v", reqURI, err)
	}
	trimmedResp := outputTrim(out)
	resp := struct {
		Status string
		Data   struct {
			Alerts []PromAlert
		}
	}{}
	if err = json.Unmarshal([]byte(trimmedResp), &resp); err != nil {
		return nil, err
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("Resp to <%s> is %s instead of success", reqURI, resp.Status)
	}
	alerts := []edgeproto.Alert{}
	for _, pa := range resp.Data.Alerts {
		alert := edgeproto.Alert{}
		alert.Labels = pa.Labels
		alert.Annotations = pa.Annotations
		alert.State = pa.State
		alert.Value = float64(pa.Value)
		if pa.ActiveAt != nil {
			alert.ActiveAt = cloudcommon.TimeToTimestamp(*pa.ActiveAt)
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

//this takes a float64 representation of a time(in sec) given to use by prometheus
//and turns it into a type.Timestamp format for writing into influxDB
func parseTime(timeFloat float64) *types.Timestamp {
	sec, dec := math.Modf(timeFloat)
	time := time.Unix(int64(sec), int64(dec*(1e9)))
	ts, _ := types.TimestampProto(time)
	return ts
}

func collectAppPrometheusMetrics(ctx context.Context, p *K8sClusterStats) map[shepherd_common.MetricAppInstKey]*shepherd_common.AppMetrics {
	appStatsMap := make(map[shepherd_common.MetricAppInstKey]*shepherd_common.AppMetrics)

	appKey := shepherd_common.MetricAppInstKey{
		ClusterInstKey: p.key,
	}
	// Get Pod CPU usage percentage
	resp, err := getPromMetrics(ctx, p.promAddr, promQCpuPod, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.Pod = metric.Labels.PodName
			stat, found := appStatsMap[appKey]
			if !found {
				stat = &shepherd_common.AppMetrics{}
				appStatsMap[appKey] = stat
			}
			stat.CpuTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.Cpu = val
			}
		}
	}
	// Get Pod Mem usage
	resp, err = getPromMetrics(ctx, p.promAddr, promQMemPod, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.Pod = metric.Labels.PodName
			stat, found := appStatsMap[appKey]
			if !found {
				stat = &shepherd_common.AppMetrics{}
				appStatsMap[appKey] = stat
			}
			stat.MemTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				stat.Mem = val
			}
		}
	}
	// Get Pod Disk usage
	resp, err = getPromMetrics(ctx, p.promAddr, promQDiskPod, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.Pod = metric.Labels.PodName
			stat, found := appStatsMap[appKey]
			if !found {
				stat = &shepherd_common.AppMetrics{}
				appStatsMap[appKey] = stat
			}
			stat.DiskTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				stat.Disk = val
			}
		}
	}
	// Get Pod NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(ctx, p.promAddr, promQNetRecvRate, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.Pod = metric.Labels.PodName
			stat, found := appStatsMap[appKey]
			if !found {
				stat = &shepherd_common.AppMetrics{}
				appStatsMap[appKey] = stat
			}
			stat.NetRecvTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.NetRecv = uint64(val)
			}
		}
	}
	// Get Pod NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(ctx, p.promAddr, promQNetSentRate, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.Pod = metric.Labels.PodName
			stat, found := appStatsMap[appKey]
			if !found {
				stat = &shepherd_common.AppMetrics{}
				appStatsMap[appKey] = stat
			}
			//copy only if we can parse the value
			stat.NetSentTS = parseTime(metric.Values[0].(float64))
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.NetSent = uint64(val)
			}
		}
	}
	return appStatsMap
}

func collectClusterPrometheusMetrics(ctx context.Context, p *K8sClusterStats) error {
	// Get Cluster CPU usage
	resp, err := getPromMetrics(ctx, p.promAddr, promQCpuClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.CpuTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.Cpu = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster Mem usage
	resp, err = getPromMetrics(ctx, p.promAddr, promQMemClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.MemTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.Mem = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster Disk usage percentage
	resp, err = getPromMetrics(ctx, p.promAddr, promQDiskClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.DiskTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.Disk = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(ctx, p.promAddr, promQRecvBytesRateClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.NetRecvTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.NetRecv = uint64(val)
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster NetSent bytes rate averaged over 1m
	resp, err = getPromMetrics(ctx, p.promAddr, promQSentBytesRateClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.NetSentTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.NetSent = uint64(val)
				// We should have only one value here
				break
			}
		}
	}

	// Get Cluster Established TCP connections
	resp, err = getPromMetrics(ctx, p.promAddr, promQTcpConnClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.TcpConnsTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.TcpConns = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster TCP retransmissions
	resp, err = getPromMetrics(ctx, p.promAddr, promQTcpRetransClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.TcpRetransTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.TcpRetrans = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Sent Datagrams
	resp, err = getPromMetrics(ctx, p.promAddr, promQUdpSentPktsClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.UdpSentTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.UdpSent = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Recv Datagrams
	resp, err = getPromMetrics(ctx, p.promAddr, promQUdpRecvPktsClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.UdpRecvTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.UdpRecv = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Recv Errors
	resp, err = getPromMetrics(ctx, p.promAddr, promQUdpRecvErr, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.UdpRecvErrTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.UdpRecvErr = val
				// We should have only one value here
				break
			}
		}
	}
	return nil
}
