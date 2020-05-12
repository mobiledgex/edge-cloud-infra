// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package ormapi

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
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

// Request summary for ShowCloudletManifest
// swagger:parameters ShowCloudletManifest
type swaggerShowCloudletManifest struct {
	// in: body
	Body RegionCloudlet
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
