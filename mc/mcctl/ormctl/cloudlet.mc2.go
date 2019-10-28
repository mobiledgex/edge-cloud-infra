// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

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

var CreateCloudletCmd = &cli.Command{
	Use:                  "CreateCloudlet",
	RequiredArgs:         strings.Join(append([]string{"region"}, CloudletRequiredArgs...), " "),
	OptionalArgs:         strings.Join(CloudletOptionalArgs, " "),
	AliasArgs:            strings.Join(CloudletAliasArgs, " "),
	SpecialArgs:          &CloudletSpecialArgs,
	Comments:             addRegionComment(CloudletComments),
	ReqData:              &ormapi.RegionCloudlet{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/CreateCloudlet"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var DeleteCloudletCmd = &cli.Command{
	Use:                  "DeleteCloudlet",
	RequiredArgs:         strings.Join(append([]string{"region"}, CloudletRequiredArgs...), " "),
	OptionalArgs:         strings.Join(CloudletOptionalArgs, " "),
	AliasArgs:            strings.Join(CloudletAliasArgs, " "),
	SpecialArgs:          &CloudletSpecialArgs,
	Comments:             addRegionComment(CloudletComments),
	ReqData:              &ormapi.RegionCloudlet{},
	ReplyData:            &edgeproto.Result{},
	Run:                  runRest("/auth/ctrl/DeleteCloudlet"),
	StreamOut:            true,
	StreamOutIncremental: true,
}

var UpdateCloudletCmd = &cli.Command{
	Use:          "UpdateCloudlet",
	RequiredArgs: strings.Join(append([]string{"region"}, CloudletRequiredArgs...), " "),
	OptionalArgs: strings.Join(CloudletOptionalArgs, " "),
	AliasArgs:    strings.Join(CloudletAliasArgs, " "),
	SpecialArgs:  &CloudletSpecialArgs,
	Comments:     addRegionComment(CloudletComments),
	ReqData:      &ormapi.RegionCloudlet{},
	ReplyData:    &edgeproto.Result{},
	Run: runRest("/auth/ctrl/UpdateCloudlet",
		withSetFieldsFunc(setUpdateCloudletFields),
	),
	StreamOut:            true,
	StreamOutIncremental: true,
}

func setUpdateCloudletFields(in map[string]interface{}) {
	// get map for edgeproto object in region struct
	obj := in[strings.ToLower("Cloudlet")]
	if obj == nil {
		return
	}
	objmap, ok := obj.(map[string]interface{})
	if !ok {
		return
	}
	objmap["fields"] = cli.GetSpecifiedFields(objmap, &edgeproto.Cloudlet{}, cli.JsonNamespace)
}

var ShowCloudletCmd = &cli.Command{
	Use:          "ShowCloudlet",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(CloudletRequiredArgs, CloudletOptionalArgs...), " "),
	AliasArgs:    strings.Join(CloudletAliasArgs, " "),
	SpecialArgs:  &CloudletSpecialArgs,
	Comments:     addRegionComment(CloudletComments),
	ReqData:      &ormapi.RegionCloudlet{},
	ReplyData:    &edgeproto.Cloudlet{},
	Run:          runRest("/auth/ctrl/ShowCloudlet"),
	StreamOut:    true,
}

var CloudletApiCmds = []*cli.Command{
	CreateCloudletCmd,
	DeleteCloudletCmd,
	UpdateCloudletCmd,
	ShowCloudletCmd,
}

var ShowCloudletInfoCmd = &cli.Command{
	Use:          "ShowCloudletInfo",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(CloudletInfoRequiredArgs, CloudletInfoOptionalArgs...), " "),
	AliasArgs:    strings.Join(CloudletInfoAliasArgs, " "),
	SpecialArgs:  &CloudletInfoSpecialArgs,
	Comments:     addRegionComment(CloudletInfoComments),
	ReqData:      &ormapi.RegionCloudletInfo{},
	ReplyData:    &edgeproto.CloudletInfo{},
	Run:          runRest("/auth/ctrl/ShowCloudletInfo"),
	StreamOut:    true,
}

var CloudletInfoApiCmds = []*cli.Command{
	ShowCloudletInfoCmd,
}

