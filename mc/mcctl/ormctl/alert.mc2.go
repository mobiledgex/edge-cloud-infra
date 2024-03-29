// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: alert.proto

package ormctl

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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

var ShowAlertCmd = &ApiCommand{
	Name:         "ShowAlert",
	Use:          "show",
	Short:        "Show alerts",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(AlertRequiredArgs, AlertOptionalArgs...), " "),
	AliasArgs:    strings.Join(AlertAliasArgs, " "),
	SpecialArgs:  &AlertSpecialArgs,
	Comments:     addRegionComment(AlertComments),
	ReqData:      &ormapi.RegionAlert{},
	ReplyData:    &edgeproto.Alert{},
	Path:         "/auth/ctrl/ShowAlert",
	StreamOut:    true,
	ProtobufApi:  true,
}
var AlertApiCmds = []*ApiCommand{
	ShowAlertCmd,
}

const AlertGroup = "Alert"

func init() {
	AllApis.AddGroup(AlertGroup, "Manage Alerts", AlertApiCmds)
}

var AlertRequiredArgs = []string{}
var AlertOptionalArgs = []string{
	"labels",
	"annotations",
	"state",
	"activeat",
	"value",
	"notifyid",
	"controller",
}
var AlertAliasArgs = []string{
	"labels=alert.labels",
	"annotations=alert.annotations",
	"state=alert.state",
	"activeat=alert.activeat",
	"value=alert.value",
	"notifyid=alert.notifyid",
	"controller=alert.controller",
}
var AlertComments = map[string]string{
	"labels":      "Labels uniquely define the alert",
	"annotations": "Annotations are extra information about the alert",
	"state":       "State of the alert",
	"activeat":    "When alert became active",
	"value":       "Any value associated with alert",
	"notifyid":    "Id of client assigned by server (internal use only)",
	"controller":  "Connected controller unique id",
}
var AlertSpecialArgs = map[string]string{
	"alert.annotations": "StringToString",
	"alert.labels":      "StringToString",
}
