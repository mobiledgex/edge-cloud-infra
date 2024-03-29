// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

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

// Request summary for CreateClusterInst
// swagger:parameters CreateClusterInst
type swaggerCreateClusterInst struct {
	// in: body
	Body RegionClusterInst
}

type RegionClusterInst struct {
	// Region name
	// required: true
	Region string
	// ClusterInst in region
	ClusterInst edgeproto.ClusterInst
}

func (s *RegionClusterInst) GetRegion() string {
	return s.Region
}

func (s *RegionClusterInst) GetObj() interface{} {
	return &s.ClusterInst
}

func (s *RegionClusterInst) GetObjName() string {
	return "ClusterInst"
}
func (s *RegionClusterInst) GetObjFields() []string {
	return s.ClusterInst.Fields
}

func (s *RegionClusterInst) SetObjFields(fields []string) {
	s.ClusterInst.Fields = fields
}

// Request summary for DeleteClusterInst
// swagger:parameters DeleteClusterInst
type swaggerDeleteClusterInst struct {
	// in: body
	Body RegionClusterInst
}

// Request summary for UpdateClusterInst
// swagger:parameters UpdateClusterInst
type swaggerUpdateClusterInst struct {
	// in: body
	Body RegionClusterInst
}

// Request summary for ShowClusterInst
// swagger:parameters ShowClusterInst
type swaggerShowClusterInst struct {
	// in: body
	Body RegionClusterInst
}

// Request summary for DeleteIdleReservableClusterInsts
// swagger:parameters DeleteIdleReservableClusterInsts
type swaggerDeleteIdleReservableClusterInsts struct {
	// in: body
	Body RegionIdleReservableClusterInsts
}

type RegionIdleReservableClusterInsts struct {
	// Region name
	// required: true
	Region string
	// IdleReservableClusterInsts in region
	IdleReservableClusterInsts edgeproto.IdleReservableClusterInsts
}

func (s *RegionIdleReservableClusterInsts) GetRegion() string {
	return s.Region
}

func (s *RegionIdleReservableClusterInsts) GetObj() interface{} {
	return &s.IdleReservableClusterInsts
}

func (s *RegionIdleReservableClusterInsts) GetObjName() string {
	return "IdleReservableClusterInsts"
}
