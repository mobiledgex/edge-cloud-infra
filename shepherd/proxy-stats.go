package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/dockermgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	ssh "github.com/mobiledgex/golang-ssh"
)

var ProxyMap map[string]ProxyScrapePoint
var ProxyMutex sync.Mutex

const (
	// TCP stat names in envoy
	envoyTcpClusterName = "cluster.backend"
	envoyTcpActive      = "upstream_cx_active"
	envoyTcpTotal       = "upstream_cx_total"
	envoyTcpDropped     = "upstream_cx_connect_fail" // this one might not be right/enough
	envoyTcpBytesSent   = "upstream_cx_tx_bytes_total"
	envoyTcpBytesRecvd  = "upstream_cx_rx_bytes_total"
	envoyTcpSessionTime = "upstream_cx_length_ms"

	// UDP stat names in envoy
	// So tx, and rx is from the point of view of envoy to the backend
	// but we want to give the POV of the backend to the client so we flip these.
	envoyUdpClusterName             = "cluster.udp_backend"
	envoyUdpRecvBytes               = "upstream_cx_tx_bytes_total"
	envoyUdpSentBytes               = "upstream_cx_rx_bytes_total"
	envoyUdpOverflow                = "upstream_cx_overflow"     // # number of datagrams dropped due to hitting the max connections limit
	envoyUdpNoneHealthy             = "upstream_cx_none_healthy" // # of datagrams dropped due to no healthy hosts
	envoyUdpSessionAdditionalPrefix = "udp."
	envoyUdpRecvDatagrams           = "sess_tx_datagrams"
	envoyUdpSentDatagrams           = "sess_rx_datagrams"
	envoyUdpRecvErrs                = "sess_tx_errors"
	envoyUdpSentErrs                = "sess_rx_errors"

	envoyUnseen = "No recorded values"

	maxReconnectSkipDuration = 5 * time.Minute // try to re-connect every 5 mins
)

// no support for const slices
var envoyHistogramBuckets = []string{"P0", "P25", "P50", "P75", "P90", "P95", "P99", "P99.5", "P99.9", "P100"}

type ProxyScrapePoint struct {
	Key                edgeproto.AppInstKey
	FailedChecksCount  int
	App                string
	Cluster            string
	ClusterOrg         string
	TcpPorts           []int32
	UdpPorts           []int32
	LastConnectAttempt time.Time
	Client             ssh.Client
	ProxyContainer     string
	ListenIP           string
}

type DockerNetworkType string

const (
	DockerNetworkHost   DockerNetworkType = "host"
	DockerNetworkBridge DockerNetworkType = "bridge"
)

func InitProxyScraper() {
	ProxyMap = make(map[string]ProxyScrapePoint)
}

func StartProxyScraper(done chan bool) {
	if ProxyMap == nil {
		return
	}
	go ProxyScraper(done)
}

// Figure out envoy proxy container name and network type
func getProxyContainerAndNetworkType(ctx context.Context, scrapePoint ProxyScrapePoint) (string, DockerNetworkType, error) {
	pfType := pf.GetType(*platformName)
	log.SpanLog(ctx, log.DebugLevelMetrics, "getProxyContainerName", "type", pfType)
	if pfType == "fake" {
		return "fakeEnvoy", DockerNetworkHost, nil
	}
	container := proxy.GetEnvoyContainerName(scrapePoint.App)
	request := fmt.Sprintf("docker exec %s echo hello", container)
	resp, err := scrapePoint.Client.Output(request)
	if err != nil && strings.Contains(resp, "No such container") {
		// try the docker name if it fails
		container = proxy.GetEnvoyContainerName(dockermgmt.GetContainerName(&scrapePoint.Key.AppKey))
		request = fmt.Sprintf("docker exec %s echo hello", container)
		resp, err = scrapePoint.Client.Output(request)
		// Perhaps this is nginx
		if err != nil && strings.Contains(resp, "No such container") {
			container = "nginx"
			request := fmt.Sprintf("docker exec %s echo hello", scrapePoint.App)
			resp, err = scrapePoint.Client.Output(request)
		}
	}
	if err != nil {
		log.ForceLogSpan(log.SpanFromContext(ctx))
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to find envoy proxy for app", "scrapepoint", scrapePoint.Key, "err", err, "resp", resp)
		return "", "", err
	}

	log.SpanLog(ctx, log.DebugLevelMetrics, "finding network type for container", "container", container)

	out, err := scrapePoint.Client.Output("docker inspect -f '{{ .NetworkSettings.Networks }}' " + container)
	if err != nil {
		return "", DockerNetworkHost, fmt.Errorf("Unable to find proxy docker network type - %s, %v", out, err)
	}
	if strings.Contains(out, "host:") {
		log.SpanLog(ctx, log.DebugLevelMetrics, "docker host network", "container", container)
		return container, DockerNetworkHost, nil
	}
	// all proxies are started in host mode, but existing proxies may exist running in bridge
	log.SpanLog(ctx, log.DebugLevelMetrics, "docker bridge network - legacy case", "container", container)
	return container, DockerNetworkBridge, nil
}

