// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

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

var CreateAppCmd = &ApiCommand{
	Name:         "CreateApp",
	Use:          "create",
	Short:        "Create Application. Creates a definition for an application instance for Cloudlet deployment.",
	RequiredArgs: "region " + strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	NoConfig:     "DeletePrepare,CreatedAt,UpdatedAt,DelOpt,AutoProvPolicy",
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/CreateApp",
	ProtobufApi:  true,
}

var DeleteAppCmd = &ApiCommand{
	Name:         "DeleteApp",
	Use:          "delete",
	Short:        "Delete Application. Deletes a definition of an Application instance. Make sure no other application instances exist with that definition. If they do exist, you must delete those Application instances first.",
	RequiredArgs: "region " + strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	NoConfig:     "DeletePrepare,CreatedAt,UpdatedAt,DelOpt,AutoProvPolicy",
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/DeleteApp",
	ProtobufApi:  true,
}

var UpdateAppCmd = &ApiCommand{
	Name:          "UpdateApp",
	Use:           "update",
	Short:         "Update Application. Updates the definition of an Application instance.",
	RequiredArgs:  "region " + strings.Join(AppRequiredArgs, " "),
	OptionalArgs:  strings.Join(AppOptionalArgs, " "),
	AliasArgs:     strings.Join(AppAliasArgs, " "),
	SpecialArgs:   &AppSpecialArgs,
	Comments:      addRegionComment(AppComments),
	NoConfig:      "DeletePrepare,CreatedAt,UpdatedAt,DelOpt,AutoProvPolicy",
	ReqData:       &ormapi.RegionApp{},
	ReplyData:     &edgeproto.Result{},
	Path:          "/auth/ctrl/UpdateApp",
	SetFieldsFunc: SetUpdateAppFields,
	ProtobufApi:   true,
}

func SetUpdateAppFields(in map[string]interface{}) {
	// get map for edgeproto object in region struct
	obj := in["App"]
	if obj == nil {
		return
	}
	objmap, ok := obj.(map[string]interface{})
	if !ok {
		return
	}
	fields := cli.GetSpecifiedFields(objmap, &edgeproto.App{}, cli.JsonNamespace)
	// include fields already specified
	if inFields, found := objmap["fields"]; found {
		if fieldsArr, ok := inFields.([]string); ok {
			fields = append(fields, fieldsArr...)
		}
	}
	objmap["fields"] = fields
}

