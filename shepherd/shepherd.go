package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/openstack"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vmpool"
	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vsphere"
	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	platform "github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_edgebox"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_fake"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_platform/shepherd_vmprovider"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
	"github.com/mobiledgex/edge-cloud/util/tasks"
	opentracing "github.com/opentracing/opentracing-go"
)

var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:51001", "CRM notify listener addresses")
var metricsAddr = flag.String("metricsAddr", "0.0.0.0:9091", "Metrics Proxy Address")
var platformName = flag.String("platform", "", "Platform type of Cloudlet")
var physicalName = flag.String("physicalName", "", "Physical infrastructure cloudlet name, defaults to cloudlet name in cloudletKey")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")
var name = flag.String("name", "shepherd", "Unique name to identify a process")
var parentSpan = flag.String("span", "", "Use parent span for logging")
var region = flag.String("region", "local", "Region name")
var promTargetsFile = flag.String("targetsFile", "/tmp/prom_targets.json", "Prometheus targets file")
var appDNSRoot = flag.String("appDNSRoot", "mobiledgex.net", "App domain name root")
var deploymentTag = flag.String("deploymentTag", "", "Tag to indicate type of deployment setup. Ex: production, staging, etc")
var chefServerPath = flag.String("chefServerPath", "", "Chef server path")

var defaultPrometheusPort = cloudcommon.PrometheusPort

//map keeping track of all the currently running prometheuses
var workerMap map[string]*ClusterWorker
var workerMapMutex *sync.Mutex
var vmAppWorkerMap map[string]*AppInstWorker
var MEXPrometheusAppName = cloudcommon.MEXPrometheusAppName
var AppInstCache edgeproto.AppInstCache
var ClusterInstCache edgeproto.ClusterInstCache
var AppCache edgeproto.AppCache
var VMPoolCache edgeproto.VMPoolCache
var VMPoolInfoCache edgeproto.VMPoolInfoCache
var CloudletCache edgeproto.CloudletCache
var CloudletInfoCache edgeproto.CloudletInfoCache
var MetricSender *notify.MetricSend
var AlertCache edgeproto.AlertCache
var AutoProvPoliciesCache edgeproto.AutoProvPolicyCache
var SettingsCache edgeproto.SettingsCache
var settings edgeproto.Settings
var AppInstByAutoProvPolicy edgeproto.AppInstLookupByPolicyKey
var targetFileWorkers tasks.KeyWorkers
var appInstAlertWorkers tasks.KeyWorkers

var cloudletKey edgeproto.CloudletKey
var myPlatform platform.Platform
var nodeMgr node.NodeMgr

var sigChan chan os.Signal
var notifyClient *notify.Client
var cloudletWait = make(chan bool, 1)

var targetsFileWorkerKey = "write-targets"

var CRMTimeout = 1 * time.Minute

