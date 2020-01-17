package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	platform "github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_edgebox"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_fake"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_openstack"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	opentracing "github.com/opentracing/opentracing-go"
)

var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:51001", "CRM notify listener addresses")
var platformName = flag.String("platform", "", "Platform type of Cloudlet")
var vaultAddr = flag.String("vaultAddr", "", "Address to vault")
var physicalName = flag.String("physicalName", "", "Physical infrastructure cloudlet name, defaults to cloudlet name in cloudletKey")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var name = flag.String("name", "shepherd", "Unique name to identify a process")
var parentSpan = flag.String("span", "", "Use parent span for logging")
var region = flag.String("region", "local", "Region name")
var defaultPrometheusPort = cloudcommon.PrometheusPort

//map keeping track of all the currently running prometheuses
var workerMap map[string]*ClusterWorker
var vmAppWorkerMap map[string]*AppInstWorker
var MEXPrometheusAppName = cloudcommon.MEXPrometheusAppName
var AppInstCache edgeproto.AppInstCache
var ClusterInstCache edgeproto.ClusterInstCache
var AppCache edgeproto.AppCache
var MetricSender *notify.MetricSend
var AlertCache edgeproto.AlertCache

var cloudletKey edgeproto.CloudletKey
var myPlatform platform.Platform

var sigChan chan os.Signal
var collectInterval time.Duration

func appInstCb(ctx context.Context, old *edgeproto.AppInst, new *edgeproto.AppInst) {
	// LB metrics are not supported in fake mode
	if myPlatform.GetType() != "fake" {
		CollectProxyStats(ctx, new)
	}
	var port int32
	var exists bool
	var mapKey string

	// check cluster name if this is a VM App
	if new.Key.ClusterInstKey.ClusterKey.Name == cloudcommon.DefaultVMCluster {
		mapKey = new.Key.GetKeyString()
		stats, exists := vmAppWorkerMap[mapKey]
		if new.State == edgeproto.TrackedState_READY && !exists {
			// Add/Create
			stats, err := NewAppInstWorker(ctx, collectInterval, MetricSender.Update, new, myPlatform)
			if err == nil {
				vmAppWorkerMap[mapKey] = stats
				stats.Start(ctx)
			}
		} else if new.State != edgeproto.TrackedState_READY && exists {
			delete(vmAppWorkerMap, mapKey)
			stats.Stop(ctx)
		}
		// Done for VM Apps
		return
	} else if new.Key.AppKey.Name == MEXPrometheusAppName {
		// check for prometheus
		mapKey = k8smgmt.GetK8sNodeNameSuffix(&new.Key.ClusterInstKey)
	} else {
		return
	}
	stats, exists := workerMap[mapKey]
	if new.State == edgeproto.TrackedState_READY {
		log.SpanLog(ctx, log.DebugLevelMetrics, "New Prometheus instance detected", "clustername", mapKey, "appInst", new)
		// get address of prometheus.
		clusterInst := edgeproto.ClusterInst{}
		found := ClusterInstCache.Get(&new.Key.ClusterInstKey, &clusterInst)
		if !found {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find clusterInst for prometheus")
			return
		}
		clustIP, err := myPlatform.GetClusterIP(ctx, &clusterInst)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "error getting clusterIP", "err", err.Error())
			return
		}
		// We don't actually expose prometheus ports - we should default to 9090
		if len(new.MappedPorts) > 0 {
			port = new.MappedPorts[0].PublicPort
		} else {
			port = defaultPrometheusPort
		}
		promAddress := fmt.Sprintf("%s:%d", clustIP, port)
		log.SpanLog(ctx, log.DebugLevelMetrics, "prometheus found", "promAddress", promAddress)
		if !exists {
			stats, err = NewClusterWorker(ctx, promAddress, collectInterval, MetricSender.Update, &clusterInst, myPlatform)
			if err == nil {
				workerMap[mapKey] = stats
				stats.Start(ctx)
			}
		} else { //somehow this cluster's prometheus was already registered
			log.SpanLog(ctx, log.DebugLevelMetrics, "Error, Prometheus app already registered for this cluster")
		}
	} else { //if its anything other than ready just stop it
		//try to remove it from the workerMap
		if exists {
			delete(workerMap, mapKey)
			stats.Stop(ctx)
		}
	}
}

