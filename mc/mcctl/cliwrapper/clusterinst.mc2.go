// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

package cliwrapper

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
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

func (s *Client) CreateClusterInst(uri, token string, in *ormapi.RegionClusterInst) ([]edgeproto.Result, int, error) {
	args := []string{"region", "CreateClusterInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Liveness,Auto,MasterNodeFlavor,NodeFlavor,ExternalVolumeSize,AllocatedIp,Status,ReservedBy", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) DeleteClusterInst(uri, token string, in *ormapi.RegionClusterInst) ([]edgeproto.Result, int, error) {
	args := []string{"region", "DeleteClusterInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Liveness,Auto,MasterNodeFlavor,NodeFlavor,ExternalVolumeSize,AllocatedIp,Status,ReservedBy", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) UpdateClusterInst(uri, token string, in *ormapi.RegionClusterInst) ([]edgeproto.Result, int, error) {
	args := []string{"region", "UpdateClusterInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Liveness,Auto,MasterNodeFlavor,NodeFlavor,ExternalVolumeSize,AllocatedIp,Status,ReservedBy,Flavor,NumMasters,AvailabilityZone,Reservable,SharedVolumeSize,PrivacyPolicy", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowClusterInst(uri, token string, in *ormapi.RegionClusterInst) ([]edgeproto.ClusterInst, int, error) {
	args := []string{"region", "ShowClusterInst"}
	outlist := []edgeproto.ClusterInst{}
	noconfig := strings.Split("Liveness,Auto,MasterNodeFlavor,NodeFlavor,ExternalVolumeSize,AllocatedIp,Status,ReservedBy", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