func appInstCb(ctx context.Context, old *edgeproto.AppInst, new *edgeproto.AppInst) {
	if target := CollectProxyStats(ctx, new); target != "" {
		targetFileWorkers.NeedsWork(ctx, targetsFileWorkerKey)
		appInstAlertWorkers.NeedsWork(ctx, new.Key)
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
	if app.Deployment == cloudcommon.DeploymentTypeVM {
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
	workerMapMutex.Lock()
	defer workerMapMutex.Unlock()
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

// It's possible that we may miss the transition from AppInst READY to another
// state before it gets deleted, so we need to handle delete as well.
func appInstDeletedCb(ctx context.Context, old *edgeproto.AppInst) {
	old.State = edgeproto.TrackedState_NOT_PRESENT
	appInstCb(ctx, old, old)
}

func clusterInstCb(ctx context.Context, old *edgeproto.ClusterInst, new *edgeproto.ClusterInst) {
	var mapKey = k8smgmt.GetK8sNodeNameSuffix(&new.Key)
	workerMapMutex.Lock()
	defer workerMapMutex.Unlock()
	stats, exists := workerMap[mapKey]
	if new.State == edgeproto.TrackedState_READY && exists {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Update cluster details", "old", old, "new", new)
		if new.Reservable {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Update reserved-by setting")
			stats.reservedBy = new.ReservedBy
			workerMap[mapKey] = stats
		}
		return
	}
	// This is for Docker deployments only
	if new.Deployment != cloudcommon.DeploymentTypeDocker {
		log.SpanLog(ctx, log.DebugLevelMetrics, "New cluster instace", "clusterInst", new)
		return
	}
	collectInterval := settings.ShepherdMetricsCollectionInterval.TimeDuration()
	if new.State == edgeproto.TrackedState_READY {
		log.SpanLog(ctx, log.DebugLevelMetrics, "New Docker cluster detected", "clustername", mapKey, "clusterInst", new)
		stats, err := NewClusterWorker(ctx, "", collectInterval, MetricSender.Update, new, myPlatform)
		if err == nil {
			workerMap[mapKey] = stats
			stats.Start(ctx)
		}
	} else { //if its anything other than ready just stop it
		//try to remove it from the workerMap
		if exists {
			delete(workerMap, mapKey)
			stats.Stop(ctx)
		}
	}
}

func autoProvPolicyCb(ctx context.Context, old *edgeproto.AutoProvPolicy, new *edgeproto.AutoProvPolicy) {
	// we only care if undeploy policy changed.
	if old != nil && old.UndeployClientCount == new.UndeployClientCount && old.UndeployIntervalCount == new.UndeployIntervalCount {
		return
	}
	instKeys := AppInstByAutoProvPolicy.Find(new.Key)
	for _, key := range instKeys {
		appInstAlertWorkers.NeedsWork(ctx, key)
	}
}

func settingsCb(ctx context.Context, _ *edgeproto.Settings, new *edgeproto.Settings) {
	old := settings
	settings = *new
	if old.ShepherdMetricsCollectionInterval !=
		new.ShepherdMetricsCollectionInterval ||
		old.ShepherdAlertEvaluationInterval !=
			new.ShepherdAlertEvaluationInterval {
		// re-write Cloudlet Prometheus config and reload
		err := intprocess.WriteCloudletPromConfig(ctx, new)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelNotify, "Failed to write cloudlet prometheus config", "err", err)
		} else {
			reloadCloudletProm(ctx)
		}
	}
	if old.AutoDeployIntervalSec != new.AutoDeployIntervalSec {
		// re-write undeploy rules since they all depend on AutoDeployIntervalSec
		s := &AppInstByAutoProvPolicy
		s.Mux.Lock()
		for _, insts := range s.PolicyKeys {
			for appInstKey, _ := range insts {
				appInstAlertWorkers.NeedsWork(ctx, appInstKey)
			}
		}
		s.Mux.Unlock()
	}
}

func vmPoolInfoCb(ctx context.Context, old *edgeproto.VMPoolInfo, new *edgeproto.VMPoolInfo) {
	vmPool := edgeproto.VMPool{}
	vmPool.Key = new.Key
	vmPool.Vms = []edgeproto.VM{}
	for _, infoVM := range new.Vms {
		vmPool.Vms = append(vmPool.Vms, infoVM)
	}
	vmPool.State = new.State
	vmPool.Errors = new.Errors
	myPlatform.SetVMPool(ctx, &vmPool)
}

func cloudletCb(ctx context.Context, old *edgeproto.Cloudlet, new *edgeproto.Cloudlet) {
	select {
	case cloudletWait <- true:
		// Got cloudlet object
	default:
	}
}

func getPlatform() (platform.Platform, error) {
	var plat platform.Platform
	var err error
	switch *platformName {
	case "PLATFORM_TYPE_EDGEBOX":
		plat = &shepherd_edgebox.Platform{}
	case "PLATFORM_TYPE_OPENSTACK":
		osProvider := openstack.OpenstackPlatform{}
		vmPlatform := vmlayer.VMPlatform{
			Type:       vmlayer.VMProviderOpenstack,
			VMProvider: &osProvider,
		}
		plat = &shepherd_vmprovider.ShepherdPlatform{
			VMPlatform: &vmPlatform,
		}
	case "PLATFORM_TYPE_VSPHERE":
		vsphereProvider := vsphere.VSpherePlatform{}
		vmPlatform := vmlayer.VMPlatform{
			Type:       vmlayer.VMProviderVSphere,
			VMProvider: &vsphereProvider,
		}
		plat = &shepherd_vmprovider.ShepherdPlatform{
			VMPlatform: &vmPlatform,
		}
	case "PLATFORM_TYPE_VM_POOL":
		vmpoolProvider := vmpool.VMPoolPlatform{}
		vmPlatform := vmlayer.VMPlatform{
			Type:       vmlayer.VMProviderVMPool,
			VMProvider: &vmpoolProvider,
		}
		plat = &shepherd_vmprovider.ShepherdPlatform{
			VMPlatform: &vmPlatform,
		}
	case "PLATFORM_TYPE_FAKEINFRA":
		plat = &shepherd_fake.Platform{}
	default:
		err = fmt.Errorf("Platform %s not supported", *platformName)
	}
	return plat, err
}

func main() {
	nodeMgr.InitFlags()
	flag.Parse()
	start()
	defer stop()

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}

func start() {
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer(nodeMgr.TlsCertFile)

	var span opentracing.Span
	if *parentSpan != "" {
		span = log.NewSpanFromString(log.DebugLevelInfo, *parentSpan, "main")
	} else {
		span = log.StartSpan(log.DebugLevelInfo, "main")
	}
	defer span.Finish()
	ctx := log.ContextWithSpan(context.Background(), span)
	settings = *edgeproto.GetDefaultSettings()

	cloudcommon.ParseMyCloudletKey(false, cloudletKeyStr, &cloudletKey)

	err := nodeMgr.Init(ctx, "shepherd", node.WithCloudletKey(&cloudletKey), node.WithRegion(*region))
	if err != nil {
		log.FatalLog(err.Error())
	}
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

	targetFileWorkers.Init("cloudlet-prom-targets", writePrometheusTargetsFile)
	appInstAlertWorkers.Init("alert-file-writer", writePrometheusAlertRuleForAppInst)

	if err = startPrometheusMetricsProxy(ctx); err != nil {
		log.FatalLog("Failed to start prometheus metrics proxy", "err", err)
	}

	// register shepherd to receive appinst and clusterinst notifications from crm
	edgeproto.InitAppInstCache(&AppInstCache)
	AppInstCache.SetUpdatedCb(appInstCb)
	AppInstCache.SetDeletedCb(appInstDeletedCb)
	edgeproto.InitClusterInstCache(&ClusterInstCache)
	ClusterInstCache.SetUpdatedCb(clusterInstCb)
	edgeproto.InitAppCache(&AppCache)
	edgeproto.InitAutoProvPolicyCache(&AutoProvPoliciesCache)
	AutoProvPoliciesCache.SetUpdatedCb(autoProvPolicyCb)
	edgeproto.InitSettingsCache(&SettingsCache)
	AppInstByAutoProvPolicy.Init()
	// also register to receive cloudlet details
	edgeproto.InitVMPoolCache(&VMPoolCache)
	edgeproto.InitVMPoolInfoCache(&VMPoolInfoCache)
	edgeproto.InitCloudletCache(&CloudletCache)

	addrs := strings.Split(*notifyAddrs, ",")
	notifyClient = notify.NewClient(nodeMgr.Name(), addrs, tls.GetGrpcDialOption(clientTlsConfig))
	notifyClient.SetFilterByCloudletKey()
	notifyClient.RegisterRecvSettingsCache(&SettingsCache)
	notifyClient.RegisterRecvVMPoolCache(&VMPoolCache)
	notifyClient.RegisterRecvVMPoolInfoCache(&VMPoolInfoCache)
	notifyClient.RegisterRecvAppInstCache(&AppInstCache)
	notifyClient.RegisterRecvClusterInstCache(&ClusterInstCache)
	notifyClient.RegisterRecvAppCache(&AppCache)
	notifyClient.RegisterRecvCloudletCache(&CloudletCache)
	notifyClient.RegisterRecvAutoProvPolicyCache(&AutoProvPoliciesCache)
	SettingsCache.SetUpdatedCb(settingsCb)
	VMPoolInfoCache.SetUpdatedCb(vmPoolInfoCb)
	CloudletCache.SetUpdatedCb(cloudletCb)
	// register to send metrics
	MetricSender = notify.NewMetricSend()
	notifyClient.RegisterSend(MetricSender)
	edgeproto.InitAlertCache(&AlertCache)
	notifyClient.RegisterSendAlertCache(&AlertCache)
	// register to send cloudletInfo, to receive appinst/clusterinst/cloudlet notifications from crm
	edgeproto.InitCloudletInfoCache(&CloudletInfoCache)
	notifyClient.RegisterSendCloudletInfoCache(&CloudletInfoCache)

	nodeMgr.RegisterClient(notifyClient)
	notifyClient.RegisterSendAllRecv(&sendAllRecv{})

	notifyClient.Start()

	cloudletInfo := edgeproto.CloudletInfo{
		Key: cloudletKey,
	}

	// Send state INIT to get cloudlet obj from crm
	cloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_INIT
	CloudletInfoCache.Update(ctx, &cloudletInfo, 0)

	var cloudlet edgeproto.Cloudlet

	// Fetch cloudlet cache from controller->crm->shepherd
	// This also ensures that cloudlet is up before we start collecting metrics
	log.SpanLog(ctx, log.DebugLevelInfo, "wait for cloudlet cache", "key", cloudletKey)
	select {
	case <-cloudletWait:
		if !CloudletCache.Get(&cloudletKey, &cloudlet) {
			log.FatalLog("failed to fetch cloudlet cache from controller")
		}
	case <-time.After(CRMTimeout):
		log.FatalLog("Timed out waiting for cloudlet cache from controller")
	}
	log.SpanLog(ctx, log.DebugLevelInfo, "fetched cloudlet cache from controller", "cloudlet", cloudlet)

	if cloudlet.PlatformType == edgeproto.PlatformType_PLATFORM_TYPE_VM_POOL {
		if cloudlet.VmPool == "" {
			log.FatalLog("Cloudlet is missing VM pool name")
		}
		vmPoolKey := edgeproto.VMPoolKey{
			Name:         cloudlet.VmPool,
			Organization: cloudlet.Key.Organization,
		}
		var vmPool edgeproto.VMPool
		if !VMPoolCache.Get(&vmPoolKey, &vmPool) {
			log.FatalLog("failed to fetch vm pool cache from controller")
		}
	}

	pc := pf.PlatformConfig{
		CloudletKey:    &cloudletKey,
		VaultAddr:      nodeMgr.VaultAddr,
		Region:         *region,
		EnvVars:        cloudlet.EnvVar,
		DeploymentTag:  *deploymentTag,
		PhysicalName:   *physicalName,
		AppDNSRoot:     *appDNSRoot,
		ChefServerPath: *chefServerPath,
	}

	err = myPlatform.Init(ctx, &pc)
	if err != nil {
		log.FatalLog("Failed to initialize platform", "platformName", platformName, "err", err)
	}
	workerMap = make(map[string]*ClusterWorker)
	workerMapMutex = &sync.Mutex{}
	vmAppWorkerMap = make(map[string]*AppInstWorker)
	// LB metrics are not supported in fake mode
	InitProxyScraper()
	if myPlatform.GetType() != "fake" {
		StartProxyScraper()
	}
	InitPlatformMetrics()

	// Send state READY to get AppInst/ClusterInst objs from crm
	cloudletInfo.State = edgeproto.CloudletState_CLOUDLET_STATE_READY
	CloudletInfoCache.Update(ctx, &cloudletInfo, 0)

	log.SpanLog(ctx, log.DebugLevelMetrics, "Ready")
}

func stop() {
	if notifyClient != nil {
		notifyClient.Stop()
	}
	log.FinishTracer()
}

type sendAllRecv struct{}

func (s *sendAllRecv) RecvAllStart() {}

func (s *sendAllRecv) RecvAllEnd(ctx context.Context) {
	targetFileWorkers.NeedsWork(ctx, targetsFileWorkerKey)
}
