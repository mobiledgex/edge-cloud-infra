// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

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
	Name:         "UpdateApp",
	Use:          "update",
	Short:        "Update Application. Updates the definition of an Application instance.",
	RequiredArgs: "region " + strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	NoConfig:     "DeletePrepare,CreatedAt,UpdatedAt,DelOpt,AutoProvPolicy",
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/UpdateApp",
	ProtobufApi:  true,
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

var AddAppAlertPolicyCmd = &ApiCommand{
	Name:         "AddAppAlertPolicy",
	Use:          "addalertpolicy",
	Short:        "Add an AlertPolicy to the App",
	RequiredArgs: "region " + strings.Join(AppAlertPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAlertPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAlertPolicyAliasArgs, " "),
	SpecialArgs:  &AppAlertPolicySpecialArgs,
	Comments:     addRegionComment(AppAlertPolicyComments),
	ReqData:      &ormapi.RegionAppAlertPolicy{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/AddAppAlertPolicy",
	ProtobufApi:  true,
}

var RemoveAppAlertPolicyCmd = &ApiCommand{
	Name:         "RemoveAppAlertPolicy",
	Use:          "removealertpolicy",
	Short:        "Remove an AlertPolicy from the App",
	RequiredArgs: "region " + strings.Join(AppAlertPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAlertPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAlertPolicyAliasArgs, " "),
	SpecialArgs:  &AppAlertPolicySpecialArgs,
	Comments:     addRegionComment(AppAlertPolicyComments),
	ReqData:      &ormapi.RegionAppAlertPolicy{},
	ReplyData:    &edgeproto.Result{},
	Path:         "/auth/ctrl/RemoveAppAlertPolicy",
	ProtobufApi:  true,
}

var ShowCloudletsForAppDeploymentCmd = &ApiCommand{
	Name:                 "ShowCloudletsForAppDeployment",
	Use:                  "showcloudletsfordeployment",
	Short:                "Discover cloudlets supporting deployments of App.DefaultFlavor",
	RequiredArgs:         "region",
	OptionalArgs:         strings.Join(append(DeploymentCloudletRequestRequiredArgs, DeploymentCloudletRequestOptionalArgs...), " "),
	AliasArgs:            strings.Join(DeploymentCloudletRequestAliasArgs, " "),
	SpecialArgs:          &DeploymentCloudletRequestSpecialArgs,
	Comments:             addRegionComment(DeploymentCloudletRequestComments),
	NoConfig:             "App.DeletePrepare,App.CreatedAt,App.UpdatedAt,App.DelOpt,App.AutoProvPolicy",
	ReqData:              &ormapi.RegionDeploymentCloudletRequest{},
	ReplyData:            &edgeproto.CloudletKey{},
	Path:                 "/auth/ctrl/ShowCloudletsForAppDeployment",
	StreamOut:            true,
	StreamOutIncremental: true,
	ProtobufApi:          true,
}
var AppApiCmds = []*ApiCommand{
	CreateAppCmd,
	DeleteAppCmd,
	UpdateAppCmd,
	ShowAppCmd,
	AddAppAutoProvPolicyCmd,
	RemoveAppAutoProvPolicyCmd,
	AddAppAlertPolicyCmd,
	RemoveAppAlertPolicyCmd,
	ShowCloudletsForAppDeploymentCmd,
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
	"configs:empty",
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
	"requiredoutboundconnections:empty",
	"requiredoutboundconnections:#.protocol",
	"requiredoutboundconnections:#.port",
	"requiredoutboundconnections:#.remoteip",
	"allowserverless",
	"serverlessconfig.vcpus",
	"serverlessconfig.ram",
	"serverlessconfig.minreplicas",
	"vmappostype",
	"alertpolicies",
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
	"configs:empty=app.configs:empty",
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
	"requiredoutboundconnections:empty=app.requiredoutboundconnections:empty",
	"requiredoutboundconnections:#.protocol=app.requiredoutboundconnections:#.protocol",
	"requiredoutboundconnections:#.port=app.requiredoutboundconnections:#.port",
	"requiredoutboundconnections:#.remoteip=app.requiredoutboundconnections:#.remoteip",
	"allowserverless=app.allowserverless",
	"serverlessconfig.vcpus=app.serverlessconfig.vcpus",
	"serverlessconfig.ram=app.serverlessconfig.ram",
	"serverlessconfig.minreplicas=app.serverlessconfig.minreplicas",
	"vmappostype=app.vmappostype",
	"alertpolicies=app.alertpolicies",
}
var AppComments = map[string]string{
	"fields":                                 "Fields are used for the Update API to specify which fields to apply",
	"app-org":                                "App developer organization",
	"appname":                                "App name",
	"appvers":                                "App version",
	"imagepath":                              "URI of where image resides",
	"imagetype":                              "Image type (see ImageType), one of Unknown, Docker, Qcow, Helm, Ovf",
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
	"configs:empty":                          "Customization files passed through to implementing services, specify configs:empty=true to clear",
	"configs:#.kind":                         "Kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"configs:#.config":                       "Config file contents or URI reference",
	"scalewithcluster":                       "Option to run App on all nodes of the cluster",
	"internalports":                          "Should this app have access to outside world?",
	"revision":                               "Revision can be specified or defaults to current timestamp when app is updated",
	"officialfqdn":                           "Official FQDN is the FQDN that the app uses to connect by default",
	"md5sum":                                 "MD5Sum of the VM-based app image",
	"autoprovpolicy":                         "(_deprecated_) Auto provisioning policy name",
	"accesstype":                             "(Deprecated) Access type, one of DefaultForDeployment, Direct, LoadBalancer",
	"deleteprepare":                          "Preparing to be deleted",
	"autoprovpolicies":                       "Auto provisioning policy names, may be specified multiple times, specify autoprovpolicies:empty=true to clear",
	"templatedelimiter":                      "Delimiter to be used for template parsing, defaults to [[ ]]",
	"skiphcports":                            "Comma separated list of protocol:port pairs that we should not run health check on. Should be configured in case app does not always listen on these ports. all can be specified if no health check to be run for this app. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443.",
	"trusted":                                "Indicates that an instance of this app can be started on a trusted cloudlet",
	"requiredoutboundconnections:empty":      "Connections this app require to determine if the app is compatible with a trust policy, specify requiredoutboundconnections:empty=true to clear",
	"requiredoutboundconnections:#.protocol": "tcp, udp or icmp",
	"requiredoutboundconnections:#.port":     "TCP or UDP port",
	"requiredoutboundconnections:#.remoteip": "remote IP X.X.X.X",
	"allowserverless":                        "App is allowed to deploy as serverless containers",
	"serverlessconfig.vcpus":                 "Virtual CPUs allocation per container when serverless, may be fractional in increments of 0.001",
	"serverlessconfig.ram":                   "RAM allocation in megabytes per container when serverless",
	"serverlessconfig.minreplicas":           "Minimum number of replicas when serverless",
	"vmappostype":                            "OS Type for VM Apps, one of Unknown, Linux, Windows10, Windows2012, Windows2016, Windows2019",
	"alertpolicies":                          "Alerts Policies, specify alertpolicies:empty=true to clear",
}
var AppSpecialArgs = map[string]string{
	"app.alertpolicies":    "StringArray",
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
var AppAlertPolicyRequiredArgs = []string{
	"app-org",
	"appname",
	"appvers",
	"alertpolicy-name",
}
var AppAlertPolicyOptionalArgs = []string{}
var AppAlertPolicyAliasArgs = []string{
	"app-org=appalertpolicy.appkey.organization",
	"appname=appalertpolicy.appkey.name",
	"appvers=appalertpolicy.appkey.version",
	"alertpolicy-name=appalertpolicy.alertpolicy",
}
var AppAlertPolicyComments = map[string]string{
	"app-org":          "App developer organization",
	"appname":          "App name",
	"appvers":          "App version",
	"alertpolicy-name": "Alert name",
}
var AppAlertPolicySpecialArgs = map[string]string{}
var DeploymentCloudletRequestRequiredArgs = []string{}
var DeploymentCloudletRequestOptionalArgs = []string{
	"app.fields",
	"app.key.organization",
	"appname",
	"appvers",
	"app.imagepath",
	"app.imagetype",
	"app.accessports",
	"app.defaultflavor.name",
	"app.authpublickey",
	"app.command",
	"app.annotations",
	"app.deployment",
	"app.deploymentmanifest",
	"app.deploymentgenerator",
	"app.androidpackagename",
	"app.configs:#.kind",
	"app.configs:#.config",
	"app.scalewithcluster",
	"app.internalports",
	"app.revision",
	"app.officialfqdn",
	"app.md5sum",
	"app.accesstype",
	"app.autoprovpolicies",
	"app.templatedelimiter",
	"app.skiphcports",
	"app.trusted",
	"app.requiredoutboundconnections:#.protocol",
	"app.requiredoutboundconnections:#.port",
	"app.requiredoutboundconnections:#.remoteip",
	"app.allowserverless",
	"app.serverlessconfig.vcpus",
	"app.serverlessconfig.ram",
	"app.serverlessconfig.minreplicas",
	"app.vmappostype",
	"app.alertpolicies",
	"dryrundeploy",
	"numnodes",
}
var DeploymentCloudletRequestAliasArgs = []string{
	"app.fields=deploymentcloudletrequest.app.fields",
	"app.key.organization=deploymentcloudletrequest.app.key.organization",
	"appname=deploymentcloudletrequest.app.key.name",
	"appvers=deploymentcloudletrequest.app.key.version",
	"app.imagepath=deploymentcloudletrequest.app.imagepath",
	"app.imagetype=deploymentcloudletrequest.app.imagetype",
	"app.accessports=deploymentcloudletrequest.app.accessports",
	"app.defaultflavor.name=deploymentcloudletrequest.app.defaultflavor.name",
	"app.authpublickey=deploymentcloudletrequest.app.authpublickey",
	"app.command=deploymentcloudletrequest.app.command",
	"app.annotations=deploymentcloudletrequest.app.annotations",
	"app.deployment=deploymentcloudletrequest.app.deployment",
	"app.deploymentmanifest=deploymentcloudletrequest.app.deploymentmanifest",
	"app.deploymentgenerator=deploymentcloudletrequest.app.deploymentgenerator",
	"app.androidpackagename=deploymentcloudletrequest.app.androidpackagename",
	"app.delopt=deploymentcloudletrequest.app.delopt",
	"app.configs:#.kind=deploymentcloudletrequest.app.configs:#.kind",
	"app.configs:#.config=deploymentcloudletrequest.app.configs:#.config",
	"app.scalewithcluster=deploymentcloudletrequest.app.scalewithcluster",
	"app.internalports=deploymentcloudletrequest.app.internalports",
	"app.revision=deploymentcloudletrequest.app.revision",
	"app.officialfqdn=deploymentcloudletrequest.app.officialfqdn",
	"app.md5sum=deploymentcloudletrequest.app.md5sum",
	"app.autoprovpolicy=deploymentcloudletrequest.app.autoprovpolicy",
	"app.accesstype=deploymentcloudletrequest.app.accesstype",
	"app.deleteprepare=deploymentcloudletrequest.app.deleteprepare",
	"app.autoprovpolicies=deploymentcloudletrequest.app.autoprovpolicies",
	"app.templatedelimiter=deploymentcloudletrequest.app.templatedelimiter",
	"app.skiphcports=deploymentcloudletrequest.app.skiphcports",
	"app.createdat.seconds=deploymentcloudletrequest.app.createdat.seconds",
	"app.createdat.nanos=deploymentcloudletrequest.app.createdat.nanos",
	"app.updatedat.seconds=deploymentcloudletrequest.app.updatedat.seconds",
	"app.updatedat.nanos=deploymentcloudletrequest.app.updatedat.nanos",
	"app.trusted=deploymentcloudletrequest.app.trusted",
	"app.requiredoutboundconnections:#.protocol=deploymentcloudletrequest.app.requiredoutboundconnections:#.protocol",
	"app.requiredoutboundconnections:#.port=deploymentcloudletrequest.app.requiredoutboundconnections:#.port",
	"app.requiredoutboundconnections:#.remoteip=deploymentcloudletrequest.app.requiredoutboundconnections:#.remoteip",
	"app.allowserverless=deploymentcloudletrequest.app.allowserverless",
	"app.serverlessconfig.vcpus=deploymentcloudletrequest.app.serverlessconfig.vcpus",
	"app.serverlessconfig.ram=deploymentcloudletrequest.app.serverlessconfig.ram",
	"app.serverlessconfig.minreplicas=deploymentcloudletrequest.app.serverlessconfig.minreplicas",
	"app.vmappostype=deploymentcloudletrequest.app.vmappostype",
	"app.alertpolicies=deploymentcloudletrequest.app.alertpolicies",
	"dryrundeploy=deploymentcloudletrequest.dryrundeploy",
	"numnodes=deploymentcloudletrequest.numnodes",
}
var DeploymentCloudletRequestComments = map[string]string{
	"app.fields":              "Fields are used for the Update API to specify which fields to apply",
	"app.key.organization":    "App developer organization",
	"appname":                 "App name",
	"appvers":                 "App version",
	"app.imagepath":           "URI of where image resides",
	"app.imagetype":           "Image type (see ImageType), one of Unknown, Docker, Qcow, Helm, Ovf",
	"app.accessports":         "Comma separated list of protocol:port pairs that the App listens on. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443",
	"app.defaultflavor.name":  "Flavor name",
	"app.authpublickey":       "Public key used for authentication",
	"app.command":             "Command that the container runs to start service",
	"app.annotations":         "Annotations is a comma separated map of arbitrary key value pairs, for example: key1=val1,key2=val2,key3=val 3",
	"app.deployment":          "Deployment type (kubernetes, docker, or vm)",
	"app.deploymentmanifest":  "Deployment manifest is the deployment specific manifest file/config. For docker deployment, this can be a docker-compose or docker run file. For kubernetes deployment, this can be a kubernetes yaml or helm chart file.",
	"app.deploymentgenerator": "Deployment generator target to generate a basic deployment manifest",
	"app.androidpackagename":  "Android package name used to match the App name from the Android package",
	"app.delopt":              "Override actions to Controller, one of NoAutoDelete, AutoDelete",
	"app.configs:#.kind":      "Kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"app.configs:#.config":    "Config file contents or URI reference",
	"app.scalewithcluster":    "Option to run App on all nodes of the cluster",
	"app.internalports":       "Should this app have access to outside world?",
	"app.revision":            "Revision can be specified or defaults to current timestamp when app is updated",
	"app.officialfqdn":        "Official FQDN is the FQDN that the app uses to connect by default",
	"app.md5sum":              "MD5Sum of the VM-based app image",
	"app.autoprovpolicy":      "(_deprecated_) Auto provisioning policy name",
	"app.accesstype":          "(Deprecated) Access type, one of DefaultForDeployment, Direct, LoadBalancer",
	"app.deleteprepare":       "Preparing to be deleted",
	"app.autoprovpolicies":    "Auto provisioning policy names, may be specified multiple times",
	"app.templatedelimiter":   "Delimiter to be used for template parsing, defaults to [[ ]]",
	"app.skiphcports":         "Comma separated list of protocol:port pairs that we should not run health check on. Should be configured in case app does not always listen on these ports. all can be specified if no health check to be run for this app. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443.",
	"app.trusted":             "Indicates that an instance of this app can be started on a trusted cloudlet",
	"app.requiredoutboundconnections:#.protocol": "tcp, udp or icmp",
	"app.requiredoutboundconnections:#.port":     "TCP or UDP port",
	"app.requiredoutboundconnections:#.remoteip": "remote IP X.X.X.X",
	"app.allowserverless":                        "App is allowed to deploy as serverless containers",
	"app.serverlessconfig.vcpus":                 "Virtual CPUs allocation per container when serverless, may be fractional in increments of 0.001",
	"app.serverlessconfig.ram":                   "RAM allocation in megabytes per container when serverless",
	"app.serverlessconfig.minreplicas":           "Minimum number of replicas when serverless",
	"app.vmappostype":                            "OS Type for VM Apps, one of Unknown, Linux, Windows10, Windows2012, Windows2016, Windows2019",
	"app.alertpolicies":                          "Alerts Policies",
	"dryrundeploy":                               "Attempt to qualify cloudlet resources for deployment",
	"numnodes":                                   "Optional number of worker VMs in dry run K8s Cluster, default = 2",
}
var DeploymentCloudletRequestSpecialArgs = map[string]string{
	"deploymentcloudletrequest.app.alertpolicies":    "StringArray",
	"deploymentcloudletrequest.app.autoprovpolicies": "StringArray",
	"deploymentcloudletrequest.app.fields":           "StringArray",
}
