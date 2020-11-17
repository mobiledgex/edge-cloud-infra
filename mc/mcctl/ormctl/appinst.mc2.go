// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinst.proto

package ormctl

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cli"
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

var CreateAppInstCmd = &cli.Command{
	Use:                  "CreateAppInst",
	RequiredArgs:         "region " + strings.Join(CreateAppInstRequiredArgs, " "),
	OptionalArgs:         strings.Join(CreateAppInstOptionalArgs, " "),
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
	RequiredArgs:         "region " + strings.Join(DeleteAppInstRequiredArgs, " "),
	OptionalArgs:         strings.Join(DeleteAppInstOptionalArgs, " "),
	AliasArgs:            strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:          &AppInstSpecialArgs,
	Comments:             addRegionComment(AppInstComments),
	ReqData:              &ormapi.RegionAppInst{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/DeleteAppInst"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var RefreshAppInstCmd = &cli.Command{
	Use:                  "RefreshAppInst",
	RequiredArgs:         "region " + strings.Join(RefreshAppInstRequiredArgs, " "),
	OptionalArgs:         strings.Join(RefreshAppInstOptionalArgs, " "),
	AliasArgs:            strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:          &AppInstSpecialArgs,
	Comments:             addRegionComment(AppInstComments),
	ReqData:              &ormapi.RegionAppInst{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/RefreshAppInst"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var UpdateAppInstCmd = &cli.Command{
	Use:          "UpdateAppInst",
	RequiredArgs: "region " + strings.Join(UpdateAppInstRequiredArgs, " "),
	OptionalArgs: strings.Join(UpdateAppInstOptionalArgs, " "),
	AliasArgs:    strings.Join(AppInstAliasArgs, " "),
	SpecialArgs:  &AppInstSpecialArgs,
	Comments:     addRegionComment(AppInstComments),
	ReqData:      &ormapi.RegionAppInst{},
	ReplyData:    &edgeproto.Result{},
	Run: runRest("/auth/ctrl/UpdateAppInst",
		withSetFieldsFunc(setUpdateAppInstFields),
	),
	StreamOut:            true,
	StreamOutIncremental: true,
}

func setUpdateAppInstFields(in map[string]interface{}) {
	// get map for edgeproto object in region struct
	obj := in[strings.ToLower("AppInst")]
	if obj == nil {
		return
	}
	objmap, ok := obj.(map[string]interface{})
	if !ok {
		return
	}
	fields := cli.GetSpecifiedFields(objmap, &edgeproto.AppInst{}, cli.JsonNamespace)
	// include fields already specified
	if inFields, found := objmap["fields"]; found {
		if fieldsArr, ok := inFields.([]string); ok {
			fields = append(fields, fieldsArr...)
		}
	}
	objmap["fields"] = fields
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
	RefreshAppInstCmd,
	UpdateAppInstCmd,
	ShowAppInstCmd,
}

var CreateAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var CreateAppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"flavor",
	"crmoverride",
	"autoclusteripaccess",
	"configs:#.kind",
	"configs:#.config",
	"sharedvolumesize",
	"healthcheck",
	"privacypolicy",
	"vmflavor",
	"optres",
}
var DeleteAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var DeleteAppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"flavor",
	"crmoverride",
	"autoclusteripaccess",
	"forceupdate",
	"updatemultiple",
	"configs:#.kind",
	"configs:#.config",
	"sharedvolumesize",
	"healthcheck",
	"privacypolicy",
	"vmflavor",
	"optres",
}
var RefreshAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
}
var RefreshAppInstOptionalArgs = []string{
	"cluster",
	"cloudlet-org",
	"cloudlet",
	"cluster-org",
	"crmoverride",
	"forceupdate",
	"updatemultiple",
}
var UpdateAppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var UpdateAppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"crmoverride",
	"configs:#.kind",
	"configs:#.config",
	"powerstate",
}
var AppInstKeyRequiredArgs = []string{}
var AppInstKeyOptionalArgs = []string{
	"appkey.organization",
	"appkey.name",
	"appkey.version",
	"clusterinstkey.clusterkey.name",
	"clusterinstkey.cloudletkey.organization",
	"clusterinstkey.cloudletkey.name",
	"clusterinstkey.organization",
}
var AppInstKeyAliasArgs = []string{
	"appkey.organization=appinstkey.appkey.organization",
	"appkey.name=appinstkey.appkey.name",
	"appkey.version=appinstkey.appkey.version",
	"clusterinstkey.clusterkey.name=appinstkey.clusterinstkey.clusterkey.name",
	"clusterinstkey.cloudletkey.organization=appinstkey.clusterinstkey.cloudletkey.organization",
	"clusterinstkey.cloudletkey.name=appinstkey.clusterinstkey.cloudletkey.name",
	"clusterinstkey.organization=appinstkey.clusterinstkey.organization",
}
var AppInstKeyComments = map[string]string{
	"appkey.organization":                     "App developer organization",
	"appkey.name":                             "App name",
	"appkey.version":                          "App version",
	"clusterinstkey.clusterkey.name":          "Cluster name",
	"clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
}
var AppInstKeySpecialArgs = map[string]string{}
var AppInstRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"cloudlet-org",
	"cloudlet",
}
var AppInstOptionalArgs = []string{
	"cluster",
	"cluster-org",
	"flavor",
	"crmoverride",
	"autoclusteripaccess",
	"forceupdate",
	"updatemultiple",
	"configs:#.kind",
	"configs:#.config",
	"sharedvolumesize",
	"healthcheck",
	"privacypolicy",
	"powerstate",
	"vmflavor",
	"optres",
}
var AppInstAliasArgs = []string{
	"fields=appinst.fields",
	"app-org=appinst.key.appkey.organization",
	"appname=appinst.key.appkey.name",
	"appvers=appinst.key.appkey.version",
	"cluster=appinst.key.clusterinstkey.clusterkey.name",
	"cloudlet-org=appinst.key.clusterinstkey.cloudletkey.organization",
	"cloudlet=appinst.key.clusterinstkey.cloudletkey.name",
	"cluster-org=appinst.key.clusterinstkey.organization",
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
	"mappedports:#.proto=appinst.mappedports:#.proto",
	"mappedports:#.internalport=appinst.mappedports:#.internalport",
	"mappedports:#.publicport=appinst.mappedports:#.publicport",
	"mappedports:#.fqdnprefix=appinst.mappedports:#.fqdnprefix",
	"mappedports:#.endport=appinst.mappedports:#.endport",
	"mappedports:#.tls=appinst.mappedports:#.tls",
	"mappedports:#.nginx=appinst.mappedports:#.nginx",
	"flavor=appinst.flavor.name",
	"state=appinst.state",
	"errors=appinst.errors",
	"crmoverride=appinst.crmoverride",
	"runtimeinfo.containerids=appinst.runtimeinfo.containerids",
	"createdat.seconds=appinst.createdat.seconds",
	"createdat.nanos=appinst.createdat.nanos",
	"autoclusteripaccess=appinst.autoclusteripaccess",
	"status.tasknumber=appinst.status.tasknumber",
	"status.maxtasks=appinst.status.maxtasks",
	"status.taskname=appinst.status.taskname",
	"status.stepname=appinst.status.stepname",
	"status.msgcount=appinst.status.msgcount",
	"status.msgs=appinst.status.msgs",
	"revision=appinst.revision",
	"forceupdate=appinst.forceupdate",
	"updatemultiple=appinst.updatemultiple",
	"configs:#.kind=appinst.configs:#.kind",
	"configs:#.config=appinst.configs:#.config",
	"sharedvolumesize=appinst.sharedvolumesize",
	"healthcheck=appinst.healthcheck",
	"privacypolicy=appinst.privacypolicy",
	"powerstate=appinst.powerstate",
	"externalvolumesize=appinst.externalvolumesize",
	"availabilityzone=appinst.availabilityzone",
	"vmflavor=appinst.vmflavor",
	"optres=appinst.optres",
}
var AppInstComments = map[string]string{
	"fields":                         "Fields are used for the Update API to specify which fields to apply",
	"app-org":                        "App developer organization",
	"appname":                        "App name",
	"appvers":                        "App version",
	"cluster":                        "Cluster name",
	"cloudlet-org":                   "Organization of the cloudlet site",
	"cloudlet":                       "Name of the cloudlet",
	"cluster-org":                    "Name of Developer organization that this cluster belongs to",
	"cloudletloc.latitude":           "latitude in WGS 84 coordinates",
	"cloudletloc.longitude":          "longitude in WGS 84 coordinates",
	"cloudletloc.horizontalaccuracy": "horizontal accuracy (radius in meters)",
	"cloudletloc.verticalaccuracy":   "vertical accuracy (meters)",
	"cloudletloc.altitude":           "On android only lat and long are guaranteed to be supplied altitude in meters",
	"cloudletloc.course":             "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"cloudletloc.speed":              "speed (IOS) / velocity (Android) (meters/sec)",
	"uri":                            "Base FQDN (not really URI) for the App. See Service FQDN for endpoint access.",
	"liveness":                       "Liveness of instance (see Liveness), one of LivenessUnknown, LivenessStatic, LivenessDynamic, LivenessAutoprov",
	"mappedports:#.proto":            "TCP (L4) or UDP (L4) protocol, one of LProtoUnknown, LProtoTcp, LProtoUdp",
	"mappedports:#.internalport":     "Container port",
	"mappedports:#.publicport":       "Public facing port for TCP/UDP (may be mapped on shared LB reverse proxy)",
	"mappedports:#.fqdnprefix":       "skip 4 to preserve the numbering. 4 was path_prefix but was removed since we dont need it after removed http FQDN prefix to append to base FQDN in FindCloudlet response. May be empty.",
	"mappedports:#.endport":          "A non-zero end port indicates a port range from internal port to end port, inclusive.",
	"mappedports:#.tls":              "TLS termination for this port",
	"mappedports:#.nginx":            "use nginx proxy for this port if you really need a transparent proxy (udp only)",
	"flavor":                         "Flavor name",
	"state":                          "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies, DeleteDone",
	"errors":                         "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"crmoverride":                    "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"runtimeinfo.containerids":       "List of container names",
	"autoclusteripaccess":            "IpAccess for auto-clusters. Ignored otherwise., one of IpAccessUnknown, IpAccessDedicated, IpAccessShared",
	"revision":                       "Revision changes each time the App is updated.  Refreshing the App Instance will sync the revision with that of the App",
	"forceupdate":                    "Force Appinst refresh even if revision number matches App revision number.",
	"updatemultiple":                 "Allow multiple instances to be updated at once",
	"configs:#.kind":                 "kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"configs:#.config":               "config file contents or URI reference",
	"sharedvolumesize":               "shared volume size when creating auto cluster",
	"healthcheck":                    "Health Check status, one of HealthCheckUnknown, HealthCheckFailRootlbOffline, HealthCheckFailServerFail, HealthCheckOk",
	"privacypolicy":                  "Optional privacy policy name",
	"powerstate":                     "Power State of the AppInst, one of PowerOn, PowerOff, Reboot",
	"externalvolumesize":             "Size of external volume to be attached to nodes.  This is for the root partition",
	"availabilityzone":               "Optional Availability Zone if any",
	"vmflavor":                       "OS node flavor to use",
	"optres":                         "Optional Resources required by OS flavor if any",
}
var AppInstSpecialArgs = map[string]string{
	"appinst.errors":                   "StringArray",
	"appinst.fields":                   "StringArray",
	"appinst.runtimeinfo.containerids": "StringArray",
	"appinst.status.msgs":              "StringArray",
}
var AppInstRuntimeRequiredArgs = []string{}
var AppInstRuntimeOptionalArgs = []string{
	"containerids",
}
var AppInstRuntimeAliasArgs = []string{
	"containerids=appinstruntime.containerids",
}
var AppInstRuntimeComments = map[string]string{
	"containerids": "List of container names",
}
var AppInstRuntimeSpecialArgs = map[string]string{
	"appinstruntime.containerids": "StringArray",
}
var AppInstInfoRequiredArgs = []string{
	"key.appkey.organization",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization",
}
var AppInstInfoOptionalArgs = []string{
	"notifyid",
	"state",
	"errors",
	"runtimeinfo.containerids",
	"status.tasknumber",
	"status.maxtasks",
	"status.taskname",
	"status.stepname",
	"status.msgcount",
	"status.msgs",
	"powerstate",
}
var AppInstInfoAliasArgs = []string{
	"fields=appinstinfo.fields",
	"key.appkey.organization=appinstinfo.key.appkey.organization",
	"key.appkey.name=appinstinfo.key.appkey.name",
	"key.appkey.version=appinstinfo.key.appkey.version",
	"key.clusterinstkey.clusterkey.name=appinstinfo.key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization=appinstinfo.key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name=appinstinfo.key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization=appinstinfo.key.clusterinstkey.organization",
	"notifyid=appinstinfo.notifyid",
	"state=appinstinfo.state",
	"errors=appinstinfo.errors",
	"runtimeinfo.containerids=appinstinfo.runtimeinfo.containerids",
	"status.tasknumber=appinstinfo.status.tasknumber",
	"status.maxtasks=appinstinfo.status.maxtasks",
	"status.taskname=appinstinfo.status.taskname",
	"status.stepname=appinstinfo.status.stepname",
	"status.msgcount=appinstinfo.status.msgcount",
	"status.msgs=appinstinfo.status.msgs",
	"powerstate=appinstinfo.powerstate",
}
var AppInstInfoComments = map[string]string{
	"fields":                                      "Fields are used for the Update API to specify which fields to apply",
	"key.appkey.organization":                     "App developer organization",
	"key.appkey.name":                             "App name",
	"key.appkey.version":                          "App version",
	"key.clusterinstkey.clusterkey.name":          "Cluster name",
	"key.clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"key.clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
	"notifyid":                                    "Id of client assigned by server (internal use only)",
	"state":                                       "Current state of the AppInst on the Cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies, DeleteDone",
	"errors":                                      "Any errors trying to create, update, or delete the AppInst on the Cloudlet",
	"runtimeinfo.containerids":                    "List of container names",
	"powerstate":                                  "Power State of the AppInst, one of PowerOn, PowerOff, Reboot",
}
var AppInstInfoSpecialArgs = map[string]string{
	"appinstinfo.errors":                   "StringArray",
	"appinstinfo.fields":                   "StringArray",
	"appinstinfo.runtimeinfo.containerids": "StringArray",
	"appinstinfo.status.msgs":              "StringArray",
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
var AppInstLookupRequiredArgs = []string{
	"key.appkey.organization",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization",
}
var AppInstLookupOptionalArgs = []string{
	"policykey.organization",
	"policykey.name",
}
var AppInstLookupAliasArgs = []string{
	"key.appkey.organization=appinstlookup.key.appkey.organization",
	"key.appkey.name=appinstlookup.key.appkey.name",
	"key.appkey.version=appinstlookup.key.appkey.version",
	"key.clusterinstkey.clusterkey.name=appinstlookup.key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization=appinstlookup.key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name=appinstlookup.key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization=appinstlookup.key.clusterinstkey.organization",
	"policykey.organization=appinstlookup.policykey.organization",
	"policykey.name=appinstlookup.policykey.name",
}
var AppInstLookupComments = map[string]string{
	"key.appkey.organization":                     "App developer organization",
	"key.appkey.name":                             "App name",
	"key.appkey.version":                          "App version",
	"key.clusterinstkey.clusterkey.name":          "Cluster name",
	"key.clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"key.clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
	"policykey.organization":                      "Name of the organization for the cluster that this policy will apply to",
	"policykey.name":                              "Policy name",
}
var AppInstLookupSpecialArgs = map[string]string{}
var AppInstLookup2RequiredArgs = []string{
	"key.appkey.organization",
	"key.appkey.name",
	"key.appkey.version",
	"key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization",
}
var AppInstLookup2OptionalArgs = []string{
	"cloudletkey.organization",
	"cloudletkey.name",
}
var AppInstLookup2AliasArgs = []string{
	"key.appkey.organization=appinstlookup2.key.appkey.organization",
	"key.appkey.name=appinstlookup2.key.appkey.name",
	"key.appkey.version=appinstlookup2.key.appkey.version",
	"key.clusterinstkey.clusterkey.name=appinstlookup2.key.clusterinstkey.clusterkey.name",
	"key.clusterinstkey.cloudletkey.organization=appinstlookup2.key.clusterinstkey.cloudletkey.organization",
	"key.clusterinstkey.cloudletkey.name=appinstlookup2.key.clusterinstkey.cloudletkey.name",
	"key.clusterinstkey.organization=appinstlookup2.key.clusterinstkey.organization",
	"cloudletkey.organization=appinstlookup2.cloudletkey.organization",
	"cloudletkey.name=appinstlookup2.cloudletkey.name",
}
var AppInstLookup2Comments = map[string]string{
	"key.appkey.organization":                     "App developer organization",
	"key.appkey.name":                             "App name",
	"key.appkey.version":                          "App version",
	"key.clusterinstkey.clusterkey.name":          "Cluster name",
	"key.clusterinstkey.cloudletkey.organization": "Organization of the cloudlet site",
	"key.clusterinstkey.cloudletkey.name":         "Name of the cloudlet",
	"key.clusterinstkey.organization":             "Name of Developer organization that this cluster belongs to",
	"cloudletkey.organization":                    "Organization of the cloudlet site",
	"cloudletkey.name":                            "Name of the cloudlet",
}
var AppInstLookup2SpecialArgs = map[string]string{}
