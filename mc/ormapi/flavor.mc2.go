// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flavor.proto

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

// Request summary for CreateFlavor
// swagger:parameters CreateFlavor
type swaggerCreateFlavor struct {
	// in: body
	Body RegionFlavor
}
type RegionFlavor struct {
	// required: true
	// Region name
	Region string
	Flavor edgeproto.Flavor
}

// Request summary for DeleteFlavor
// swagger:parameters DeleteFlavor
type swaggerDeleteFlavor struct {
	// in: body
	Body RegionFlavor
}

// Request summary for UpdateFlavor
// swagger:parameters UpdateFlavor
type swaggerUpdateFlavor struct {
	// in: body
	Body RegionFlavor
}

// Request summary for ShowFlavor
// swagger:parameters ShowFlavor
type swaggerShowFlavor struct {
	// in: body
	Body RegionFlavor
}

// Request summary for AddFlavorRes
// swagger:parameters AddFlavorRes
type swaggerAddFlavorRes struct {
	// in: body
	Body RegionFlavor
}

// Request summary for RemoveFlavorRes
// swagger:parameters RemoveFlavorRes
type swaggerRemoveFlavorRes struct {
	// in: body
	Body RegionFlavor
}
