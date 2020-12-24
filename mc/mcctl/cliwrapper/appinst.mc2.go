// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinst.proto

package cliwrapper

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.Result, int, error) {
	args := []string{"region", "CreateAppInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("CloudletLoc,Uri,MappedPorts,Liveness,CreatedAt,Status,Revision,Errors,RuntimeInfo,NodeFlavor,ExternalVolumeSize,AvailabilityZone,State,UpdatedAt,UpdateMultiple,ForceUpdate,PowerState", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) DeleteAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.Result, int, error) {
	args := []string{"region", "DeleteAppInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("CloudletLoc,Uri,MappedPorts,Liveness,CreatedAt,Status,Revision,Errors,RuntimeInfo,NodeFlavor,ExternalVolumeSize,AvailabilityZone,State,UpdatedAt,PowerState", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) RefreshAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.Result, int, error) {
	args := []string{"region", "RefreshAppInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("CloudletLoc,Uri,MappedPorts,Liveness,CreatedAt,Status,Revision,Errors,RuntimeInfo,NodeFlavor,ExternalVolumeSize,AvailabilityZone,State,UpdatedAt,Flavor,AutoClusterIpAccess,Configs,PowerState,TrustPolicy,HealthCheck,SharedVolumeSize,VmFlavor,OptRes", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) UpdateAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.Result, int, error) {
	args := []string{"region", "UpdateAppInst"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("CloudletLoc,Uri,MappedPorts,Liveness,CreatedAt,Status,Revision,Errors,RuntimeInfo,NodeFlavor,ExternalVolumeSize,AvailabilityZone,State,UpdatedAt,Flavor,AutoClusterIpAccess,UpdateMultiple,ForceUpdate,TrustPolicy,HealthCheck,SharedVolumeSize,VmFlavor,OptRes", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowAppInst(uri, token string, in *ormapi.RegionAppInst) ([]edgeproto.AppInst, int, error) {
	args := []string{"region", "ShowAppInst"}
	outlist := []edgeproto.AppInst{}
	noconfig := strings.Split("CloudletLoc,Uri,MappedPorts,Liveness,CreatedAt,Status,Revision,Errors,RuntimeInfo,NodeFlavor,ExternalVolumeSize,AvailabilityZone,State,UpdatedAt", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
