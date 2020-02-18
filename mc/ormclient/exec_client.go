// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: exec.proto

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

func (s *Client) RunCommand(uri, token string, in *ormapi.RegionExecRequest) (*edgeproto.ExecRequest, int, error) {
	out := edgeproto.ExecRequest{}
	status, err := s.PostJson(uri+"/auth/ctrl/RunCommand", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) RunConsole(uri, token string, in *ormapi.RegionExecRequest) (*edgeproto.ExecRequest, int, error) {
	out := edgeproto.ExecRequest{}
	status, err := s.PostJson(uri+"/auth/ctrl/RunConsole", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) ShowLogs(uri, token string, in *ormapi.RegionExecRequest) (*edgeproto.ExecRequest, int, error) {
	out := edgeproto.ExecRequest{}
	status, err := s.PostJson(uri+"/auth/ctrl/ShowLogs", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

type ExecApiClient interface {
	RunCommand(uri, token string, in *ormapi.RegionExecRequest) (*edgeproto.ExecRequest, int, error)
	RunConsole(uri, token string, in *ormapi.RegionExecRequest) (*edgeproto.ExecRequest, int, error)
	ShowLogs(uri, token string, in *ormapi.RegionExecRequest) (*edgeproto.ExecRequest, int, error)
}
