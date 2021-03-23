// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

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

var CreateClusterInstCmd = &cli.Command{
	Use:                  "create",
	Short:                "Create Cluster Instance. Creates an instance of a Cluster on a Cloudlet, defined by a Cluster Key and a Cloudlet Key. ClusterInst is a collection of compute resources on a Cloudlet on which AppInsts are deployed.",
	RequiredArgs:         "region " + strings.Join(ClusterInstRequiredArgs, " "),
	OptionalArgs:         strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:            strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:          &ClusterInstSpecialArgs,
	Comments:             addRegionComment(ClusterInstComments),
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/CreateClusterInst"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var DeleteClusterInstCmd = &cli.Command{
	Use:                  "delete",
	Short:                "Delete Cluster Instance. Deletes an instance of a Cluster deployed on a Cloudlet.",
	RequiredArgs:         "region " + strings.Join(ClusterInstRequiredArgs, " "),
	OptionalArgs:         strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:            strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:          &ClusterInstSpecialArgs,
	Comments:             addRegionComment(ClusterInstComments),
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/DeleteClusterInst"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var UpdateClusterInstCmd = &cli.Command{
	Use:          "update",
	Short:        "Update Cluster Instance. Updates an instance of a Cluster deployed on a Cloudlet.",
	RequiredArgs: "region " + strings.Join(UpdateClusterInstRequiredArgs, " "),
	OptionalArgs: strings.Join(UpdateClusterInstOptionalArgs, " "),
	AliasArgs:    strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:  &ClusterInstSpecialArgs,
	Comments:     addRegionComment(ClusterInstComments),
	ReqData:      &ormapi.RegionClusterInst{},
	ReplyData:    &edgeproto.Result{},
	Run: runRest("/auth/ctrl/UpdateClusterInst",
		withSetFieldsFunc(setUpdateClusterInstFields),
	),
	StreamOut:            true,
	StreamOutIncremental: true,
}

func setUpdateClusterInstFields(in map[string]interface{}) {
	// get map for edgeproto object in region struct
	obj := in[strings.ToLower("ClusterInst")]
	if obj == nil {
		return
	}
	objmap, ok := obj.(map[string]interface{})
	if !ok {
		return
	}
	fields := cli.GetSpecifiedFields(objmap, &edgeproto.ClusterInst{}, cli.JsonNamespace)
	// include fields already specified
	if inFields, found := objmap["fields"]; found {
		if fieldsArr, ok := inFields.([]string); ok {
			fields = append(fields, fieldsArr...)
		}
	}
	objmap["fields"] = fields
}

var ShowClusterInstCmd = &cli.Command{
	Use:          "show",
	Short:        "Show Cluster Instances. Lists all the cluster instances managed by Edge Controller.",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(ClusterInstRequiredArgs, ClusterInstOptionalArgs...), " "),
	AliasArgs:    strings.Join(ClusterInstAliasArgs, " "),
	SpecialArgs:  &ClusterInstSpecialArgs,
	Comments:     addRegionComment(ClusterInstComments),
	ReqData:      &ormapi.RegionClusterInst{},
	ReplyData:    &edgeproto.ClusterInst{},
	Run:          runRest("/auth/ctrl/ShowClusterInst"),
	StreamOut:    true,
}