var ShowAppCmd = &ApiCommand{
	Name:         "ShowApp",
	Use:          "show",
	Short:        "Show Applications. Lists all Application definitions managed from the Edge Controller. Any fields specified will be used to filter results.",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(AppRequiredArgs, AppOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	NoConfig:     "DeletePrepare,CreatedAt,UpdatedAt,DelOpt,AutoProvPolicy",
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.App{},
	Path:         "/auth/ctrl/ShowApp",
	StreamOut:    true,
	ProtobufApi:  true,
}

var AddAppAutoProvPolicyCmd = &ApiCommand{
	Name:         "AddAppAutoProvPolicy",
	Use:          "addautoprovpolicy",
	Short:        "Add an AutoProvPolicy to the App",
	RequiredArgs: "region " + strings.Join(AppAutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AppAutoProvPolicySpecialArgs,
	Comments:     addRegionComment(AppAutoProvPolicyComments),
	ReqData:      &ormapi.RegionAppAutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/AddAppAutoProvPolicy",
	ProtobufApi:  true,
}

var RemoveAppAutoProvPolicyCmd = &ApiCommand{
	Name:         "RemoveAppAutoProvPolicy",
	Use:          "removeautoprovpolicy",
	Short:        "Remove an AutoProvPolicy from the App",
	RequiredArgs: "region " + strings.Join(AppAutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AppAutoProvPolicySpecialArgs,
	Comments:     addRegionComment(AppAutoProvPolicyComments),
	ReqData:      &ormapi.RegionAppAutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/RemoveAppAutoProvPolicy",
	ProtobufApi:  true,
}

var AppApiCmds = []*ApiCommand{
	CreateAppCmd,
	DeleteAppCmd,
	UpdateAppCmd,
	ShowAppCmd,
	AddAppAutoProvPolicyCmd,
	RemoveAppAutoProvPolicyCmd,
}

const AppGroup = "App"

func init() {
	AllApis.AddGroup(AppGroup, "Manage Apps", AppApiCmds)
}

var AppRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
}
var AppOptionalArgs = []string{
	"imagepath",
	"imagetype",
	"accessports",
	"defaultflavor",
	"authpublickey",
	"command",
	"annotations",
	"deployment",
	"deploymentmanifest",
	"deploymentgenerator",
	"androidpackagename",
	"configs:#.kind",
	"configs:#.config",
	"scalewithcluster",
	"internalports",
	"revision",
	"officialfqdn",
	"md5sum",
	"accesstype",
	"autoprovpolicies",
	"templatedelimiter",
	"skiphcports",
	"trusted",
	"requiredoutboundconnections:#.protocol",
	"requiredoutboundconnections:#.port",
	"requiredoutboundconnections:#.remoteip",
	"allowserverless",
	"serverlessconfig.vcpus",
	"serverlessconfig.ram",
	"serverlessconfig.minreplicas",
	"vmappostype",
}
var AppAliasArgs = []string{
	"fields=app.fields",
	"app-org=app.key.organization",
	"appname=app.key.name",
	"appvers=app.key.version",
	"imagepath=app.imagepath",
	"imagetype=app.imagetype",
	"accessports=app.accessports",
	"defaultflavor=app.defaultflavor.name",
	"authpublickey=app.authpublickey",
	"command=app.command",
	"annotations=app.annotations",
	"deployment=app.deployment",
	"deploymentmanifest=app.deploymentmanifest",
	"deploymentgenerator=app.deploymentgenerator",
	"androidpackagename=app.androidpackagename",
	"delopt=app.delopt",
	"configs:#.kind=app.configs:#.kind",
	"configs:#.config=app.configs:#.config",
	"scalewithcluster=app.scalewithcluster",
	"internalports=app.internalports",
	"revision=app.revision",
	"officialfqdn=app.officialfqdn",
	"md5sum=app.md5sum",
	"autoprovpolicy=app.autoprovpolicy",
	"accesstype=app.accesstype",
	"deleteprepare=app.deleteprepare",
	"autoprovpolicies=app.autoprovpolicies",
	"templatedelimiter=app.templatedelimiter",
	"skiphcports=app.skiphcports",
	"createdat.seconds=app.createdat.seconds",
	"createdat.nanos=app.createdat.nanos",
	"updatedat.seconds=app.updatedat.seconds",
	"updatedat.nanos=app.updatedat.nanos",
	"trusted=app.trusted",
	"requiredoutboundconnections:#.protocol=app.requiredoutboundconnections:#.protocol",
	"requiredoutboundconnections:#.port=app.requiredoutboundconnections:#.port",
	"requiredoutboundconnections:#.remoteip=app.requiredoutboundconnections:#.remoteip",
	"allowserverless=app.allowserverless",
	"serverlessconfig.vcpus=app.serverlessconfig.vcpus",
	"serverlessconfig.ram=app.serverlessconfig.ram",
	"serverlessconfig.minreplicas=app.serverlessconfig.minreplicas",
	"vmappostype=app.vmappostype",
}
var AppComments = map[string]string{
	"fields":                                 "Fields are used for the Update API to specify which fields to apply",
	"app-org":                                "App developer organization",
	"appname":                                "App name",
	"appvers":                                "App version",
	"imagepath":                              "URI of where image resides",
	"imagetype":                              "Image type (see ImageType), one of ImageTypeUnknown, ImageTypeDocker, ImageTypeQcow, ImageTypeHelm, ImageTypeOvf",
	"accessports":                            "Comma separated list of protocol:port pairs that the App listens on. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443",
	"defaultflavor":                          "Flavor name",
	"authpublickey":                          "Public key used for authentication",
	"command":                                "Command that the container runs to start service",
	"annotations":                            "Annotations is a comma separated map of arbitrary key value pairs, for example: key1=val1,key2=val2,key3=val 3",
	"deployment":                             "Deployment type (kubernetes, docker, or vm)",
	"deploymentmanifest":                     "Deployment manifest is the deployment specific manifest file/config. For docker deployment, this can be a docker-compose or docker run file. For kubernetes deployment, this can be a kubernetes yaml or helm chart file.",
	"deploymentgenerator":                    "Deployment generator target to generate a basic deployment manifest",
	"androidpackagename":                     "Android package name used to match the App name from the Android package",
	"delopt":                                 "Override actions to Controller, one of NoAutoDelete, AutoDelete",
	"configs:#.kind":                         "Kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"configs:#.config":                       "Config file contents or URI reference",
	"scalewithcluster":                       "Option to run App on all nodes of the cluster",
	"internalports":                          "Should this app have access to outside world?",
	"revision":                               "Revision can be specified or defaults to current timestamp when app is updated",
	"officialfqdn":                           "Official FQDN is the FQDN that the app uses to connect by default",
	"md5sum":                                 "MD5Sum of the VM-based app image",
	"autoprovpolicy":                         "(_deprecated_) Auto provisioning policy name",
	"accesstype":                             "(Deprecated) Access type, one of AccessTypeDefaultForDeployment, AccessTypeDirect, AccessTypeLoadBalancer",
	"deleteprepare":                          "Preparing to be deleted",
	"autoprovpolicies":                       "Auto provisioning policy names, may be specified multiple times",
	"templatedelimiter":                      "Delimiter to be used for template parsing, defaults to [[ ]]",
	"skiphcports":                            "Comma separated list of protocol:port pairs that we should not run health check on. Should be configured in case app does not always listen on these ports. all can be specified if no health check to be run for this app. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443.",
	"trusted":                                "Indicates that an instance of this app can be started on a trusted cloudlet",
	"requiredoutboundconnections:#.protocol": "tcp, udp or icmp",
	"requiredoutboundconnections:#.port":     "TCP or UDP port",
	"requiredoutboundconnections:#.remoteip": "remote IP X.X.X.X",
	"allowserverless":                        "App is allowed to deploy as serverless containers",
	"serverlessconfig.vcpus":                 "Virtual CPUs allocation per container when serverless, may be fractional in increments of 0.001",
	"serverlessconfig.ram":                   "RAM allocation in megabytes per container when serverless",
	"serverlessconfig.minreplicas":           "Minimum number of replicas when serverless",
	"vmappostype":                            "OS Type for VM Apps, one of VmAppOsUnknown, VmAppOsLinux, VmAppOsWindows10, VmAppOsWindows2012, VmAppOsWindows2016, VmAppOsWindows2019",
}
var AppSpecialArgs = map[string]string{
	"app.autoprovpolicies": "StringArray",
	"app.fields":           "StringArray",
}
var AppAutoProvPolicyRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"autoprovpolicy",
}
var AppAutoProvPolicyOptionalArgs = []string{}
var AppAutoProvPolicyAliasArgs = []string{
	"app-org=appautoprovpolicy.appkey.organization",
	"appname=appautoprovpolicy.appkey.name",
	"appvers=appautoprovpolicy.appkey.version",
	"autoprovpolicy=appautoprovpolicy.autoprovpolicy",
}
var AppAutoProvPolicyComments = map[string]string{
	"app-org":        "App developer organization",
	"appname":        "App name",
	"appvers":        "App version",
	"autoprovpolicy": "Auto provisioning policy name",
}
var AppAutoProvPolicySpecialArgs = map[string]string{}
