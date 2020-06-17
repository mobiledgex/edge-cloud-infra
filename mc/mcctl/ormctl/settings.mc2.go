// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: settings.proto

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
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var UpdateSettingsCmd = &cli.Command{
	Use:          "UpdateSettings",
	RequiredArgs: "region " + strings.Join(SettingsRequiredArgs, " "),
	OptionalArgs: strings.Join(SettingsOptionalArgs, " "),
	AliasArgs:    strings.Join(SettingsAliasArgs, " "),
	SpecialArgs:  &SettingsSpecialArgs,
	Comments:     addRegionComment(SettingsComments),
	ReqData:      &ormapi.RegionSettings{},
	ReplyData:    &edgeproto.Result{},
	Run: runRest("/auth/ctrl/UpdateSettings",
		withSetFieldsFunc(setUpdateSettingsFields),
	),
}

func setUpdateSettingsFields(in map[string]interface{}) {
	// get map for edgeproto object in region struct
	obj := in[strings.ToLower("Settings")]
	if obj == nil {
		return
	}
	objmap, ok := obj.(map[string]interface{})
	if !ok {
		return
	}
	fields := cli.GetSpecifiedFields(objmap, &edgeproto.Settings{}, cli.JsonNamespace)
	// include fields already specified
	if inFields, found := objmap["fields"]; found {
		if fieldsArr, ok := inFields.([]string); ok {
			fields = append(fields, fieldsArr...)
		}
	}
	objmap["fields"] = fields
}

var ResetSettingsCmd = &cli.Command{
	Use:          "ResetSettings",
	RequiredArgs: "region " + strings.Join(SettingsRequiredArgs, " "),
	OptionalArgs: strings.Join(SettingsOptionalArgs, " "),
	AliasArgs:    strings.Join(SettingsAliasArgs, " "),
	SpecialArgs:  &SettingsSpecialArgs,
	Comments:     addRegionComment(SettingsComments),
	ReqData:      &ormapi.RegionSettings{},
	ReplyData:    &edgeproto.Result{},
	Run:          runRest("/auth/ctrl/ResetSettings"),
}

var ShowSettingsCmd = &cli.Command{
	Use:          "ShowSettings",
	RequiredArgs: "region " + strings.Join(SettingsRequiredArgs, " "),
	OptionalArgs: strings.Join(SettingsOptionalArgs, " "),
	AliasArgs:    strings.Join(SettingsAliasArgs, " "),
	SpecialArgs:  &SettingsSpecialArgs,
	Comments:     addRegionComment(SettingsComments),
	ReqData:      &ormapi.RegionSettings{},
	ReplyData:    &edgeproto.Settings{},
	Run:          runRest("/auth/ctrl/ShowSettings"),
}

var SettingsApiCmds = []*cli.Command{
	UpdateSettingsCmd,
	ResetSettingsCmd,
	ShowSettingsCmd,
}

var SettingsRequiredArgs = []string{}
var SettingsOptionalArgs = []string{
	"shepherdmetricscollectioninterval",
	"shepherdhealthcheckretries",
	"shepherdhealthcheckinterval",
	"autodeployintervalsec",
	"autodeployoffsetsec",
	"autodeploymaxintervals",
	"createappinsttimeout",
	"updateappinsttimeout",
	"deleteappinsttimeout",
	"createclusterinsttimeout",
	"updateclusterinsttimeout",
	"deleteclusterinsttimeout",
	"masternodeflavor",
	"loadbalancermaxportrange",
	"maxtrackeddmeclients",
	"chefclientinterval",
	"influxdbmetricsretention",
	"cloudletmaintenancetimeout",
}
var SettingsAliasArgs = []string{
	"fields=settings.fields",
	"shepherdmetricscollectioninterval=settings.shepherdmetricscollectioninterval",
	"shepherdhealthcheckretries=settings.shepherdhealthcheckretries",
	"shepherdhealthcheckinterval=settings.shepherdhealthcheckinterval",
	"autodeployintervalsec=settings.autodeployintervalsec",
	"autodeployoffsetsec=settings.autodeployoffsetsec",
	"autodeploymaxintervals=settings.autodeploymaxintervals",
	"createappinsttimeout=settings.createappinsttimeout",
	"updateappinsttimeout=settings.updateappinsttimeout",
	"deleteappinsttimeout=settings.deleteappinsttimeout",
	"createclusterinsttimeout=settings.createclusterinsttimeout",
	"updateclusterinsttimeout=settings.updateclusterinsttimeout",
	"deleteclusterinsttimeout=settings.deleteclusterinsttimeout",
	"masternodeflavor=settings.masternodeflavor",
	"loadbalancermaxportrange=settings.loadbalancermaxportrange",
	"maxtrackeddmeclients=settings.maxtrackeddmeclients",
	"chefclientinterval=settings.chefclientinterval",
	"influxdbmetricsretention=settings.influxdbmetricsretention",
	"cloudletmaintenancetimeout=settings.cloudletmaintenancetimeout",
}
var SettingsComments = map[string]string{
	"fields":                            "Fields are used for the Update API to specify which fields to apply",
	"shepherdmetricscollectioninterval": "Shepherd metrics collection interval for k8s and docker appInstances (duration)",
	"shepherdhealthcheckretries":        "Number of times Shepherd Health Check fails before we mark appInst down",
	"shepherdhealthcheckinterval":       "Health Checking probing frequency (duration)",
	"autodeployintervalsec":             "Auto Provisioning Stats push and analysis interval (seconds)",
	"autodeployoffsetsec":               "Auto Provisioning analysis offset from interval (seconds)",
	"autodeploymaxintervals":            "Auto Provisioning Policy max allowed intervals",
	"createappinsttimeout":              "Create AppInst timeout (duration)",
	"updateappinsttimeout":              "Update AppInst timeout (duration)",
	"deleteappinsttimeout":              "Delete AppInst timeout (duration)",
	"createclusterinsttimeout":          "Create ClusterInst timeout (duration)",
	"updateclusterinsttimeout":          "Update ClusterInst timeout (duration)",
	"deleteclusterinsttimeout":          "Delete ClusterInst timeout (duration)",
	"masternodeflavor":                  "Default flavor for k8s master VM and > 0  workers",
	"loadbalancermaxportrange":          "Max IP Port range when using a load balancer",
	"maxtrackeddmeclients":              "Max DME clients to be tracked at the same time.",
	"chefclientinterval":                "Default chef client interval (duration)",
	"influxdbmetricsretention":          "Default influxDB metrics retention policy (duration)",
	"cloudletmaintenancetimeout":        "Default Cloudlet Maintenance timeout (used twice for AutoProv and Cloudlet)",
}
var SettingsSpecialArgs = map[string]string{
	"settings.fields": "StringArray",
}