// Init cluster client for a scrape point
func initClient(ctx context.Context, app *edgeproto.App, appInst *edgeproto.AppInst, clusterInst *edgeproto.ClusterInst, scrapePoint *ProxyScrapePoint) error {
	var err error

	// record last connection attempt
	scrapePoint.LastConnectAttempt = time.Now()
	if app.Deployment == cloudcommon.DeploymentTypeVM && app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER {
		scrapePoint.Client, err = myPlatform.GetVmAppRootLbClient(ctx, &appInst.Key)
		if err != nil {
			// If we cannot get a platform client no point in trying to get metrics
			log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to acquire platform client", "VmApp", appInst.Key, "error", err)
			return err
		}
	} else {
		scrapePoint.Client, err = myPlatform.GetClusterPlatformClient(ctx, clusterInst, cloudcommon.ClientTypeRootLB)
		if err != nil {
			// If we cannot get a platform client no point in trying to get metrics
			log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to acquire platform client", "cluster", clusterInst.Key, "error", err)
			return err
		}
	}
	// Now that we have a client - figure out what container name we should ping
	var netType DockerNetworkType
	scrapePoint.ProxyContainer, netType, err = getProxyContainerAndNetworkType(ctx, *scrapePoint)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to find envoy proxy for app", "scrapepoint", scrapePoint.Key,
			"LastConnectAttempt", scrapePoint.LastConnectAttempt, "err", err)
		scrapePoint.Client.StopPersistentConn()
		return err
	}
	if netType == DockerNetworkHost {
		scrapePoint.ListenIP = infracommon.GetUniqueLoopbackIp(ctx, appInst.MappedPorts)
	} else {
		// legacy case
		scrapePoint.ListenIP = cloudcommon.ProxyMetricsDefaultListenIP
	}
	return nil
}

