package main

import (
	"context"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	platform "github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// For each cluster the notify worker is created
type AppInstWorker struct {
	pf         platform.Platform
	appInstKey edgeproto.AppInstKey
	interval   time.Duration
	send       func(ctx context.Context, metric *edgeproto.Metric) bool
	waitGrp    sync.WaitGroup
	stop       chan struct{}
}

func NewAppInstWorker(ctx context.Context, interval time.Duration, send func(ctx context.Context, metric *edgeproto.Metric) bool, appinst *edgeproto.AppInst, pf platform.Platform) (*AppInstWorker, error) {
	p := AppInstWorker{}
	p.pf = pf
	if int64(interval) > int64(pf.GetMetricsCollectInterval()) {
		p.interval = interval
	} else {
		p.interval = pf.GetMetricsCollectInterval()
	}
	p.send = send
	p.appInstKey = appinst.Key
	log.SpanLog(ctx, log.DebugLevelMetrics, "NewAppInstWorker", "app", appinst)
	return &p, nil
}

func (p *AppInstWorker) Start(ctx context.Context) {
	p.stop = make(chan struct{})
	p.waitGrp.Add(1)
	go p.RunNotify()
	log.SpanLog(ctx, log.DebugLevelMetrics, "Started AppInstWorker thread\n")
}

func (p *AppInstWorker) Stop(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelMetrics, "Stopping AppInstWorker thread\n")
	close(p.stop)
	p.waitGrp.Wait()
}

func (p *AppInstWorker) RunNotify() {
	done := false
	for !done {
		select {
		case <-time.After(p.interval):
			span := log.StartSpan(log.DebugLevelSampled, "send-metric")
			span.SetTag("operator", p.appInstKey.ClusterInstKey.CloudletKey.OperatorKey.Name)
			span.SetTag("cloudlet", p.appInstKey.ClusterInstKey.CloudletKey.Name)
			span.SetTag("cluster", cloudcommon.DefaultVMCluster)
			ctx := log.ContextWithSpan(context.Background(), span)
			key := shepherd_common.MetricAppInstKey{
				ClusterInstKey: p.appInstKey.ClusterInstKey,
				Pod:            p.appInstKey.AppKey.Name,
			}
			log.SpanLog(ctx, log.DebugLevelMetrics, "Collecting metrics for app", "key", key)
			stat, err := p.pf.GetVmStats(ctx, &p.appInstKey)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Failed to get metrics from VM", "app", p.appInstKey, "err", err)
				span.Finish()
				continue
			}
			log.SpanLog(ctx, log.DebugLevelMetrics, "metrics for app", "key", key, "metrics", stat)
			appMetrics := MarshalAppMetrics(&key, &stat)
			for _, metric := range appMetrics {
				p.send(ctx, metric)
			}
			span.Finish()
		case <-p.stop:
			done = true
		}
	}
	p.waitGrp.Done()
}