var DeleteIdleReservableClusterInstsCmd = &cli.Command{
	Use:          "delete",
	Short:        "Cleanup Reservable Cluster Instances. Deletes reservable cluster instances that are not in use.",
	RequiredArgs: "region " + strings.Join(IdleReservableClusterInstsRequiredArgs, " "),
	OptionalArgs: strings.Join(IdleReservableClusterInstsOptionalArgs, " "),
	AliasArgs:    strings.Join(IdleReservableClusterInstsAliasArgs, " "),
	SpecialArgs:  &IdleReservableClusterInstsSpecialArgs,
	Comments:     addRegionComment(IdleReservableClusterInstsComments),
	ReqData:      &ormapi.RegionIdleReservableClusterInsts{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/DeleteIdleReservableClusterInsts"),
}

var ClusterInstApiCmds = []*cli.Command{
	CreateClusterInstCmd,
	DeleteClusterInstCmd,
	UpdateClusterInstCmd,
	ShowClusterInstCmd,
	DeleteIdleReservableClusterInstsCmd,
}

var ClusterInstApiCmdsGroup = cli.GenGroup("clusterinst", "Manage ClusterInsts", ClusterInstApiCmds)

var UpdateClusterInstRequiredArgs = []string{
	"cluster",
	"cloudlet-org",
	"cloudlet",
	"cluster-org",
}
var UpdateClusterInstOptionalArgs = []string{
	"crmoverride",
	"numnodes",
	"autoscalepolicy",
	"skipcrmcleanuponfailure",
	"reservationendedat.seconds",
	"reservationendedat.nanos",
}
var ClusterInstKeyRequiredArgs = []string{}
var ClusterInstKeyOptionalArgs = []string{
	"clusterkey.name",
	"cloudletkey.organization",
	"cloudletkey.name",
	"organization",
}
var ClusterInstKeyAliasArgs = []string{
	"clusterkey.name=clusterinstkey.clusterkey.name",
	"cloudletkey.organization=clusterinstkey.cloudletkey.organization",
	"cloudletkey.name=clusterinstkey.cloudletkey.name",
	"organization=clusterinstkey.organization",
}
var ClusterInstKeyComments = map[string]string{
	"clusterkey.name":          "Cluster name",
	"cloudletkey.organization": "Organization of the cloudlet site",
	"cloudletkey.name":         "Name of the cloudlet",
	"organization":             "Name of Developer organization that this cluster belongs to",
}
var ClusterInstKeySpecialArgs = map[string]string{}
var ClusterInstRequiredArgs = []string{
	"cluster",
	"cloudlet-org",
	"cloudlet",
	"cluster-org",
}
var ClusterInstOptionalArgs = []string{
	"flavor",
	"crmoverride",
	"ipaccess",
	"deployment",
	"nummasters",
	"numnodes",
	"autoscalepolicy",
	"imagename",
	"reservable",
	"sharedvolumesize",
	"skipcrmcleanuponfailure",
	"reservationendedat.seconds",
	"reservationendedat.nanos",
}
var ClusterInstAliasArgs = []string{
	"fields=clusterinst.fields",
	"cluster=clusterinst.key.clusterkey.name",
	"cloudlet-org=clusterinst.key.cloudletkey.organization",
	"cloudlet=clusterinst.key.cloudletkey.name",
	"cluster-org=clusterinst.key.organization",
	"flavor=clusterinst.flavor.name",
	"liveness=clusterinst.liveness",
	"auto=clusterinst.auto",
	"state=clusterinst.state",
	"errors=clusterinst.errors",
	"crmoverride=clusterinst.crmoverride",
	"ipaccess=clusterinst.ipaccess",
	"allocatedip=clusterinst.allocatedip",
	"nodeflavor=clusterinst.nodeflavor",
	"deployment=clusterinst.deployment",
	"nummasters=clusterinst.nummasters",
	"numnodes=clusterinst.numnodes",
	"status.tasknumber=clusterinst.status.tasknumber",
	"status.maxtasks=clusterinst.status.maxtasks",
	"status.taskname=clusterinst.status.taskname",
	"status.stepname=clusterinst.status.stepname",
	"status.msgcount=clusterinst.status.msgcount",
	"status.msgs=clusterinst.status.msgs",
	"externalvolumesize=clusterinst.externalvolumesize",
	"autoscalepolicy=clusterinst.autoscalepolicy",
	"availabilityzone=clusterinst.availabilityzone",
	"imagename=clusterinst.imagename",
	"reservable=clusterinst.reservable",
	"reservedby=clusterinst.reservedby",
	"sharedvolumesize=clusterinst.sharedvolumesize",
	"masternodeflavor=clusterinst.masternodeflavor",
	"skipcrmcleanuponfailure=clusterinst.skipcrmcleanuponfailure",
	"optres=clusterinst.optres",
	"resources.vms:#.name=clusterinst.resources.vms:#.name",
	"resources.vms:#.type=clusterinst.resources.vms:#.type",
	"resources.vms:#.status=clusterinst.resources.vms:#.status",
	"resources.vms:#.infraflavor=clusterinst.resources.vms:#.infraflavor",
	"resources.vms:#.ipaddresses:#.externalip=clusterinst.resources.vms:#.ipaddresses:#.externalip",
	"resources.vms:#.ipaddresses:#.internalip=clusterinst.resources.vms:#.ipaddresses:#.internalip",
	"resources.vms:#.containers:#.name=clusterinst.resources.vms:#.containers:#.name",
	"resources.vms:#.containers:#.type=clusterinst.resources.vms:#.containers:#.type",
	"resources.vms:#.containers:#.status=clusterinst.resources.vms:#.containers:#.status",
	"resources.vms:#.containers:#.clusterip=clusterinst.resources.vms:#.containers:#.clusterip",
	"resources.vms:#.containers:#.restarts=clusterinst.resources.vms:#.containers:#.restarts",
	"createdat.seconds=clusterinst.createdat.seconds",
	"createdat.nanos=clusterinst.createdat.nanos",
	"updatedat.seconds=clusterinst.updatedat.seconds",
	"updatedat.nanos=clusterinst.updatedat.nanos",
	"reservationendedat.seconds=clusterinst.reservationendedat.seconds",
	"reservationendedat.nanos=clusterinst.reservationendedat.nanos",
}
var ClusterInstComments = map[string]string{
	"fields":                                 "Fields are used for the Update API to specify which fields to apply",
	"cluster":                                "Cluster name",
	"cloudlet-org":                           "Organization of the cloudlet site",
	"cloudlet":                               "Name of the cloudlet",
	"cluster-org":                            "Name of Developer organization that this cluster belongs to",
	"flavor":                                 "Flavor name",
	"liveness":                               "Liveness of instance (see Liveness), one of LivenessUnknown, LivenessStatic, LivenessDynamic, LivenessAutoprov",
	"auto":                                   "Auto is set to true when automatically created by back-end (internal use only)",
	"state":                                  "State of the cluster instance, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies, DeleteDone",
	"errors":                                 "Any errors trying to create, update, or delete the ClusterInst on the Cloudlet.",
	"crmoverride":                            "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"ipaccess":                               "IP access type (RootLB Type), one of IpAccessUnknown, IpAccessDedicated, IpAccessShared",
	"allocatedip":                            "Allocated IP for dedicated access",
	"nodeflavor":                             "Cloudlet specific node flavor",
	"deployment":                             "Deployment type (kubernetes or docker)",
	"nummasters":                             "Number of k8s masters (In case of docker deployment, this field is not required)",
	"numnodes":                               "Number of k8s nodes (In case of docker deployment, this field is not required)",
	"externalvolumesize":                     "Size of external volume to be attached to nodes.  This is for the root partition",
	"autoscalepolicy":                        "Auto scale policy name",
	"availabilityzone":                       "Optional Resource AZ if any",
	"imagename":                              "Optional resource specific image to launch",
	"reservable":                             "If ClusterInst is reservable",
	"reservedby":                             "For reservable MobiledgeX ClusterInsts, the current developer tenant",
	"sharedvolumesize":                       "Size of an optional shared volume to be mounted on the master",
	"masternodeflavor":                       "Generic flavor for k8s master VM when worker nodes > 0",
	"skipcrmcleanuponfailure":                "Prevents cleanup of resources on failure within CRM, used for diagnostic purposes",
	"optres":                                 "Optional Resources required by OS flavor if any",
	"resources.vms:#.name":                   "Virtual machine name",
	"resources.vms:#.type":                   "Type can be platform, rootlb, cluster-master, cluster-node, vmapp",
	"resources.vms:#.status":                 "Runtime status of the VM",
	"resources.vms:#.infraflavor":            "Flavor allocated within the cloudlet infrastructure, distinct from the control plane flavor",
	"resources.vms:#.containers:#.name":      "Name of the container",
	"resources.vms:#.containers:#.type":      "Type can be docker or kubernetes",
	"resources.vms:#.containers:#.status":    "Runtime status of the container",
	"resources.vms:#.containers:#.clusterip": "IP within the CNI and is applicable to kubernetes only",
	"resources.vms:#.containers:#.restarts":  "Restart count, applicable to kubernetes only",
}
var ClusterInstSpecialArgs = map[string]string{
	"clusterinst.errors":      "StringArray",
	"clusterinst.fields":      "StringArray",
	"clusterinst.status.msgs": "StringArray",
}
var IdleReservableClusterInstsRequiredArgs = []string{}
var IdleReservableClusterInstsOptionalArgs = []string{
	"idletime",
}
var IdleReservableClusterInstsAliasArgs = []string{
	"idletime=idlereservableclusterinsts.idletime",
}
var IdleReservableClusterInstsComments = map[string]string{
	"idletime": "Idle time (duration)",
}
var IdleReservableClusterInstsSpecialArgs = map[string]string{}
var ClusterInstInfoRequiredArgs = []string{
	"key.clusterkey.name",
	"key.cloudletkey.organization",
	"key.cloudletkey.name",
	"key.organization",
}
var ClusterInstInfoOptionalArgs = []string{
	"notifyid",
	"state",
	"errors",
	"status.tasknumber",
	"status.maxtasks",
	"status.taskname",
	"status.stepname",
	"status.msgcount",
	"status.msgs",
	"resources.vms:#.name",
	"resources.vms:#.type",
	"resources.vms:#.status",
	"resources.vms:#.infraflavor",
	"resources.vms:#.ipaddresses:#.externalip",
	"resources.vms:#.ipaddresses:#.internalip",
	"resources.vms:#.containers:#.name",
	"resources.vms:#.containers:#.type",
	"resources.vms:#.containers:#.status",
	"resources.vms:#.containers:#.clusterip",
	"resources.vms:#.containers:#.restarts",
}
var ClusterInstInfoAliasArgs = []string{
	"fields=clusterinstinfo.fields",
	"key.clusterkey.name=clusterinstinfo.key.clusterkey.name",
	"key.cloudletkey.organization=clusterinstinfo.key.cloudletkey.organization",
	"key.cloudletkey.name=clusterinstinfo.key.cloudletkey.name",
	"key.organization=clusterinstinfo.key.organization",
	"notifyid=clusterinstinfo.notifyid",
	"state=clusterinstinfo.state",
	"errors=clusterinstinfo.errors",
	"status.tasknumber=clusterinstinfo.status.tasknumber",
	"status.maxtasks=clusterinstinfo.status.maxtasks",
	"status.taskname=clusterinstinfo.status.taskname",
	"status.stepname=clusterinstinfo.status.stepname",
	"status.msgcount=clusterinstinfo.status.msgcount",
	"status.msgs=clusterinstinfo.status.msgs",
	"resources.vms:#.name=clusterinstinfo.resources.vms:#.name",
	"resources.vms:#.type=clusterinstinfo.resources.vms:#.type",
	"resources.vms:#.status=clusterinstinfo.resources.vms:#.status",
	"resources.vms:#.infraflavor=clusterinstinfo.resources.vms:#.infraflavor",
	"resources.vms:#.ipaddresses:#.externalip=clusterinstinfo.resources.vms:#.ipaddresses:#.externalip",
	"resources.vms:#.ipaddresses:#.internalip=clusterinstinfo.resources.vms:#.ipaddresses:#.internalip",
	"resources.vms:#.containers:#.name=clusterinstinfo.resources.vms:#.containers:#.name",
	"resources.vms:#.containers:#.type=clusterinstinfo.resources.vms:#.containers:#.type",
	"resources.vms:#.containers:#.status=clusterinstinfo.resources.vms:#.containers:#.status",
	"resources.vms:#.containers:#.clusterip=clusterinstinfo.resources.vms:#.containers:#.clusterip",
	"resources.vms:#.containers:#.restarts=clusterinstinfo.resources.vms:#.containers:#.restarts",
}
var ClusterInstInfoComments = map[string]string{
	"fields":                                 "Fields are used for the Update API to specify which fields to apply",
	"key.clusterkey.name":                    "Cluster name",
	"key.cloudletkey.organization":           "Organization of the cloudlet site",
	"key.cloudletkey.name":                   "Name of the cloudlet",
	"key.organization":                       "Name of Developer organization that this cluster belongs to",
	"notifyid":                               "Id of client assigned by server (internal use only)",
	"state":                                  "State of the cluster instance, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare, CrmInitok, CreatingDependencies, DeleteDone",
	"errors":                                 "Any errors trying to create, update, or delete the ClusterInst on the Cloudlet.",
	"resources.vms:#.name":                   "Virtual machine name",
	"resources.vms:#.type":                   "Type can be platform, rootlb, cluster-master, cluster-node, vmapp",
	"resources.vms:#.status":                 "Runtime status of the VM",
	"resources.vms:#.infraflavor":            "Flavor allocated within the cloudlet infrastructure, distinct from the control plane flavor",
	"resources.vms:#.containers:#.name":      "Name of the container",
	"resources.vms:#.containers:#.type":      "Type can be docker or kubernetes",
	"resources.vms:#.containers:#.status":    "Runtime status of the container",
	"resources.vms:#.containers:#.clusterip": "IP within the CNI and is applicable to kubernetes only",
	"resources.vms:#.containers:#.restarts":  "Restart count, applicable to kubernetes only",
}
var ClusterInstInfoSpecialArgs = map[string]string{
	"clusterinstinfo.errors":      "StringArray",
	"clusterinstinfo.fields":      "StringArray",
	"clusterinstinfo.status.msgs": "StringArray",
}