func CollectProxyStats(ctx context.Context, appInst *edgeproto.AppInst) string {
	// ignore apps not exposed to the outside world as they don't have a envoy/nginx proxy
	app := edgeproto.App{}
	found := AppCache.Get(&appInst.Key.AppKey, &app)
	if !found {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find app", "app", appInst.Key.AppKey.Name)
		return ""
	}
	if !shepherd_common.ShouldRunEnvoy(&app, appInst) {
		return ""
	}
	ProxyMapKey := shepherd_common.GetProxyKey(appInst.GetKey())
	// add/remove from the list of proxy endpoints to hit
	if appInst.State == edgeproto.TrackedState_READY {
		// if we already have this in the map, don't create a new one
		ProxyMutex.Lock()
		if scrapePoint, found := ProxyMap[ProxyMapKey]; found {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Key already exists", "key", ProxyMapKey,
				"scrape", scrapePoint.Key)
			// Re-create the scrape point if the client is not initialized
			if !clientReady(scrapePoint) {
				delete(ProxyMap, ProxyMapKey)
			} else {
				ProxyMutex.Unlock()
				return ""
			}
		}
		ProxyMutex.Unlock()

		scrapePoint := ProxyScrapePoint{
			Key:        appInst.Key,
			App:        k8smgmt.NormalizeName(appInst.Key.AppKey.Name),
			Cluster:    appInst.Key.ClusterInstKey.ClusterKey.Name,
			ClusterOrg: appInst.Key.ClusterInstKey.Organization,
			TcpPorts:   make([]int32, 0),
			UdpPorts:   make([]int32, 0),
		}

		for _, p := range appInst.MappedPorts {
			if p.Proto == dme.LProto_L_PROTO_TCP {
				scrapePoint.TcpPorts = append(scrapePoint.TcpPorts, p.InternalPort)
			}
			if p.Proto == dme.LProto_L_PROTO_UDP && !p.Nginx {
				scrapePoint.UdpPorts = append(scrapePoint.UdpPorts, p.InternalPort)
			}
		}
		// Don't need to scrape anything if no ports are trackable
		if len(scrapePoint.TcpPorts) == 0 && len(scrapePoint.UdpPorts) == 0 {
			log.SpanLog(ctx, log.DebugLevelMetrics, "No ports to scrape", "key", appInst.Key)
			return ""
		}

		clusterInst := edgeproto.ClusterInst{}
		found := ClusterInstCache.Get(appInst.ClusterInstKey(), &clusterInst)
		// lb vm apps continue anyway
		if !found && !(app.Deployment == cloudcommon.DeploymentTypeVM && app.AccessType == edgeproto.AccessType_ACCESS_TYPE_LOAD_BALANCER) {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find clusterInst for AppInst", "AppInstKey", appInst.Key)
			return ""
		}
		err := initClient(ctx, &app, appInst, &clusterInst, &scrapePoint)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to init platform client - do it later", "app", appInst.Key, "cluster", clusterInst.Key, "error", err)
		}
		// If this was created between last check and now
		ProxyMutex.Lock()
		if scrape, found := ProxyMap[ProxyMapKey]; found {
			ProxyMutex.Unlock()
			log.SpanLog(ctx, log.DebugLevelMetrics, "Instance was created somewhere else", "key", ProxyMapKey, "scrape", scrape.Key)
			if scrapePoint.Client != nil {
				scrapePoint.Client.StopPersistentConn()
			}
			return ""
		}
		log.SpanLog(ctx, log.DebugLevelMetrics, "Creating Proxy Stats", "app inst", appInst.Key, "scrape point key", ProxyMapKey, "container", scrapePoint.ProxyContainer)
		ProxyMap[ProxyMapKey] = scrapePoint
		ProxyMutex.Unlock()
		return ProxyMapKey
	}
	// if the app is anything other than ready, stop tracking it if it exists
	ProxyMutex.Lock()
	defer ProxyMutex.Unlock()
	scrapePoint, found := ProxyMap[ProxyMapKey]
	if !found {
		return ""
	}

	log.SpanLog(ctx, log.DebugLevelMetrics, "Proxy ScrapePoint deleted", "key", ProxyMapKey, "app inst", scrapePoint.Key)
	if scrapePoint.Client != nil {
		// Close the ssh session
		scrapePoint.Client.StopPersistentConn()
	}
	delete(ProxyMap, ProxyMapKey)
	return ProxyMapKey
}

func copyMapValues() []ProxyScrapePoint {
	ProxyMutex.Lock()
	scrapePoints := make([]ProxyScrapePoint, 0, len(ProxyMap))
	for _, value := range ProxyMap {
		scrapePoints = append(scrapePoints, value)
	}
	ProxyMutex.Unlock()
	return scrapePoints
}

func getProxyScrapePoint(key string) *ProxyScrapePoint {
	ProxyMutex.Lock()
	defer ProxyMutex.Unlock()
	scrapePoint, found := ProxyMap[key]
	if !found {
		return nil
	}
	return &scrapePoint
}

func updateProxyScrapeClient(key edgeproto.AppInstKey) {
	span := log.StartSpan(log.DebugLevelMetrics, "update-proxyClient")
	defer span.Finish()
	log.SetTags(span, cloudletKey.GetTags())
	span.SetTag("app", key.AppKey.Name)
	span.SetTag("cluster", key.ClusterInstKey.ClusterKey.Name)
	span.SetTag("cloudlet", key.ClusterInstKey.CloudletKey.Name)
	ctx := log.ContextWithSpan(context.Background(), span)

	// find appInst in the cache
	appInst := edgeproto.AppInst{}
	found := AppInstCache.Get(&key, &appInst)
	if !found {
		return
	}
	// trigger re-init of the client
	scrapeKey := CollectProxyStats(ctx, &appInst)
	log.SpanLog(ctx, log.DebugLevelMetrics, "Re-init client for proxy scrape", "app", appInst,
		"key", scrapeKey)
}

