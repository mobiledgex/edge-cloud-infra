package ormctl

import (
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

func GetUsageCommand() *cobra.Command {
	cmds := []*cli.Command{&cli.Command{
		Use:          "app",
		RequiredArgs: strings.Join(append([]string{"region"}, AppUsageRequiredArgs...), " "),
		OptionalArgs: strings.Join(AppUsageOptionalArgs, " "),
		AliasArgs:    strings.Join(AppUsageAliasArgs, " "),
		Comments:     addRegionComment(AppUsageComments),
		ReqData:      &ormapi.RegionAppInstUsage{},
		ReplyData:    &ormapi.AllUsage{},
		Run:          runRest("/auth/usage/app"),
	}, &cli.Command{
		Use:          "cluster",
		RequiredArgs: strings.Join(append([]string{"region"}, ClusterUsageRequiredArgs...), " "),
		OptionalArgs: strings.Join(ClusterUsageOptionalArgs, " "),
		AliasArgs:    strings.Join(ClusterUsageAliasArgs, " "),
		Comments:     addRegionComment(ClusterUsageComments),
		ReqData:      &ormapi.RegionClusterInstUsage{},
		ReplyData:    &ormapi.AllUsage{},
		Run:          runRest("/auth/usage/cluster"),
	}}
	return cli.GenGroup("usage", "view usage ", cmds)
}

var AppUsageRequiredArgs = []string{
	"starttime",
	"endtime",
}

var AppUsageOptionalArgs = []string{
	"appname",
	"apporg",
	"appvers",
	"cluster",
	"cloudlet",
	"cloudletorg",
	"vmonly",
}

var AppUsageAliasArgs = []string{
	"apporg=appinst.appkey.organization",
	"appname=appinst.appkey.name",
	"appvers=appinst.appkey.version",
	"cluster=appinst.clusterinstkey.clusterkey.name",
	"cloudletorg=appinst.clusterinstkey.cloudletkey.organization",
	"cloudlet=appinst.clusterinstkey.cloudletkey.name",
}

var ClusterUsageRequiredArgs = []string{
	"starttime",
	"endtime",
}

var ClusterUsageOptionalArgs = []string{
	"cluster",
	"clusterorg",
	"cloudletorg",
	"cloudlet",
}

var ClusterUsageAliasArgs = []string{
	"clusterorg=clusterinst.organization",
	"cluster=clusterinst.clusterkey.name",
	"cloudletorg=clusterinst.cloudletkey.organization",
	"cloudlet=clusterinst.cloudletkey.name",
}

var AppUsageComments = map[string]string{
	"apporg":      "Organization or Company Name that a Developer is part of",
	"appname":     "App name",
	"appvers":     "App version",
	"cluster":     "Cluster name",
	"cloudletorg": "Organization name owning of the cloudlet",
	"cloudlet":    "Name of the cloudlet",
	"starttime":   "Time to start displaying usage from",
	"endtime":     "Time up to which to display usage",
	"vmonly":      "Only show VM based apps",
}

var ClusterUsageComments = map[string]string{
	"clusterorg":  "Organization or Company Name that a Developer is part of",
	"cluster":     "Cluster name",
	"cloudletorg": "Organization name owning of the cloudlet",
	"cloudlet":    "Name of the cloudlet",
	"starttime":   "Time to start displaying usage from",
	"endtime":     "Time up to which to display usage",
}
