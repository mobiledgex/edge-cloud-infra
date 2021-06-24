package alert

import (
	"bytes"
	"context"
	"text/template"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

// TODO - add a loop for alerts
var MEXPrometheusUserAlertsT = `additionalPrometheusRules:
- name: autoscalepolicy
  groups:
  - name: autoscale.rules
    rules:
    - expr: 1 - avg(rate(node_cpu_seconds_total{job="node-exporter",mode="idle"}[1m]))
      record: :node_cpu_utilisation:avg1m
    - expr: |-
        1 - avg by (node) (
          rate(node_cpu_seconds_total{job="node-exporter",mode="idle"}[1m])
        * on (namespace, pod) group_left(node)
          node_namespace_pod:kube_pod_info:)
      record: node:node_cpu_utilisation:avg1m
    - expr: sum(node:node_cpu_utilisation:avg1m{node=~"[[.NodePrefix]].*"} > bool [[printf "%.2f" .ScaleUpCpuThreshF]])
      record: 'node_cpu_high_count'
    - expr: sum(node:node_cpu_utilisation:avg1m{node=~"[[.NodePrefix]].*"} < bool [[printf "%.2f" .ScaleDownCpuThreshF]])
      record: 'node_cpu_low_count'
    - expr: count(kube_node_info) - count(kube_node_spec_taint)
      record: 'node_count'
    - alert: [[.AutoScaleUpName]]
      expr: node_cpu_high_count == node_count and node_count < [[.MaxNodes]]
      for: [[.TriggerTimeSec]]s
      labels:
        severity: none
      annotations:
        message: High cpu greater than [[.ScaleUpCpuThresh]]% for all nodes
        [[.NodeCountName]]: '{{ with query "node_count" }}{{ . | first | value | humanize }}{{ end }}'
    - alert: [[.AutoScaleDownName]]
      expr: node_cpu_low_count > 0 and node_count > [[.MinNodes]]
      for: [[.TriggerTimeSec]]s
      labels:
        severity: none
      annotations:
        message: Low cpu less than [[.ScaleDownCpuThresh]]% for some nodes
        [[.LowCpuNodeCountName]]: '{{ $value }}'
        [[.NodeCountName]]: '{{ with query "node_count" }}{{ . | first | value | humanize }}{{ end }}'
        [[.MinNodesName]]: '[[.MinNodes]]'
`

// TODO - fix me
type AlertArgs struct {
	AutoScalePolicy     string
	AutoScaleUpName     string
	AutoScaleDownName   string
	ScaleUpCpuThresh    uint32
	ScaleDownCpuThresh  uint32
	ScaleUpCpuThreshF   float32
	ScaleDownCpuThreshF float32
	TriggerTimeSec      uint32
	MaxNodes            int
	MinNodes            int
	NodeCountName       string
	LowCpuNodeCountName string
	MinNodesName        string
	NodePrefix          string // master node must have different prefix
}

func GetAlertsRules(ctx context.Context, alerts []*edgeproto.UserAlert) (string, error) {
	// change delims because Prometheus triggers off of golang delims
	t := template.Must(template.New("policy").Delims("[[", "]]").Parse(MEXPrometheusUserAlertsT))
	args := AlertArgs{
		// TODO - fill me
	}
	buf := bytes.Buffer{}
	err := t.Execute(&buf, &args)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
