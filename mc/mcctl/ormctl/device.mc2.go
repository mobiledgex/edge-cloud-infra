// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: device.proto

package ormctl

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
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

var InjectDeviceCmd = &ApiCommand{
	Name:         "InjectDevice",
	Use:          "inject",
	Short:        "Inject a device",
	RequiredArgs: "region " + strings.Join(DeviceRequiredArgs, " "),
	OptionalArgs: strings.Join(DeviceOptionalArgs, " "),
	AliasArgs:    strings.Join(DeviceAliasArgs, " "),
	SpecialArgs:  &DeviceSpecialArgs,
	Comments:     addRegionComment(DeviceComments),
	ReqData:      &ormapi.RegionDevice{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/InjectDevice",
	ProtobufApi:  true,
}

var ShowDeviceCmd = &ApiCommand{
	Name:         "ShowDevice",
	Use:          "show",
	Short:        "Show devices",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(DeviceRequiredArgs, DeviceOptionalArgs...), " "),
	AliasArgs:    strings.Join(DeviceAliasArgs, " "),
	SpecialArgs:  &DeviceSpecialArgs,
	Comments:     addRegionComment(DeviceComments),
	ReqData:      &ormapi.RegionDevice{},
	ReplyData:    &edgeproto.Device{},
	Path:         "/auth/ctrl/ShowDevice",
	StreamOut:    true,
	ProtobufApi:  true,
}

var EvictDeviceCmd = &ApiCommand{
	Name:         "EvictDevice",
	Use:          "evict",
	Short:        "Evict a device",
	RequiredArgs: "region " + strings.Join(DeviceRequiredArgs, " "),
	OptionalArgs: strings.Join(DeviceOptionalArgs, " "),
	AliasArgs:    strings.Join(DeviceAliasArgs, " "),
	SpecialArgs:  &DeviceSpecialArgs,
	Comments:     addRegionComment(DeviceComments),
	ReqData:      &ormapi.RegionDevice{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/EvictDevice",
	ProtobufApi:  true,
}

var ShowDeviceReportCmd = &ApiCommand{
	Name:         "ShowDeviceReport",
	Use:          "showreport",
	Short:        "Device Reports API.",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(DeviceReportRequiredArgs, DeviceReportOptionalArgs...), " "),
	AliasArgs:    strings.Join(DeviceReportAliasArgs, " "),
	SpecialArgs:  &DeviceReportSpecialArgs,
	Comments:     addRegionComment(DeviceReportComments),
	ReqData:      &ormapi.RegionDeviceReport{},
	ReplyData:    &edgeproto.Device{},
	Path:         "/auth/ctrl/ShowDeviceReport",
	StreamOut:    true,
	ProtobufApi:  true,
}
var DeviceApiCmds = []*ApiCommand{
	InjectDeviceCmd,
	ShowDeviceCmd,
	EvictDeviceCmd,
	ShowDeviceReportCmd,
}

const DeviceGroup = "Device"

func init() {
	AllApis.AddGroup(DeviceGroup, "Manage Devices", DeviceApiCmds)
}

var DeviceReportRequiredArgs = []string{
	"key.uniqueidtype",
	"key.uniqueid",
}
var DeviceReportOptionalArgs = []string{
	"begin.seconds",
	"begin.nanos",
	"end.seconds",
	"end.nanos",
}
var DeviceReportAliasArgs = []string{
	"key.uniqueidtype=devicereport.key.uniqueidtype",
	"key.uniqueid=devicereport.key.uniqueid",
	"begin.seconds=devicereport.begin.seconds",
	"begin.nanos=devicereport.begin.nanos",
	"end.seconds=devicereport.end.seconds",
	"end.nanos=devicereport.end.nanos",
}
var DeviceReportComments = map[string]string{
	"key.uniqueidtype": "Type of unique ID provided by the client",
	"key.uniqueid":     "Unique identification of the client device or user. May be overridden by the server.",
	"begin.seconds":    "Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.",
	"begin.nanos":      "Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive.",
	"end.seconds":      "Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.",
	"end.nanos":        "Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive.",
}
var DeviceReportSpecialArgs = map[string]string{}
var DeviceRequiredArgs = []string{
	"key.uniqueidtype",
	"key.uniqueid",
}
var DeviceOptionalArgs = []string{
	"firstseen.seconds",
	"firstseen.nanos",
	"lastseen.seconds",
	"lastseen.nanos",
	"notifyid",
}
var DeviceAliasArgs = []string{
	"fields=device.fields",
	"key.uniqueidtype=device.key.uniqueidtype",
	"key.uniqueid=device.key.uniqueid",
	"firstseen.seconds=device.firstseen.seconds",
	"firstseen.nanos=device.firstseen.nanos",
	"lastseen.seconds=device.lastseen.seconds",
	"lastseen.nanos=device.lastseen.nanos",
	"notifyid=device.notifyid",
}
var DeviceComments = map[string]string{
	"key.uniqueidtype":  "Type of unique ID provided by the client",
	"key.uniqueid":      "Unique identification of the client device or user. May be overridden by the server.",
	"firstseen.seconds": "Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.",
	"firstseen.nanos":   "Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive.",
	"lastseen.seconds":  "Represents seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.",
	"lastseen.nanos":    "Non-negative fractions of a second at nanosecond resolution. Negative second values with fractions must still have non-negative nanos values that count forward in time. Must be from 0 to 999,999,999 inclusive.",
	"notifyid":          "Id of client assigned by server (internal use only)",
}
var DeviceSpecialArgs = map[string]string{
	"device.fields": "StringArray",
}
