package ormctl

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
		OptionalArgs: "timezone",
		ReqData:      &ormapi.GenerateReport{},
		Comments:     GenerateReportComments,
		Run:          runGenerateReport("/auth/report/generate"),
	}, &cli.Command{
		Use:          "show",
		Short:        "Show already generated reports",
		RequiredArgs: "org",
		ReqData:      &ormapi.DownloadReport{},
		ReplyData:    &[]string{},
		Comments:     GenerateReportComments,
		Run:          runRest("/auth/report/show"),
	}, &cli.Command{
		Use:          "download",
		Short:        "Download generated report",
		RequiredArgs: "org filename",
		ReqData:      &ormapi.DownloadReport{},
		Comments:     DownloadReportComments,
		Run:          runDownloadReport("/auth/report/download"),
	}}
	return cli.GenGroup("report", "Manage reports", cmds)
}

var ReporterComments = map[string]string{
	"org":          `Org name`,
	"email":        `Email to send generated reports`,
	"schedule":     `Report schedule, one of EveryWeek, Every15Days, Every30Days`,
	"scheduledate": `Date when the next report is scheduled to be generated (default: now)`,
}

var DownloadReportComments = map[string]string{
	"org":      `Org name`,
	"filename": `Name of the report file to be downloaded`,
}

var GenerateReportComments = map[string]string{
	"org":       `Org name`,
	"starttime": `Absolute time to start report capture in UTC`,
	"endtime":   `Absolute time to end report capture in UTC`,
	"timezone":  `Timezone in which to show the reports, defaults to either user setting or UTC`,
}

func runGenerateReport(path string) func(c *cli.Command, args []string) error {
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
		client.Debug = cli.Debug

		report, ok := c.ReqData.(*ormapi.GenerateReport)
		if !ok {
			return fmt.Errorf("unable to fetch report args: %v", c.ReqData)
		}

		uri := getUri() + path
		resp, err := client.PostJsonSend(uri, Token, in)
		if err != nil {
			return fmt.Errorf("post %s client do failed, %s", uri, err.Error())
		}
		defer resp.Body.Close()
		filename := ormapi.GetReportFileName(report)
		if resp.StatusCode == http.StatusOK {
			err = downloadPDF(filename, resp)
		}
		c.ReplyData = &ormapi.Result{}
		return check(c, resp.StatusCode, err, c.ReplyData)
	}
}

func runDownloadReport(path string) func(c *cli.Command, args []string) error {
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
		client.Debug = cli.Debug
		report, ok := c.ReqData.(*ormapi.DownloadReport)
		if !ok {
			return fmt.Errorf("unable to fetch report args: %v", c.ReqData)
		}

		uri := getUri() + path
		resp, err := client.PostJsonSend(uri, Token, in)
		if err != nil {
			return fmt.Errorf("post %s client do failed, %s", uri, err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			err = downloadPDF(report.Filename, resp)
		}
		c.ReplyData = &ormapi.Result{}
		return check(c, resp.StatusCode, err, c.ReplyData)
	}
}

func downloadPDF(filename string, resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Save blob to file
	pdfFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to created file %s, %v", filename, err)
	}
	defer pdfFile.Close()
	if _, err = pdfFile.Write(body); err != nil {
		return fmt.Errorf("failed to write data to file %s, %v", filename, err)
	}
	fmt.Printf("Saved PDF report to %s\n", filename)
	return nil
}
