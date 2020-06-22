// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: refs.proto

package ormapi

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

// Request summary for ShowCloudletRefs
// swagger:parameters ShowCloudletRefs
type swaggerShowCloudletRefs struct {
	// in: body
	Body RegionCloudletRefs
}

type RegionCloudletRefs struct {
	// required: true
	// Region name
	Region       string
	CloudletRefs edgeproto.CloudletRefs
}

// Request summary for ShowClusterRefs
// swagger:parameters ShowClusterRefs
type swaggerShowClusterRefs struct {
	// in: body
	Body RegionClusterRefs
}

type RegionClusterRefs struct {
	// required: true
	// Region name
	Region      string
	ClusterRefs edgeproto.ClusterRefs
}

// Request summary for ShowAppInstRefs
// swagger:parameters ShowAppInstRefs
type swaggerShowAppInstRefs struct {
	// in: body
	Body RegionAppInstRefs
}

type RegionAppInstRefs struct {
	// required: true
	// Region name
	Region      string
	AppInstRefs edgeproto.AppInstRefs
}
