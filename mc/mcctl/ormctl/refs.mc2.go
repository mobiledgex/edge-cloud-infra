// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: refs.proto

package ormctl

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
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

var ShowCloudletRefsCmd = &ApiCommand{
	Name:         "ShowCloudletRefs",
	Use:          "show",
	Short:        "Show CloudletRefs (debug only)",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(CloudletRefsRequiredArgs, CloudletRefsOptionalArgs...), " "),
	AliasArgs:    strings.Join(CloudletRefsAliasArgs, " "),
	SpecialArgs:  &CloudletRefsSpecialArgs,
	Comments:     addRegionComment(CloudletRefsComments),
	ReqData:      &ormapi.RegionCloudletRefs{},
	ReplyData:    &edgeproto.CloudletRefs{},
	Path:         "/auth/ctrl/ShowCloudletRefs",
	StreamOut:    true,
	ProtobufApi:  true,
}
var CloudletRefsApiCmds = []*ApiCommand{
	ShowCloudletRefsCmd,
}

const CloudletRefsGroup = "CloudletRefs"

func init() {
	AllApis.AddGroup(CloudletRefsGroup, "Manage CloudletRefs", CloudletRefsApiCmds)
}

var ShowClusterRefsCmd = &ApiCommand{
	Name:         "ShowClusterRefs",
	Use:          "show",
	Short:        "Show ClusterRefs (debug only)",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(ClusterRefsRequiredArgs, ClusterRefsOptionalArgs...), " "),
	AliasArgs:    strings.Join(ClusterRefsAliasArgs, " "),
	SpecialArgs:  &ClusterRefsSpecialArgs,
	Comments:     addRegionComment(ClusterRefsComments),
	ReqData:      &ormapi.RegionClusterRefs{},
	ReplyData:    &edgeproto.ClusterRefs{},
	Path:         "/auth/ctrl/ShowClusterRefs",
	StreamOut:    true,
	ProtobufApi:  true,
}
var ClusterRefsApiCmds = []*ApiCommand{
	ShowClusterRefsCmd,
}

const ClusterRefsGroup = "ClusterRefs"

func init() {
	AllApis.AddGroup(ClusterRefsGroup, "Manage ClusterRefs", ClusterRefsApiCmds)
}

var ShowAppInstRefsCmd = &ApiCommand{
	Name:         "ShowAppInstRefs",
	Use:          "show",
	Short:        "Show AppInstRefs (debug only)",
	RequiredArgs: "region",
	OptionalArgs: strings.Join(append(AppInstRefsRequiredArgs, AppInstRefsOptionalArgs...), " "),
	AliasArgs:    strings.Join(AppInstRefsAliasArgs, " "),
	SpecialArgs:  &AppInstRefsSpecialArgs,
	Comments:     addRegionComment(AppInstRefsComments),
	ReqData:      &ormapi.RegionAppInstRefs{},
	ReplyData:    &edgeproto.AppInstRefs{},
	Path:         "/auth/ctrl/ShowAppInstRefs",
	StreamOut:    true,
	ProtobufApi:  true,
}
var AppInstRefsApiCmds = []*ApiCommand{
	ShowAppInstRefsCmd,
}

const AppInstRefsGroup = "AppInstRefs"

func init() {
	AllApis.AddGroup(AppInstRefsGroup, "Manage AppInstRefs", AppInstRefsApiCmds)
}