var CloudletKeyRequiredArgs = []string{}
var CloudletKeyOptionalArgs = []string{
	"operator",
	"name",
}
var CloudletKeyAliasArgs = []string{
	"operator=cloudletkey.operatorkey.name",
	"name=cloudletkey.name",
}
var CloudletKeyComments = map[string]string{
	"operator": "Company or Organization name of the operator",
	"name":     "Name of the cloudlet",
}
var CloudletKeySpecialArgs = map[string]string{}
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
var OperationTimeLimitsComments = map[string]string{
	"createclusterinsttimeout": "max time to create a cluster instance",
	"updateclusterinsttimeout": "max time to update a cluster instance",
	"deleteclusterinsttimeout": "max time to delete a cluster instance",
	"createappinsttimeout":     "max time to create an app instance",
	"updateappinsttimeout":     "max time to update an app instance",
	"deleteappinsttimeout":     "max time to delete an app instance",
}
var OperationTimeLimitsSpecialArgs = map[string]string{}
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
var CloudletInfraCommonComments = map[string]string{
	"dockerregistry":       "the mex docker registry, e.g.  registry.mobiledgex.net:5000.",
	"dnszone":              "DNS Zone",
	"registryfileserver":   "registry file server contains files which get pulled on instantiation such as certs and images",
	"cfkey":                "Cloudflare key",
	"cfuser":               "Cloudflare key",
	"dockerregpass":        "Docker registry password",
	"networkscheme":        "network scheme",
	"dockerregistrysecret": "the name of the docker registry secret, e.g. mexgitlabsecret",
}
var CloudletInfraCommonSpecialArgs = map[string]string{}
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
var AzurePropertiesComments = map[string]string{
	"location":      "azure region e.g. uswest2",
	"resourcegroup": "azure resource group",
	"username":      "azure username",
	"password":      "azure password",
}
var AzurePropertiesSpecialArgs = map[string]string{}
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
var GcpPropertiesComments = map[string]string{
	"project":        "gcp project for billing",
	"zone":           "availability zone",
	"serviceaccount": "service account to login with",
	"gcpauthkeyurl":  "vault credentials link",
}
var GcpPropertiesSpecialArgs = map[string]string{}
var OpenStackPropertiesRequiredArgs = []string{}
var OpenStackPropertiesOptionalArgs = []string{
	"osexternalnetworkname",
	"osimagename",
	"osexternalroutername",
	"osmexnetwork",
	"openrcvars",
}
var OpenStackPropertiesAliasArgs = []string{
	"osexternalnetworkname=openstackproperties.osexternalnetworkname",
	"osimagename=openstackproperties.osimagename",
	"osexternalroutername=openstackproperties.osexternalroutername",
	"osmexnetwork=openstackproperties.osmexnetwork",
	"openrcvars=openstackproperties.openrcvars",
}
var OpenStackPropertiesComments = map[string]string{
	"osexternalnetworkname": "name of the external network, e.g. external-network-shared",
	"osimagename":           "openstack image , e.g. mobiledgex",
	"osexternalroutername":  "openstack router",
	"osmexnetwork":          "openstack internal network",
	"openrcvars":            "openrc env vars",
}
var OpenStackPropertiesSpecialArgs = map[string]string{
	"openrcvars": "StringToString",
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
var OpenRcVarsEntryComments = map[string]string{}
var OpenRcVarsEntrySpecialArgs = map[string]string{}
var CloudletInfraPropertiesRequiredArgs = []string{}
var CloudletInfraPropertiesOptionalArgs = []string{
	"cloudletkind",
	"mexoscontainerimagename",
	"openstackproperties.osexternalnetworkname",
	"openstackproperties.osimagename",
	"openstackproperties.osexternalroutername",
	"openstackproperties.osmexnetwork",
	"openstackproperties.openrcvars",
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
	"openstackproperties.openrcvars=cloudletinfraproperties.openstackproperties.openrcvars",
	"azureproperties.location=cloudletinfraproperties.azureproperties.location",
	"azureproperties.resourcegroup=cloudletinfraproperties.azureproperties.resourcegroup",
	"azureproperties.username=cloudletinfraproperties.azureproperties.username",
	"azureproperties.password=cloudletinfraproperties.azureproperties.password",
	"gcpproperties.project=cloudletinfraproperties.gcpproperties.project",
	"gcpproperties.zone=cloudletinfraproperties.gcpproperties.zone",
	"gcpproperties.serviceaccount=cloudletinfraproperties.gcpproperties.serviceaccount",
	"gcpproperties.gcpauthkeyurl=cloudletinfraproperties.gcpproperties.gcpauthkeyurl",
}
var CloudletInfraPropertiesComments = map[string]string{
	"cloudletkind":                              "what kind of infrastructure: Azure, GCP, Openstack",
	"mexoscontainerimagename":                   "name and version of the docker image container image that mexos runs in",
	"openstackproperties.osexternalnetworkname": "name of the external network, e.g. external-network-shared",
	"openstackproperties.osimagename":           "openstack image , e.g. mobiledgex",
	"openstackproperties.osexternalroutername":  "openstack router",
	"openstackproperties.osmexnetwork":          "openstack internal network",
	"openstackproperties.openrcvars":            "openrc env vars",
	"azureproperties.location":                  "azure region e.g. uswest2",
	"azureproperties.resourcegroup":             "azure resource group",
	"azureproperties.username":                  "azure username",
	"azureproperties.password":                  "azure password",
	"gcpproperties.project":                     "gcp project for billing",
	"gcpproperties.zone":                        "availability zone",
	"gcpproperties.serviceaccount":              "service account to login with",
	"gcpproperties.gcpauthkeyurl":               "vault credentials link",
}
var CloudletInfraPropertiesSpecialArgs = map[string]string{
	"openstackproperties.openrcvars": "StringToString",
}
var PlatformConfigRequiredArgs = []string{}
var PlatformConfigOptionalArgs = []string{
	"registrypath",
	"imagepath",
	"notifyctrladdrs",
	"vaultaddr",
	"tlscertfile",
	"crmroleid",
	"crmsecretid",
	"platformtag",
	"testmode",
	"span",
}
var PlatformConfigAliasArgs = []string{
	"registrypath=platformconfig.registrypath",
	"imagepath=platformconfig.imagepath",
	"notifyctrladdrs=platformconfig.notifyctrladdrs",
	"vaultaddr=platformconfig.vaultaddr",
	"tlscertfile=platformconfig.tlscertfile",
	"crmroleid=platformconfig.crmroleid",
	"crmsecretid=platformconfig.crmsecretid",
	"platformtag=platformconfig.platformtag",
	"testmode=platformconfig.testmode",
	"span=platformconfig.span",
}
var PlatformConfigComments = map[string]string{
	"registrypath":    "Path to Docker registry holding edge-cloud image",
	"imagepath":       "Path to platform base image",
	"notifyctrladdrs": "Address of controller notify port (can be multiple of these)",
	"vaultaddr":       "Vault address",
	"tlscertfile":     "TLS cert file",
	"crmroleid":       "Vault role ID for CRM",
	"crmsecretid":     "Vault secret ID for CRM",
	"platformtag":     "Tag of edge-cloud image",
	"testmode":        "Internal Test Flag",
	"span":            "Span string",
}
var PlatformConfigSpecialArgs = map[string]string{}
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
	"errors",
	"state",
	"crmoverride",
	"deploymentlocal",
	"platformtype",
	"flavor.name",
	"physicalname",
	"envvar",
	"upgrade",
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
	"errors=cloudlet.errors",
	"status.tasknumber=cloudlet.status.tasknumber",
	"status.maxtasks=cloudlet.status.maxtasks",
	"status.taskname=cloudlet.status.taskname",
	"status.stepname=cloudlet.status.stepname",
	"state=cloudlet.state",
	"crmoverride=cloudlet.crmoverride",
	"deploymentlocal=cloudlet.deploymentlocal",
	"platformtype=cloudlet.platformtype",
	"notifysrvaddr=cloudlet.notifysrvaddr",
	"flavor.name=cloudlet.flavor.name",
	"physicalname=cloudlet.physicalname",
	"envvar=cloudlet.envvar",
	"upgrade=cloudlet.upgrade",
	"config.registrypath=cloudlet.config.registrypath",
	"config.imagepath=cloudlet.config.imagepath",
	"config.notifyctrladdrs=cloudlet.config.notifyctrladdrs",
	"config.vaultaddr=cloudlet.config.vaultaddr",
	"config.tlscertfile=cloudlet.config.tlscertfile",
	"config.crmroleid=cloudlet.config.crmroleid",
	"config.crmsecretid=cloudlet.config.crmsecretid",
	"config.platformtag=cloudlet.config.platformtag",
	"config.testmode=cloudlet.config.testmode",
	"config.span=cloudlet.config.span",
}
var CloudletComments = map[string]string{
	"operator":                            "Company or Organization name of the operator",
	"name":                                "Name of the cloudlet",
	"accesscredentials":                   "Placeholder for cloudlet access credentials, i.e. openstack keys, passwords, etc",
	"location.latitude":                   "latitude in WGS 84 coordinates",
	"location.longitude":                  "longitude in WGS 84 coordinates",
	"location.horizontalaccuracy":         "horizontal accuracy (radius in meters)",
	"location.verticalaccuracy":           "veritical accuracy (meters)",
	"location.altitude":                   "On android only lat and long are guaranteed to be supplied altitude in meters",
	"location.course":                     "course (IOS) / bearing (Android) (degrees east relative to true north)",
	"location.speed":                      "speed (IOS) / velocity (Android) (meters/sec)",
	"ipsupport":                           "Type of IP support provided by Cloudlet (see IpSupport), one of IpSupportUnknown, IpSupportStatic, IpSupportDynamic",
	"staticips":                           "List of static IPs for static IP support",
	"numdynamicips":                       "Number of dynamic IPs available for dynamic IP support",
	"timelimits.createclusterinsttimeout": "max time to create a cluster instance",
	"timelimits.updateclusterinsttimeout": "max time to update a cluster instance",
	"timelimits.deleteclusterinsttimeout": "max time to delete a cluster instance",
	"timelimits.createappinsttimeout":     "max time to create an app instance",
	"timelimits.updateappinsttimeout":     "max time to update an app instance",
	"timelimits.deleteappinsttimeout":     "max time to delete an app instance",
	"errors":                              "Any errors trying to create, update, or delete the Cloudlet.",
	"state":                               "Current state of the cloudlet, one of TrackedStateUnknown, NotPresent, CreateRequested, Creating, CreateError, Ready, UpdateRequested, Updating, UpdateError, DeleteRequested, Deleting, DeleteError, DeletePrepare",
	"crmoverride":                         "Override actions to CRM, one of NoOverride, IgnoreCrmErrors, IgnoreCrm, IgnoreTransientState, IgnoreCrmAndTransientState",
	"deploymentlocal":                     "Deploy cloudlet services locally",
	"platformtype":                        "Platform type, one of PlatformTypeFake, PlatformTypeDind, PlatformTypeOpenstack, PlatformTypeAzure, PlatformTypeGcp, PlatformTypeMexdind",
	"notifysrvaddr":                       "Address for the CRM notify listener to run on",
	"flavor.name":                         "Flavor name",
	"physicalname":                        "Physical infrastructure cloudlet name",
	"envvar":                              "Single Key-Value pair of env var to be passed to CRM",
	"upgrade":                             "Upgrade cloudlet services",
	"config.registrypath":                 "Path to Docker registry holding edge-cloud image",
	"config.imagepath":                    "Path to platform base image",
	"config.notifyctrladdrs":              "Address of controller notify port (can be multiple of these)",
	"config.vaultaddr":                    "Vault address",
	"config.tlscertfile":                  "TLS cert file",
	"config.crmroleid":                    "Vault role ID for CRM",
	"config.crmsecretid":                  "Vault secret ID for CRM",
	"config.platformtag":                  "Tag of edge-cloud image",
	"config.testmode":                     "Internal Test Flag",
	"config.span":                         "Span string",
}
var CloudletSpecialArgs = map[string]string{
	"envvar": "StringToString",
	"errors": "StringArray",
}
var EnvVarEntryRequiredArgs = []string{}
var EnvVarEntryOptionalArgs = []string{
	"key",
	"value",
}
var EnvVarEntryAliasArgs = []string{
	"key=envvarentry.key",
	"value=envvarentry.value",
}
var EnvVarEntryComments = map[string]string{}
var EnvVarEntrySpecialArgs = map[string]string{}
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
var FlavorInfoComments = map[string]string{
	"name":  "Name of the flavor on the Cloudlet",
	"vcpus": "Number of VCPU cores on the Cloudlet",
	"ram":   "Ram in MB on the Cloudlet",
	"disk":  "Amount of disk in GB on the Cloudlet",
}
var FlavorInfoSpecialArgs = map[string]string{}
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
	"status.tasknumber",
	"status.maxtasks",
	"status.taskname",
	"status.stepname",
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
	"status.tasknumber=cloudletinfo.status.tasknumber",
	"status.maxtasks=cloudletinfo.status.maxtasks",
	"status.taskname=cloudletinfo.status.taskname",
	"status.stepname=cloudletinfo.status.stepname",
}
var CloudletInfoComments = map[string]string{
	"operator":      "Company or Organization name of the operator",
	"name":          "Name of the cloudlet",
	"state":         "State of cloudlet, one of CloudletStateUnknown, CloudletStateErrors, CloudletStateReady, CloudletStateOffline, CloudletStateNotPresent, CloudletStateInit, CloudletStateUpgrade",
	"notifyid":      "Id of client assigned by server (internal use only)",
	"controller":    "Connected controller unique id",
	"osmaxram":      "Maximum Ram in MB on the Cloudlet",
	"osmaxvcores":   "Maximum number of VCPU cores on the Cloudlet",
	"osmaxvolgb":    "Maximum amount of disk in GB on the Cloudlet",
	"errors":        "Any errors encountered while making changes to the Cloudlet",
	"flavors.name":  "Name of the flavor on the Cloudlet",
	"flavors.vcpus": "Number of VCPU cores on the Cloudlet",
	"flavors.ram":   "Ram in MB on the Cloudlet",
	"flavors.disk":  "Amount of disk in GB on the Cloudlet",
}
var CloudletInfoSpecialArgs = map[string]string{
	"errors": "StringArray",
}
var CloudletMetricsRequiredArgs = []string{}
var CloudletMetricsOptionalArgs = []string{
	"foo",
}
var CloudletMetricsAliasArgs = []string{
	"foo=cloudletmetrics.foo",
}
var CloudletMetricsComments = map[string]string{
	"foo": "what goes here?",
}
var CloudletMetricsSpecialArgs = map[string]string{}