func clusterInstCb(ctx context.Context, old *edgeproto.ClusterInst, new *edgeproto.ClusterInst) {
	// This is for Docker deployments only
	if new.Deployment != cloudcommon.AppDeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelMetrics, "New cluster instace", "clusterInst", new)
		return
	}
	var mapKey = k8smgmt.GetK8sNodeNameSuffix(&new.Key)
	stats, exists := workerMap[mapKey]
	if new.State == edgeproto.TrackedState_READY {
		log.SpanLog(ctx, log.DebugLevelMetrics, "New Docker cluster detected", "clustername", mapKey, "clusterInst", new)
		if !exists {
			stats, err := NewClusterWorker(ctx, "", collectInterval, MetricSender.Update, new, myPlatform)
			if err == nil {
				workerMap[mapKey] = stats
				stats.Start(ctx)
			}
		} else { //somehow this cluster's prometheus was already registered
			log.SpanLog(ctx, log.DebugLevelMetrics, "Error, This cluster is already registered")
		}
	} else { //if its anything other than ready just stop it
		//try to remove it from the workerMap
		if exists {
			delete(workerMap, mapKey)
			stats.Stop(ctx)
		}
	}
}

func getPlatform() (platform.Platform, error) {
	var plat platform.Platform
	var err error
	switch *platformName {
	case "PLATFORM_TYPE_EDGEBOX":
		plat = &shepherd_edgebox.Platform{}
	case "PLATFORM_TYPE_OPENSTACK":
		plat = &shepherd_openstack.Platform{}
	case "PLATFORM_TYPE_FAKEINFRA":
		// change the scrape interval to 1s so we dont have to wait as long for e2e tests to go
		collectInterval = time.Second
		plat = &shepherd_fake.Platform{}
	default:
		err = fmt.Errorf("Platform %s not supported", *platformName)
	}
	return plat, err
}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(*tlsCertFile)
	defer log.FinishTracer()
	collectInterval = cloudcommon.ShepherdMetricsCollectionInterval
	var span opentracing.Span
	if *parentSpan != "" {
		span = log.NewSpanFromString(log.DebugLevelInfo, *parentSpan, "main")
	} else {
		span = log.StartSpan(log.DebugLevelInfo, "main")
	}
	ctx := log.ContextWithSpan(context.Background(), span)

	cloudcommon.ParseMyCloudletKey(false, cloudletKeyStr, &cloudletKey)
	log.SpanLog(ctx, log.DebugLevelMetrics, "Metrics collection", "interval", collectInterval)
	var err error
	myPlatform, err = getPlatform()
	if err != nil {
		log.FatalLog("Failed to get platform", "platformName", platformName, "err", err)
	}
	err = myPlatform.Init(ctx, &cloudletKey, *region, *physicalName, *vaultAddr)
	if err != nil {
		log.FatalLog("Failed to initialize platform", "platformName", platformName, "err", err)
	}
	workerMap = make(map[string]*ClusterWorker)
	vmAppWorkerMap = make(map[string]*AppInstWorker)
	// LB metrics are not supported in fake mode
	if myPlatform.GetType() != "fake" {
		InitProxyScraper()
		StartProxyScraper()
	}
	InitPlatformMetrics()

	// register shepherd to receive appinst and clusterinst notifications from crm
	edgeproto.InitAppInstCache(&AppInstCache)
	AppInstCache.SetUpdatedCb(appInstCb)
	ClusterInstCache.SetUpdatedCb(clusterInstCb)
	edgeproto.InitClusterInstCache(&ClusterInstCache)
	edgeproto.InitAppCache(&AppCache)
	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient := notify.NewClient(addrs, *tlsCertFile)
	notifyClient.RegisterRecvAppInstCache(&AppInstCache)
	notifyClient.RegisterRecvClusterInstCache(&ClusterInstCache)
	notifyClient.RegisterRecvAppCache(&AppCache)
	// register to send metrics
	MetricSender = notify.NewMetricSend()
	notifyClient.RegisterSend(MetricSender)
	edgeproto.InitAlertCache(&AlertCache)
	notifyClient.RegisterSendAlertCache(&AlertCache)

	notifyClient.Start()
	defer notifyClient.Stop()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.SpanLog(ctx, log.DebugLevelMetrics, "Ready")
	span.Finish()

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
