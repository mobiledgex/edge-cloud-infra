package ormctl

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

func GetReportCommand() *cobra.Command {
	cmds := []*cli.Command{&cli.Command{
		Use:          "generate",
		Short:        "Generate new report",
		RequiredArgs: "org region starttime endtime",
		ReqData:      &ormapi.GenerateReport{},
		Run:          runRest("/auth/report/generate"),
	}}
	return cli.GenGroup("report", "Manage reports", cmds)
}
