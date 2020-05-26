package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	baselog "log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var CloudletPrometheusAddr = "0.0.0.0:" + intprocess.CloudletPrometheusPort

var promTargetTemplate, promAlertTemplate *template.Template

var promTargetT = `
{
	"targets": ["{{.MetricsProxyAddr}}"],
	"labels": {
		"` + cloudcommon.AlertLabelApp + `": "{{.Key.AppKey.Name}}",
		"` + cloudcommon.AlertLabelAppVer + `": "{{.Key.AppKey.Version}}",
		"` + cloudcommon.AlertLabelAppOrg + `": "{{.Key.AppKey.Organization}}",
		"` + cloudcommon.AlertLabelCluster + `": "{{.Key.ClusterInstKey.ClusterKey.Name}}",
		"` + cloudcommon.AlertLabelClusterOrg + `": "{{.Key.ClusterInstKey.Organization}}",
		"` + cloudcommon.AlertLabelCloudlet + `": "{{.Key.ClusterInstKey.CloudletKey.Name}}",
		"` + cloudcommon.AlertLabelCloudletOrg + `": "{{.Key.ClusterInstKey.CloudletKey.Organization}}",
		"__metrics_path__":"{{.EnvoyMetricsPath}}"
	}
}`

var promHealthCheckAlerts = `groups:
- name: StaticRules
  rules:
  - alert: ` + cloudcommon.AlertAppInstDown + `
    expr: up == 0
    for: 15s
    annotations:
      ` + cloudcommon.AlertHealthCheckStatus + ": " + strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_ROOTLB_OFFLINE)) + `
  - alert: ` + cloudcommon.AlertAppInstDown + `
    expr: envoy_cluster_health_check_healthy == 0
    annotations:
      ` + cloudcommon.AlertHealthCheckStatus + ": " + strconv.Itoa(int(edgeproto.HealthCheck_HEALTH_CHECK_FAIL_SERVER_FAIL))

var promAutoProvAlertT = `groups:
- name: ` + cloudcommon.AlertAutoProvDown + `
  rules:
  - alert: ScaleDown
    expr: envoy_cluster_upstream_cx_active{` + cloudcommon.AlertLabelApp + `="{{.AppKey.Name}}",` + cloudcommon.AlertLabelAppVer + `="{{.AppKey.Version}}",` + cloudcommon.AlertLabelAppOrg + `="{{.AppKey.Organization}}"} == 0
    for: 5m
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
	ioutil.WriteFile(*promTargetsFile, []byte(targets), 0644)
}

// Delete Alert file and reload rules
func deleteCloudletPrometheusAlertFile(ctx context.Context, file string) error {
	// remove alerting rules
	err := os.Remove(file)
	if err != nil {
		return err
	}
	// need to force prometheus to re-read the rules file
	resp, err := http.Post("http://0.0.0.0:9092/-/reload", "", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to reload prometheus", "err", err)
		return nil
	}
	resp.Body.Close()
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
	resp, err := http.Post("http://0.0.0.0:9092/-/reload", "", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to reload prometheus", "err", err)
		return nil
	}
	resp.Body.Close()
	return nil
}

func targetsList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>%s</h1>", "List all targets")
	targets := copyMapValues()
	for ii, v := range targets {
		fmt.Fprintf(w, "<h1>Target %d</h1><div>%s</div>", ii, getProxyKey(&v.Key))
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

func getPromtheusFileName(name string) string {
	return "/tmp/" + intprocess.PormtheusRulesPrefix + name + ".yml"
}

// Starts Cloudlet Prometheus MetricsProxy thread to serve as a target for metrics
func startPrometheusMetricsProxy(ctx context.Context) error {
	// This works for edgebox and openstack cloudlets for now
	if *platformName == "PLATFORM_TYPE_EDGEBOX" ||
		*platformName == "PLATFORM_TYPE_OPENSTACK" {
		// Init prometheus targets and alert templates
		promTargetTemplate = template.Must(template.New("prometheustarget").Parse(promTargetT))
		promAlertTemplate = template.Must(template.New("alert").Parse(promAutoProvAlertT))
		healthCeckFile := getPromtheusFileName("healthcheck")
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
	}
	return nil
}

func shouldAddAutoDeprovPolicy(ctx context.Context, appInst *edgeproto.AppInst, app *edgeproto.App) bool {
	// if the clusterInst is not reservable, no policies for this appInst
	if appInst.Key.ClusterInstKey.Organization != cloudcommon.OrganizationMobiledgeX {
		return false
	}
	policy := edgeproto.AutoProvPolicy{}
	for _, polName := range app.AutoProvPolicies {
		polKey := edgeproto.PolicyKey{
			Organization: app.Key.Organization,
			Name:         polName,
		}
		log.SpanLog(ctx, log.DebugLevelMetrics, "Eval policy", "app", app, "policy", polKey)
		found := AutoProvPoliciesCache.Get(&polKey, &policy)
		if !found {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find polocy", "policy", polKey)
			continue
		}
		// Check if the cloudlets should be this onee are one of those
		for _, cloudlet := range policy.Cloudlets {
			if cloudletKey.Matches(&cloudlet.Key) {
				return true
			}
		}
	}
	// Didn't find any policies that should be enacted on this cloudlet
	return false
}

func writePromtheusAlertRuleForAppInst(ctx context.Context, appInst *edgeproto.AppInst) {
	// AppInst is being deleted - delete rules
	if appInst.State != edgeproto.TrackedState_READY {
		fileName := getPromtheusFileName(k8smgmt.NormalizeName(appInst.Key.AppKey.Name))
		if err := deleteCloudletPrometheusAlertFile(ctx, fileName); err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to delete prometheus rules", "file", fileName, "err", err)
		}
		return
	}
	// check cluster name if this is a VM App
	app := edgeproto.App{}
	found := AppCache.Get(&appInst.Key.AppKey, &app)
	if !found {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to find app", "app", appInst.Key.AppKey.Name)
		return
	}
	// check if there is an auto-scale policy first
	if !shouldAddAutoDeprovPolicy(ctx, appInst, &app) {
		log.SpanLog(ctx, log.DebugLevelMetrics, "no autoprovisioning for this AppInst", "appInst", appInst, "app", app)
		return
	}
	buf := bytes.Buffer{}
	if err := promAlertTemplate.Execute(&buf, appInst.Key); err != nil {
		log.DebugLog(log.DebugLevelMetrics, "Failed to create autoprov alerts", "template", promAlertTemplate,
			"data", appInst, "error", err)
		return
	}

	fileName := getPromtheusFileName(k8smgmt.NormalizeName(appInst.Key.AppKey.Name))
	err := writeCloudletPrometheusAlerts(ctx, fileName, buf.Bytes())
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Failed to write prometheus rules", "file", fileName, "err", err)
	}
}
