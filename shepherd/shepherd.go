package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	baselog "log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"text/template"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	platform "github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_edgebox"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_fake"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_openstack"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	opentracing "github.com/opentracing/opentracing-go"
)

var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:51001", "CRM notify listener addresses")
var metricsAddr = flag.String("metricsAddr", "0.0.0.0:9091", "Metrics Porxy Address")
var platformName = flag.String("platform", "", "Platform type of Cloudlet")
var physicalName = flag.String("physicalName", "", "Physical infrastructure cloudlet name, defaults to cloudlet name in cloudletKey")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var name = flag.String("name", "shepherd", "Unique name to identify a process")
var parentSpan = flag.String("span", "", "Use parent span for logging")
var region = flag.String("region", "local", "Region name")
var promTargetsFile = flag.String("targetsFile", "/tmp/prom_targets.json", "Prometheus targets file")
var promAlertsFile = flag.String("alertsFile", "/tmp/prom_rules.yml", "Prometheus alerts file")

var defaultPrometheusPort = cloudcommon.PrometheusPort

//map keeping track of all the currently running prometheuses
var workerMap map[string]*ClusterWorker
var vmAppWorkerMap map[string]*AppInstWorker
var MEXPrometheusAppName = cloudcommon.MEXPrometheusAppName
var AppInstCache edgeproto.AppInstCache
var ClusterInstCache edgeproto.ClusterInstCache
var AppCache edgeproto.AppCache
var CloudletCache edgeproto.CloudletCache
var CloudletInfoCache edgeproto.CloudletInfoCache
var MetricSender *notify.MetricSend
var AlertCache edgeproto.AlertCache
var settings edgeproto.Settings

var cloudletKey edgeproto.CloudletKey
var myPlatform platform.Platform
var nodeMgr node.NodeMgr

var sigChan chan os.Signal

var promTargetTemplate *template.Template

var promTargetT = `
{
	"targets": ["{{.MetricsProxyAddr}}"],
	"labels": {
		"app": "{{.Key.AppKey.Name}}",
		"ver": "{{.Key.AppKey.Version}}",
		"apporg": "{{.Key.AppKey.Organization}}",
		"cluster": "{{.Key.ClusterInstKey.ClusterKey.Name}}",
		"clusterorg": "{{.Key.ClusterInstKey.Organization}}",
		"cloudlet": "{{.Key.ClusterInstKey.CloudletKey.Name}}",
		"cloudletorg": "{{.Key.ClusterInstKey.CloudletKey.Organization}}",
		"__metrics_path__":"{{.EnvoyMetricsPath}}"
	}
}`

var promHealthCheckAlerts = `groups:
- name: StaticRules
  rules:
  - alert: RootLbProxyDown
    expr: up == 0
    for: 1m
  - alert: HealthCheck
    expr: envoy_cluster_health_check_healthy == 0
`

type targetData struct {
	MetricsProxyAddr string
	Key              edgeproto.AppInstKey
	EnvoyMetricsPath string
}

func getAppInstPrometheusTargetString(appInst *edgeproto.AppInst) (string, error) {
	host := *metricsAddr
	switch *platformName {
	case "PLATFORM_TYPE_EDGEBOX":
		fallthrough
	case "PLATFORM_TYPE_FAKEINFRA":
		host = "host.docker.internal:9091"
	}
	target := targetData{
		MetricsProxyAddr: host,
		Key:              appInst.GetKeyVal(),
		EnvoyMetricsPath: "/metrics/" + getProxyKey(&appInst.Key),
	}
	buf := bytes.Buffer{}
	if err := promTargetTemplate.Execute(&buf, target); err != nil {
		log.DebugLog(log.DebugLevelMetrics, "Failed to create a target", "template", promTargetTemplate,
			"data", target, "error", err)
		return "", err
	}
	return buf.String(), nil
}

// Walk through AppInstances and write out the targets
func writePrometheusTargetsFile() {
	var targets = "["
	AppInstCache.Show(&edgeproto.AppInst{}, func(obj *edgeproto.AppInst) error {
		if targets != "[" {
			targets += ","
		}
		promTargetJson, err := getAppInstPrometheusTargetString(obj)
		if err == nil {
			targets += promTargetJson
		}
		// just skip the targets that we are unable to fill
		return nil
	})
	targets += "]"
	ioutil.WriteFile(*promTargetsFile, []byte(targets), 0600)
}

func appInstCb(ctx context.Context, old *edgeproto.AppInst, new *edgeproto.AppInst) {
	// LB metrics are not supported in fake mode
	if myPlatform.GetType() != "fake" {
		if target := CollectProxyStats(ctx, new); target != "" {
			go writePrometheusTargetsFile()
		}
	}
	var port int32
	var exists bool
	var mapKey string

	collectInterval := settings.ShepherdMetricsCollectionInterval.TimeDuration()
	// check cluster name if this is a VM App
	app := edgeproto.App{}
	found := AppCache.Get(&new.Key.AppKey, &app)
	if !found {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find app", "app", new.Key.AppKey.Name)
		return
	}
	if app.Deployment == cloudcommon.AppDeploymentTypeVM {
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
	collectInterval := settings.ShepherdMetricsCollectionInterval.TimeDuration()
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
		plat = &shepherd_openstack.ShepherdPlatform{}
	case "PLATFORM_TYPE_FAKEINFRA":
		plat = &shepherd_fake.Platform{}
	default:
		err = fmt.Errorf("Platform %s not supported", *platformName)
	}
	return plat, err
}

