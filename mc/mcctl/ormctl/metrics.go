package ormctl

import (
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

func GetMetricsCommand() *cobra.Command {
	cmds := []*cli.Command{&cli.Command{
		Use:          "app",
		RequiredArgs: strings.Join(append([]string{"region"}, AppMetricRequiredArgs...), " "),
		OptionalArgs: strings.Join(AppMetricOptionalArgs, " "),
		AliasArgs:    strings.Join(AppMetricAliasArgs, " "),
		Comments:     addRegionComment(MetricComments),
		ReqData:      &ormapi.RegionAppInstMetrics{},
		ReplyData:    &ormapi.AllMetrics{},
		Run:          runRest("/auth/metrics/app"),
	}, &cli.Command{
		Use:          "cluster",
		RequiredArgs: strings.Join(append([]string{"region"}, ClusterMetricRequiredArgs...), " "),
		OptionalArgs: strings.Join(ClusterMetricOptionalArgs, " "),
		AliasArgs:    strings.Join(ClusterMetricAliasArgs, " "),
		Comments:     addRegionComment(MetricComments),
		ReqData:      &ormapi.RegionClusterInstMetrics{},
		ReplyData:    &ormapi.AllMetrics{},
		Run:          runRest("/auth/metrics/cluster"),
	}, &cli.Command{
		Use:          "cloudlet",
		RequiredArgs: strings.Join(append([]string{"region"}, CloudletMetricRequiredArgs...), " "),
		OptionalArgs: strings.Join(CloudletMetricOptionalArgs, " "),
		AliasArgs:    strings.Join(CloudletMetricAliasArgs, " "),
		Comments:     addRegionComment(MetricComments),
		ReqData:      &ormapi.RegionCloudletMetrics{},
		ReplyData:    &ormapi.AllMetrics{},
		Run:          runRest("/auth/metrics/cloudlet"),
	}}
	return cli.GenGroup("metrics", "view metrics ", cmds)
}

var AppMetricRequiredArgs = []string{
	"developerorg",
	"selector",
}

var AppMetricOptionalArgs = []string{
	"appname",
	"appvers",
	"cluster",
	"cloudlet",
	"operatororg",
	"last",
	"starttime",
	"endtime",
}

var AppMetricAliasArgs = []string{
	"developerorg=appinst.appkey.organization",
	"appname=appinst.appkey.name",
	"appvers=appinst.appkey.version",
	"cluster=appinst.clusterinstkey.clusterkey.name",
	"operatororg=appinst.clusterinstkey.cloudletkey.organization",
	"cloudlet=appinst.clusterinstkey.cloudletkey.name",
}

var ClusterMetricRequiredArgs = []string{
	"developerorg",
	"selector",
}

var ClusterMetricOptionalArgs = []string{
	"cluster",
	"operatororg",
	"cloudlet",
	"last",
	"starttime",
	"endtime",
}

var ClusterMetricAliasArgs = []string{
	"developerorg=clusterinst.organization",
	"cluster=clusterinst.clusterkey.name",
	"operatororg=clusterinst.cloudletkey.organization",
	"cloudlet=clusterinst.cloudletkey.name",
}

var CloudletMetricRequiredArgs = []string{
	"operatororg",
	"selector",
}

var CloudletMetricOptionalArgs = []string{
	"cloudlet",
	"last",
	"starttime",
	"endtime",
}

var CloudletMetricAliasArgs = []string{
	"operator=cloudlet.operatorkey.name",
	"cloudlet=cloudlet.name",
}

var MetricComments = map[string]string{
	"developerorg": "Organization or Company Name that a Developer is part of",
	"appname":      "App name",
	"appvers":      "App version",
	"cluster":      "Cluster name",
	"operatororg":  "Company or Organization name of the operator",
	"cloudlet":     "Name of the cloudlet",
	"selector":     "Comma separated list of metrics to view",
	"last":         "Display the last X metrics",
	"starttime":    "Time to start displaying stats from",
	"endtime":      "Time up to which to display stats",
}
