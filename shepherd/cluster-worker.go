package main

import (
	"context"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	platform "github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// For each cluster the notify worker is created
type ClusterWorker struct {
	clusterInstKey edgeproto.ClusterInstKey
	deployment     string
	promAddr       string
	interval       time.Duration
	appStatsMap    map[MetricAppInstKey]*AppMetrics
	clusterStat    *ClusterMetrics
	send           func(ctx context.Context, metric *edgeproto.Metric) bool
	waitGrp        sync.WaitGroup
	stop           chan struct{}
	client         pc.PlatformClient
}

func NewClusterWorker(promAddr string, interval time.Duration, send func(ctx context.Context, metric *edgeproto.Metric) bool, clusterInst *edgeproto.ClusterInst, pf platform.Platform) (*ClusterWorker, error) {
	var err error
	p := ClusterWorker{}
	p.promAddr = promAddr
	p.appStatsMap = make(map[MetricAppInstKey]*AppMetrics)
	p.clusterStat = &ClusterMetrics{}
	p.interval = interval
	p.send = send
	p.clusterInstKey = clusterInst.Key
	p.deployment = clusterInst.Deployment
	p.client, err = pf.GetPlatformClient(clusterInst)
	if err != nil {
		// If we cannot get a platform client no point in trying to get metrics
		log.DebugLog(log.DebugLevelMetrics, "Failed to acquire platform client", "cluster", clusterInst.Key, "error", err)
		return nil, err
	}
	return &p, nil
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
			if collectClusterPormetheusMetrics(p) != nil {
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

	if stat.cpuTS != nil {
		metric = newMetric(key.clusterInstKey, "appinst-cpu", key, stat.cpuTS)
		metric.AddDoubleVal("cpu", stat.cpu)
		metrics = append(metrics, metric)
		stat.cpuTS = nil
	}

	if stat.memTS != nil {
		metric = newMetric(key.clusterInstKey, "appinst-mem", key, stat.memTS)
		metric.AddIntVal("mem", stat.mem)
		metrics = append(metrics, metric)
		stat.memTS = nil
	}

	if stat.diskTS != nil {
		metric = newMetric(key.clusterInstKey, "appinst-disk", key, stat.diskTS)
		metric.AddIntVal("disk", stat.disk)
		metrics = append(metrics, metric)
		stat.diskTS = nil
	}

	if stat.netSendTS != nil && stat.netRecvTS != nil {
		//for measurements with multiple values just pick one timestamp to use
		metric = newMetric(key.clusterInstKey, "appinst-network", key, stat.netSendTS)
		metric.AddIntVal("sendBytes", stat.netSend)
		metric.AddIntVal("recvBytes", stat.netRecv)
		metrics = append(metrics, metric)
	}
	stat.netSendTS = nil
	stat.netRecvTS = nil

	return metrics
}
