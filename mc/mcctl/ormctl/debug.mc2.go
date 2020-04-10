// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: debug.proto

package ormctl

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import "github.com/mobiledgex/edge-cloud/cli"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var EnableDebugLevelsCmd = &cli.Command{
	Use:          "EnableDebugLevels",
	RequiredArgs: strings.Join(EnableDebugLevelsRequiredArgs, " "),
	OptionalArgs: strings.Join(EnableDebugLevelsOptionalArgs, " "),
	AliasArgs:    strings.Join(DebugRequestAliasArgs, " "),
	SpecialArgs:  &DebugRequestSpecialArgs,
	Comments:     addRegionComment(DebugRequestComments),
	ReqData:      &ormapi.RegionDebugRequest{},
	ReplyData:    &edgeproto.DebugReply{},
	Run:          runRest("/auth/ctrl/EnableDebugLevels"),
	StreamOut:    true,
}

var DisableDebugLevelsCmd = &cli.Command{
	Use:          "DisableDebugLevels",
	RequiredArgs: strings.Join(DisableDebugLevelsRequiredArgs, " "),
	OptionalArgs: strings.Join(DisableDebugLevelsOptionalArgs, " "),
	AliasArgs:    strings.Join(DebugRequestAliasArgs, " "),
	SpecialArgs:  &DebugRequestSpecialArgs,
	Comments:     addRegionComment(DebugRequestComments),
	ReqData:      &ormapi.RegionDebugRequest{},
	ReplyData:    &edgeproto.DebugReply{},
	Run:          runRest("/auth/ctrl/DisableDebugLevels"),
	StreamOut:    true,
}

var ShowDebugLevelsCmd = &cli.Command{
	Use:          "ShowDebugLevels",
	RequiredArgs: strings.Join(ShowDebugLevelsRequiredArgs, " "),
	OptionalArgs: strings.Join(ShowDebugLevelsOptionalArgs, " "),
	AliasArgs:    strings.Join(DebugRequestAliasArgs, " "),
	SpecialArgs:  &DebugRequestSpecialArgs,
	Comments:     addRegionComment(DebugRequestComments),
	ReqData:      &ormapi.RegionDebugRequest{},
	ReplyData:    &edgeproto.DebugReply{},
	Run:          runRest("/auth/ctrl/ShowDebugLevels"),
	StreamOut:    true,
}

var RunDebugCmd = &cli.Command{
	Use:          "RunDebug",
	RequiredArgs: strings.Join(RunDebugRequiredArgs, " "),
	OptionalArgs: strings.Join(RunDebugOptionalArgs, " "),
	AliasArgs:    strings.Join(DebugRequestAliasArgs, " "),
	SpecialArgs:  &DebugRequestSpecialArgs,
	Comments:     addRegionComment(DebugRequestComments),
	ReqData:      &ormapi.RegionDebugRequest{},
	ReplyData:    &edgeproto.DebugReply{},
	Run:          runRest("/auth/ctrl/RunDebug"),
	StreamOut:    true,
}

var DebugApiCmds = []*cli.Command{
	EnableDebugLevelsCmd,
	DisableDebugLevelsCmd,
	ShowDebugLevelsCmd,
	RunDebugCmd,
}

