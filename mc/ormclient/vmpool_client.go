// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: vmpool.proto

package ormclient

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
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
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/CreateVMPool", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) DeleteVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/DeleteVMPool", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) UpdateVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/UpdateVMPool", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) ShowVMPool(uri, token string, in *ormapi.RegionVMPool) ([]edgeproto.VMPool, int, error) {
	out := edgeproto.VMPool{}
	outlist := []edgeproto.VMPool{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowVMPool", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) AddVMPoolMember(uri, token string, in *ormapi.RegionVMPoolMember) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/AddVMPoolMember", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) RemoveVMPoolMember(uri, token string, in *ormapi.RegionVMPoolMember) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/RemoveVMPoolMember", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

type VMPoolApiClient interface {
	CreateVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error)
	DeleteVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error)
	UpdateVMPool(uri, token string, in *ormapi.RegionVMPool) (*edgeproto.Result, int, error)
	ShowVMPool(uri, token string, in *ormapi.RegionVMPool) ([]edgeproto.VMPool, int, error)
	AddVMPoolMember(uri, token string, in *ormapi.RegionVMPoolMember) (*edgeproto.Result, int, error)
	RemoveVMPoolMember(uri, token string, in *ormapi.RegionVMPoolMember) (*edgeproto.Result, int, error)
}