func clientReady(scrape ProxyScrapePoint) bool {
	if scrape.Client != nil && scrape.ProxyContainer != "" {
		return true
	}
	return false
}

func shouldUpdateFailedClient(scrape *ProxyScrapePoint) bool {
	var skipDuration time.Duration

	// skip either 5(5+current) intervals, or maxSkipDuration, whichever is shortest
	if 6*settings.ShepherdMetricsCollectionInterval.TimeDuration() > maxReconnectSkipDuration {
		skipDuration = maxReconnectSkipDuration
	} else {
		skipDuration = 6 * settings.ShepherdMetricsCollectionInterval.TimeDuration()
	}

	if time.Since(scrape.LastConnectAttempt) > skipDuration {
		return true
	}
	return false
}

func checkAndSetLastPushLbMetrics(lastPushed *time.Time) bool {
	if time.Since(*lastPushed) > settings.ShepherdMetricsCollectionInterval.TimeDuration() {
		*lastPushed = time.Now()
		return true
	}
	return false
}

func ProxyScraper(done chan bool) {
	var metricsLastPushed time.Time
	for {
		// check if there are any new apps we need to start/stop scraping for
		select {
		case <-time.After(metricsScrapingInterval):
			scrapePoints := copyMapValues()
			for _, v := range scrapePoints {
				if !clientReady(v) {
					if shouldUpdateFailedClient(&v) {
						// Update this in the background
						go updateProxyScrapeClient(v.Key)
					}
					// no need to actually collect metrics
					continue
				}
				span := log.StartSpan(log.DebugLevelSampled, "send-metric")
				log.SetTags(span, cloudletKey.GetTags())
				span.SetTag("cluster", v.Cluster)
				ctx := log.ContextWithSpan(context.Background(), span)

				metrics, err := QueryProxy(ctx, &v)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelMetrics, "Error retrieving proxy metrics", "appinst", v.App, "error", err.Error())
				} else if checkAndSetLastPushLbMetrics(&metricsLastPushed) {
					log.DebugLog(log.DebugLevelInfo, "Pushing proxy metrics")
					// send to crm->controller->influx
					influxData := MarshallTcpProxyMetric(v, metrics)
					influxData = append(influxData, MarshallUdpProxyMetric(v, metrics)...)
					for _, datapoint := range influxData {
						MetricSender.Update(context.Background(), datapoint)
					}
				}
				span.Finish()
			}
		case <-done:
			// process killed/interrupted, so quit
			return
		}
	}
}

func getProxyMetricsRequest(target *ProxyScrapePoint, path string) string {
	execStr := ""
	if target.ListenIP == cloudcommon.ProxyMetricsDefaultListenIP {
		// legacy case, need to exec into the container
		execStr = "docker exec " + target.ProxyContainer
	}
	return fmt.Sprintf("%s curl -s -S http://%s:%d/%s", execStr, target.ListenIP, cloudcommon.ProxyMetricsPort, path)
}

func QueryProxy(ctx context.Context, scrapePoint *ProxyScrapePoint) (*shepherd_common.ProxyMetrics, error) {
	if scrapePoint.Client == nil {
		return nil, fmt.Errorf("ScrapePoint client is not initialized")
	}
	// query envoy
	if scrapePoint.ProxyContainer == "nginx" {
		return QueryNginx(ctx, scrapePoint) //if envoy isn't there(for legacy apps) query nginx
	}
	request := getProxyMetricsRequest(scrapePoint, "stats")
	resp, err := scrapePoint.Client.OutputWithTimeout(request, shepherd_common.ShepherdSshConnectTimeout)
	if err != nil {
		log.ForceLogSpan(log.SpanFromContext(ctx))
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to run request", "request", request, "err", err.Error(), "resp", resp)
		return nil, err
	}
	metrics := &shepherd_common.ProxyMetrics{Nginx: false}
	respMap := parseEnvoyResp(ctx, resp)
	err = envoyTcpConnections(ctx, respMap, scrapePoint.TcpPorts, metrics)
	if err != nil {
		return nil, fmt.Errorf("Error parsing response: %v", err)
	}
	err = envoyUdpConnections(ctx, respMap, scrapePoint.UdpPorts, metrics)
	if err != nil {
		return nil, fmt.Errorf("Error parsing response: %v", err)
	}
	return metrics, nil
}

