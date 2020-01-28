// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: privacypolicy.proto

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

// Request summary for CreatePrivacyPolicy
// swagger:parameters CreatePrivacyPolicy
type swaggerCreatePrivacyPolicy struct {
	// in: body
	Body RegionPrivacyPolicy
}
type RegionPrivacyPolicy struct {
	// required: true
	// Region name
	Region        string
	PrivacyPolicy edgeproto.PrivacyPolicy
}

// Request summary for DeletePrivacyPolicy
// swagger:parameters DeletePrivacyPolicy
type swaggerDeletePrivacyPolicy struct {
	// in: body
	Body RegionPrivacyPolicy
}

// Request summary for UpdatePrivacyPolicy
// swagger:parameters UpdatePrivacyPolicy
type swaggerUpdatePrivacyPolicy struct {
	// in: body
	Body RegionPrivacyPolicy
}

// Request summary for ShowPrivacyPolicy
// swagger:parameters ShowPrivacyPolicy
type swaggerShowPrivacyPolicy struct {
	// in: body
	Body RegionPrivacyPolicy
}
