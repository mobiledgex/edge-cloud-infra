// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package ormapi

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

// Request summary for CreateGPUDriver
// swagger:parameters CreateGPUDriver
type swaggerCreateGPUDriver struct {
	// in: body
	Body RegionGPUDriver
}

type RegionGPUDriver struct {
	// required: true
	// Region name
	Region    string
	GPUDriver edgeproto.GPUDriver
}

// Request summary for DeleteGPUDriver
// swagger:parameters DeleteGPUDriver
type swaggerDeleteGPUDriver struct {
	// in: body
	Body RegionGPUDriver
}

// Request summary for UpdateGPUDriver
// swagger:parameters UpdateGPUDriver
type swaggerUpdateGPUDriver struct {
	// in: body
	Body RegionGPUDriver
}

// Request summary for ShowGPUDriver
// swagger:parameters ShowGPUDriver
type swaggerShowGPUDriver struct {
	// in: body
	Body RegionGPUDriver
}

// Request summary for AddGPUDriverBuild
// swagger:parameters AddGPUDriverBuild
type swaggerAddGPUDriverBuild struct {
	// in: body
	Body RegionGPUDriverBuildMember
}

type RegionGPUDriverBuildMember struct {
	// required: true
	// Region name
	Region               string
	GPUDriverBuildMember edgeproto.GPUDriverBuildMember
}

// Request summary for RemoveGPUDriverBuild
// swagger:parameters RemoveGPUDriverBuild
type swaggerRemoveGPUDriverBuild struct {
	// in: body
	Body RegionGPUDriverBuildMember
}

// Request summary for GetGPUDriverBuildURL
// swagger:parameters GetGPUDriverBuildURL
type swaggerGetGPUDriverBuildURL struct {
	// in: body
	Body RegionGPUDriverBuildMember
}

// Request summary for CreateCloudlet
// swagger:parameters CreateCloudlet
type swaggerCreateCloudlet struct {
	// in: body
	Body RegionCloudlet
}

type RegionCloudlet struct {
	// required: true
	// Region name
	Region   string
	Cloudlet edgeproto.Cloudlet
}

// Request summary for DeleteCloudlet
// swagger:parameters DeleteCloudlet
type swaggerDeleteCloudlet struct {
	// in: body
	Body RegionCloudlet
}

// Request summary for UpdateCloudlet
// swagger:parameters UpdateCloudlet
type swaggerUpdateCloudlet struct {
	// in: body
	Body RegionCloudlet
}

// Request summary for ShowCloudlet
// swagger:parameters ShowCloudlet
type swaggerShowCloudlet struct {
	// in: body
	Body RegionCloudlet
}

// Request summary for GetCloudletManifest
// swagger:parameters GetCloudletManifest
type swaggerGetCloudletManifest struct {
	// in: body
	Body RegionCloudletKey
}

type RegionCloudletKey struct {
	// required: true
	// Region name
	Region      string
	CloudletKey edgeproto.CloudletKey
}

// Request summary for GetCloudletProps
// swagger:parameters GetCloudletProps
type swaggerGetCloudletProps struct {
	// in: body
	Body RegionCloudletProps
}

type RegionCloudletProps struct {
	// required: true
	// Region name
	Region        string
	CloudletProps edgeproto.CloudletProps
}

// Request summary for GetCloudletResourceQuotaProps
// swagger:parameters GetCloudletResourceQuotaProps
type swaggerGetCloudletResourceQuotaProps struct {
	// in: body
	Body RegionCloudletResourceQuotaProps
}

type RegionCloudletResourceQuotaProps struct {
	// required: true
	// Region name
	Region                     string
	CloudletResourceQuotaProps edgeproto.CloudletResourceQuotaProps
}

// Request summary for GetCloudletResourceUsage
// swagger:parameters GetCloudletResourceUsage
type swaggerGetCloudletResourceUsage struct {
	// in: body
	Body RegionCloudletResourceUsage
}

type RegionCloudletResourceUsage struct {
	// required: true
	// Region name
	Region                string
	CloudletResourceUsage edgeproto.CloudletResourceUsage
}

// Request summary for AddCloudletResMapping
// swagger:parameters AddCloudletResMapping
type swaggerAddCloudletResMapping struct {
	// in: body
	Body RegionCloudletResMap
}

type RegionCloudletResMap struct {
	// required: true
	// Region name
	Region         string
	CloudletResMap edgeproto.CloudletResMap
}

// Request summary for RemoveCloudletResMapping
// swagger:parameters RemoveCloudletResMapping
type swaggerRemoveCloudletResMapping struct {
	// in: body
	Body RegionCloudletResMap
}

// Request summary for FindFlavorMatch
// swagger:parameters FindFlavorMatch
type swaggerFindFlavorMatch struct {
	// in: body
	Body RegionFlavorMatch
}

type RegionFlavorMatch struct {
	// required: true
	// Region name
	Region      string
	FlavorMatch edgeproto.FlavorMatch
}

// Request summary for ShowFlavorsForCloudlet
// swagger:parameters ShowFlavorsForCloudlet
type swaggerShowFlavorsForCloudlet struct {
	// in: body
	Body RegionCloudletKey
}

// Request summary for RevokeAccessKey
// swagger:parameters RevokeAccessKey
type swaggerRevokeAccessKey struct {
	// in: body
	Body RegionCloudletKey
}

// Request summary for GenerateAccessKey
// swagger:parameters GenerateAccessKey
type swaggerGenerateAccessKey struct {
	// in: body
	Body RegionCloudletKey
}

// Request summary for ShowCloudletInfo
// swagger:parameters ShowCloudletInfo
type swaggerShowCloudletInfo struct {
	// in: body
	Body RegionCloudletInfo
}

type RegionCloudletInfo struct {
	// required: true
	// Region name
	Region       string
	CloudletInfo edgeproto.CloudletInfo
}

// Request summary for InjectCloudletInfo
// swagger:parameters InjectCloudletInfo
type swaggerInjectCloudletInfo struct {
	// in: body
	Body RegionCloudletInfo
}

// Request summary for EvictCloudletInfo
// swagger:parameters EvictCloudletInfo
type swaggerEvictCloudletInfo struct {
	// in: body
	Body RegionCloudletInfo
}
