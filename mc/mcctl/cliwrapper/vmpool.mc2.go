// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: vmpool.proto

package cliwrapper

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"
import _ "github.com/gogo/protobuf/types"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error) {
	args := []string{"region", "CreateVMPool"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Vms:#.GroupName,Vms:#.InternalName,Vms:#.State,Vms:#.UpdatedAt.Seconds,Vms:#.UpdatedAt.Nanos,Action,Error", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) DeleteVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error) {
	args := []string{"region", "DeleteVMPool"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Vms:#.GroupName,Vms:#.InternalName,Vms:#.State,Vms:#.UpdatedAt.Seconds,Vms:#.UpdatedAt.Nanos,Action,Error", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) UpdateVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error) {
	args := []string{"region", "UpdateVMPool"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Vms:#.GroupName,Vms:#.InternalName,Vms:#.State,Vms:#.UpdatedAt.Seconds,Vms:#.UpdatedAt.Nanos,Action,Error", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) ShowVMPool(uri, token string, in *ormapi.RegionVMPool) ([]edgeproto.VMPool, int, error) {
	args := []string{"region", "ShowVMPool"}
	outlist := []edgeproto.VMPool{}
	noconfig := strings.Split("Vms:#.GroupName,Vms:#.InternalName,Vms:#.State,Vms:#.UpdatedAt.Seconds,Vms:#.UpdatedAt.Nanos,Action,Error", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) AddVMPoolMember(uri, token string, in *ormapi.RegionVMPoolMember) (*edgeproto.Result, int, error) {
	args := []string{"region", "AddVMPoolMember"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Vm.GroupName,Vm.State,Vm.UpdatedAt.Seconds,Vm.UpdatedAt.Nanos", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) RemoveVMPoolMember(uri, token string, in *ormapi.RegionVMPoolMember) (*edgeproto.Result, int, error) {
	args := []string{"region", "RemoveVMPoolMember"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Vm.GroupName,Vm.State,Vm.UpdatedAt.Seconds,Vm.UpdatedAt.Nanos", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}
