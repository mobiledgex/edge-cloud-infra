// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var CreateCloudletCmd = &Command{
	Use:          "CreateCloudlet",
	RequiredArgs: strings.Join(append([]string{"region"}, CloudletRequiredArgs...), " "),
	OptionalArgs: strings.Join(CloudletOptionalArgs, " "),
	AliasArgs:    strings.Join(CloudletAliasArgs, " "),
	ReqData:      &ormapi.RegionCloudlet{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/CreateCloudlet",
	StreamOut:    true,
}

var DeleteCloudletCmd = &Command{
	Use:                  "DeleteCloudlet",
	RequiredArgs:         strings.Join(append([]string{"region"}, CloudletRequiredArgs...), " "),
	OptionalArgs:         strings.Join(CloudletOptionalArgs, " "),
	AliasArgs:            strings.Join(CloudletAliasArgs, " "),
	ReqData:              &ormapi.RegionCloudlet{},
	ReplyData:            &edgeproto.Result{},
	Path:                 "/auth/ctrl/DeleteCloudlet",
	StreamOut:            true,
	StreamOutIncremental: true,
}

var UpdateCloudletCmd = &Command{
	Use:          "UpdateCloudlet",
	RequiredArgs: strings.Join(append([]string{"region"}, CloudletRequiredArgs...), " "),
	OptionalArgs: strings.Join(CloudletOptionalArgs, " "),
	AliasArgs:    strings.Join(CloudletAliasArgs, " "),
	ReqData:      &ormapi.RegionCloudlet{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/UpdateCloudlet",
	StreamOut:    true,
}

var ShowCloudletCmd = &Command{
	Use:          "ShowCloudlet",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(CloudletRequiredArgs, CloudletOptionalArgs...), " "),
	AliasArgs:    strings.Join(CloudletAliasArgs, " "),
	ReqData:      &ormapi.RegionCloudlet{},
	ReplyData:    &edgeproto.Cloudlet{},
	Path:         "/auth/ctrl/ShowCloudlet",
	StreamOut:    true,
}
var CloudletApiCmds = []*Command{
	CreateCloudletCmd,
	DeleteCloudletCmd,
	UpdateCloudletCmd,
	ShowCloudletCmd,
}

var ShowCloudletInfoCmd = &Command{
	Use:          "ShowCloudletInfo",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(CloudletInfoRequiredArgs, CloudletInfoOptionalArgs...), " "),
	AliasArgs:    strings.Join(CloudletInfoAliasArgs, " "),
	ReqData:      &ormapi.RegionCloudletInfo{},
	ReplyData:    &edgeproto.CloudletInfo{},
	Path:         "/auth/ctrl/ShowCloudletInfo",
	StreamOut:    true,
}
var CloudletInfoApiCmds = []*Command{
	ShowCloudletInfoCmd,
}

