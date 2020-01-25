// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: settings.proto

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

// Request summary for UpdateSettings
// swagger:parameters UpdateSettings
type swaggerUpdateSettings struct {
	// in: body
	Body RegionSettings
}
type RegionSettings struct {
	// Region name
	Region   string
	Settings edgeproto.Settings
}

// Request summary for ResetSettings
// swagger:parameters ResetSettings
type swaggerResetSettings struct {
	// in: body
	Body RegionSettings
}

// Request summary for ShowSettings
// swagger:parameters ShowSettings
type swaggerShowSettings struct {
	// in: body
	Body RegionSettings
}
