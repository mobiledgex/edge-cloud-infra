package main

import (
	"context"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type ClusterWorker struct {
	clusterInstKey edgeproto.ClusterInstKey
	promAddr       string
	interval       time.Duration
	appStatsMap    map[MetricAppInstKey]*AppMetrics
	clusterStat    *ClusterMetrics
	send           func(ctx context.Context, metric *edgeproto.Metric) bool
	waitGrp        sync.WaitGroup
	stop           chan struct{}
	client         pc.PlatformClient
}

// TODO - move all the ClusterWorker functions here
