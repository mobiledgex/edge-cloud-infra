// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stream.proto

package ormapi

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

// Request summary for StreamAppInst
// swagger:parameters StreamAppInst
type swaggerStreamAppInst struct {
	// in: body
	Body RegionAppInstKey
}

type RegionAppInstKey struct {
	// Region name
	// required: true
	Region string
	// AppInstKey in region
	AppInstKey edgeproto.AppInstKey
}

func (s *RegionAppInstKey) GetRegion() string {
	return s.Region
}

func (s *RegionAppInstKey) GetObj() interface{} {
	return &s.AppInstKey
}

func (s *RegionAppInstKey) GetObjName() string {
	return "AppInstKey"
}

// Request summary for StreamClusterInst
// swagger:parameters StreamClusterInst
type swaggerStreamClusterInst struct {
	// in: body
	Body RegionClusterInstKey
}

// Request summary for StreamCloudlet
// swagger:parameters StreamCloudlet
type swaggerStreamCloudlet struct {
	// in: body
	Body RegionCloudletKey
}

// Request summary for StreamGPUDriver
// swagger:parameters StreamGPUDriver
type swaggerStreamGPUDriver struct {
	// in: body
	Body RegionGPUDriverKey
}
