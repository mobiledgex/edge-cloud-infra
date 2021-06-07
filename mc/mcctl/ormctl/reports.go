package ormctl

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

const ReporterGroup = "Reporter"
const ReportGroup = "Report"
const ReportDataGroup = "ReportData"

func init() {
	cmds := []*ApiCommand{&ApiCommand{
		Name:         "CreateReporter",
		Use:          "create",
		Short:        "Create new reporter",
		RequiredArgs: "name org",
		OptionalArgs: "email schedule startscheduledate timezone",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		Path:         "/auth/reporter/create",
	}, &ApiCommand{
		Name:         "UpdateReporter",
		Use:          "update",
		Short:        "Update reporter",
		RequiredArgs: "name org",
		OptionalArgs: "email schedule startscheduledate timezone",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		Path:         "/auth/reporter/update",
	}, &ApiCommand{
		Name:         "DeleteReporter",
		Use:          "delete",
		Short:        "Delete reporter",
		RequiredArgs: "name org",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		Path:         "/auth/reporter/delete",
	}, &ApiCommand{
		Name:         "ShowReporter",
		Use:          "show",
		Short:        "Show reporters",
		OptionalArgs: "name org",
		ReqData:      &ormapi.Reporter{},
		Comments:     ormapi.ReporterComments,
		ReplyData:    &[]ormapi.Reporter{},
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
		Short:        "Show already generated reports for an org",
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
	AllApis.AddGroup(ReportGroup, "Manage reports", cmds)

	cmds = []*ApiCommand{&ApiCommand{
		Name:         "GenerateReportData",
		Use:          "generate",
		Short:        "Generate report data for an org of all regions",
		RequiredArgs: "org starttime endtime",
		ReqData:      &ormapi.GenerateReport{},
		ReplyData:    &map[string]interface{}{},
		Comments:     ormapi.GenerateReportComments,
		Path:         "/auth/report/generatedata",
	}}
	AllApis.AddGroup(ReportDataGroup, "Access report data", cmds)
}
