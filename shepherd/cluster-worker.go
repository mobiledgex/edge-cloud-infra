package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	platform "github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

// For each cluster the notify worker is created
type ClusterWorker struct {
	clusterInstKey edgeproto.ClusterInstKey
	deployment     string
	promAddr       string
	interval       time.Duration
	clusterStat    shepherd_common.ClusterStats
	send           func(ctx context.Context, metric *edgeproto.Metric) bool
	waitGrp        sync.WaitGroup
	stop           chan struct{}
	client         ssh.Client
}

func NewClusterWorker(ctx context.Context, promAddr string, interval time.Duration, send func(ctx context.Context, metric *edgeproto.Metric) bool, clusterInst *edgeproto.ClusterInst, pf platform.Platform) (*ClusterWorker, error) {
	var err error
	p := ClusterWorker{}
	p.promAddr = promAddr
	p.deployment = clusterInst.Deployment
	p.interval = interval
	p.send = send
	p.clusterInstKey = clusterInst.Key
	p.client, err = pf.GetPlatformClient(ctx, clusterInst)
	if err != nil {
		// If we cannot get a platform client no point in trying to get metrics
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to acquire platform client", "cluster", clusterInst.Key, "error", err)
		return nil, err
	}
	log.SpanLog(ctx, log.DebugLevelMetrics, "NewClusterWorker", "cluster", clusterInst.Key, "promAddr", promAddr)
	// only support K8s deployments
	if p.deployment == cloudcommon.AppDeploymentTypeKubernetes {
		p.clusterStat = &K8sClusterStats{
			key:      p.clusterInstKey,
			client:   p.client,
			promAddr: p.promAddr,
		}
	} else if p.deployment == cloudcommon.AppDeploymentTypeDocker {
		p.clusterStat = &DockerClusterStats{
			key:    p.clusterInstKey,
			client: p.client,
		}
	} else {
		return nil, fmt.Errorf("Unsupported deployment %s", clusterInst.Deployment)
	}

	return &p, nil
}

func (p *ClusterWorker) Start(ctx context.Context) {
	p.stop = make(chan struct{})
	p.waitGrp.Add(1)
	go p.RunNotify()
	log.SpanLog(ctx, log.DebugLevelMetrics, "Started ClusterWorker thread\n")
}

func (p *ClusterWorker) Stop(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelMetrics, "Stopping ClusterWorker thread\n")
	close(p.stop)
	p.waitGrp.Wait()
	flushAlerts(ctx, &p.clusterInstKey)
}

func (p *ClusterWorker) RunNotify() {
	done := false
	for !done {
		select {
		case <-time.After(p.interval):
			span := log.StartSpan(log.DebugLevelSampled, "send-metric")
			span.SetTag("operator", p.clusterInstKey.CloudletKey.Organization)
			span.SetTag("cloudlet", p.clusterInstKey.CloudletKey.Name)
			span.SetTag("cluster", p.clusterInstKey.ClusterKey.Name)
			ctx := log.ContextWithSpan(context.Background(), span)
			clusterStats := p.clusterStat.GetClusterStats(ctx)
			appStatsMap := p.clusterStat.GetAppStats(ctx)

			// create another span for alerts that is always logged
			aspan := log.StartSpan(log.DebugLevelMetrics, "alerts check")
			aspan.SetTag("operator", p.clusterInstKey.CloudletKey.Organization)
			aspan.SetTag("cloudlet", p.clusterInstKey.CloudletKey.Name)
			aspan.SetTag("cluster", p.clusterInstKey.ClusterKey.Name)
			actx := log.ContextWithSpan(context.Background(), aspan)

			for key, stat := range appStatsMap {
				appMetrics := MarshalAppMetrics(&key, stat)
				for _, metric := range appMetrics {
					p.send(ctx, metric)
				}
			}
			clusterMetrics := MarshalClusterMetrics(p.clusterInstKey, clusterStats)
			for _, metric := range clusterMetrics {
				p.send(ctx, metric)
			}

			clusterAlerts := p.clusterStat.GetAlerts(actx)
			updateAlerts(actx, &p.clusterInstKey, clusterAlerts)

			span.Finish()
			aspan.Finish()
		case <-p.stop:
			done = true
		}
	}
	p.waitGrp.Done()
}

