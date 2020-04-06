// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: device.proto

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
import _ "github.com/gogo/protobuf/types"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) InjectDevice(uri, token string, in *ormapi.RegionDevice) (*edgeproto.Result, int, error) {
	args := []string{"region", "InjectDevice"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) ShowDevice(uri, token string, in *ormapi.RegionDevice) ([]edgeproto.Device, int, error) {
	args := []string{"region", "ShowDevice"}
	outlist := []edgeproto.Device{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) EvictDevice(uri, token string, in *ormapi.RegionDevice) (*edgeproto.Result, int, error) {
	args := []string{"region", "EvictDevice"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) ShowDeviceReport(uri, token string, in *ormapi.RegionDeviceReport) ([]edgeproto.Device, int, error) {
	args := []string{"region", "ShowDeviceReport"}
	outlist := []edgeproto.Device{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
