// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stream.proto

package cliwrapper

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
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

func (s *Client) StreamAppInst(uri, token string, in *ormapi.RegionAppInstKey) ([]edgeproto.StreamMsg, int, error) {
	args := []string{"region", "StreamAppInst"}
	outlist := []edgeproto.StreamMsg{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) StreamClusterInst(uri, token string, in *ormapi.RegionClusterInstKey) ([]edgeproto.StreamMsg, int, error) {
	args := []string{"region", "StreamClusterInst"}
	outlist := []edgeproto.StreamMsg{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) StreamCloudlet(uri, token string, in *ormapi.RegionCloudletKey) ([]edgeproto.StreamMsg, int, error) {
	args := []string{"region", "StreamCloudlet"}
	outlist := []edgeproto.StreamMsg{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
