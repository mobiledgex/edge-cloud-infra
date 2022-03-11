package ormctl

import (
	fmt "fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
)

const AuditGroup = "Audit"

func init() {
	cmds := []*ApiCommand{&ApiCommand{
		Name:         "ShowAuditSelf",
		Use:          "showself",
		OptionalArgs: "limit operation tags starttime endtime startage endage",
		AliasArgs:    strings.Join(AuditAliasArgs, " "),
		Comments:     AuditSelfComments,
		SpecialArgs:  &AuditSpecialArgs,
		ReqData:      &ormapi.AuditQuery{},
		ReplyData:    &[]ormapi.AuditResponse{},
		Path:         "/auth/audit/showself",
	}, &ApiCommand{
		Name:         "ShowAuditOrg",
		Use:          "showorg",
		OptionalArgs: "org limit operation tags starttime endtime startage endage",
		AliasArgs:    strings.Join(AuditAliasArgs, " "),
		Comments:     AuditOrgComments,
		SpecialArgs:  &AuditSpecialArgs,
		ReqData:      &ormapi.AuditQuery{},
		ReplyData:    &[]ormapi.AuditResponse{},
		Path:         "/auth/audit/showorg",
	}, &ApiCommand{
		Name:      "ShowAuditOperations",
		Use:       "operations",
		ReplyData: &[]string{},
		Path:      "/auth/audit/operations",
	}}
	AllApis.AddGroup(AuditGroup, "Show audit logs", cmds)
}

var tagsComment = fmt.Sprintf("key=value tag, may be specified multiple times, key may include %s", strings.Join(edgeproto.AllKeyTags, ", "))

var AuditOrgComments = map[string]string{
	"username":  "filter by user name",
	"org":       "filter by organization name",
	"limit":     "limit the number of returned results (default 100)",
	"operation": "operation name (see operations command)",
	"tags":      tagsComment,
	"starttime": "absolute time of search range start (RFC3339)",
	"endtime":   "absolute time of search range end (RFC3339)",
	"startage":  "relative age from now of search range start (default 48h)",
	"endage":    "relative age from now of search range end (default 0)",
}

var AuditSelfComments = map[string]string{
	"org":       "filter by organization name",
	"limit":     "limit the number of returned results (default 100)",
	"operation": "operation name (see operations command)",
	"tags":      tagsComment,
	"starttime": "absolute time of search range start",
	"endtime":   "absolute time of search range end",
	"startage":  "relative age from now of search range start (default 48h)",
	"endage":    "relative age from now of search range end (default 0)",
}

var AuditSpecialArgs = map[string]string{
	"tags": "StringToString",
}

var AuditAliasArgs = []string{
	"starttime=timerange.starttime",
	"endtime=timerange.endtime",
	"startage=timerange.startage",
	"endage=timerange.endage",
}