// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

package cliwrapper

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.Result, int, error) {
	args := []string{"ctrl", "CreateAppInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("CloudletLoc,MappedPorts,Liveness,ClusterInstKey.CloudletKey,CreatedAt,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) DeleteAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.Result, int, error) {
	args := []string{"ctrl", "DeleteAppInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("CloudletLoc,MappedPorts,Liveness,ClusterInstKey.CloudletKey,CreatedAt,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) UpdateAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.Result, int, error) {
	args := []string{"ctrl", "UpdateAppInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("CloudletLoc,MappedPorts,Liveness,ClusterInstKey.CloudletKey,CreatedAt,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.AppInst, int, error) {
	args := []string{"ctrl", "ShowAppInst"}
	outlist := []edgeproto.AppInst{}
	noconfig := strings.Split("CloudletLoc,MappedPorts,Liveness,ClusterInstKey.CloudletKey,CreatedAt,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
