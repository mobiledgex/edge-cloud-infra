package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	baselog "log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"text/template"

	"github.com/mobiledgex/edge-cloud-infra/autoprov/autorules"
	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/prommgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"gopkg.in/yaml.v2"
)

const HealthCheckRulesPrefix = "healthcheck"

var CloudletPrometheusAddr = "0.0.0.0:" + intprocess.CloudletPrometheusPort

var promTargetTemplate *template.Template
var targetsLock sync.Mutex

var promTargetT = `
{
	"targets": ["{{.MetricsProxyAddr}}"],
	"labels": {
		"` + edgeproto.AppKeyTagName + `": "{{.Key.AppKey.Name}}",
		"` + edgeproto.AppKeyTagVersion + `": "{{.Key.AppKey.Version}}",
		"` + edgeproto.AppKeyTagOrganization + `": "{{.Key.AppKey.Organization}}",
		"` + edgeproto.ClusterKeyTagName + `": "{{.Key.ClusterInstKey.ClusterKey.Name}}",
		"` + edgeproto.ClusterInstKeyTagOrganization + `": "{{.Key.ClusterInstKey.Organization}}",
		"` + edgeproto.CloudletKeyTagName + `": "{{.Key.ClusterInstKey.CloudletKey.Name}}",
		"` + edgeproto.CloudletKeyTagOrganization + `": "{{.Key.ClusterInstKey.CloudletKey.Organization}}",
		"__metrics_path__":"{{.EnvoyMetricsPath}}"
	}
}`

var promHealthCheckAlerts = `groups:
- name: StaticRules
  rules:
  - alert: ` + cloudcommon.AlertAppInstDown + `
    expr: up == 0
    for: 15s
    labels:
      ` + cloudcommon.AlertHealthCheckStatus + ": " + strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE)) + `
  - alert: ` + cloudcommon.AlertAppInstDown + `
    expr: envoy_cluster_health_check_healthy == 0
    labels:
      ` + cloudcommon.AlertHealthCheckStatus + ": " + strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL))

type targetData struct {
	MetricsProxyAddr string
	Key              edgeproto.AppInstKey
	EnvoyMetricsPath string
}

func init() {
	promTargetTemplate = template.Must(template.New("prometheustarget").Parse(promTargetT))
}

func getAppInstPrometheusTargetString(appInstKey *edgeproto.AppInstKey) (string, error) {
	host := *metricsAddr
	switch *platformName {
	case "PLATFORM_TYPE_EDGEBOX":
		fallthrough
	case "PLATFORM_TYPE_FAKEINFRA":
		host = "host.docker.internal:9091"
	}
	target := targetData{
		MetricsProxyAddr: host,
		Key:              *appInstKey,
		EnvoyMetricsPath: "/metrics/" + shepherd_common.GetProxyKey(appInstKey),
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
func writePrometheusTargetsFile(ctx context.Context, key interface{}) {
	targetsLock.Lock()
	defer targetsLock.Unlock()
	var targets = "["
	proxyScrapePoints := copyMapValues()
	for _, val := range proxyScrapePoints {
		if targets != "[" {
			targets += ","
		}
		promTargetJson, err := getAppInstPrometheusTargetString(&val.Key)
		if err == nil {
			targets += promTargetJson
		}
	}
	targets += "]"
	err := ioutil.WriteFile(*promTargetsFile, []byte(targets), 0644)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to write prom targets file", "file", *promTargetsFile, "err", err)
	}
	if runtime.GOOS == "darwin" {
		// probably because of the way docker uses VMs on mac,
		// the file watch doesn't detect changes done to the targets
		// file in the host.
		cmd := exec.Command("docker", "exec", intprocess.PrometheusContainer, "touch", *promTargetsFile)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to touch prom targets file in container to trigger refresh in Prometheus", "out", string(out), "err", err)
		}
	}
}

// Delete Alert file and reload rules
func deleteCloudletPrometheusAlertFile(ctx context.Context, file string) error {
	// remove alerting rules
	err := os.Remove(file)
	if err != nil {
		return err
	}
	// need to force prometheus to re-read the rules file
	reloadCloudletProm(ctx)
	return nil
}

// Write prometheus rules file and reload rules
func writeCloudletPrometheusAlerts(ctx context.Context, file string, alertsBuf []byte) error {
	// write alerting rules
	err := ioutil.WriteFile(file, alertsBuf, 0644)
	if err != nil {
		return err
	}
	// need to force prometheus to re-read the rules file
	reloadCloudletProm(ctx)
	return nil
}

func reloadCloudletProm(ctx context.Context) {
	resp, err := http.Post("http://0.0.0.0:9092/-/reload", "", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to reload prometheus", "err", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to read prometheus reload response", "code", resp.StatusCode, "err", err)
		} else {
			log.SpanLog(ctx, log.DebugLevelInfo, "Failed to reload prometheus", "code", resp.StatusCode, "err", string(data))
		}
	}
}

func targetsList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>%s</h1>", "List all targets")
	targets := copyMapValues()
	for ii, v := range targets {
		fmt.Fprintf(w, "<h1>Target %d</h1><div>%s</div>", ii, shepherd_common.GetProxyKey(&v.Key))
	}
}