func envoyTcpConnections(ctx context.Context, respMap map[string]string, ports []int32, metrics *shepherd_common.ProxyMetrics) error {
	var err error
	var droppedVal uint64
	metrics.EnvoyTcpStats = make(map[int32]shepherd_common.TcpConnectionsMetric)
	for _, port := range ports {
		new := shepherd_common.TcpConnectionsMetric{}
		//active, accepts, handled conn, bytes sent/recvd
		envoyCluster := envoyTcpClusterName + strconv.Itoa(int(port)) + "."
		activeSearch := envoyCluster + envoyTcpActive
		droppedSearch := envoyCluster + envoyTcpDropped
		totalSearch := envoyCluster + envoyTcpTotal
		bytesSentSearch := envoyCluster + envoyTcpBytesSent
		bytesRecvdSearch := envoyCluster + envoyTcpBytesRecvd
		sessionTimeSearch := envoyCluster + envoyTcpSessionTime
		new.ActiveConn, err = getUIntStat(respMap, activeSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy active connections stats: %v", err)
		}
		new.Accepts, err = getUIntStat(respMap, totalSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy accepts connections stats: %v", err)
		}
		droppedVal, err = getUIntStat(respMap, droppedSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy handled connections stats: %v", err)
		}
		new.HandledConn = new.Accepts - droppedVal
		new.BytesSent, err = getUIntStat(respMap, bytesSentSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy bytes_sent connections stats: %v", err)
		}
		new.BytesRecvd, err = getUIntStat(respMap, bytesRecvdSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy bytes_recvd connections stats: %v", err)
		}

		// session time histogram
		var sessionTimeHistogram map[string]float64
		sessionTimeHistogram, err = getHistogramIntStats(respMap, sessionTimeSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy session time connections stats: %v", err)
		}
		new.SessionTime = sessionTimeHistogram
		metrics.Ts, _ = types.TimestampProto(time.Now())
		metrics.EnvoyTcpStats[port] = new
	}
	return nil
}

func envoyUdpConnections(ctx context.Context, respMap map[string]string, ports []int32, metrics *shepherd_common.ProxyMetrics) error {
	var err error
	metrics.EnvoyUdpStats = make(map[int32]shepherd_common.UdpConnectionsMetric)
	for _, port := range ports {
		new := shepherd_common.UdpConnectionsMetric{}

		envoyCluster := envoyUdpClusterName + strconv.Itoa(int(port)) + "."
		recvBytesSearch := envoyCluster + envoyUdpRecvBytes
		sentBytesSearch := envoyCluster + envoyUdpSentBytes
		overflowSearch := envoyCluster + envoyUdpOverflow
		missedSearch := envoyCluster + envoyUdpNoneHealthy

		envoyCluster = envoyCluster + envoyUdpSessionAdditionalPrefix
		recvDatagramsSearch := envoyCluster + envoyUdpRecvDatagrams
		sentDatagramsSearch := envoyCluster + envoyUdpSentDatagrams
		recvErrSearch := envoyCluster + envoyUdpRecvErrs
		sentErrSearch := envoyCluster + envoyUdpSentErrs

		new.RecvBytes, err = getUIntStat(respMap, recvBytesSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy recvBytes connections stats: %v", err)
		}
		new.SentBytes, err = getUIntStat(respMap, sentBytesSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy sentBytes connections stats: %v", err)
		}
		new.Overflow, err = getUIntStat(respMap, overflowSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy overflow connections stats: %v", err)
		}
		new.Missed, err = getUIntStat(respMap, missedSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy missed connections stats: %v", err)
		}
		new.RecvDatagrams, err = getUIntStat(respMap, recvDatagramsSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy recvDatagrams connections stats: %v", err)
		}
		new.SentDatagrams, err = getUIntStat(respMap, sentDatagramsSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy sentDatagrams connections stats: %v", err)
		}
		new.RecvErrs, err = getUIntStat(respMap, recvErrSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy recvErrs connections stats: %v", err)
		}
		new.SentErrs, err = getUIntStat(respMap, sentErrSearch)
		if err != nil {
			return fmt.Errorf("Error retrieving envoy sentErrs connections stats: %v", err)
		}
		metrics.Ts, _ = types.TimestampProto(time.Now())
		metrics.EnvoyUdpStats[port] = new
	}
	return nil
}

