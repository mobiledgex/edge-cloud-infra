// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoscalepolicy.proto

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

// Request summary for CreateAutoScalePolicy
// swagger:parameters CreateAutoScalePolicy
type swaggerCreateAutoScalePolicy struct {
	// in: body
	Body RegionAutoScalePolicy
}

type RegionAutoScalePolicy struct {
	// Region name
	// required: true
	Region string
	// AutoScalePolicy in region
	AutoScalePolicy edgeproto.AutoScalePolicy
}

func (s *RegionAutoScalePolicy) GetRegion() string {
	return s.Region
}

func (s *RegionAutoScalePolicy) GetObj() interface{} {
	return &s.AutoScalePolicy
}

func (s *RegionAutoScalePolicy) GetObjName() string {
	return "AutoScalePolicy"
}
func (s *RegionAutoScalePolicy) GetObjFields() []string {
	return s.AutoScalePolicy.Fields
}

func (s *RegionAutoScalePolicy) SetObjFields(fields []string) {
	s.AutoScalePolicy.Fields = fields
}

// Request summary for DeleteAutoScalePolicy
// swagger:parameters DeleteAutoScalePolicy
type swaggerDeleteAutoScalePolicy struct {
	// in: body
	Body RegionAutoScalePolicy
}

// Request summary for UpdateAutoScalePolicy
// swagger:parameters UpdateAutoScalePolicy
type swaggerUpdateAutoScalePolicy struct {
	// in: body
	Body RegionAutoScalePolicy
}

// Request summary for ShowAutoScalePolicy
// swagger:parameters ShowAutoScalePolicy
type swaggerShowAutoScalePolicy struct {
	// in: body
	Body RegionAutoScalePolicy
}