var EnableDebugLevelsRequiredArgs = []string{
	"levels",
}
var EnableDebugLevelsOptionalArgs = []string{
	"name",
	"type",
	"organization",
	"cloudlet",
	"region",
	"pretty",
	"id",
	"args",
}
var DisableDebugLevelsRequiredArgs = []string{
	"levels",
}
var DisableDebugLevelsOptionalArgs = []string{
	"name",
	"type",
	"organization",
	"cloudlet",
	"region",
	"pretty",
	"id",
	"args",
}
var ShowDebugLevelsRequiredArgs = []string{}
var ShowDebugLevelsOptionalArgs = []string{
	"name",
	"type",
	"organization",
	"cloudlet",
	"region",
	"pretty",
	"id",
	"args",
}
var RunDebugRequiredArgs = []string{}
var RunDebugOptionalArgs = []string{
	"name",
	"type",
	"organization",
	"cloudlet",
	"region",
	"cmd",
	"pretty",
	"id",
	"args",
}
var DebugRequestRequiredArgs = []string{}
var DebugRequestOptionalArgs = []string{
	"name",
	"type",
	"organization",
	"cloudlet",
	"region",
	"levels",
	"cmd",
	"pretty",
	"id",
	"args",
}
var DebugRequestAliasArgs = []string{
	"name=debugrequest.node.name",
	"type=debugrequest.node.type",
	"organization=debugrequest.node.cloudletkey.organization",
	"cloudlet=debugrequest.node.cloudletkey.name",
	"region=debugrequest.node.region",
	"levels=debugrequest.levels",
	"cmd=debugrequest.cmd",
	"pretty=debugrequest.pretty",
	"id=debugrequest.id",
	"args=debugrequest.args",
}
var DebugRequestComments = map[string]string{
	"name":         "Name or hostname of node",
	"type":         "Node type",
	"organization": "Organization of the cloudlet site",
	"cloudlet":     "Name of the cloudlet",
	"region":       "Region the node is in",
	"levels":       "Comma separated list of debug level names: etcd,api,notify,dmereq,locapi,mexos,metrics,upgrade,info,sampled",
	"cmd":          "Debug command",
	"pretty":       "if possible, make output pretty",
	"id":           "Id used internally",
	"args":         "Additional arguments for cmd",
}
var DebugRequestSpecialArgs = map[string]string{}
var DebugReplyRequiredArgs = []string{}
var DebugReplyOptionalArgs = []string{
	"node.name",
	"node.type",
	"node.cloudletkey.organization",
	"node.cloudletkey.name",
	"node.region",
	"output",
	"id",
}
var DebugReplyAliasArgs = []string{
	"node.name=debugreply.node.name",
	"node.type=debugreply.node.type",
	"node.cloudletkey.organization=debugreply.node.cloudletkey.organization",
	"node.cloudletkey.name=debugreply.node.cloudletkey.name",
	"node.region=debugreply.node.region",
	"output=debugreply.output",
	"id=debugreply.id",
}
var DebugReplyComments = map[string]string{
	"node.name":                     "Name or hostname of node",
	"node.type":                     "Node type",
	"node.cloudletkey.organization": "Organization of the cloudlet site",
	"node.cloudletkey.name":         "Name of the cloudlet",
	"node.region":                   "Region the node is in",
	"output":                        "Debug output, if any",
	"id":                            "Id used internally",
}
var DebugReplySpecialArgs = map[string]string{}
var DebugDataRequiredArgs = []string{}
var DebugDataOptionalArgs = []string{
	"requests[#].node.name",
	"requests[#].node.type",
	"requests[#].node.cloudletkey.organization",
	"requests[#].node.cloudletkey.name",
	"requests[#].node.region",
	"requests[#].levels",
	"requests[#].cmd",
	"requests[#].pretty",
	"requests[#].id",
	"requests[#].args",
}
var DebugDataAliasArgs = []string{
	"requests[#].node.name=debugdata.requests[#].node.name",
	"requests[#].node.type=debugdata.requests[#].node.type",
	"requests[#].node.cloudletkey.organization=debugdata.requests[#].node.cloudletkey.organization",
	"requests[#].node.cloudletkey.name=debugdata.requests[#].node.cloudletkey.name",
	"requests[#].node.region=debugdata.requests[#].node.region",
	"requests[#].levels=debugdata.requests[#].levels",
	"requests[#].cmd=debugdata.requests[#].cmd",
	"requests[#].pretty=debugdata.requests[#].pretty",
	"requests[#].id=debugdata.requests[#].id",
	"requests[#].args=debugdata.requests[#].args",
}
var DebugDataComments = map[string]string{
	"requests[#].node.name":                     "Name or hostname of node",
	"requests[#].node.type":                     "Node type",
	"requests[#].node.cloudletkey.organization": "Organization of the cloudlet site",
	"requests[#].node.cloudletkey.name":         "Name of the cloudlet",
	"requests[#].node.region":                   "Region the node is in",
	"requests[#].levels":                        "Comma separated list of debug level names: etcd,api,notify,dmereq,locapi,mexos,metrics,upgrade,info,sampled",
	"requests[#].cmd":                           "Debug command",
	"requests[#].pretty":                        "if possible, make output pretty",
	"requests[#].id":                            "Id used internally",
	"requests[#].args":                          "Additional arguments for cmd",
}
var DebugDataSpecialArgs = map[string]string{}