func metricsProxy(w http.ResponseWriter, r *http.Request) {
	// Sanity check
	if len(r.URL.Path) < len("/metrics/")+1 {
		return
	}
	app := r.URL.Path[len("/metrics/"):]
	if app != "" {
		// Search ProxyMap for the names
		target := getProxyScrapePoint(app)
		if target.Client == nil {
			// if client is not initialized trigger health-check failure
			http.Error(w, "Client is not initialized", http.StatusInternalServerError)
			return
		}
		if target.ProxyContainer == "nginx" {
			return
		}
		request := fmt.Sprintf("docker exec %s curl -s -S http://127.0.0.1:%d/stats/prometheus", target.ProxyContainer, cloudcommon.ProxyMetricsPort)
		if myPlatform.GetType() == "fake" {
			sock := "/tmp/envoy_" + app + ".sock"
			request = fmt.Sprintf("curl -s --unix-socket %s http:/sock/stats/prometheus", sock)
		}
		resp, err := target.Client.OutputWithTimeout(request, shepherd_common.ShepherdSshConnectTimeout)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(resp))
	}
}

func getAppInstRulesFileName(key edgeproto.AppInstKey) string {
	name := cloudcommon.NormalizeName(key.AppKey.Name)
	return getPrometheusFileName(name)
}

func getPrometheusFileName(name string) string {
	return "/tmp/" + intprocess.PrometheusRulesPrefix + name + ".yml"
}

// Starts Cloudlet Prometheus MetricsProxy thread to serve as a target for metrics
func startPrometheusMetricsProxy(ctx context.Context) error {
	// Init prometheus targets and alert templates
	healthCeckFile := getPrometheusFileName(HealthCheckRulesPrefix)
	err := writeCloudletPrometheusAlerts(ctx, healthCeckFile, []byte(promHealthCheckAlerts))
	if err != nil {
		return fmt.Errorf("Failed to write prometheus rules to %s, err: %s",
			healthCeckFile, err.Error())
	}
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
	return nil
}

func getAutoProvPolicy(ctx context.Context, appInst *edgeproto.AppInst, app *edgeproto.App) (*edgeproto.AutoProvPolicy, bool) {
	for polKey, _ := range app.GetAutoProvPolicys() {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Eval policy", "app", app, "policy", polKey)
		policy := edgeproto.AutoProvPolicy{}
		found := AutoProvPoliciesCache.Get(&polKey, &policy)
		if !found {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find policy", "policy", polKey)
			continue
		}
		// Check if one of the cloudlets in the policy matches ours
		for _, cloudlet := range policy.Cloudlets {
			if cloudletKey.Matches(&cloudlet.Key) {
				return &policy, true
			}
		}
	}
	// Didn't find any policies that should be enacted on this cloudlet
	return nil, false
}

func writePrometheusAlertRuleForAppInst(ctx context.Context, k interface{}) {
	key, ok := k.(edgeproto.AppInstKey)
	if !ok {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unexpected failure, key not AppInstKey", "key", key)
		return
	}

	appInst := edgeproto.AppInst{}
	found := AppInstCache.Get(&key, &appInst)
	if !found || appInst.State != edgeproto.TrackedState_READY {
		log.SpanLog(ctx, log.DebugLevelApi, "delete rules for AppInst", "AppInst", key)
		untrackAppInstByPolicy(key)
		// AppInst is being deleted - delete rules
		fileName := getAppInstRulesFileName(key)
		if err := deleteCloudletPrometheusAlertFile(ctx, fileName); err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to delete prometheus rules", "file", fileName, "err", err)
		}
		return
	}
	// check cluster name if this is a VM App
	app := edgeproto.App{}
	found = AppCache.Get(&appInst.Key.AppKey, &app)
	if !found {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find app", "app", appInst.Key.AppKey.Name)
		return
	}

	log.SpanLog(ctx, log.DebugLevelApi, "write rules for AppInst", "AppInst", key)

	// get any rules for AppInst
	grps := prommgmt.GroupsData{}

	if appInst.Liveness == edgeproto.Liveness_LIVENESS_AUTOPROV {
		// auto-provisioned AppInst, check policy.
		policy, found := getAutoProvPolicy(ctx, &appInst, &app)
		if !found {
			log.SpanLog(ctx, log.DebugLevelMetrics, "No AutoProvPolicy found", "app", app.Key, "cloudlet", appInst.Key.ClusterInstKey.CloudletKey)
		} else {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Apply AutoProvPolicy", "app", app.Key, "cloudlet", appInst.Key.ClusterInstKey.CloudletKey, "policy", policy.Key)
			ruleGrp := autorules.GetAutoUndeployRules(ctx, settings, &app.Key, policy)
			if ruleGrp != nil {
				grps.Groups = append(grps.Groups, *ruleGrp)
			}
			trackAppInstByPolicy(appInst.Key, policy.Key)
		}
	}

	if len(grps.Groups) == 0 {
		log.SpanLog(ctx, log.DebugLevelApi, "no rules for AppInst", "AppInst", key)
		// no rules
		return
	}
	byt, err := yaml.Marshal(grps)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to marshal prom rule groups", "AppInst", appInst.Key, "rules", grps, "err", err)
		return
	}

	fileName := getAppInstRulesFileName(appInst.Key)
	err = writeCloudletPrometheusAlerts(ctx, fileName, byt)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to write prometheus rules", "file", fileName, "err", err)
	}
}

func trackAppInstByPolicy(appInstKey edgeproto.AppInstKey, policyKey edgeproto.PolicyKey) {
	obj := edgeproto.AppInstLookup{
		Key:       appInstKey,
		PolicyKey: policyKey,
	}
	AppInstByAutoProvPolicy.Updated(&obj)
}

// Unfortunately during removal we may not have the policy used, so we walk
// the data to remove any references to the AppInst. This is ok since we should
// only have a small amount of data just for this Cloudlet.
func untrackAppInstByPolicy(appInstKey edgeproto.AppInstKey) {
	s := &AppInstByAutoProvPolicy
	s.Mux.Lock()
	defer s.Mux.Unlock()
	for policyKey, insts := range s.PolicyKeys {
		delete(insts, appInstKey)
		if len(insts) == 0 {
			delete(s.PolicyKeys, policyKey)
		}
	}
}
