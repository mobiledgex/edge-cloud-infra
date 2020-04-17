// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: exec.proto

package testutil

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "context"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func TestRunCommand(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, int, error) {
	dat := &ormapi.RegionExecRequest{}
	dat.Region = region
	dat.ExecRequest = *in
	return mcClient.RunCommand(uri, token, dat)
}
func TestPermRunCommand(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.ExecRequest, int, error) {
	in := &edgeproto.ExecRequest{}
	in.AppInstKey.AppKey.Organization = org
	return TestRunCommand(mcClient, uri, token, region, in)
}

func TestRunConsole(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, int, error) {
	dat := &ormapi.RegionExecRequest{}
	dat.Region = region
	dat.ExecRequest = *in
	return mcClient.RunConsole(uri, token, dat)
}
func TestPermRunConsole(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.ExecRequest, int, error) {
	in := &edgeproto.ExecRequest{}
	in.AppInstKey.AppKey.Organization = org
	return TestRunConsole(mcClient, uri, token, region, in)
}

func TestShowLogs(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, int, error) {
	dat := &ormapi.RegionExecRequest{}
	dat.Region = region
	dat.ExecRequest = *in
	return mcClient.ShowLogs(uri, token, dat)
}
func TestPermShowLogs(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.ExecRequest, int, error) {
	in := &edgeproto.ExecRequest{}
	in.AppInstKey.AppKey.Organization = org
	return TestShowLogs(mcClient, uri, token, region, in)
}

func TestAccessCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, int, error) {
	dat := &ormapi.RegionExecRequest{}
	dat.Region = region
	dat.ExecRequest = *in
	return mcClient.AccessCloudlet(uri, token, dat)
}
func TestPermAccessCloudlet(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.ExecRequest, int, error) {
	in := &edgeproto.ExecRequest{}
	return TestAccessCloudlet(mcClient, uri, token, region, in)
}

func (s *TestClient) RunCommand(ctx context.Context, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	inR := &ormapi.RegionExecRequest{
		Region:      s.Region,
		ExecRequest: *in,
	}
	out, status, err := s.McClient.RunCommand(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) RunConsole(ctx context.Context, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	inR := &ormapi.RegionExecRequest{
		Region:      s.Region,
		ExecRequest: *in,
	}
	out, status, err := s.McClient.RunConsole(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) ShowLogs(ctx context.Context, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	inR := &ormapi.RegionExecRequest{
		Region:      s.Region,
		ExecRequest: *in,
	}
	out, status, err := s.McClient.ShowLogs(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) AccessCloudlet(ctx context.Context, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	inR := &ormapi.RegionExecRequest{
		Region:      s.Region,
		ExecRequest: *in,
	}
	out, status, err := s.McClient.AccessCloudlet(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) SendLocalRequest(ctx context.Context, in *edgeproto.ExecRequest) (*edgeproto.ExecRequest, error) {
	return nil, nil
}
