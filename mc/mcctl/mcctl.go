package main

import (
	"os"

	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/ormctl"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "mcctl",
	PersistentPreRunE: ormctl.PreRunE,
}

func main() {
	rootCmd.AddCommand(ormctl.GetLoginCmd())
	rootCmd.AddCommand(ormctl.GetUserCommand())
	rootCmd.AddCommand(ormctl.GetRoleCommand())
	rootCmd.AddCommand(ormctl.GetOrgCommand())
	rootCmd.AddCommand(ormctl.GetControllerCommand())
	rootCmd.AddCommand(ormctl.GetAllDataCommand())
	regionCmds := ormctl.GetRegionCommand()
	regionCmds.AddCommand(ormctl.GetRunCommandCmd())
	rootCmd.AddCommand(regionCmds)
	rootCmd.AddCommand(ormctl.GetConfigCommand())
	rootCmd.AddCommand(ormctl.GetAuditCommand())
	rootCmd.AddCommand(ormctl.GetOrgCloudletCommand())
	rootCmd.AddCommand(ormctl.GetOrgCloudletPoolCommand())

	rootCmd.PersistentFlags().StringVar(&ormctl.Addr, "addr", "http://127.0.0.1:9900", "MC address")
	rootCmd.PersistentFlags().StringVar(&ormctl.Token, "token", "", "JWT token")
	cli.AddInputFlags(rootCmd.PersistentFlags())
	cli.AddOutputFlags(rootCmd.PersistentFlags())
	cli.AddDebugFlag(rootCmd.PersistentFlags())
	rootCmd.PersistentFlags().BoolVar(&ormctl.SkipVerify, "skipverify", false, "don't verify cert for TLS connections")

	cobra.EnableCommandSorting = false
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