// newMetric is called for both Cluster and App stats
func newMetric(clusterInstKey edgeproto.ClusterInstKey, name string, key *shepherd_common.MetricAppInstKey, ts *types.Timestamp) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Name = name
	metric.Timestamp = *ts
	metric.AddTag("cloudletorg", clusterInstKey.CloudletKey.Organization)
	metric.AddTag("cloudlet", clusterInstKey.CloudletKey.Name)
	metric.AddTag("cluster", clusterInstKey.ClusterKey.Name)
	metric.AddTag("clusterorg", clusterInstKey.Organization)
	if key != nil {
		metric.AddTag("pod", key.Pod)
		metric.AddTag("app", key.App)
		metric.AddTag("ver", key.Version)
		//TODO: this should be changed when we have the actual app key
		metric.AddTag("apporg", key.ClusterInstKey.Organization)
	}
	return &metric
}

func MarshalClusterMetrics(key edgeproto.ClusterInstKey, cm *shepherd_common.ClusterMetrics) []*edgeproto.Metric {
	var metrics []*edgeproto.Metric
	var metric *edgeproto.Metric

	// bail out if we get no metrics
	if cm == nil {
		return nil
	}

	// nil timestamps mean the curl request failed. So do not write the metric in
	if cm.CpuTS != nil {
		metric = newMetric(key, "cluster-cpu", nil, cm.CpuTS)
		metric.AddDoubleVal("cpu", cm.Cpu)
		metrics = append(metrics, metric)
		//reset to nil for the next collection
		cm.CpuTS = nil
	}

	if cm.MemTS != nil {
		metric = newMetric(key, "cluster-mem", nil, cm.MemTS)
		metric.AddDoubleVal("mem", cm.Mem)
		metrics = append(metrics, metric)
		cm.MemTS = nil
	}

	if cm.DiskTS != nil {
		metric = newMetric(key, "cluster-disk", nil, cm.DiskTS)
		metric.AddDoubleVal("disk", cm.Disk)
		metrics = append(metrics, metric)
		cm.DiskTS = nil
	}

	if cm.NetSentTS != nil && cm.NetRecvTS != nil {
		//for measurements with multiple values just pick one timestamp to use
		metric = newMetric(key, "cluster-network", nil, cm.NetSentTS)
		metric.AddIntVal("sendBytes", cm.NetSent)
		metric.AddIntVal("recvBytes", cm.NetRecv)
		metrics = append(metrics, metric)
	}
	cm.NetSentTS = nil
	cm.NetRecvTS = nil

	if cm.TcpConnsTS != nil && cm.TcpRetransTS != nil {
		metric = newMetric(key, "cluster-tcp", nil, cm.TcpConnsTS)
		metric.AddIntVal("tcpConns", cm.TcpConns)
		metric.AddIntVal("tcpRetrans", cm.TcpRetrans)
		metrics = append(metrics, metric)
	}
	cm.NetSentTS = nil
	cm.NetRecvTS = nil

	if cm.UdpSentTS != nil && cm.UdpRecvTS != nil && cm.UdpRecvErrTS != nil {
		metric = newMetric(key, "cluster-udp", nil, cm.UdpSentTS)
		metric.AddIntVal("udpSent", cm.UdpSent)
		metric.AddIntVal("udpRecv", cm.UdpRecv)
		metric.AddIntVal("udpRecvErr", cm.UdpRecvErr)
		metrics = append(metrics, metric)
	}
	cm.UdpSentTS = nil
	cm.UdpRecvTS = nil
	cm.UdpRecvErrTS = nil

	return metrics
}

func MarshalAppMetrics(key *shepherd_common.MetricAppInstKey, stat *shepherd_common.AppMetrics) []*edgeproto.Metric {
	var metrics []*edgeproto.Metric
	var metric *edgeproto.Metric

	// bail out if we get no metrics
	if stat == nil {
		return nil
	}

	if stat.CpuTS != nil {
		metric = newMetric(key.ClusterInstKey, "appinst-cpu", key, stat.CpuTS)
		metric.AddDoubleVal("cpu", stat.Cpu)
		metrics = append(metrics, metric)
		stat.CpuTS = nil
	}

	if stat.MemTS != nil {
		metric = newMetric(key.ClusterInstKey, "appinst-mem", key, stat.MemTS)
		metric.AddIntVal("mem", stat.Mem)
		metrics = append(metrics, metric)
		stat.MemTS = nil
	}

	if stat.DiskTS != nil {
		metric = newMetric(key.ClusterInstKey, "appinst-disk", key, stat.DiskTS)
		metric.AddIntVal("disk", stat.Disk)
		metrics = append(metrics, metric)
		stat.DiskTS = nil
	}

	if stat.NetSentTS != nil && stat.NetRecvTS != nil {
		//for measurements with multiple values just pick one timestamp to use
		metric = newMetric(key.ClusterInstKey, "appinst-network", key, stat.NetSentTS)
		metric.AddIntVal("sendBytes", stat.NetSent)
		metric.AddIntVal("recvBytes", stat.NetRecv)
		metrics = append(metrics, metric)
	}
	stat.NetSentTS = nil
	stat.NetRecvTS = nil

	return metrics
}

