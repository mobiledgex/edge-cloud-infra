package common

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

var TestClusterUserDefAlertsRules = `additionalPrometheusRules:
- name: userdefinedalerts
  groups:
  - name: useralerts.rules
    rules:
    - alert: testAlert1
      expr: max(kube_pod_labels{label_mexAppName="pokemongo",label_mexAppVersion="100"})by(label_mexAppName,label_mexAppVersion,pod)*on(pod)group_right(label_mexAppName,label_mexAppVersion)(sum(rate(container_cpu_usage_seconds_total{image!=""}[1m]))by(pod)) > 80 and max(kube_pod_labels{label_mexAppName="pokemongo",label_mexAppVersion="100"})by(label_mexAppName,label_mexAppVersion,pod)*on(pod)group_right(label_mexAppName,label_mexAppVersion)(sum(container_memory_working_set_bytes{image!=""})by(pod)) > 123456 and max(kube_pod_labels{label_mexAppName="pokemongo",label_mexAppVersion="100"})by(label_mexAppName,label_mexAppVersion,pod)*on(pod)group_right(label_mexAppName,label_mexAppVersion)(sum(container_fs_usage_bytes{image!=""})by(pod)) > 123456
      for: 30s
      labels:
        severity: warning
        type: "User Defined"
        app: "Pokemon Go!"
        apporg: "NianticInc"
        appver: "1.0.0"
        cloudlet: "San Jose Site"
        cloudletorg: "AT&T Inc."
        cluster: "Pokemons"
        clusterorg: "NianticInc"
        scope: "Application"
        type: "UserDefined"
    - alert: testAlert3
      expr: max(kube_pod_labels{label_mexAppName="pokemongo",label_mexAppVersion="100"})by(label_mexAppName,label_mexAppVersion,pod)*on(pod)group_right(label_mexAppName,label_mexAppVersion)(sum(rate(container_cpu_usage_seconds_total{image!=""}[1m]))by(pod)) > 100
      for: 30s
      labels:
        severity: error
        type: "User Defined"
        app: "Pokemon Go!"
        apporg: "NianticInc"
        appver: "1.0.0"
        cloudlet: "San Jose Site"
        cloudletorg: "AT&T Inc."
        cluster: "Pokemons"
        clusterorg: "NianticInc"
        scope: "Application"
        testLabel1: "testValue1"
        testLabel2: "testValue2"
        type: "UserDefined"
      annotations:
        testAnnotation1: "description1"
        testAnnotation2: "description2"
`

// TestAutoScaleT primarily checks that AutoScale template parsing works, because
// otherwise cluster-svc could crash during runtime if template has an issue.
func TestAutoScaleT(t *testing.T) {
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	clusterInst := testutil.ClusterInstData[0]

	policy := edgeproto.AutoScalePolicy{}
	policy.Key.Organization = clusterInst.Key.Organization
	policy.Key.Name = "test-policy"
	policy.TriggerTimeSec = 120

	clusterInst.AutoScalePolicy = policy.Key.Name

	userAlerts := testutil.UserAlertData
	settings := edgeproto.GetDefaultSettings()

	configExpected := `additionalPrometheusRules:
- name: autoscalepolicy
  groups:
  - name: autoscale.rules
    rules:
    - expr: |-
        1 - avg by (node) (
          rate(node_cpu_seconds_total{job="node-exporter",mode="idle"}[60s])
        * on (namespace, pod) group_left(node)
          node_namespace_pod:kube_pod_info:)
      record: node:node_cpu_utilisation:avg
    - expr: sum(node:node_cpu_utilisation:avg unless (kube_node_spec_taint{effect="NoSchedule"} * on(node) kube_node_spec_taint))
      record: 'total_worker_node_cpu_utilisation'
    - expr: max_over_time(total_worker_node_cpu_utilisation[120s])
      record: 'stabilized_max_total_worker_node_cpu_utilisation'

    - expr: avg_over_time(instance:node_memory_utilisation:ratio[60s])
      record: 'node_memory_utilisation:ratio:avg'
    - expr: sum(node_memory_utilisation:ratio:avg unless (kube_node_spec_taint{effect="NoSchedule"} * on(node) kube_node_spec_taint))
      record: 'total_worker_node_mem_utilisation'
    - expr: max_over_time(total_worker_node_mem_utilisation[120s])
      record: 'stabilized_max_total_worker_node_mem_utilisation'
`
	testClusterRulesT(t, ctx, &clusterInst, &policy, settings, userAlerts, configExpected, TestClusterUserDefAlertsRules)

	policy.StabilizationWindowSec = 300
	settings.ClusterAutoScaleAveragingDurationSec = 30

	configExpected = `additionalPrometheusRules:
- name: autoscalepolicy
  groups:
  - name: autoscale.rules
    rules:
    - expr: |-
        1 - avg by (node) (
          rate(node_cpu_seconds_total{job="node-exporter",mode="idle"}[30s])
        * on (namespace, pod) group_left(node)
          node_namespace_pod:kube_pod_info:)
      record: node:node_cpu_utilisation:avg
    - expr: sum(node:node_cpu_utilisation:avg unless (kube_node_spec_taint{effect="NoSchedule"} * on(node) kube_node_spec_taint))
      record: 'total_worker_node_cpu_utilisation'
    - expr: max_over_time(total_worker_node_cpu_utilisation[300s])
      record: 'stabilized_max_total_worker_node_cpu_utilisation'

    - expr: avg_over_time(instance:node_memory_utilisation:ratio[30s])
      record: 'node_memory_utilisation:ratio:avg'
    - expr: sum(node_memory_utilisation:ratio:avg unless (kube_node_spec_taint{effect="NoSchedule"} * on(node) kube_node_spec_taint))
      record: 'total_worker_node_mem_utilisation'
    - expr: max_over_time(total_worker_node_mem_utilisation[300s])
      record: 'stabilized_max_total_worker_node_mem_utilisation'
`
	testClusterRulesT(t, ctx, &clusterInst, &policy, settings, userAlerts, configExpected, TestClusterUserDefAlertsRules)
}

func testClusterRulesT(t *testing.T, ctx context.Context, clusterInst *edgeproto.ClusterInst, policy *edgeproto.AutoScalePolicy, settings *edgeproto.Settings, alerts []edgeproto.UserAlert, expectedAutoProvRules string, expectedUserAlertsRules string) {
	clusterSvc := ClusterSvc{}
	appInst := testutil.AppInstData[0]

	configs, err := clusterSvc.GetAppInstConfigs(ctx, clusterInst, &appInst, policy, settings, alerts)
	require.Nil(t, err)
	require.Equal(t, 2, len(configs))
	ioutil.WriteFile("foo", []byte(configs[0].Config), 0644)
	require.Equal(t, expectedAutoProvRules, configs[0].Config)
	ioutil.WriteFile("bar", []byte(configs[1].Config), 0644)
	require.Equal(t, expectedUserAlertsRules, configs[1].Config)
}
