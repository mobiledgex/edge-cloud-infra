// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

package ormctl

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var CreateClusterInstCmd = &Command{
	Use:                  "CreateClusterInst",
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Path:                 "/auth/ctrl/CreateClusterInst",
	OptionalArgs:         "region",
	StreamOut:            true,
	StreamOutIncremental: true,
}

var DeleteClusterInstCmd = &Command{
	Use:                  "DeleteClusterInst",
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Path:                 "/auth/ctrl/DeleteClusterInst",
	OptionalArgs:         "region",
	StreamOut:            true,
	StreamOutIncremental: true,
}

var UpdateClusterInstCmd = &Command{
	Use:                  "UpdateClusterInst",
	ReqData:              &ormapi.RegionClusterInst{},
	ReplyData:            &edgeproto.Result{},
	Path:                 "/auth/ctrl/UpdateClusterInst",
	OptionalArgs:         "region",
	StreamOut:            true,
	StreamOutIncremental: true,
}

var ShowClusterInstCmd = &Command{
	Use:          "ShowClusterInst",
	ReqData:      &ormapi.RegionClusterInst{},
	ReplyData:    &edgeproto.ClusterInst{},
	Path:         "/auth/ctrl/ShowClusterInst",
	OptionalArgs: "region",
	StreamOut:    true,
}
var ClusterInstApiCmds = []*Command{
	CreateClusterInstCmd,
	DeleteClusterInstCmd,
	UpdateClusterInstCmd,
	ShowClusterInstCmd,
}
