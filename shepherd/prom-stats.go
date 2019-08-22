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
	platform "github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type PromResp struct {
	Status string   `json:"status,omitempty"`
	Data   PromData `json:"data,omitempty"`
}
type PromData struct {
	ResType string       `json:"resultType,omitempty"`
	Result  []PromMetric `json:"result,omitempty"`
}
type PromMetric struct {
	Labels PromLables    `json:"metric,omitempty"`
	Values []interface{} `json:"value,omitempty"`
}
type PromLables struct {
	PodName string `json:"pod_name,omitempty"`
}

const platformClientHeaderSize = 3

func NewClusterWorker(promAddr string, interval time.Duration, send func(ctx context.Context, metric *edgeproto.Metric) bool, clusterInst *edgeproto.ClusterInst, pf platform.Platform) (*ClusterWorker, error) {
	var err error
	p := ClusterWorker{}
	p.promAddr = promAddr
	p.appStatsMap = make(map[MetricAppInstKey]*AppMetrics)
	p.clusterStat = &ClusterMetrics{}
	p.interval = interval
	p.send = send
	p.clusterInstKey = clusterInst.Key
	p.client, err = pf.GetPlatformClient(clusterInst)
	if err != nil {
		// If we cannot get a platform client no point in trying to get metrics
		log.DebugLog(log.DebugLevelMetrics, "Failed to acquire platform client", "cluster", clusterInst.Key, "error", err)
		return nil, err
	}
	return &p, nil
}

