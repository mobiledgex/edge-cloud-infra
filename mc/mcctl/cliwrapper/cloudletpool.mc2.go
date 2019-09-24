// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudletpool.proto

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

func (s *Client) CreateCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (edgeproto.Result, int, error) {
	args := []string{"region", "CreateCloudletPool"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Members", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) DeleteCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) (edgeproto.Result, int, error) {
	args := []string{"region", "DeleteCloudletPool"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Members", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) ShowCloudletPool(uri, token string, in *ormapi.RegionCloudletPool) ([]edgeproto.CloudletPool, int, error) {
	args := []string{"region", "ShowCloudletPool"}
	outlist := []edgeproto.CloudletPool{}
	noconfig := strings.Split("Members", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) CreateCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (edgeproto.Result, int, error) {
	args := []string{"region", "CreateCloudletPoolMember"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) DeleteCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) (edgeproto.Result, int, error) {
	args := []string{"region", "DeleteCloudletPoolMember"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) ShowCloudletPoolMember(uri, token string, in *ormapi.RegionCloudletPoolMember) ([]edgeproto.CloudletPoolMember, int, error) {
	args := []string{"region", "ShowCloudletPoolMember"}
	outlist := []edgeproto.CloudletPoolMember{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowPoolsForCloudlet(uri, token string, in *ormapi.RegionCloudletKey) ([]edgeproto.CloudletPool, int, error) {
	args := []string{"region", "ShowPoolsForCloudlet"}
	outlist := []edgeproto.CloudletPool{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowCloudletsForPool(uri, token string, in *ormapi.RegionCloudletPoolKey) ([]edgeproto.Cloudlet, int, error) {
	args := []string{"region", "ShowCloudletsForPool"}
	outlist := []edgeproto.Cloudlet{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowCloudletsForPoolList(uri, token string, in *ormapi.RegionCloudletPoolList) ([]edgeproto.Cloudlet, int, error) {
	args := []string{"region", "ShowCloudletsForPoolList"}
	outlist := []edgeproto.Cloudlet{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
