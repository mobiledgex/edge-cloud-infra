package ormctl

import (
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

// We don't use the auto-generated Command because the client
// must implement the webrtc protocol.

func GetRunCommandCmd() *cobra.Command {
	RunCommandCmd.Run = runExecRequest("/auth/ctrl/RunCommand")
	return RunCommandCmd.GenCmd()
}

func GetRunConsoleCmd() *cobra.Command {
	RunConsoleCmd.Run = runExecRequest("/auth/ctrl/RunConsole")
	return RunConsoleCmd.GenCmd()
}

func GetShowLogsCmd() *cobra.Command {
	ShowLogsCmd.Run = runExecRequest("/auth/ctrl/ShowLogs")
	return ShowLogsCmd.GenCmd()
}

func runExecRequest(path string) func(c *cli.Command, args []string) error {
	return func(c *cli.Command, args []string) error {
		input := cli.Input{
			RequiredArgs: strings.Split(c.RequiredArgs, " "),
			AliasArgs:    strings.Split(c.AliasArgs, " "),
		}
		req := ormapi.RegionExecRequest{}
		_, err := input.ParseArgs(args, &req)
		if err != nil {
			return err
		}
		out := ormapi.WSStreamPayload{}

		// print streamed data as it comes
		st, err := client.PostJsonStreamOut(getWSUri()+path, Token, &req, &out, func() {
			check(c, 0, nil, out)
		})
		return check(c, st, err, nil)
	}
}
