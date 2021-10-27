// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: trustpolicyexception.proto

package ormctl

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var CreateTrustPolicyExceptionCmd = &ApiCommand{
	Name:         "CreateTrustPolicyException",
	Use:          "create",
	Short:        "Create a Trust Policy Exception, by App Developer Organization",
	RequiredArgs: "region " + strings.Join(TrustPolicyExceptionRequiredArgs, " "),
	OptionalArgs: strings.Join(TrustPolicyExceptionOptionalArgs, " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     addRegionComment(TrustPolicyExceptionComments),
	ReqData:      &ormapi.RegionTrustPolicyException{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/CreateTrustPolicyException",
	ProtobufApi:  true,
}

var UpdateTrustPolicyExceptionCmd = &ApiCommand{
	Name:         "UpdateTrustPolicyException",
	Use:          "update",
	Short:        "Update a Trust Policy Exception, by Operator Organization",
	RequiredArgs: "region " + strings.Join(UpdateTrustPolicyExceptionRequiredArgs, " "),
	OptionalArgs: strings.Join(UpdateTrustPolicyExceptionOptionalArgs, " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     addRegionComment(TrustPolicyExceptionComments),
	ReqData:      &ormapi.RegionTrustPolicyException{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/UpdateTrustPolicyException",
	ProtobufApi:  true,
}

var DeleteTrustPolicyExceptionCmd = &ApiCommand{
	Name:         "DeleteTrustPolicyException",
	Use:          "delete",
	Short:        "Delete a Trust Policy Exception, by App Developer Organization",
	RequiredArgs: "region " + strings.Join(TrustPolicyExceptionRequiredArgs, " "),
	OptionalArgs: strings.Join(TrustPolicyExceptionOptionalArgs, " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     addRegionComment(TrustPolicyExceptionComments),
	ReqData:      &ormapi.RegionTrustPolicyException{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/DeleteTrustPolicyException",
	ProtobufApi:  true,
}

var ShowTrustPolicyExceptionCmd = &ApiCommand{
	Name:         "ShowTrustPolicyException",
	Use:          "show",
	Short:        "Show Trust Policy Exceptions. Any fields specified will be used to filter results.",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(TrustPolicyExceptionRequiredArgs, TrustPolicyExceptionOptionalArgs...), " "),
	AliasArgs:    strings.Join(TrustPolicyExceptionAliasArgs, " "),
	SpecialArgs:  &TrustPolicyExceptionSpecialArgs,
	Comments:     addRegionComment(TrustPolicyExceptionComments),
	ReqData:      &ormapi.RegionTrustPolicyException{},
	ReplyData:    &edgeproto.TrustPolicyException{},
	Path:         "/auth/ctrl/ShowTrustPolicyException",
	StreamOut:    true,
	ProtobufApi:  true,
}
var TrustPolicyExceptionApiCmds = []*ApiCommand{
	CreateTrustPolicyExceptionCmd,
	UpdateTrustPolicyExceptionCmd,
	DeleteTrustPolicyExceptionCmd,
	ShowTrustPolicyExceptionCmd,
}

const TrustPolicyExceptionGroup = "TrustPolicyException"

func init() {
	AllApis.AddGroup(TrustPolicyExceptionGroup, "Manage TrustPolicyExceptions", TrustPolicyExceptionApiCmds)
}

var UpdateTrustPolicyExceptionRequiredArgs = []string{
	"app-org",
	"app-name",
	"app-ver",
	"cloudletpool-org",
	"cloudletpool-name",
	"name",
	"state",
}
var UpdateTrustPolicyExceptionOptionalArgs = []string{
	"outboundsecurityrules:empty",
	"outboundsecurityrules:#.protocol",
	"outboundsecurityrules:#.portrangemin",
	"outboundsecurityrules:#.portrangemax",
	"outboundsecurityrules:#.remotecidr",
}
var TrustPolicyExceptionRequiredArgs = []string{
	"app-org",
	"app-name",
	"app-ver",
	"cloudletpool-org",
	"cloudletpool-name",
	"name",
}
var TrustPolicyExceptionOptionalArgs = []string{
	"state",
	"outboundsecurityrules:empty",
	"outboundsecurityrules:#.protocol",
	"outboundsecurityrules:#.portrangemin",
	"outboundsecurityrules:#.portrangemax",
	"outboundsecurityrules:#.remotecidr",
}
var TrustPolicyExceptionAliasArgs = []string{
	"fields=trustpolicyexception.fields",
	"app-org=trustpolicyexception.key.appkey.organization",
	"app-name=trustpolicyexception.key.appkey.name",
	"app-ver=trustpolicyexception.key.appkey.version",
	"cloudletpool-org=trustpolicyexception.key.cloudletpoolkey.organization",
	"cloudletpool-name=trustpolicyexception.key.cloudletpoolkey.name",
	"name=trustpolicyexception.key.name",
	"state=trustpolicyexception.state",
	"outboundsecurityrules:empty=trustpolicyexception.outboundsecurityrules:empty",
	"outboundsecurityrules:#.protocol=trustpolicyexception.outboundsecurityrules:#.protocol",
	"outboundsecurityrules:#.portrangemin=trustpolicyexception.outboundsecurityrules:#.portrangemin",
	"outboundsecurityrules:#.portrangemax=trustpolicyexception.outboundsecurityrules:#.portrangemax",
	"outboundsecurityrules:#.remotecidr=trustpolicyexception.outboundsecurityrules:#.remotecidr",
}
var TrustPolicyExceptionComments = map[string]string{
	"fields":                               "Fields are used for the Update API to specify which fields to apply",
	"app-org":                              "App developer organization",
	"app-name":                             "App name",
	"app-ver":                              "App version",
	"cloudletpool-org":                     "Name of the organization this pool belongs to",
	"cloudletpool-name":                    "CloudletPool Name",
	"name":                                 "TrustPolicyExceptionKey name",
	"state":                                "State of the exception within the approval process, one of Unknown, ApprovalRequested, Active, Rejected",
	"outboundsecurityrules:empty":          "List of outbound security rules for whitelisting traffic, specify outboundsecurityrules:empty=true to clear",
	"outboundsecurityrules:#.protocol":     "tcp, udp, icmp",
	"outboundsecurityrules:#.portrangemin": "TCP or UDP port range start",
	"outboundsecurityrules:#.portrangemax": "TCP or UDP port range end",
	"outboundsecurityrules:#.remotecidr":   "remote CIDR X.X.X.X/X",
}
var TrustPolicyExceptionSpecialArgs = map[string]string{
	"trustpolicyexception.fields": "StringArray",
}
