// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudletpool.proto

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

func (s *Client) CreateCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (*edgeproto.Result, int, error) {
	args := []string{"cloudletpool", "create"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Members,CreatedAt,UpdatedAt", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) DeleteCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (*edgeproto.Result, int, error) {
	args := []string{"cloudletpool", "delete"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Members,CreatedAt,UpdatedAt", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) UpdateCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (*edgeproto.Result, int, error) {
	args := []string{"cloudletpool", "update"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Members,CreatedAt,UpdatedAt", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) ShowCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) ([]edgeproto.CloudletPool, int, error) {
	args := []string{"cloudletpool", "show"}
	outlist := []edgeproto.CloudletPool{}
	noconfig := strings.Split("Members,CreatedAt,UpdatedAt", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) AddCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (*edgeproto.Result, int, error) {
	args := []string{"cloudletpool", "addmember"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) RemoveCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (*edgeproto.Result, int, error) {
	args := []string{"cloudletpool", "removemember"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}
