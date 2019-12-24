// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package ormapi

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

type RegionCloudlet struct {
	Region   string             `json:"region"`
	Cloudlet edgeproto.Cloudlet `json:"cloudlet"`
}

type RegionCloudletResMap struct {
	Region         string                   `json:"region"`
	CloudletResMap edgeproto.CloudletResMap `json:"cloudletresmap"`
}

type RegionFlavorMatch struct {
	Region      string                `json:"region"`
	FlavorMatch edgeproto.FlavorMatch `json:"flavormatch"`
}

type RegionCloudletInfo struct {
	Region       string                 `json:"region"`
	CloudletInfo edgeproto.CloudletInfo `json:"cloudletinfo"`
}