func metricsProxy(w http.ResponseWriter, r *http.Request) {
	app := r.URL.Path[len("/metrics/"):]
	if app != "" {
		// Search ProxyMap for the names
		target := getProxyScrapePoint(app)
		if target.ProxyContainer == "nginx" {
			return
		}
		request := fmt.Sprintf("docker exec %s curl -s -S http://127.0.0.1:%d/stats/prometheus", target.ProxyContainer, cloudcommon.ProxyMetricsPort)
		resp, err := target.Client.OutputWithTimeout(request, shepherd_common.ShepherdSshConnectTimeout)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(resp))
	}
}

func targetsList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>%s</h1>", "List all targets")
	targets := copyMapValues()
	for ii, v := range targets {
		fmt.Fprintf(w, "<h1>Target %d</h1><div>%s</div>", ii, getProxyKey(&v.Key))
	}
}

func main() {
	nodeMgr.InitFlags()
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(nodeMgr.TlsCertFile)
	defer log.FinishTracer()
	var span opentracing.Span
	if *parentSpan != "" {
		span = log.NewSpanFromString(log.DebugLevelInfo, *parentSpan, "main")
	} else {
		span = log.StartSpan(log.DebugLevelInfo, "main")
	}
	ctx := log.ContextWithSpan(context.Background(), span)
	settings = *edgeproto.GetDefaultSettings()

	cloudcommon.ParseMyCloudletKey(false, cloudletKeyStr, &cloudletKey)

	err := nodeMgr.Init(ctx, "shepherd", node.WithCloudletKey(&cloudletKey), node.WithRegion(*region))
	clientTlsConfig, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegionalCloudlet,
		[]node.MatchCA{node.SameRegionalCloudletMatchCA()})
	if err != nil {
		log.FatalLog("Failed to get internal pki tls config", "err", err)
	}

	myPlatform, err = getPlatform()
	if err != nil {
		log.FatalLog("Failed to get platform", "platformName", platformName, "err", err)
	}

	// This works for edgebox and openstack cloudlets for now
	if *platformName == "PLATFORM_TYPE_EDGEBOX" ||
		*platformName == "PLATFORM_TYPE_OPENSTACK" {
		// Init prometheus targets template
		promTargetTemplate = template.Must(template.New("prometheustarget").Parse(promTargetT))
		// write alerting rules
		ioutil.WriteFile(*promAlertsFile, []byte(promHealthCheckAlerts), 0600)

		// Init http metricsProxy for Prometheus API endpoints
		var nullLogger baselog.Logger
		nullLogger.SetOutput(ioutil.Discard)

		http.HandleFunc("/list", targetsList)
		http.HandleFunc("/metrics/", metricsProxy)
		httpServer := &http.Server{
			Addr:     *metricsAddr,
			ErrorLog: &nullLogger,
		}
		go func() {
			err = httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.FatalLog("Failed to serve metrics", "err", err)
			}
		}()
	}

	// register shepherd to receive appinst and clusterinst notifications from crm
	edgeproto.InitAppInstCache(&AppInstCache)
	AppInstCache.SetUpdatedCb(appInstCb)
	edgeproto.InitClusterInstCache(&ClusterInstCache)
	ClusterInstCache.SetUpdatedCb(clusterInstCb)
	edgeproto.InitAppCache(&AppCache)
	// also register to receive cloudlet details
	edgeproto.InitCloudletCache(&CloudletCache)

	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient := notify.NewClient(addrs, tls.GetGrpcDialOption(clientTlsConfig))
	notifyClient.SetFilterByCloudletKey()
	notifyClient.RegisterRecvAppInstCache(&AppInstCache)
	notifyClient.RegisterRecvClusterInstCache(&ClusterInstCache)
	notifyClient.RegisterRecvAppCache(&AppCache)
	notifyClient.RegisterRecvCloudletCache(&CloudletCache)
	// register to send metrics
	MetricSender = notify.NewMetricSend()
	notifyClient.RegisterSend(MetricSender)
	edgeproto.InitAlertCache(&AlertCache)
	notifyClient.RegisterSendAlertCache(&AlertCache)
	// register to send cloudletInfo, to receive appinst/clusterinst/cloudlet notifications from crm
	edgeproto.InitCloudletInfoCache(&CloudletInfoCache)
	notifyClient.RegisterSendCloudletInfoCache(&CloudletInfoCache)

	nodeMgr.RegisterClient(notifyClient)

	notifyClient.Start()
	defer notifyClient.Stop()

	cloudletInfo := edgeproto.CloudletInfo{
		Key: cloudletKey,
	}

	// Send state INIT to get cloudlet obj from crm
	cloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_INIT
	CloudletInfoCache.Update(ctx, &cloudletInfo, 0)

	var cloudlet edgeproto.Cloudlet

	// Fetch cloudlet cache from controller
	// This also ensures that cloudlet is up before we start collecting metrics
	found := false
	log.SpanLog(ctx, log.DebugLevelInfo, "wait for cloudlet cache", "key", cloudletKey)
	for i := 0; i < 50; i++ {
		if CloudletCache.Get(&cloudletKey, &cloudlet) {
			found = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !found {
		log.FatalLog("failed to fetch cloudlet cache from controller")
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "fetched cloudlet cache from controller", "cloudlet", cloudlet)

	err = myPlatform.Init(ctx, &cloudletKey, *region, *physicalName, nodeMgr.VaultAddr, cloudlet.EnvVar)
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

	// Send state READY to get AppInst/ClusterInst objs from crm
	cloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_READY
	CloudletInfoCache.Update(ctx, &cloudletInfo, 0)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.SpanLog(ctx, log.DebugLevelMetrics, "Ready")
	span.Finish()

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
