// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: client.proto

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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var ShowAppInstClientCmd = &cli.Command{
	Use:                  "ShowAppInstClient",
	RequiredArgs:         "region",
	OptionalArgs:         strings.Join(append(AppInstClientRequiredArgs, AppInstClientOptionalArgs...), " "),
	AliasArgs:            strings.Join(AppInstClientAliasArgs, " "),
	SpecialArgs:          &AppInstClientSpecialArgs,
	Comments:             addRegionComment(AppInstClientComments),
	ReqData:              &ormapi.RegionAppInstClient{},
	ReplyData:            &edgeproto.AppInstClient{},
	Run:                  runRest("/auth/ctrl/ShowAppInstClient"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var AppInstClientApiCmds = []*cli.Command{
	ShowAppInstClientCmd,
}

var AppInstClientKeyRequiredArgs = []string{}
var AppInstClientKeyOptionalArgs = []string{
	"appinstkey.appkey.developerkey.name",
	"appinstkey.appkey.name",
	"appinstkey.appkey.version",
	"appinstkey.clusterinstkey.clusterkey.name",
	"appinstkey.clusterinstkey.cloudletkey.operatorkey.name",
	"appinstkey.clusterinstkey.cloudletkey.name",
	"appinstkey.clusterinstkey.developer",
	"uuid",
}
var AppInstClientKeyAliasArgs = []string{
	"appinstkey.appkey.developerkey.name=appinstclientkey.appinstkey.appkey.developerkey.name",
	"appinstkey.appkey.name=appinstclientkey.appinstkey.appkey.name",
	"appinstkey.appkey.version=appinstclientkey.appinstkey.appkey.version",
	"appinstkey.clusterinstkey.clusterkey.name=appinstclientkey.appinstkey.clusterinstkey.clusterkey.name",
	"appinstkey.clusterinstkey.cloudletkey.operatorkey.name=appinstclientkey.appinstkey.clusterinstkey.cloudletkey.operatorkey.name",
	"appinstkey.clusterinstkey.cloudletkey.name=appinstclientkey.appinstkey.clusterinstkey.cloudletkey.name",
	"appinstkey.clusterinstkey.developer=appinstclientkey.appinstkey.clusterinstkey.developer",
	"uuid=appinstclientkey.uuid",
}
var AppInstClientKeyComments = map[string]string{
	"appinstkey.appkey.developerkey.name":                    "Organization or Company Name that a Developer is part of",
	"appinstkey.appkey.name":                                 "App name",
	"appinstkey.appkey.version":                              "App version",
	"appinstkey.clusterinstkey.clusterkey.name":              "Cluster name",
	"appinstkey.clusterinstkey.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"appinstkey.clusterinstkey.cloudletkey.name":             "Name of the cloudlet",
	"appinstkey.clusterinstkey.developer":                    "Name of Developer that this cluster belongs to",
	"uuid":                                                   "App name",
}
var AppInstClientKeySpecialArgs = map[string]string{}
var AppInstClientRequiredArgs = []string{}
var AppInstClientOptionalArgs = []string{
	"clientkey.appinstkey.appkey.developerkey.name",
	"clientkey.appinstkey.appkey.name",
	"clientkey.appinstkey.appkey.version",
	"clientkey.appinstkey.clusterinstkey.clusterkey.name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.operatorkey.name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.name",
	"clientkey.appinstkey.clusterinstkey.developer",
	"clientkey.uuid",
	"location.latitude",
	"location.longitude",
	"location.horizontalaccuracy",
	"location.verticalaccuracy",
	"location.altitude",
	"location.course",
	"location.speed",
	"location.timestamp.seconds",
	"location.timestamp.nanos",
	"notifyid",
	"status",
}
var AppInstClientAliasArgs = []string{
	"clientkey.appinstkey.appkey.developerkey.name=appinstclient.clientkey.appinstkey.appkey.developerkey.name",
	"clientkey.appinstkey.appkey.name=appinstclient.clientkey.appinstkey.appkey.name",
	"clientkey.appinstkey.appkey.version=appinstclient.clientkey.appinstkey.appkey.version",
	"clientkey.appinstkey.clusterinstkey.clusterkey.name=appinstclient.clientkey.appinstkey.clusterinstkey.clusterkey.name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.operatorkey.name=appinstclient.clientkey.appinstkey.clusterinstkey.cloudletkey.operatorkey.name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.name=appinstclient.clientkey.appinstkey.clusterinstkey.cloudletkey.name",
	"clientkey.appinstkey.clusterinstkey.developer=appinstclient.clientkey.appinstkey.clusterinstkey.developer",
	"clientkey.uuid=appinstclient.clientkey.uuid",
	"location.latitude=appinstclient.location.latitude",
	"location.longitude=appinstclient.location.longitude",
	"location.horizontalaccuracy=appinstclient.location.horizontalaccuracy",
	"location.verticalaccuracy=appinstclient.location.verticalaccuracy",
	"location.altitude=appinstclient.location.altitude",
	"location.course=appinstclient.location.course",
	"location.speed=appinstclient.location.speed",
	"location.timestamp.seconds=appinstclient.location.timestamp.seconds",
	"location.timestamp.nanos=appinstclient.location.timestamp.nanos",
	"notifyid=appinstclient.notifyid",
	"status=appinstclient.status",
}
var AppInstClientComments = map[string]string{
	"clientkey.appinstkey.appkey.developerkey.name":                    "Organization or Company Name that a Developer is part of",
	"clientkey.appinstkey.appkey.name":                                 "App name",
	"clientkey.appinstkey.appkey.version":                              "App version",
	"clientkey.appinstkey.clusterinstkey.clusterkey.name":              "Cluster name",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"clientkey.appinstkey.clusterinstkey.cloudletkey.name":             "Name of the cloudlet",
	"clientkey.appinstkey.clusterinstkey.developer":                    "Name of Developer that this cluster belongs to",
	"clientkey.uuid":              "App name",
	"location.latitude":           "latitude in WGS 84 coordinates",
	"location.longitude":          "longitude in WGS 84 coordinates",
	"location.horizontalaccuracy": "horizontal accuracy (radius in meters)",
	"location.verticalaccuracy":   "vertical accuracy (meters)",
	"location.altitude":           "On android only lat and long are guaranteed to be supplied altitude in meters",
	"location.course":             "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"location.speed":              "speed (IOS) / velocity (Android) (meters/sec)",
	"notifyid":                    "Id of client assigned by server (internal use only)",
	"status":                      "Status return, one of FindUnknown, FindFound, FindNotfound",
}
var AppInstClientSpecialArgs = map[string]string{}
