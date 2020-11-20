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

var CreateAppCmd = &cli.Command{
	Use:          "CreateApp",
	RequiredArgs: "region " + strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/CreateApp"),
}

var DeleteAppCmd = &cli.Command{
	Use:          "DeleteApp",
	RequiredArgs: "region " + strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/DeleteApp"),
}

var UpdateAppCmd = &cli.Command{
	Use:          "UpdateApp",
	RequiredArgs: "region " + strings.Join(AppRequiredArgs, " "),
	OptionalArgs: strings.Join(AppOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.Result{},
	Run: runRest("/auth/ctrl/UpdateApp",
		withSetFieldsFunc(setUpdateAppFields),
	),
}

func setUpdateAppFields(in map[string]interface{}) {
	// get map for edgeproto object in region struct
	obj := in[strings.ToLower("App")]
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

var ShowAppCmd = &cli.Command{
	Use:          "ShowApp",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(AppRequiredArgs, AppOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppAliasArgs, " "),
	SpecialArgs:  &AppSpecialArgs,
	Comments:     addRegionComment(AppComments),
	ReqData:      &ormapi.RegionApp{},
	ReplyData:    &edgeproto.App{},
	Run:          runRest("/auth/ctrl/ShowApp"),
	StreamOut:    true,
}

var AddAppAutoProvPolicyCmd = &cli.Command{
	Use:          "AddAppAutoProvPolicy",
	RequiredArgs: "region " + strings.Join(AppAutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AppAutoProvPolicySpecialArgs,
	Comments:     addRegionComment(AppAutoProvPolicyComments),
	ReqData:      &ormapi.RegionAppAutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/AddAppAutoProvPolicy"),
}

var RemoveAppAutoProvPolicyCmd = &cli.Command{
	Use:          "RemoveAppAutoProvPolicy",
	RequiredArgs: "region " + strings.Join(AppAutoProvPolicyRequiredArgs, " "),
	OptionalArgs: strings.Join(AppAutoProvPolicyOptionalArgs, " "),
	AliasArgs:    strings.Join(AppAutoProvPolicyAliasArgs, " "),
	SpecialArgs:  &AppAutoProvPolicySpecialArgs,
	Comments:     addRegionComment(AppAutoProvPolicyComments),
	ReqData:      &ormapi.RegionAppAutoProvPolicy{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/RemoveAppAutoProvPolicy"),
}

var AppApiCmds = []*cli.Command{
	CreateAppCmd,
	DeleteAppCmd,
	UpdateAppCmd,
	ShowAppCmd,
	AddAppAutoProvPolicyCmd,
	RemoveAppAutoProvPolicyCmd,
}

var AppKeyRequiredArgs = []string{}
var AppKeyOptionalArgs = []string{
	"organization",
	"name",
	"version",
}
var AppKeyAliasArgs = []string{
	"organization=appkey.organization",
	"name=appkey.name",
	"version=appkey.version",
}
var AppKeyComments = map[string]string{
	"organization": "App developer organization",
	"name":         "App name",
	"version":      "App version",
}
var AppKeySpecialArgs = map[string]string{}
var ConfigFileRequiredArgs = []string{}
var ConfigFileOptionalArgs = []string{
	"kind",
	"config",
}
var ConfigFileAliasArgs = []string{
	"kind=configfile.kind",
	"config=configfile.config",
}
var ConfigFileComments = map[string]string{
	"kind":   "kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"config": "config file contents or URI reference",
}
var ConfigFileSpecialArgs = map[string]string{}
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
	"delopt",
	"configs:#.kind",
	"configs:#.config",
	"scalewithcluster",
	"internalports",
	"revision",
	"officialfqdn",
	"md5sum",
	"defaultsharedvolumesize",
	"autoprovpolicy",
	"accesstype",
	"defaultprivacypolicy",
	"autoprovpolicies",
	"templatedelimiter",
	"skiphcports",
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
	"defaultsharedvolumesize=app.defaultsharedvolumesize",
	"autoprovpolicy=app.autoprovpolicy",
	"accesstype=app.accesstype",
	"defaultprivacypolicy=app.defaultprivacypolicy",
	"deleteprepare=app.deleteprepare",
	"autoprovpolicies=app.autoprovpolicies",
	"templatedelimiter=app.templatedelimiter",
	"skiphcports=app.skiphcports",
	"createdat.seconds=app.createdat.seconds",
	"createdat.nanos=app.createdat.nanos",
	"updatedat.seconds=app.updatedat.seconds",
	"updatedat.nanos=app.updatedat.nanos",
}
var AppComments = map[string]string{
	"fields":                  "Fields are used for the Update API to specify which fields to apply",
	"app-org":                 "App developer organization",
	"appname":                 "App name",
	"appvers":                 "App version",
	"imagepath":               "URI of where image resides",
	"imagetype":               "Image type (see ImageType), one of ImageTypeUnknown, ImageTypeDocker, ImageTypeQcow, ImageTypeHelm",
	"accessports":             "Comma separated list of protocol:port pairs that the App listens on. Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443",
	"defaultflavor":           "Flavor name",
	"authpublickey":           "public key used for authentication",
	"command":                 "Command that the container runs to start service",
	"annotations":             "Annotations is a comma separated map of arbitrary key value pairs, for example: key1=val1,key2=val2,key3=val 3",
	"deployment":              "Deployment type (kubernetes, docker, or vm)",
	"deploymentmanifest":      "Deployment manifest is the deployment specific manifest file/config For docker deployment, this can be a docker-compose or docker run file For kubernetes deployment, this can be a kubernetes yaml or helm chart file",
	"deploymentgenerator":     "Deployment generator target to generate a basic deployment manifest",
	"androidpackagename":      "Android package name used to match the App name from the Android package",
	"delopt":                  "Override actions to Controller, one of NoAutoDelete, AutoDelete",
	"configs:#.kind":          "kind (type) of config, i.e. envVarsYaml, helmCustomizationYaml",
	"configs:#.config":        "config file contents or URI reference",
	"scalewithcluster":        "Option to run App on all nodes of the cluster",
	"internalports":           "Should this app have access to outside world?",
	"revision":                "Revision can be specified or defaults to current timestamp when app is updated",
	"officialfqdn":            "Official FQDN is the FQDN that the app uses to connect by default",
	"md5sum":                  "MD5Sum of the VM-based app image",
	"defaultsharedvolumesize": "shared volume size when creating auto cluster",
	"autoprovpolicy":          "(_deprecated_) Auto provisioning policy name",
	"accesstype":              "Access type, one of AccessTypeDefaultForDeployment, AccessTypeDirect, AccessTypeLoadBalancer",
	"defaultprivacypolicy":    "Privacy policy when creating auto cluster",
	"deleteprepare":           "Preparing to be deleted",
	"autoprovpolicies":        "Auto provisioning policy names",
	"templatedelimiter":       "Delimiter to be used for template parsing, defaults to [[ ]]",
	"skiphcports":             "Comma separated list of protocol:port pairs that we should not run health check on Should be configured in case app does not always listen on these ports all can be specified if no health check to be run for this app Numerical values must be decimal format. i.e. tcp:80,udp:10002,http:443",
}
var AppSpecialArgs = map[string]string{
	"app.autoprovpolicies": "StringArray",
	"app.fields":           "StringArray",
}
var AppAutoProvPolicyRequiredArgs = []string{}
var AppAutoProvPolicyOptionalArgs = []string{
	"appkey.organization",
	"appkey.name",
	"appkey.version",
	"autoprovpolicy",
}
var AppAutoProvPolicyAliasArgs = []string{
	"appkey.organization=appautoprovpolicy.appkey.organization",
	"appkey.name=appautoprovpolicy.appkey.name",
	"appkey.version=appautoprovpolicy.appkey.version",
	"autoprovpolicy=appautoprovpolicy.autoprovpolicy",
}
var AppAutoProvPolicyComments = map[string]string{
	"appkey.organization": "App developer organization",
	"appkey.name":         "App name",
	"appkey.version":      "App version",
	"autoprovpolicy":      "Auto provisioning policy name",
}
var AppAutoProvPolicySpecialArgs = map[string]string{}
