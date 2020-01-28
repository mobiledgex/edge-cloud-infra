// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: restagtable.proto

package ormapi

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
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

// Request summary for CreateResTagTable
// swagger:parameters CreateResTagTable
type swaggerCreateResTagTable struct {
	// in: body
	Body RegionResTagTable
}
type RegionResTagTable struct {
	// required: true
	// Region name
	Region      string
	ResTagTable edgeproto.ResTagTable
}

// Request summary for DeleteResTagTable
// swagger:parameters DeleteResTagTable
type swaggerDeleteResTagTable struct {
	// in: body
	Body RegionResTagTable
}

// Request summary for UpdateResTagTable
// swagger:parameters UpdateResTagTable
type swaggerUpdateResTagTable struct {
	// in: body
	Body RegionResTagTable
}

// Request summary for ShowResTagTable
// swagger:parameters ShowResTagTable
type swaggerShowResTagTable struct {
	// in: body
	Body RegionResTagTable
}

// Request summary for AddResTag
// swagger:parameters AddResTag
type swaggerAddResTag struct {
	// in: body
	Body RegionResTagTable
}

// Request summary for RemoveResTag
// swagger:parameters RemoveResTag
type swaggerRemoveResTag struct {
	// in: body
	Body RegionResTagTable
}

// Request summary for GetResTagTable
// swagger:parameters GetResTagTable
type swaggerGetResTagTable struct {
	// in: body
	Body RegionResTagTableKey
}
type RegionResTagTableKey struct {
	// required: true
	// Region name
	Region         string
	ResTagTableKey edgeproto.ResTagTableKey
}