var CloudletRefsRequiredArgs = []string{
	"key.organization",
	"key.name",
}
var CloudletRefsOptionalArgs = []string{
	"rootlbports:#.key",
	"rootlbports:#.value",
	"useddynamicips",
	"usedstaticips",
	"optresusedmap:#.key",
	"optresusedmap:#.value",
	"reservedautoclusterids",
	"clusterinsts:#.clusterkey.name",
	"clusterinsts:#.organization",
	"vmappinsts:#.appkey.organization",
	"vmappinsts:#.appkey.name",
	"vmappinsts:#.appkey.version",
	"vmappinsts:#.clusterinstkey.clusterkey.name",
	"vmappinsts:#.clusterinstkey.organization",
}
var CloudletRefsAliasArgs = []string{
	"key.organization=cloudletrefs.key.organization",
	"key.name=cloudletrefs.key.name",
	"rootlbports:#.key=cloudletrefs.rootlbports:#.key",
	"rootlbports:#.value=cloudletrefs.rootlbports:#.value",
	"useddynamicips=cloudletrefs.useddynamicips",
	"usedstaticips=cloudletrefs.usedstaticips",
	"optresusedmap:#.key=cloudletrefs.optresusedmap:#.key",
	"optresusedmap:#.value=cloudletrefs.optresusedmap:#.value",
	"reservedautoclusterids=cloudletrefs.reservedautoclusterids",
	"clusterinsts:#.clusterkey.name=cloudletrefs.clusterinsts:#.clusterkey.name",
	"clusterinsts:#.organization=cloudletrefs.clusterinsts:#.organization",
	"vmappinsts:#.appkey.organization=cloudletrefs.vmappinsts:#.appkey.organization",
	"vmappinsts:#.appkey.name=cloudletrefs.vmappinsts:#.appkey.name",
	"vmappinsts:#.appkey.version=cloudletrefs.vmappinsts:#.appkey.version",
	"vmappinsts:#.clusterinstkey.clusterkey.name=cloudletrefs.vmappinsts:#.clusterinstkey.clusterkey.name",
	"vmappinsts:#.clusterinstkey.organization=cloudletrefs.vmappinsts:#.clusterinstkey.organization",
}
var CloudletRefsComments = map[string]string{
	"key.organization":                            "Organization of the cloudlet site",
	"key.name":                                    "Name of the cloudlet",
	"useddynamicips":                              "Used dynamic IPs",
	"usedstaticips":                               "Used static IPs",
	"reservedautoclusterids":                      "Track reservable autoclusterinsts ids in use. This is a bitmap.",
	"clusterinsts:#.clusterkey.name":              "Cluster name",
	"clusterinsts:#.organization":                 "Name of Developer organization that this cluster belongs to",
	"vmappinsts:#.appkey.organization":            "App developer organization",
	"vmappinsts:#.appkey.name":                    "App name",
	"vmappinsts:#.appkey.version":                 "App version",
	"vmappinsts:#.clusterinstkey.clusterkey.name": "Cluster name",
	"vmappinsts:#.clusterinstkey.organization":    "Name of Developer organization that this cluster belongs to",
}
var CloudletRefsSpecialArgs = map[string]string{}
var ClusterRefsRequiredArgs = []string{
	"key.clusterkey.name",
	"key.cloudletkey.organization",
	"key.cloudletkey.name",
	"key.organization",
}
var ClusterRefsOptionalArgs = []string{
	"apps:#.organization",
	"apps:#.name",
	"apps:#.version",
	"usedram",
	"usedvcores",
	"useddisk",
}
var ClusterRefsAliasArgs = []string{
	"key.clusterkey.name=clusterrefs.key.clusterkey.name",
	"key.cloudletkey.organization=clusterrefs.key.cloudletkey.organization",
	"key.cloudletkey.name=clusterrefs.key.cloudletkey.name",
	"key.organization=clusterrefs.key.organization",
	"apps:#.organization=clusterrefs.apps:#.organization",
	"apps:#.name=clusterrefs.apps:#.name",
	"apps:#.version=clusterrefs.apps:#.version",
	"usedram=clusterrefs.usedram",
	"usedvcores=clusterrefs.usedvcores",
	"useddisk=clusterrefs.useddisk",
}
var ClusterRefsComments = map[string]string{
	"key.clusterkey.name":          "Cluster name",
	"key.cloudletkey.organization": "Organization of the cloudlet site",
	"key.cloudletkey.name":         "Name of the cloudlet",
	"key.organization":             "Name of Developer organization that this cluster belongs to",
	"apps:#.organization":          "App developer organization",
	"apps:#.name":                  "App name",
	"apps:#.version":               "App version",
	"usedram":                      "Used RAM in MB",
	"usedvcores":                   "Used VCPU cores",
	"useddisk":                     "Used disk in GB",
}
var ClusterRefsSpecialArgs = map[string]string{}
var AppInstRefsRequiredArgs = []string{
	"key.organization",
	"key.name",
	"key.version",
}
var AppInstRefsOptionalArgs = []string{
	"insts:#.key",
	"insts:#.value",
}
var AppInstRefsAliasArgs = []string{
	"key.organization=appinstrefs.key.organization",
	"key.name=appinstrefs.key.name",
	"key.version=appinstrefs.key.version",
	"insts:#.key=appinstrefs.insts:#.key",
	"insts:#.value=appinstrefs.insts:#.value",
}
var AppInstRefsComments = map[string]string{
	"key.organization": "App developer organization",
	"key.name":         "App name",
	"key.version":      "App version",
}
var AppInstRefsSpecialArgs = map[string]string{}
