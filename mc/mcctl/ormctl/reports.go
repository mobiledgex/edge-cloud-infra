package ormctl

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

func GetReporterCommand() *cobra.Command {
	cmds := []*cli.Command{&cli.Command{
		Use:          "create",
		Short:        "Create new reporter",
		RequiredArgs: "org",
		OptionalArgs: "email schedule scheduledate",
		ReqData:      &ormapi.Reporter{},
		Comments:     ReporterComments,
		Run:          runRest("/auth/reporter/create"),
	}, &cli.Command{
		Use:          "update",
		Short:        "Update reporter",
		RequiredArgs: "org",
		OptionalArgs: "email schedule scheduledate",
		ReqData:      &ormapi.Reporter{},
		Comments:     ReporterComments,
		Run:          runRest("/auth/reporter/update"),
	}, &cli.Command{
		Use:          "delete",
		Short:        "Delete reporter",
		RequiredArgs: "org",
		ReqData:      &ormapi.Reporter{},
		Comments:     ReporterComments,
		Run:          runRest("/auth/reporter/delete"),
	}, &cli.Command{
		Use:          "show",
		Short:        "Show reporters",
		RequiredArgs: "org",
		ReqData:      &ormapi.Reporter{},
		Comments:     ReporterComments,
		Run:          runRest("/auth/reporter/show"),
	}}
	return cli.GenGroup("reporter", "Manage report schedule", cmds)
}

func GetReportCommand() *cobra.Command {
	cmds := []*cli.Command{&cli.Command{
		Use:          "generate",
		Short:        "Generate new report for an org of all regions",
		RequiredArgs: "org starttime endtime",
		ReqData:      &ormapi.GenerateReport{},
		Comments:     GenerateReportComments,
		Run:          runRest("/auth/report/generate"),
	}, &cli.Command{
		Use:          "show",
		Short:        "Show generated reports",
		RequiredArgs: "org",
		ReqData:      &ormapi.GenerateReport{},
		Comments:     GenerateReportComments,
		Run:          runRest("/auth/report/show"),
	}}
	return cli.GenGroup("report", "Manage reports", cmds)
}

var ReporterComments = map[string]string{
	"org":          `Org name`,
	"email":        `Email to send generated reports`,
	"schedule":     `Report schedule, one of EveryWeek, Every15Days, Every30Days`,
	"scheduledate": `Date when the next report is scheduled to be generated (default: now)`,
}

var GenerateReportComments = map[string]string{
	"org":       `Org name`,
	"starttime": `absolute time to start report capture`,
	"endtime":   `absolute time to end report capture`,
}
