// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: gputagtable.proto

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

func (s *Client) CreateGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/CreateGpuTagTable", token, in, &out)
	return out, status, err
}

func (s *Client) DeleteGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/DeleteGpuTagTable", token, in, &out)
	return out, status, err
}

func (s *Client) UpdateGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/UpdateGpuTagTable", token, in, &out)
	return out, status, err
}

func (s *Client) ShowGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) ([]edgeproto.GpuTagTable, int, error) {
	out := edgeproto.GpuTagTable{}
	outlist := []edgeproto.GpuTagTable{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowGpuTagTable", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) AddGpuTag(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/AddGpuTag", token, in, &out)
	return out, status, err
}

func (s *Client) RemoveGpuTag(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/RemoveGpuTag", token, in, &out)
	return out, status, err
}

func (s *Client) GetGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTableKey) (edgeproto.GpuTagTable, int, error) {
	out := edgeproto.GpuTagTable{}
	status, err := s.PostJson(uri+"/auth/ctrl/GetGpuTagTable", token, in, &out)
	return out, status, err
}

type GpuTagTableApiClient interface {
	CreateGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error)
	DeleteGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error)
	UpdateGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error)
	ShowGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTable) ([]edgeproto.GpuTagTable, int, error)
	AddGpuTag(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error)
	RemoveGpuTag(uri, token string, in *ormapi.RegionGpuTagTable) (edgeproto.Result, int, error)
	GetGpuTagTable(uri, token string, in *ormapi.RegionGpuTagTableKey) (edgeproto.GpuTagTable, int, error)
}
