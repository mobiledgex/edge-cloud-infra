// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: vmpool.proto

package ormapi

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
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

// Request summary for CreateVMPool
// swagger:parameters CreateVMPool
type swaggerCreateVMPool struct {
	// in: body
	Body RegionVMPool
}

type RegionVMPool struct {
	// required: true
	// Region name
	Region string
	VMPool edgeproto.VMPool
}

// Request summary for DeleteVMPool
// swagger:parameters DeleteVMPool
type swaggerDeleteVMPool struct {
	// in: body
	Body RegionVMPool
}

// Request summary for UpdateVMPool
// swagger:parameters UpdateVMPool
type swaggerUpdateVMPool struct {
	// in: body
	Body RegionVMPool
}

// Request summary for ShowVMPool
// swagger:parameters ShowVMPool
type swaggerShowVMPool struct {
	// in: body
	Body RegionVMPool
}

// Request summary for AddVMPoolMember
// swagger:parameters AddVMPoolMember
type swaggerAddVMPoolMember struct {
	// in: body
	Body RegionVMPoolMember
}

type RegionVMPoolMember struct {
	// required: true
	// Region name
	Region       string
	VMPoolMember edgeproto.VMPoolMember
}

// Request summary for RemoveVMPoolMember
// swagger:parameters RemoveVMPoolMember
type swaggerRemoveVMPoolMember struct {
	// in: body
	Body RegionVMPoolMember
}
