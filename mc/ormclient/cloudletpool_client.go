// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudletpool.proto

package ormclient

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
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

func (s *Client) CreateCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/CreateCloudletPool", token, in, &out)
	return out, status, err
}

func (s *Client) DeleteCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/DeleteCloudletPool", token, in, &out)
	return out, status, err
}

func (s *Client) ShowCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) ([]edgeproto.CloudletPool, int, error) {
	out := edgeproto.CloudletPool{}
	outlist := []edgeproto.CloudletPool{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowCloudletPool", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

type CloudletPoolApiClient interface {
	CreateCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (edgeproto.Result, int, error)
	DeleteCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (edgeproto.Result, int, error)
	ShowCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) ([]edgeproto.CloudletPool, int, error)
}

func (s *Client) AddCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/AddCloudletPoolMember", token, in, &out)
	return out, status, err
}

func (s *Client) RemoveCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/RemoveCloudletPoolMember", token, in, &out)
	return out, status, err
}

func (s *Client) ShowCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) ([]edgeproto.CloudletPoolMember, int, error) {
	out := edgeproto.CloudletPoolMember{}
	outlist := []edgeproto.CloudletPoolMember{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowCloudletPoolMember", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) ShowPoolsForCloudlet(uri, token string, in *ormapi.RegionCloudletKey) ([]edgeproto.CloudletPool, int, error) {
	out := edgeproto.CloudletPool{}
	outlist := []edgeproto.CloudletPool{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowPoolsForCloudlet", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) ShowCloudletsForPool(uri, token string, in *ormapi.RegionCloudletPoolKey) ([]edgeproto.Cloudlet, int, error) {
	out := edgeproto.Cloudlet{}
	outlist := []edgeproto.Cloudlet{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowCloudletsForPool", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) ShowCloudletsForPoolList(uri, token string, in *ormapi.RegionCloudletPoolList) ([]edgeproto.Cloudlet, int, error) {
	out := edgeproto.Cloudlet{}
	outlist := []edgeproto.Cloudlet{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowCloudletsForPoolList", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

type CloudletPoolMemberApiClient interface {
	AddCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (edgeproto.Result, int, error)
	RemoveCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (edgeproto.Result, int, error)
	ShowCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) ([]edgeproto.CloudletPoolMember, int, error)
	ShowPoolsForCloudlet(uri, token string, in *ormapi.RegionCloudletKey) ([]edgeproto.CloudletPool, int, error)
	ShowCloudletsForPool(uri, token string, in *ormapi.RegionCloudletPoolKey) ([]edgeproto.Cloudlet, int, error)
	ShowCloudletsForPoolList(uri, token string, in *ormapi.RegionCloudletPoolList) ([]edgeproto.Cloudlet, int, error)
}
