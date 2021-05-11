package ormctl

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

const ReporterGroup = "Reporter"
const ReportGroup = "Report"

func init() {
	cmds := []*ApiCommand{&ApiCommand{
		Name:         "CreateReporter",
		Use:          "create",
		Short:        "Create new reporter",
		RequiredArgs: "org",
		OptionalArgs: "email schedule scheduledate",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		Path:         "/auth/reporter/create",
	}, &ApiCommand{
		Name:         "UpdateReporter",
		Use:          "update",
		Short:        "Update reporter",
		RequiredArgs: "org",
		OptionalArgs: "email schedule scheduledate",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		Path:         "/auth/reporter/update",
	}, &ApiCommand{
		Name:         "DeleteReporter",
		Use:          "delete",
		Short:        "Delete reporter",
		RequiredArgs: "org",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		Path:         "/auth/reporter/delete",
	}, &ApiCommand{
		Name:         "ShowReporter",
		Use:          "show",
		Short:        "Show reporters",
		RequiredArgs: "org",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		Path:         "/auth/reporter/show",
	}}
	AllApis.AddGroup(ReporterGroup, "Manage report schedule", cmds)

	cmds = []*ApiCommand{&ApiCommand{
		Name:         "GenerateReport",
		Use:          "generate",
		Short:        "Generate new report for an org of all regions",
		RequiredArgs: "org starttime endtime",
		OptionalArgs: "timezone",
		ReqData:      &ormapi.GenerateReport{},
		Comments:     ormapi.GenerateReportComments,
		Path:         "/auth/report/generate",
	}, &ApiCommand{
		Name:         "ShowReport",
		Use:          "show",
		Short:        "Show already generated reports",
		RequiredArgs: "org",
		ReqData:      &ormapi.DownloadReport{},
		ReplyData:    &[]string{},
		Comments:     ormapi.GenerateReportComments,
		Path:         "/auth/report/show",
	}, &ApiCommand{
		Name:         "DownloadReport",
		Use:          "download",
		Short:        "Download generated report",
		RequiredArgs: "org filename",
		ReqData:      &ormapi.DownloadReport{},
		Comments:     ormapi.DownloadReportComments,
		Path:         "/auth/report/download",
	}}
	AllApis.AddGroup(ReportGroup, "Manage report schedule", cmds)
}
