// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var CreateAppInstCmd = &cli.Command{
	Use:                  "CreateAppInst",
	RequiredArgs:         strings.Join(append([]string{"region"}, AppInstRequiredArgs...), " "),
	OptionalArgs:         strings.Join(AppInstOptionalArgs, " "),
	AliasArgs:            strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:          &AppInstSpecialArgs,
	Comments:             addRegionComment(AppInstComments),
	ReqData:              &ormapi.RegionAppInst{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/CreateAppInst"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var DeleteAppInstCmd = &cli.Command{
	Use:                  "DeleteAppInst",
	RequiredArgs:         strings.Join(append([]string{"region"}, AppInstRequiredArgs...), " "),
	OptionalArgs:         strings.Join(AppInstOptionalArgs, " "),
	AliasArgs:            strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:          &AppInstSpecialArgs,
	Comments:             addRegionComment(AppInstComments),
	ReqData:              &ormapi.RegionAppInst{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/DeleteAppInst"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var UpdateAppInstCmd = &cli.Command{
	Use:                  "UpdateAppInst",
	RequiredArgs:         strings.Join(append([]string{"region"}, AppInstRequiredArgs...), " "),
	OptionalArgs:         strings.Join(AppInstOptionalArgs, " "),
	AliasArgs:            strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:          &AppInstSpecialArgs,
	Comments:             addRegionComment(AppInstComments),
	ReqData:              &ormapi.RegionAppInst{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/UpdateAppInst"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var ShowAppInstCmd = &cli.Command{
	Use:          "ShowAppInst",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(AppInstRequiredArgs, AppInstOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     addRegionComment(AppInstComments),
	ReqData:      &ormapi.RegionAppInst{},
	ReplyData:    &edgeproto.AppInst{},
	Run:          runRest("/auth/ctrl/ShowAppInst"),
	StreamOut:    true,
}
var AppInstApiCmds = []*cli.Command{
	CreateAppInstCmd,
	DeleteAppInstCmd,
	UpdateAppInstCmd,
	ShowAppInstCmd,
}

var AppInstKeyRequiredArgs = []string{}
var AppInstKeyOptionalArgs = []string{
	"appkey.developerkey.name",
	"appkey.name",
	"appkey.version",
	"clusterinstkey.clusterkey.name",
	"clusterinstkey.cloudletkey.operatorkey.name",
	"clusterinstkey.cloudletkey.name",
	"clusterinstkey.developer",
}
var AppInstKeyAliasArgs = []string{
	"appkey.developerkey.name=appinstkey.appkey.developerkey.name",
	"appkey.name=appinstkey.appkey.name",
	"appkey.version=appinstkey.appkey.version",
	"clusterinstkey.clusterkey.name=appinstkey.clusterinstkey.clusterkey.name",
	"clusterinstkey.cloudletkey.operatorkey.name=appinstkey.clusterinstkey.cloudletkey.operatorkey.name",
	"clusterinstkey.cloudletkey.name=appinstkey.clusterinstkey.cloudletkey.name",
	"clusterinstkey.developer=appinstkey.clusterinstkey.developer",
}
var AppInstKeyComments = map[string]string{
	"appkey.developerkey.name":                    "Organization or Company Name that a Developer is part of",
	"appkey.name":                                 "App name",
	"appkey.version":                              "App version",
	"clusterinstkey.clusterkey.name":              "Cluster name",
	"clusterinstkey.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"clusterinstkey.cloudletkey.name":             "Name of the cloudlet",
	"clusterinstkey.developer":                    "Name of Developer that this cluster belongs to",
}
var AppInstKeySpecialArgs = map[string]string{}
var AppInstRequiredArgs = []string{
	"developer",
	"appname",
	"appvers",
	"cluster",
	"operator",
	"cloudlet",
}
var AppInstOptionalArgs = []string{
	"clusterdeveloper",
	"flavor.name",
	"state",
	"errors",
	"crmoverride",
	"runtimeinfo.containerids",
	"runtimeinfo.consoleurl",
	"autoclusteripaccess",
	"forceupdate",
	"updatemultiple",
}
var AppInstAliasArgs = []string{
	"developer=appinst.key.appkey.developerkey.name",
	"appname=appinst.key.appkey.name",
	"appvers=appinst.key.appkey.version",
	"cluster=appinst.key.clusterinstkey.clusterkey.name",
	"operator=appinst.key.clusterinstkey.cloudletkey.operatorkey.name",
	"cloudlet=appinst.key.clusterinstkey.cloudletkey.name",
	"clusterdeveloper=appinst.key.clusterinstkey.developer",
	"cloudletloc.latitude=appinst.cloudletloc.latitude",
	"cloudletloc.longitude=appinst.cloudletloc.longitude",
	"cloudletloc.horizontalaccuracy=appinst.cloudletloc.horizontalaccuracy",
	"cloudletloc.verticalaccuracy=appinst.cloudletloc.verticalaccuracy",
	"cloudletloc.altitude=appinst.cloudletloc.altitude",
	"cloudletloc.course=appinst.cloudletloc.course",
	"cloudletloc.speed=appinst.cloudletloc.speed",
	"cloudletloc.timestamp.seconds=appinst.cloudletloc.timestamp.seconds",
	"cloudletloc.timestamp.nanos=appinst.cloudletloc.timestamp.nanos",
	"uri=appinst.uri",
	"liveness=appinst.liveness",
	"mappedports.proto=appinst.mappedports.proto",
	"mappedports.internalport=appinst.mappedports.internalport",
	"mappedports.publicport=appinst.mappedports.publicport",
	"mappedports.pathprefix=appinst.mappedports.pathprefix",
	"mappedports.fqdnprefix=appinst.mappedports.fqdnprefix",
	"mappedports.endport=appinst.mappedports.endport",
	"flavor.name=appinst.flavor.name",
	"state=appinst.state",
	"errors=appinst.errors",
	"crmoverride=appinst.crmoverride",
	"runtimeinfo.containerids=appinst.runtimeinfo.containerids",
	"runtimeinfo.consoleurl=appinst.runtimeinfo.consoleurl",
	"createdat.seconds=appinst.createdat.seconds",
	"createdat.nanos=appinst.createdat.nanos",
	"autoclusteripaccess=appinst.autoclusteripaccess",
	"status.tasknumber=appinst.status.tasknumber",
	"status.maxtasks=appinst.status.maxtasks",
	"status.taskname=appinst.status.taskname",
	"status.stepname=appinst.status.stepname",
	"revision=appinst.revision",
	"forceupdate=appinst.forceupdate",
	"updatemultiple=appinst.updatemultiple",
}
var AppInstComments = map[string]string{
	"developer":                      "Organization or Company Name that a Developer is part of",
	"appname":                        "App name",
	"appvers":                        "App version",
	"cluster":                        "Cluster name",
	"operator":                       "Company or Organization name of the operator",
	"cloudlet":                       "Name of the cloudlet",
	"clusterdeveloper":               "Name of Developer that this cluster belongs to",
	"cloudletloc.latitude":           "latitude in WGS 84 coordinates",
	"cloudletloc.longitude":          "longitude in WGS 84 coordinates",
	"cloudletloc.horizontalaccuracy": "horizontal accuracy (radius in meters)",
	"cloudletloc.verticalaccuracy":   "veritical accuracy (meters)",
	"cloudletloc.altitude":           "On android only lat and long are guaranteed to be supplied altitude in meters",
	"cloudletloc.course":             "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"cloudletloc.speed":              "speed (IOS) / velocity (Android) (meters/sec)",
	"uri":                            "Base FQDN (not really URI) for the App. See Service FQDN for endpoint access.",
	"liveness":                       "Liveness of instance (see Liveness), one of LivenessUnknown, LivenessStatic, LivenessDynamic",
	"mappedports.proto":              "TCP (L4), UDP (L4), or HTTP (L7) protocol, one of LProtoUnknown, LProtoTcp, LProtoUdp, LProtoHttp",
	"mappedports.internalport":       "Container port",
	"mappedports.publicport":         "Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)",
	"mappedports.pathprefix":         "Public facing path for HTTP L7 access.",
	"mappedports.fqdnprefix":         "FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.",
	"mappedports.endport":            "A non-zero end port indicates this is a port range from internal port to end port, inclusive.",
	"flavor.name":                    "Flavor name",
	"state":                          "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare",
	"errors":                         "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"crmoverride":                    "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"runtimeinfo.containerids":       "List of container names",
	"runtimeinfo.consoleurl":         "URL to access VM console",
	"autoclusteripaccess":            "IpAccess for auto-clusters. Ignored otherwise., one of IpAccessUnknown, IpAccessDedicated, IpAccessDedicatedOrShared, IpAccessShared",
	"revision":                       "Revision increments each time the App is updated.  Updating the App Instance will sync the revision with that of the App",
	"forceupdate":                    "Force Appinst update when UpdateAppInst is done if revision matches",
	"updatemultiple":                 "Allow multiple instances to be updated at once",
}
var AppInstSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"runtimeinfo.containerids": "StringArray",
}
var AppInstRuntimeRequiredArgs = []string{}
var AppInstRuntimeOptionalArgs = []string{
	"containerids",
	"consoleurl",
}
var AppInstRuntimeAliasArgs = []string{
	"containerids=appinstruntime.containerids",
	"consoleurl=appinstruntime.consoleurl",
}
var AppInstRuntimeComments = map[string]string{
	"containerids": "List of container names",
	"consoleurl":   "URL to access VM console",
}
var AppInstRuntimeSpecialArgs = map[string]string{
	"containerids": "StringArray",
}
var AppInstInfoRequiredArgs = []string{
	"key.appkey.developerkey.name",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.operatorkey.name",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.developer",
}
var AppInstInfoOptionalArgs = []string{
	"notifyid",
	"state",
	"errors",
	"runtimeinfo.containerids",
	"runtimeinfo.consoleurl",
	"status.tasknumber",
	"status.maxtasks",
	"status.taskname",
	"status.stepname",
}
var AppInstInfoAliasArgs = []string{
	"key.appkey.developerkey.name=appinstinfo.key.appkey.developerkey.name",
	"key.appkey.name=appinstinfo.key.appkey.name",
	"key.appkey.version=appinstinfo.key.appkey.version",
	"key.clusterinstkey.clusterkey.name=appinstinfo.key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.operatorkey.name=appinstinfo.key.clusterinstkey.cloudletkey.operatorkey.name",
	"key.clusterinstkey.cloudletkey.name=appinstinfo.key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.developer=appinstinfo.key.clusterinstkey.developer",
	"notifyid=appinstinfo.notifyid",
	"state=appinstinfo.state",
	"errors=appinstinfo.errors",
	"runtimeinfo.containerids=appinstinfo.runtimeinfo.containerids",
	"runtimeinfo.consoleurl=appinstinfo.runtimeinfo.consoleurl",
	"status.tasknumber=appinstinfo.status.tasknumber",
	"status.maxtasks=appinstinfo.status.maxtasks",
	"status.taskname=appinstinfo.status.taskname",
	"status.stepname=appinstinfo.status.stepname",
}
var AppInstInfoComments = map[string]string{
	"key.appkey.developerkey.name":                    "Organization or Company Name that a Developer is part of",
	"key.appkey.name":                                 "App name",
	"key.appkey.version":                              "App version",
	"key.clusterinstkey.clusterkey.name":              "Cluster name",
	"key.clusterinstkey.cloudletkey.operatorkey.name": "Company or Organization name of the operator",
	"key.clusterinstkey.cloudletkey.name":             "Name of the cloudlet",
	"key.clusterinstkey.developer":                    "Name of Developer that this cluster belongs to",
	"notifyid":                                        "Id of client assigned by server (internal use only)",
	"state":                                           "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare",
	"errors":                                          "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"runtimeinfo.containerids":                        "List of container names",
	"runtimeinfo.consoleurl":                          "URL to access VM console",
}
var AppInstInfoSpecialArgs = map[string]string{
	"errors":                   "StringArray",
	"runtimeinfo.containerids": "StringArray",
}
var AppInstMetricsRequiredArgs = []string{}
var AppInstMetricsOptionalArgs = []string{
	"something",
}
var AppInstMetricsAliasArgs = []string{
	"something=appinstmetrics.something",
}
var AppInstMetricsComments = map[string]string{
	"something": "what goes here? Note that metrics for grpc calls can be done by a prometheus interceptor in grpc, so adding call metrics here may be redundant unless theyre needed for billing.",
}
var AppInstMetricsSpecialArgs = map[string]string{}
