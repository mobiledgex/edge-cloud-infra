// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: privacypolicy.proto

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

var CreatePrivacyPolicyCmd = &cli.Command{
	Use:          "CreatePrivacyPolicy",
	RequiredArgs: "region " + strings.Join(PrivacyPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(PrivacyPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     addRegionComment(PrivacyPolicyComments),
	ReqData:      &ormapi.RegionPrivacyPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/CreatePrivacyPolicy"),
}

var DeletePrivacyPolicyCmd = &cli.Command{
	Use:          "DeletePrivacyPolicy",
	RequiredArgs: "region " + strings.Join(PrivacyPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(PrivacyPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     addRegionComment(PrivacyPolicyComments),
	ReqData:      &ormapi.RegionPrivacyPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/DeletePrivacyPolicy"),
}

var UpdatePrivacyPolicyCmd = &cli.Command{
	Use:          "UpdatePrivacyPolicy",
	RequiredArgs: "region " + strings.Join(PrivacyPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(PrivacyPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     addRegionComment(PrivacyPolicyComments),
	ReqData:      &ormapi.RegionPrivacyPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run: runRest("/auth/ctrl/UpdatePrivacyPolicy",
		withSetFieldsFunc(setUpdatePrivacyPolicyFields),
	),
}

func setUpdatePrivacyPolicyFields(in map[string]interface{}) {
	// get map for edgeproto object in region struct
	obj := in[strings.ToLower("PrivacyPolicy")]
	if obj == nil {
		return
	}
	objmap, ok := obj.(map[string]interface{})
	if !ok {
		return
	}
	fields := cli.GetSpecifiedFields(objmap, &edgeproto.PrivacyPolicy{}, cli.JsonNamespace)
	// include fields already specified
	if inFields, found := objmap["fields"]; found {
		if fieldsArr, ok := inFields.([]string); ok {
			fields = append(fields, fieldsArr...)
		}
	}
	objmap["fields"] = fields
}

var ShowPrivacyPolicyCmd = &cli.Command{
	Use:          "ShowPrivacyPolicy",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(PrivacyPolicyRequiredArgs, PrivacyPolicyOptionalArgs...), " "),
	AliasArgs:    strings.Join(PrivacyPolicyAliasArgs, " "),
	SpecialArgs:  &PrivacyPolicySpecialArgs,
	Comments:     addRegionComment(PrivacyPolicyComments),
	ReqData:      &ormapi.RegionPrivacyPolicy{},
	ReplyData:    &edgeproto.PrivacyPolicy{},
	Run:          runRest("/auth/ctrl/ShowPrivacyPolicy"),
	StreamOut:    true,
}

var PrivacyPolicyApiCmds = []*cli.Command{
	CreatePrivacyPolicyCmd,
	DeletePrivacyPolicyCmd,
	UpdatePrivacyPolicyCmd,
	ShowPrivacyPolicyCmd,
}

var OutboundSecurityRuleRequiredArgs = []string{}
var OutboundSecurityRuleOptionalArgs = []string{
	"protocol",
	"portrangemin",
	"portrangemax",
	"remotecidr",
}
var OutboundSecurityRuleAliasArgs = []string{
	"protocol=outboundsecurityrule.protocol",
	"portrangemin=outboundsecurityrule.portrangemin",
	"portrangemax=outboundsecurityrule.portrangemax",
	"remotecidr=outboundsecurityrule.remotecidr",
}
var OutboundSecurityRuleComments = map[string]string{
	"protocol":     "tcp, udp, icmp",
	"portrangemin": "TCP or UDP port range start",
	"portrangemax": "TCP or UDP port range end",
	"remotecidr":   "remote CIDR X.X.X.X/X",
}
var OutboundSecurityRuleSpecialArgs = map[string]string{}
var PrivacyPolicyRequiredArgs = []string{
	"cluster-org",
	"name",
}
var PrivacyPolicyOptionalArgs = []string{
	"outboundsecurityrules:#.protocol",
	"outboundsecurityrules:#.portrangemin",
	"outboundsecurityrules:#.portrangemax",
	"outboundsecurityrules:#.remotecidr",
}
var PrivacyPolicyAliasArgs = []string{
	"fields=privacypolicy.fields",
	"cluster-org=privacypolicy.key.organization",
	"name=privacypolicy.key.name",
	"outboundsecurityrules:#.protocol=privacypolicy.outboundsecurityrules:#.protocol",
	"outboundsecurityrules:#.portrangemin=privacypolicy.outboundsecurityrules:#.portrangemin",
	"outboundsecurityrules:#.portrangemax=privacypolicy.outboundsecurityrules:#.portrangemax",
	"outboundsecurityrules:#.remotecidr=privacypolicy.outboundsecurityrules:#.remotecidr",
}
var PrivacyPolicyComments = map[string]string{
	"fields":                               "Fields are used for the Update API to specify which fields to apply",
	"cluster-org":                          "Name of the organization for the cluster that this policy will apply to",
	"name":                                 "Policy name",
	"outboundsecurityrules:#.protocol":     "tcp, udp, icmp",
	"outboundsecurityrules:#.portrangemin": "TCP or UDP port range start",
	"outboundsecurityrules:#.portrangemax": "TCP or UDP port range end",
	"outboundsecurityrules:#.remotecidr":   "remote CIDR X.X.X.X/X",
}
var PrivacyPolicySpecialArgs = map[string]string{
	"privacypolicy.fields": "StringArray",
}