// Don't consider alerts, which are not destined for this cluster Instance and not clusterInst alerts
func pruneForeignAlerts(clusterInstKey *edgeproto.ClusterInstKey, keys *map[edgeproto.AlertKey]context.Context) {
	alertFromKey := edgeproto.Alert{}
	for key, _ := range *keys {
		edgeproto.AlertKeyStringParse(string(key), &alertFromKey)
		if _, found := alertFromKey.Labels[cloudcommon.AlertLabelApp]; found ||
			alertFromKey.Labels[cloudcommon.AlertLabelClusterOrg] != clusterInstKey.Organization ||
			alertFromKey.Labels[cloudcommon.AlertLabelCloudletOrg] != clusterInstKey.CloudletKey.Organization ||
			alertFromKey.Labels[cloudcommon.AlertLabelCloudlet] != clusterInstKey.CloudletKey.Name ||
			alertFromKey.Labels[cloudcommon.AlertLabelCluster] != clusterInstKey.ClusterKey.Name {
			delete(*keys, key)
		}
	}
}

func updateAlerts(ctx context.Context, clusterInstKey *edgeproto.ClusterInstKey, alerts []edgeproto.Alert) {
	if alerts == nil {
		// some error occurred, do not modify existing cache set
		return
	}

	stale := make(map[edgeproto.AlertKey]context.Context)
	AlertCache.GetAllKeys(ctx, stale)

	changeCount := 0
	for ii, _ := range alerts {
		alert := &alerts[ii]
		alert.Labels[cloudcommon.AlertLabelClusterOrg] = clusterInstKey.Organization
		alert.Labels[cloudcommon.AlertLabelCloudletOrg] = clusterInstKey.CloudletKey.Organization
		alert.Labels[cloudcommon.AlertLabelCloudlet] = clusterInstKey.CloudletKey.Name
		alert.Labels[cloudcommon.AlertLabelCluster] = clusterInstKey.ClusterKey.Name

		AlertCache.UpdateModFunc(ctx, alert.GetKey(), 0, func(old *edgeproto.Alert) (*edgeproto.Alert, bool) {
			if old == nil {
				log.SpanLog(ctx, log.DebugLevelMetrics, "Update alert", "alert", alert)
				changeCount++
				return alert, true
			}
			// don't update if nothing changed
			changed := !alert.Matches(old)
			if changed {
				changeCount++
				log.SpanLog(ctx, log.DebugLevelMetrics, "Update alert", "alert", alert)
			}
			return alert, changed
		})
		delete(stale, alert.GetKeyVal())
	}
	// delete our stale entries
	pruneForeignAlerts(clusterInstKey, &stale)
	for key, _ := range stale {
		buf := edgeproto.Alert{}
		buf.SetKey(&key)
		AlertCache.Delete(ctx, &buf, 0)
		changeCount++
	}
	if changeCount == 0 {
		// suppress span log since nothing logged
		span := log.SpanFromContext(ctx)
		log.NoLogSpan(span)
	}
}

// flushAlerts removes Alerts for clusters that have been deleted
func flushAlerts(ctx context.Context, key *edgeproto.ClusterInstKey) {
	toflush := []edgeproto.AlertKey{}
	AlertCache.Mux.Lock()
	for k, v := range AlertCache.Objs {
		if v.Labels[cloudcommon.AlertLabelClusterOrg] == key.Organization &&
			v.Labels[cloudcommon.AlertLabelCloudletOrg] == key.CloudletKey.Organization &&
			v.Labels[cloudcommon.AlertLabelCloudlet] == key.CloudletKey.Name &&
			v.Labels[cloudcommon.AlertLabelCluster] == key.ClusterKey.Name {
			toflush = append(toflush, k)
		}
	}
	AlertCache.Mux.Unlock()
	for _, k := range toflush {
		buf := edgeproto.Alert{}
		buf.SetKey(&k)
		AlertCache.Delete(ctx, &buf, 0)
	}
}