//trims the output from the pc.PlatformClient.Output request so that to get rid of the header stuff tacked on by it
func outputTrim(output string) string {
	lines := strings.SplitN(output, "\n", platformClientHeaderSize+1)
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

func getPromMetrics(addr string, query string, client pc.PlatformClient) (*PromResp, error) {
	reqURI := "'http://" + addr + "/api/v1/query?query=" + query + "'"
	resp, err := client.Output("curl " + reqURI)
	if err != nil {
		errstr := fmt.Sprintf("Failed to run <%s>", reqURI)
		log.DebugLog(log.DebugLevelMetrics, errstr, "err", err.Error())
		return nil, err
	}
	trimmedResp := outputTrim(resp)
	promResp := &PromResp{}
	if err = json.Unmarshal([]byte(trimmedResp), promResp); err != nil {
		return nil, err
	}
	return promResp, nil
}

//this takes a float64 representation of a time(in sec) given to use by prometheus
//and turns it into a type.Timestamp format for writing into influxDB
func parseTime(timeFloat float64) *types.Timestamp {
	sec, dec := math.Modf(timeFloat)
	time := time.Unix(int64(sec), int64(dec*(1e9)))
	ts, _ := types.TimestampProto(time)
	return ts
}

func (p *ClusterWorker) CollectPromStats() error {
	appKey := MetricAppInstKey{
		operator:  p.clusterInstKey.CloudletKey.OperatorKey.Name,
		cloudlet:  p.clusterInstKey.CloudletKey.Name,
		cluster:   p.clusterInstKey.ClusterKey.Name,
		developer: p.clusterInstKey.Developer,
	}
	// Get Pod CPU usage percentage
	resp, err := getPromMetrics(p.promAddr, promQCpuPod, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &AppMetrics{}
				p.appStatsMap[appKey] = stat
			}
			stat.cpuTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.cpu = val
			}
		}
	}
	// Get Pod Mem usage
	resp, err = getPromMetrics(p.promAddr, promQMemPod, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &AppMetrics{}
				p.appStatsMap[appKey] = stat
			}
			stat.memTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				stat.mem = val
			}
		}
	}
	// Get Pod Disk usage
	resp, err = getPromMetrics(p.promAddr, promQDiskPod, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &AppMetrics{}
				p.appStatsMap[appKey] = stat
			}
			stat.diskTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				stat.disk = val
			}
		}
	}
	// Get Pod NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQNetRecvRate, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &AppMetrics{}
				p.appStatsMap[appKey] = stat
			}
			stat.netRecvTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.netRecv = uint64(val)
			}
		}
	}
	// Get Pod NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQNetSendRate, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			appKey.pod = metric.Labels.PodName
			stat, found := p.appStatsMap[appKey]
			if !found {
				stat = &AppMetrics{}
				p.appStatsMap[appKey] = stat
			}
			//copy only if we can parse the value
			stat.netSendTS = parseTime(metric.Values[0].(float64))
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				stat.netSend = uint64(val)
			}
		}
	}

	// Get Cluster CPU usage
	resp, err = getPromMetrics(p.promAddr, promQCpuClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.cpuTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.cpu = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster Mem usage
	resp, err = getPromMetrics(p.promAddr, promQMemClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.memTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.mem = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster Disk usage percentage
	resp, err = getPromMetrics(p.promAddr, promQDiskClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.diskTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.disk = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster NetRecv bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQRecvBytesRateClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.netRecvTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.netRecv = uint64(val)
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster NetSend bytes rate averaged over 1m
	resp, err = getPromMetrics(p.promAddr, promQSendBytesRateClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.netSendTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseFloat(metric.Values[1].(string), 64); err == nil {
				p.clusterStat.netSend = uint64(val)
				// We should have only one value here
				break
			}
		}
	}

	// Get Cluster Established TCP connections
	resp, err = getPromMetrics(p.promAddr, promQTcpConnClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.tcpConnsTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.tcpConns = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster TCP retransmissions
	resp, err = getPromMetrics(p.promAddr, promQTcpRetransClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.tcpRetransTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.tcpRetrans = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Send Datagrams
	resp, err = getPromMetrics(p.promAddr, promQUdpSendPktsClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.udpSendTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.udpSend = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Recv Datagrams
	resp, err = getPromMetrics(p.promAddr, promQUdpRecvPktsClust, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.udpRecvTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.udpRecv = val
				// We should have only one value here
				break
			}
		}
	}
	// Get Cluster UDP Recv Errors
	resp, err = getPromMetrics(p.promAddr, promQUdpRecvErr, p.client)
	if err == nil && resp.Status == "success" {
		for _, metric := range resp.Data.Result {
			p.clusterStat.udpRecvErrTS = parseTime(metric.Values[0].(float64))
			//copy only if we can parse the value
			if val, err := strconv.ParseUint(metric.Values[1].(string), 10, 64); err == nil {
				p.clusterStat.udpRecvErr = val
				// We should have only one value here
				break
			}
		}
	}

	return nil
}

func (p *ClusterWorker) Start() {
	p.stop = make(chan struct{})
	p.waitGrp.Add(1)
	go p.RunNotify()
}

func (p *ClusterWorker) Stop() {
	log.DebugLog(log.DebugLevelMetrics, "Stopping ClusterWorker thread\n")
	close(p.stop)
	p.waitGrp.Wait()
}

func (p *ClusterWorker) RunNotify() {
	log.DebugLog(log.DebugLevelMetrics, "Started ClusterWorker thread\n")
	done := false
	for !done {
		select {
		case <-time.After(p.interval):
			if p.CollectPromStats() != nil {
				continue
			}
			span := log.StartSpan(log.DebugLevelSampled, "send-metric")
			span.SetTag("operator", p.clusterInstKey.CloudletKey.OperatorKey.Name)
			span.SetTag("cloudlet", p.clusterInstKey.CloudletKey.Name)
			span.SetTag("cluster", p.clusterInstKey.ClusterKey.Name)
			ctx := log.ContextWithSpan(context.Background(), span)

			for key, stat := range p.appStatsMap {
				appMetrics := PodStatToMetrics(&key, stat)
				for _, metric := range appMetrics {
					p.send(ctx, metric)
				}
			}
			clusterMetrics := ClusterStatToMetrics(p)
			for _, metric := range clusterMetrics {
				p.send(ctx, metric)
			}
			span.Finish()
		case <-p.stop:
			done = true
		}
	}
	p.waitGrp.Done()
}

func newMetric(clusterInstKey edgeproto.ClusterInstKey, name string, key *MetricAppInstKey, ts *types.Timestamp) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Name = name
	metric.Timestamp = *ts
	metric.AddTag("operator", clusterInstKey.CloudletKey.OperatorKey.Name)
	metric.AddTag("cloudlet", clusterInstKey.CloudletKey.Name)
	metric.AddTag("cluster", clusterInstKey.ClusterKey.Name)
	metric.AddTag("dev", clusterInstKey.Developer)
	if key != nil {
		metric.AddTag("app", key.pod)
	}
	return &metric
}

func ClusterStatToMetrics(p *ClusterWorker) []*edgeproto.Metric {
	var metrics []*edgeproto.Metric
	var metric *edgeproto.Metric

	//nil timestamps mean the curl request failed. So do not write the metric in
	if p.clusterStat.cpuTS != nil {
		metric = newMetric(p.clusterInstKey, "cluster-cpu", nil, p.clusterStat.cpuTS)
		metric.AddDoubleVal("cpu", p.clusterStat.cpu)
		metrics = append(metrics, metric)
		//reset to nil for the next collection
		p.clusterStat.cpuTS = nil
	}

	if p.clusterStat.memTS != nil {
		metric = newMetric(p.clusterInstKey, "cluster-mem", nil, p.clusterStat.memTS)
		metric.AddDoubleVal("mem", p.clusterStat.mem)
		metrics = append(metrics, metric)
		p.clusterStat.memTS = nil
	}

	if p.clusterStat.diskTS != nil {
		metric = newMetric(p.clusterInstKey, "cluster-disk", nil, p.clusterStat.diskTS)
		metric.AddDoubleVal("disk", p.clusterStat.disk)
		metrics = append(metrics, metric)
		p.clusterStat.diskTS = nil
	}

	if p.clusterStat.netSendTS != nil && p.clusterStat.netRecvTS != nil {
		//for measurements with multiple values just pick one timestamp to use
		metric = newMetric(p.clusterInstKey, "cluster-network", nil, p.clusterStat.netSendTS)
		metric.AddIntVal("sendBytes", p.clusterStat.netSend)
		metric.AddIntVal("recvBytes", p.clusterStat.netRecv)
		metrics = append(metrics, metric)
	}
	p.clusterStat.netSendTS = nil
	p.clusterStat.netRecvTS = nil

	if p.clusterStat.tcpConnsTS != nil && p.clusterStat.tcpRetransTS != nil {
		metric = newMetric(p.clusterInstKey, "cluster-tcp", nil, p.clusterStat.tcpConnsTS)
		metric.AddIntVal("tcpConns", p.clusterStat.tcpConns)
		metric.AddIntVal("tcpRetrans", p.clusterStat.tcpRetrans)
		metrics = append(metrics, metric)
	}
	p.clusterStat.netSendTS = nil
	p.clusterStat.netRecvTS = nil

	if p.clusterStat.udpSendTS != nil && p.clusterStat.udpRecvTS != nil && p.clusterStat.udpRecvErrTS != nil {
		metric = newMetric(p.clusterInstKey, "cluster-udp", nil, p.clusterStat.udpSendTS)
		metric.AddIntVal("udpSend", p.clusterStat.udpSend)
		metric.AddIntVal("udpRecv", p.clusterStat.udpRecv)
		metric.AddIntVal("udpRecvErr", p.clusterStat.udpRecvErr)
		metrics = append(metrics, metric)
	}
	p.clusterStat.udpSendTS = nil
	p.clusterStat.udpRecvTS = nil
	p.clusterStat.udpRecvErrTS = nil

	return metrics
}

func PodStatToMetrics(key *MetricAppInstKey, stat *AppMetrics) []*edgeproto.Metric {
	var metrics []*edgeproto.Metric
	var metric *edgeproto.Metric

	clusterInstKey := edgeproto.ClusterInstKey{
		ClusterKey: edgeproto.ClusterKey{
			Name: key.cluster,
		},
		CloudletKey: edgeproto.CloudletKey{
			OperatorKey: edgeproto.OperatorKey{
				Name: key.operator,
			},
			Name: key.cloudlet,
		},
		Developer: key.developer,
	}

	if stat.cpuTS != nil {
		metric = newMetric(clusterInstKey, "appinst-cpu", key, stat.cpuTS)
		metric.AddDoubleVal("cpu", stat.cpu)
		metrics = append(metrics, metric)
		stat.cpuTS = nil
	}

	if stat.memTS != nil {
		metric = newMetric(clusterInstKey, "appinst-mem", key, stat.memTS)
		metric.AddIntVal("mem", stat.mem)
		metrics = append(metrics, metric)
		stat.memTS = nil
	}

	if stat.diskTS != nil {
		metric = newMetric(clusterInstKey, "appinst-disk", key, stat.diskTS)
		metric.AddIntVal("disk", stat.disk)
		metrics = append(metrics, metric)
		stat.diskTS = nil
	}

	if stat.netSendTS != nil && stat.netRecvTS != nil {
		//for measurements with multiple values just pick one timestamp to use
		metric = newMetric(clusterInstKey, "appinst-network", key, stat.netSendTS)
		metric.AddIntVal("sendBytes", stat.netSend)
		metric.AddIntVal("recvBytes", stat.netRecv)
		metrics = append(metrics, metric)
	}
	stat.netSendTS = nil
	stat.netRecvTS = nil

	return metrics
}