// converts the envoy stats page into a map for easy reading
func parseEnvoyResp(ctx context.Context, resp string) map[string]string {
	lines := strings.Split(resp, "\n")
	newMap := make(map[string]string)
	for _, line := range lines {
		keyValPair := strings.Split(line, ": ")
		if len(keyValPair) == 2 {
			newMap[keyValPair[0]] = keyValPair[1]
		} else {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Could not parse line", "line", line)
		}
	}
	return newMap
}

// converts envoy cluster ouptut into a map
// Example of cluster output string: backend8008::10.192.1.2:8008::health_flags::healthy
func parseEnvoyClusterResp(ctx context.Context, resp string) map[string]string {
	lines := strings.Split(resp, "\n")
	newMap := make(map[string]string)
	for _, line := range lines {
		items := strings.Split(line, "::")
		cnt := len(items)
		// at least 3
		if cnt < 3 {
			continue
		}
		// First is unique with respect to port, next to last is the keyname and last is value
		newMap[items[0]+"::"+items[cnt-2]] = items[cnt-1]
	}
	return newMap
}

// this function only retrieves stats from envoy that are expected to be int values
func getUIntStat(respMap map[string]string, statName string) (uint64, error) {
	stat, exists := respMap[statName]
	if !exists {
		return 0, fmt.Errorf("stat not found: %s", statName)
	}
	val, err := strconv.ParseUint(stat, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Error retrieving stats: %v", err)
	}
	return val, nil
}

// parses envoy histograms into a map form. Envoy histograms look like this:
// cluster.backend4321.upstream_cx_length_ms: P0(nan,2) P25(nan,5.1) P50(nan,11) P75(nan,105) P90(nan,182) P95(nan,186) P99(nan,189.2) P99.5(nan,189.6) P99.9(nan,189.92) P100(nan,190)
func getHistogramIntStats(respMap map[string]string, statName string) (map[string]float64, error) {
	histogramStr, exists := respMap[statName]
	if !exists {
		return nil, fmt.Errorf("stat not found: %s", statName)
	}
	histogram := make(map[string]float64)
	for _, v := range envoyHistogramBuckets {
		histogram[v] = 0
	}
	// if theres no connections yet to measure default everything to zeros
	if histogramStr == envoyUnseen {
		return histogram, nil
	}
	buckets := strings.Split(histogramStr, " ")
	if len(buckets) != len(envoyHistogramBuckets) {
		return nil, fmt.Errorf("Error parsing histogram")
	}
	for i, bucket := range buckets {
		// P0(nan,3300)
		// check if the percentile matches
		if !strings.HasPrefix(bucket, envoyHistogramBuckets[i]) {
			return nil, fmt.Errorf("Error parsing histogram")
		}
		start := strings.Index(bucket, ",")
		end := strings.Index(bucket, ")")
		if start == -1 || end == -1 {
			return nil, fmt.Errorf("Error parsing histogram")
		}
		start = start + 1
		val, err := strconv.ParseFloat(bucket[start:end], 64)
		if err != nil {
			return nil, fmt.Errorf("Error parsing histogram: %v", err)
		}
		histogram[envoyHistogramBuckets[i]] = val
	}
	return histogram, nil
}