var CloudletKeyRequiredArgs = []string{}
var CloudletKeyOptionalArgs = []string{
	"operatorkey.name",
	"name",
}
var CloudletKeyAliasArgs = []string{
	"operatorkey.name=cloudletkey.operatorkey.name",
	"name=cloudletkey.name",
}
var OperationTimeLimitsRequiredArgs = []string{}
var OperationTimeLimitsOptionalArgs = []string{
	"createclusterinsttimeout",
	"updateclusterinsttimeout",
	"deleteclusterinsttimeout",
	"createappinsttimeout",
	"updateappinsttimeout",
	"deleteappinsttimeout",
}
var OperationTimeLimitsAliasArgs = []string{
	"createclusterinsttimeout=operationtimelimits.createclusterinsttimeout",
	"updateclusterinsttimeout=operationtimelimits.updateclusterinsttimeout",
	"deleteclusterinsttimeout=operationtimelimits.deleteclusterinsttimeout",
	"createappinsttimeout=operationtimelimits.createappinsttimeout",
	"updateappinsttimeout=operationtimelimits.updateappinsttimeout",
	"deleteappinsttimeout=operationtimelimits.deleteappinsttimeout",
}
var CloudletInfraCommonRequiredArgs = []string{}
var CloudletInfraCommonOptionalArgs = []string{
	"dockerregistry",
	"dnszone",
	"registryfileserver",
	"cfkey",
	"cfuser",
	"dockerregpass",
	"networkscheme",
	"dockerregistrysecret",
}
var CloudletInfraCommonAliasArgs = []string{
	"dockerregistry=cloudletinfracommon.dockerregistry",
	"dnszone=cloudletinfracommon.dnszone",
	"registryfileserver=cloudletinfracommon.registryfileserver",
	"cfkey=cloudletinfracommon.cfkey",
	"cfuser=cloudletinfracommon.cfuser",
	"dockerregpass=cloudletinfracommon.dockerregpass",
	"networkscheme=cloudletinfracommon.networkscheme",
	"dockerregistrysecret=cloudletinfracommon.dockerregistrysecret",
}
var AzurePropertiesRequiredArgs = []string{}
var AzurePropertiesOptionalArgs = []string{
	"location",
	"resourcegroup",
	"username",
	"password",
}
var AzurePropertiesAliasArgs = []string{
	"location=azureproperties.location",
	"resourcegroup=azureproperties.resourcegroup",
	"username=azureproperties.username",
	"password=azureproperties.password",
}
var GcpPropertiesRequiredArgs = []string{}
var GcpPropertiesOptionalArgs = []string{
	"project",
	"zone",
	"serviceaccount",
	"gcpauthkeyurl",
}
var GcpPropertiesAliasArgs = []string{
	"project=gcpproperties.project",
	"zone=gcpproperties.zone",
	"serviceaccount=gcpproperties.serviceaccount",
	"gcpauthkeyurl=gcpproperties.gcpauthkeyurl",
}
var OpenStackPropertiesRequiredArgs = []string{}
var OpenStackPropertiesOptionalArgs = []string{
	"osexternalnetworkname",
	"osimagename",
	"osexternalroutername",
	"osmexnetwork",
	"openrcvars.key",
	"openrcvars.value",
}
var OpenStackPropertiesAliasArgs = []string{
	"osexternalnetworkname=openstackproperties.osexternalnetworkname",
	"osimagename=openstackproperties.osimagename",
	"osexternalroutername=openstackproperties.osexternalroutername",
	"osmexnetwork=openstackproperties.osmexnetwork",
	"openrcvars.key=openstackproperties.openrcvars.key",
	"openrcvars.value=openstackproperties.openrcvars.value",
}
var OpenRcVarsEntryRequiredArgs = []string{}
var OpenRcVarsEntryOptionalArgs = []string{
	"key",
	"value",
}
var OpenRcVarsEntryAliasArgs = []string{
	"key=openrcvarsentry.key",
	"value=openrcvarsentry.value",
}
var CloudletInfraPropertiesRequiredArgs = []string{}
var CloudletInfraPropertiesOptionalArgs = []string{
	"cloudletkind",
	"mexoscontainerimagename",
	"openstackproperties.osexternalnetworkname",
	"openstackproperties.osimagename",
	"openstackproperties.osexternalroutername",
	"openstackproperties.osmexnetwork",
	"openstackproperties.openrcvars.key",
	"openstackproperties.openrcvars.value",
	"azureproperties.location",
	"azureproperties.resourcegroup",
	"azureproperties.username",
	"azureproperties.password",
	"gcpproperties.project",
	"gcpproperties.zone",
	"gcpproperties.serviceaccount",
	"gcpproperties.gcpauthkeyurl",
}
var CloudletInfraPropertiesAliasArgs = []string{
	"cloudletkind=cloudletinfraproperties.cloudletkind",
	"mexoscontainerimagename=cloudletinfraproperties.mexoscontainerimagename",
	"openstackproperties.osexternalnetworkname=cloudletinfraproperties.openstackproperties.osexternalnetworkname",
	"openstackproperties.osimagename=cloudletinfraproperties.openstackproperties.osimagename",
	"openstackproperties.osexternalroutername=cloudletinfraproperties.openstackproperties.osexternalroutername",
	"openstackproperties.osmexnetwork=cloudletinfraproperties.openstackproperties.osmexnetwork",
	"openstackproperties.openrcvars.key=cloudletinfraproperties.openstackproperties.openrcvars.key",
	"openstackproperties.openrcvars.value=cloudletinfraproperties.openstackproperties.openrcvars.value",
	"azureproperties.location=cloudletinfraproperties.azureproperties.location",
	"azureproperties.resourcegroup=cloudletinfraproperties.azureproperties.resourcegroup",
	"azureproperties.username=cloudletinfraproperties.azureproperties.username",
	"azureproperties.password=cloudletinfraproperties.azureproperties.password",
	"gcpproperties.project=cloudletinfraproperties.gcpproperties.project",
	"gcpproperties.zone=cloudletinfraproperties.gcpproperties.zone",
	"gcpproperties.serviceaccount=cloudletinfraproperties.gcpproperties.serviceaccount",
	"gcpproperties.gcpauthkeyurl=cloudletinfraproperties.gcpproperties.gcpauthkeyurl",
}
var CloudletRequiredArgs = []string{
	"operator",
	"name",
}
var CloudletOptionalArgs = []string{
	"accesscredentials",
	"location.latitude",
	"location.longitude",
	"location.altitude",
	"location.timestamp.seconds",
	"location.timestamp.nanos",
	"ipsupport",
	"staticips",
	"numdynamicips",
}
var CloudletAliasArgs = []string{
	"operator=cloudlet.key.operatorkey.name",
	"name=cloudlet.key.name",
	"accesscredentials=cloudlet.accesscredentials",
	"location.latitude=cloudlet.location.latitude",
	"location.longitude=cloudlet.location.longitude",
	"location.horizontalaccuracy=cloudlet.location.horizontalaccuracy",
	"location.verticalaccuracy=cloudlet.location.verticalaccuracy",
	"location.altitude=cloudlet.location.altitude",
	"location.course=cloudlet.location.course",
	"location.speed=cloudlet.location.speed",
	"location.timestamp.seconds=cloudlet.location.timestamp.seconds",
	"location.timestamp.nanos=cloudlet.location.timestamp.nanos",
	"ipsupport=cloudlet.ipsupport",
	"staticips=cloudlet.staticips",
	"numdynamicips=cloudlet.numdynamicips",
	"timelimits.createclusterinsttimeout=cloudlet.timelimits.createclusterinsttimeout",
	"timelimits.updateclusterinsttimeout=cloudlet.timelimits.updateclusterinsttimeout",
	"timelimits.deleteclusterinsttimeout=cloudlet.timelimits.deleteclusterinsttimeout",
	"timelimits.createappinsttimeout=cloudlet.timelimits.createappinsttimeout",
	"timelimits.updateappinsttimeout=cloudlet.timelimits.updateappinsttimeout",
	"timelimits.deleteappinsttimeout=cloudlet.timelimits.deleteappinsttimeout",
}
var FlavorInfoRequiredArgs = []string{}
var FlavorInfoOptionalArgs = []string{
	"name",
	"vcpus",
	"ram",
	"disk",
}
var FlavorInfoAliasArgs = []string{
	"name=flavorinfo.name",
	"vcpus=flavorinfo.vcpus",
	"ram=flavorinfo.ram",
	"disk=flavorinfo.disk",
}
var CloudletInfoRequiredArgs = []string{
	"operator",
	"name",
}
var CloudletInfoOptionalArgs = []string{
	"state",
	"notifyid",
	"controller",
	"osmaxram",
	"osmaxvcores",
	"osmaxvolgb",
	"errors",
	"flavors.name",
	"flavors.vcpus",
	"flavors.ram",
	"flavors.disk",
}
var CloudletInfoAliasArgs = []string{
	"operator=cloudletinfo.key.operatorkey.name",
	"name=cloudletinfo.key.name",
	"state=cloudletinfo.state",
	"notifyid=cloudletinfo.notifyid",
	"controller=cloudletinfo.controller",
	"osmaxram=cloudletinfo.osmaxram",
	"osmaxvcores=cloudletinfo.osmaxvcores",
	"osmaxvolgb=cloudletinfo.osmaxvolgb",
	"errors=cloudletinfo.errors",
	"flavors.name=cloudletinfo.flavors.name",
	"flavors.vcpus=cloudletinfo.flavors.vcpus",
	"flavors.ram=cloudletinfo.flavors.ram",
	"flavors.disk=cloudletinfo.flavors.disk",
}
var CloudletMetricsRequiredArgs = []string{}
var CloudletMetricsOptionalArgs = []string{
	"foo",
}
var CloudletMetricsAliasArgs = []string{
	"foo=cloudletmetrics.foo",
}
