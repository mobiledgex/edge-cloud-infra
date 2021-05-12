package mccli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/ormctl"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
	"github.com/spf13/cobra"
)

func (s *RootCommand) getReportCmdGroup() *cobra.Command {
	apiGroup := ormctl.MustGetGroup("Report")
	cmds := []*cli.Command{}
	for _, c := range apiGroup.Commands {
		cliCmd := s.ConvertCmd(c)
		switch c.Name {
		case "GenerateReport":
			cliCmd.Run = s.runGenerateReport(c.Path)
		case "DownloadReport":
			cliCmd.Run = s.runDownloadReport(c.Path)
		}
		cmds = append(cmds, cliCmd)
	}
	return cli.GenGroup(strings.ToLower(apiGroup.Name), apiGroup.Desc, cmds)
}

func (s *RootCommand) runGenerateReport(path string) func(c *cli.Command, args []string) error {
	return func(c *cli.Command, args []string) error {
		c.CobraCmd.SilenceUsage = true
		in, err := c.ParseInput(args)
		if err != nil {
			if len(args) == 0 {
				// Force print usage since no args specified,
				// but obviously some are required.
				c.CobraCmd.SilenceUsage = false
			}
			return err
		}
		s.client.Debug = cli.Debug

		report, ok := c.ReqData.(*ormapi.GenerateReport)
		if !ok {
			return fmt.Errorf("unable to fetch report args: %v", c.ReqData)
		}

		uri := s.getUri() + path
		resp, err := s.client.PostJsonSend(uri, s.token, in)
		if err != nil {
			return fmt.Errorf("post %s client do failed, %s", uri, err.Error())
		}
		defer resp.Body.Close()
		filename := ormapi.GetReportFileName("", report)
		if resp.StatusCode == http.StatusOK {
			err = downloadPDF(filename, resp)
		}
		return check(c, resp.StatusCode, err, nil)
	}
}

func (s *RootCommand) runDownloadReport(path string) func(c *cli.Command, args []string) error {
	return func(c *cli.Command, args []string) error {
		c.CobraCmd.SilenceUsage = true
		in, err := c.ParseInput(args)
		if err != nil {
			if len(args) == 0 {
				// Force print usage since no args specified,
				// but obviously some are required.
				c.CobraCmd.SilenceUsage = false
			}
			return err
		}
		s.client.Debug = cli.Debug
		report, ok := c.ReqData.(*ormapi.DownloadReport)
		if !ok {
			return fmt.Errorf("unable to fetch report args: %v", c.ReqData)
		}

		uri := s.getUri() + path
		resp, err := s.client.PostJsonSend(uri, s.token, in)
		if err != nil {
			return fmt.Errorf("post %s client do failed, %s", uri, err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			err = downloadPDF(report.Filename, resp)
		}
		c.ReplyData = &ormapi.Result{}
		return check(c, resp.StatusCode, err, nil)
	}
}

func downloadPDF(filename string, resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Save blob to file
	err = ioutil.WriteFile(filename, body, 0666)
	if err != nil {
		return fmt.Errorf("failed to created file %s, %v", filename, err)
	}
	fmt.Printf("Saved PDF report to %s\n", filename)
	return nil
}