func QueryNginx(ctx context.Context, scrapePoint *ProxyScrapePoint) (*shepherd_common.ProxyMetrics, error) {
	if scrapePoint.Client == nil {
		return nil, fmt.Errorf("ScrapePoint client is not initialized")
	}
	// build the query
	request := getProxyMetricsRequest(scrapePoint, "nginx_metrics")
	resp, err := scrapePoint.Client.OutputWithTimeout(request, shepherd_common.ShepherdSshConnectTimeout)
	// if this is the first time, or the container got restarted, install curl (for old deployments)
	if strings.Contains(resp, "executable file not found") {
		log.SpanLog(ctx, log.DebugLevelInfra, "Installing curl onto docker container ", "Container", scrapePoint.App)
		installer := fmt.Sprintf("docker exec %s apt-get update; docker exec %s apt-get --assume-yes install curl", scrapePoint.App, scrapePoint.App)
		resp, err = scrapePoint.Client.Output(installer)
		if err != nil {
			return nil, fmt.Errorf("can't install curl on nginx container %s, %s, %v", *name, resp, err)
		}
		// now retry curling
		resp, err = scrapePoint.Client.OutputWithTimeout(request, shepherd_common.ShepherdSshConnectTimeout)
	}
	if err != nil {
		log.ForceLogSpan(log.SpanFromContext(ctx))
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to run request", "request", request, "err", err.Error(), "resp", "resp")
		return nil, err
	}
	metrics := &shepherd_common.ProxyMetrics{Nginx: true}
	err = parseNginxResp(resp, metrics)
	if err != nil {
		return nil, fmt.Errorf("Error parsing response: %v", err)
	}
	return metrics, nil
}

// view here: https://github.com/nginxinc/nginx-prometheus-exporter/blob/29ec94bdee98668e358efac7316bd8d12b05a130/client/nginx.go#L70
func parseNginxResp(resp string, metrics *shepherd_common.ProxyMetrics) error {
	// sometimes the response lines get cycled around, so break it up based on the start of the actual content
	trimmedResp := strings.Split(resp, "Active connections:")
	if len(trimmedResp) < 2 {
		return fmt.Errorf("Unexpected output format")
	}
	lines := strings.Split(trimmedResp[1], "\n")

	var err error
	//fix this hardcoding later
	if len(lines) < 4 {
		return fmt.Errorf("Unexpected output format")
	}

	// first line has active connections
	stats := strings.Split(lines[0], " ")
	if metrics.ActiveConn, err = strconv.ParseUint(stats[1], 10, 64); err != nil {
		return err
	}
	// third line for accepts, handled, and requests
	stats = strings.Split(lines[2], " ")
	if metrics.Accepts, err = strconv.ParseUint(stats[1], 10, 64); err != nil {
		return err
	}
	if metrics.HandledConn, err = strconv.ParseUint(stats[2], 10, 64); err != nil {
		return err
	}
	if metrics.Requests, err = strconv.ParseUint(stats[3], 10, 64); err != nil {
		return err
	}
	// last line has reading, writing, waiting
	stats = strings.Split(lines[3], " ")
	if metrics.Reading, err = strconv.ParseUint(stats[1], 10, 64); err != nil {
		return err
	}
	if metrics.Writing, err = strconv.ParseUint(stats[3], 10, 64); err != nil {
		return err
	}
	if metrics.Waiting, err = strconv.ParseUint(stats[5], 10, 64); err != nil {
		return err
	}
	metrics.Ts, _ = types.TimestampProto(time.Now())
	return nil
}

