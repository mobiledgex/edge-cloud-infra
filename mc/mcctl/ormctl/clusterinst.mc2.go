// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

package ormctl

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var CreateClusterInstCmd = &Command{
	Use:                  "CreateClusterInst",
	RequiredArgs:         strings.Join(append([]string{"region"}, ClusterInstRequiredArgs...), " "),
	OptionalArgs:         strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:            strings.Join(ClusterInstAliasArgs, " "),
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Path:                 "/auth/ctrl/CreateClusterInst",
	StreamOut:            true,
	StreamOutIncremental: true,
}

var DeleteClusterInstCmd = &Command{
	Use:                  "DeleteClusterInst",
	RequiredArgs:         strings.Join(append([]string{"region"}, ClusterInstRequiredArgs...), " "),
	OptionalArgs:         strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:            strings.Join(ClusterInstAliasArgs, " "),
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Path:                 "/auth/ctrl/DeleteClusterInst",
	StreamOut:            true,
	StreamOutIncremental: true,
}

var UpdateClusterInstCmd = &Command{
	Use:                  "UpdateClusterInst",
	RequiredArgs:         strings.Join(append([]string{"region"}, ClusterInstRequiredArgs...), " "),
	OptionalArgs:         strings.Join(ClusterInstOptionalArgs, " "),
	AliasArgs:            strings.Join(ClusterInstAliasArgs, " "),
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Path:                 "/auth/ctrl/UpdateClusterInst",
	StreamOut:            true,
	StreamOutIncremental: true,
}

var ShowClusterInstCmd = &Command{
	Use:          "ShowClusterInst",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(ClusterInstRequiredArgs, ClusterInstOptionalArgs...), " "),
	AliasArgs:    strings.Join(ClusterInstAliasArgs, " "),
	ReqData:      &ormapi.RegionClusterInst{},
	ReplyData:    &edgeproto.ClusterInst{},
	Path:         "/auth/ctrl/ShowClusterInst",
	StreamOut:    true,
}
var ClusterInstApiCmds = []*Command{
	CreateClusterInstCmd,
	DeleteClusterInstCmd,
	UpdateClusterInstCmd,
	ShowClusterInstCmd,
}

var ClusterInstKeyRequiredArgs = []string{}
var ClusterInstKeyOptionalArgs = []string{
	"clusterkey.name",
	"cloudletkey.operatorkey.name",
	"cloudletkey.name",
	"developer",
}
var ClusterInstKeyAliasArgs = []string{
	"clusterkey.name=clusterinstkey.clusterkey.name",
	"cloudletkey.operatorkey.name=clusterinstkey.cloudletkey.operatorkey.name",
	"cloudletkey.name=clusterinstkey.cloudletkey.name",
	"developer=clusterinstkey.developer",
}
var ClusterInstRequiredArgs = []string{
	"cluster",
	"operator",
	"cloudlet",
	"key.developer",
}
var ClusterInstOptionalArgs = []string{
	"flavor.name",
	"state",
	"errors",
	"crmoverride",
	"ipaccess",
	"deployment",
	"nummasters",
	"numnodes",
}
var ClusterInstAliasArgs = []string{
	"cluster=clusterinst.key.clusterkey.name",
	"operator=clusterinst.key.cloudletkey.operatorkey.name",
	"cloudlet=clusterinst.key.cloudletkey.name",
	"key.developer=clusterinst.key.developer",
	"flavor.name=clusterinst.flavor.name",
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
	"status.taskname=clusterinst.status.taskname",
	"status.stepname=clusterinst.status.stepname",
}
var ClusterInstInfoRequiredArgs = []string{
	"key.clusterkey.name",
	"key.cloudletkey.operatorkey.name",
	"key.cloudletkey.name",
	"key.developer",
}
var ClusterInstInfoOptionalArgs = []string{
	"notifyid",
	"state",
	"status.tasknumber",
	"status.taskname",
	"status.stepname",
	"errors",
}
var ClusterInstInfoAliasArgs = []string{
	"key.clusterkey.name=clusterinstinfo.key.clusterkey.name",
	"key.cloudletkey.operatorkey.name=clusterinstinfo.key.cloudletkey.operatorkey.name",
	"key.cloudletkey.name=clusterinstinfo.key.cloudletkey.name",
	"key.developer=clusterinstinfo.key.developer",
	"notifyid=clusterinstinfo.notifyid",
	"state=clusterinstinfo.state",
	"status.tasknumber=clusterinstinfo.status.tasknumber",
	"status.taskname=clusterinstinfo.status.taskname",
	"status.stepname=clusterinstinfo.status.stepname",
	"errors=clusterinstinfo.errors",
}
