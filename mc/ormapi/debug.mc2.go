// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: debug.proto

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

// Request summary for EnableDebugLevels
// swagger:parameters EnableDebugLevels
type swaggerEnableDebugLevels struct {
	// in: body
	Body RegionDebugRequest
}

type RegionDebugRequest struct {
	// Region name
	// required: true
	Region string
	// DebugRequest in region
	DebugRequest edgeproto.DebugRequest
}

func (s *RegionDebugRequest) GetRegion() string {
	return s.Region
}

func (s *RegionDebugRequest) GetObj() interface{} {
	return &s.DebugRequest
}

func (s *RegionDebugRequest) GetObjName() string {
	return "DebugRequest"
}

// Request summary for DisableDebugLevels
// swagger:parameters DisableDebugLevels
type swaggerDisableDebugLevels struct {
	// in: body
	Body RegionDebugRequest
}

// Request summary for ShowDebugLevels
// swagger:parameters ShowDebugLevels
type swaggerShowDebugLevels struct {
	// in: body
	Body RegionDebugRequest
}

// Request summary for RunDebug
// swagger:parameters RunDebug
type swaggerRunDebug struct {
	// in: body
	Body RegionDebugRequest
}