func MarshallTcpProxyMetric(scrapePoint ProxyScrapePoint, data *shepherd_common.ProxyMetrics) []*edgeproto.Metric {
	if data.Nginx {
		return []*edgeproto.Metric{MarshallNginxMetric(scrapePoint, data)}
	}
	metricList := make([]*edgeproto.Metric, 0)
	for _, port := range scrapePoint.TcpPorts {
		metric := edgeproto.Metric{}
		metric.Name = "appinst-connections"
		metric.Timestamp = *data.Ts
		metric.AddTag("cloudletorg", cloudletKey.Organization)
		metric.AddTag("cloudlet", cloudletKey.Name)
		metric.AddTag("cluster", scrapePoint.Cluster)
		metric.AddTag("clusterorg", scrapePoint.ClusterOrg)
		metric.AddTag("apporg", scrapePoint.Key.AppKey.Organization)
		metric.AddTag("app", util.DNSSanitize(scrapePoint.Key.AppKey.Name))
		metric.AddTag("ver", util.DNSSanitize(scrapePoint.Key.AppKey.Version))
		metric.AddIntVal("port", uint64(port))

		metric.AddIntVal("active", data.EnvoyTcpStats[port].ActiveConn)
		metric.AddIntVal("accepts", data.EnvoyTcpStats[port].Accepts)
		metric.AddIntVal("handled", data.EnvoyTcpStats[port].HandledConn)
		metric.AddIntVal("bytesSent", data.EnvoyTcpStats[port].BytesSent)
		metric.AddIntVal("bytesRecvd", data.EnvoyTcpStats[port].BytesRecvd)

		//session time historgram
		for k, v := range data.EnvoyTcpStats[port].SessionTime {
			metric.AddDoubleVal(k, v)
		}
		metricList = append(metricList, &metric)
	}
	return metricList
}

func MarshallUdpProxyMetric(scrapePoint ProxyScrapePoint, data *shepherd_common.ProxyMetrics) []*edgeproto.Metric {
	metricList := make([]*edgeproto.Metric, 0)
	for _, port := range scrapePoint.UdpPorts {
		metric := edgeproto.Metric{}
		metric.Name = "appinst-udp"
		metric.Timestamp = *data.Ts
		metric.AddTag("cloudletorg", cloudletKey.Organization)
		metric.AddTag("cloudlet", cloudletKey.Name)
		metric.AddTag("cluster", scrapePoint.Cluster)
		metric.AddTag("clusterorg", scrapePoint.ClusterOrg)
		metric.AddTag("apporg", scrapePoint.Key.AppKey.Organization)
		metric.AddTag("app", util.DNSSanitize(scrapePoint.Key.AppKey.Name))
		metric.AddTag("ver", util.DNSSanitize(scrapePoint.Key.AppKey.Version))
		metric.AddIntVal("port", uint64(port))

		metric.AddIntVal("bytesSent", data.EnvoyUdpStats[port].SentBytes)
		metric.AddIntVal("bytesRecvd", data.EnvoyUdpStats[port].RecvBytes)
		metric.AddIntVal("datagramsSent", data.EnvoyUdpStats[port].SentDatagrams)
		metric.AddIntVal("datagramsRecvd", data.EnvoyUdpStats[port].RecvDatagrams)
		metric.AddIntVal("sentErrs", data.EnvoyUdpStats[port].SentErrs)
		metric.AddIntVal("recvErrs", data.EnvoyUdpStats[port].RecvErrs)
		metric.AddIntVal("overflow", data.EnvoyUdpStats[port].Overflow)
		metric.AddIntVal("missed", data.EnvoyUdpStats[port].Missed)

		metricList = append(metricList, &metric)
	}
	return metricList
}

func MarshallNginxMetric(scrapePoint ProxyScrapePoint, data *shepherd_common.ProxyMetrics) *edgeproto.Metric {
	RemoveShepherdMetrics(data)
	metric := edgeproto.Metric{}
	metric.Name = "appinst-connections"
	metric.Timestamp = *data.Ts
	metric.AddTag("cloudletorg", cloudletKey.Organization)
	metric.AddTag("cloudlet", cloudletKey.Name)
	metric.AddTag("cluster", scrapePoint.Cluster)
	metric.AddTag("clusterorg", scrapePoint.ClusterOrg)
	metric.AddTag("apporg", scrapePoint.Key.AppKey.Organization)
	metric.AddTag("app", util.DNSSanitize(scrapePoint.Key.AppKey.Name))
	metric.AddTag("ver", util.DNSSanitize(scrapePoint.Key.AppKey.Version))

	metric.AddIntVal("active", data.ActiveConn)
	metric.AddIntVal("accepts", data.Accepts)
	metric.AddIntVal("handled", data.HandledConn)
	return &metric
}

func RemoveShepherdMetrics(data *shepherd_common.ProxyMetrics) {
	data.ActiveConn = data.ActiveConn - 1
	data.Accepts = data.Accepts - data.Requests
	data.HandledConn = data.HandledConn - data.Requests
}
