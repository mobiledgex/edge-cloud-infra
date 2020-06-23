// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: exec.proto

package ormctl

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import "github.com/mobiledgex/edge-cloud/cli"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var RunCommandCmd = &cli.Command{
	Use:          "RunCommand",
	RequiredArgs: "region " + strings.Join(RunCommandRequiredArgs, " "),
	OptionalArgs: strings.Join(RunCommandOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     addRegionComment(ExecRequestComments),
	ReqData:      &ormapi.RegionExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runRest("/auth/ctrl/RunCommand"),
}

var RunConsoleCmd = &cli.Command{
	Use:          "RunConsole",
	RequiredArgs: "region " + strings.Join(RunConsoleRequiredArgs, " "),
	OptionalArgs: strings.Join(RunConsoleOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     addRegionComment(ExecRequestComments),
	ReqData:      &ormapi.RegionExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runRest("/auth/ctrl/RunConsole"),
}

var ShowLogsCmd = &cli.Command{
	Use:          "ShowLogs",
	RequiredArgs: "region " + strings.Join(ShowLogsRequiredArgs, " "),
	OptionalArgs: strings.Join(ShowLogsOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     addRegionComment(ExecRequestComments),
	ReqData:      &ormapi.RegionExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runRest("/auth/ctrl/ShowLogs"),
}

var AccessCloudletCmd = &cli.Command{
	Use:          "AccessCloudlet",
	RequiredArgs: "region " + strings.Join(AccessCloudletRequiredArgs, " "),
	OptionalArgs: strings.Join(AccessCloudletOptionalArgs, " "),
	AliasArgs:    strings.Join(ExecRequestAliasArgs, " "),
	SpecialArgs:  &ExecRequestSpecialArgs,
	Comments:     addRegionComment(ExecRequestComments),
	ReqData:      &ormapi.RegionExecRequest{},
	ReplyData:    &edgeproto.ExecRequest{},
	Run:          runRest("/auth/ctrl/AccessCloudlet"),
}

var ExecApiCmds = []*cli.Command{
	RunCommandCmd,
	RunConsoleCmd,
	ShowLogsCmd,
	AccessCloudletCmd,
}

var RunCommandRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cluster",
	"cloudlet-org",
	"cloudlet",
	"command",
}
var RunCommandOptionalArgs = []string{
	"cluster-org",
	"containerid",
	"node-type",
	"node-name",
	"webrtc",
}
var RunConsoleRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var RunConsoleOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"webrtc",
}
var ShowLogsRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cluster",
	"cloudlet-org",
	"cloudlet",
}
var ShowLogsOptionalArgs = []string{
	"cluster-org",
	"containerid",
	"since",
	"tail",
	"timestamps",
	"follow",
	"webrtc",
}
var AccessCloudletRequiredArgs = []string{
	"cloudlet-org",
	"cloudlet",
}
var AccessCloudletOptionalArgs = []string{
	"command",
	"node-type",
	"node-name",
	"webrtc",
}
var CloudletMgmtNodeRequiredArgs = []string{}
var CloudletMgmtNodeOptionalArgs = []string{
	"type",
	"name",
}
var CloudletMgmtNodeAliasArgs = []string{
	"type=cloudletmgmtnode.type",
	"name=cloudletmgmtnode.name",
}
var CloudletMgmtNodeComments = map[string]string{
	"type": "Type of Cloudlet Mgmt Node",
	"name": "Name of Cloudlet Mgmt Node",
}
var CloudletMgmtNodeSpecialArgs = map[string]string{}
var RunCmdRequiredArgs = []string{}
var RunCmdOptionalArgs = []string{
	"command",
	"cloudletmgmtnode.type",
	"cloudletmgmtnode.name",
}
var RunCmdAliasArgs = []string{
	"command=runcmd.command",
	"cloudletmgmtnode.type=runcmd.cloudletmgmtnode.type",
	"cloudletmgmtnode.name=runcmd.cloudletmgmtnode.name",
}
var RunCmdComments = map[string]string{
	"command":               "Command or Shell",
	"cloudletmgmtnode.type": "Type of Cloudlet Mgmt Node",
	"cloudletmgmtnode.name": "Name of Cloudlet Mgmt Node",
}
var RunCmdSpecialArgs = map[string]string{}
var RunVMConsoleRequiredArgs = []string{}
var RunVMConsoleOptionalArgs = []string{
	"url",
}
var RunVMConsoleAliasArgs = []string{
	"url=runvmconsole.url",
}
var RunVMConsoleComments = map[string]string{
	"url": "VM Console URL",
}
var RunVMConsoleSpecialArgs = map[string]string{}
var ShowLogRequiredArgs = []string{}
var ShowLogOptionalArgs = []string{
	"since",
	"tail",
	"timestamps",
	"follow",
}
var ShowLogAliasArgs = []string{
	"since=showlog.since",
	"tail=showlog.tail",
	"timestamps=showlog.timestamps",
	"follow=showlog.follow",
}
var ShowLogComments = map[string]string{
	"since":      "Show logs since either a duration ago (5s, 2m, 3h) or a timestamp (RFC3339)",
	"tail":       "Show only a recent number of lines",
	"timestamps": "Show timestamps",
	"follow":     "Stream data",
}
var ShowLogSpecialArgs = map[string]string{}
var ExecRequestRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cluster",
	"cloudlet-org",
	"cloudlet",
	"cluster-org",
}
var ExecRequestOptionalArgs = []string{
	"containerid",
	"command",
	"node-type",
	"node-name",
	"since",
	"tail",
	"timestamps",
	"follow",
	"webrtc",
}
var ExecRequestAliasArgs = []string{
	"app-org=execrequest.appinstkey.appkey.organization",
	"appname=execrequest.appinstkey.appkey.name",
	"appvers=execrequest.appinstkey.appkey.version",
	"cluster=execrequest.appinstkey.clusterinstkey.clusterkey.name",
	"cloudlet-org=execrequest.appinstkey.clusterinstkey.cloudletkey.organization",
	"cloudlet=execrequest.appinstkey.clusterinstkey.cloudletkey.name",
	"cluster-org=execrequest.appinstkey.clusterinstkey.organization",
	"containerid=execrequest.containerid",
	"offer=execrequest.offer",
	"answer=execrequest.answer",
	"err=execrequest.err",
	"command=execrequest.cmd.command",
	"node-type=execrequest.cmd.cloudletmgmtnode.type",
	"node-name=execrequest.cmd.cloudletmgmtnode.name",
	"since=execrequest.log.since",
	"tail=execrequest.log.tail",
	"timestamps=execrequest.log.timestamps",
	"follow=execrequest.log.follow",
	"console.url=execrequest.console.url",
	"timeout=execrequest.timeout",
	"webrtc=execrequest.webrtc",
	"accessurl=execrequest.accessurl",
	"edgeturnaddr=execrequest.edgeturnaddr",
}
var ExecRequestComments = map[string]string{
	"app-org":      "App developer organization",
	"appname":      "App name",
	"appvers":      "App version",
	"cluster":      "Cluster name",
	"cloudlet-org": "Organization of the cloudlet site",
	"cloudlet":     "Name of the cloudlet",
	"cluster-org":  "Name of Developer organization that this cluster belongs to",
	"containerid":  "ContainerId is the name or ID of the target container, if applicable",
	"offer":        "WebRTC Offer",
	"answer":       "WebRTC Answer",
	"err":          "Any error message",
	"command":      "Command or Shell",
	"node-type":    "Type of Cloudlet Mgmt Node",
	"node-name":    "Name of Cloudlet Mgmt Node",
	"since":        "Show logs since either a duration ago (5s, 2m, 3h) or a timestamp (RFC3339)",
	"tail":         "Show only a recent number of lines",
	"timestamps":   "Show timestamps",
	"follow":       "Stream data",
	"console.url":  "VM Console URL",
	"timeout":      "Timeout",
	"webrtc":       "WebRTC",
	"accessurl":    "Access URL",
	"edgeturnaddr": "EdgeTurn Server Address",
}
var ExecRequestSpecialArgs = map[string]string{}
