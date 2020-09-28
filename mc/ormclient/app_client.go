// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

package ormclient

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateApp(uri, token string, in *ormapi.RegionApp) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/CreateApp", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) DeleteApp(uri, token string, in *ormapi.RegionApp) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/DeleteApp", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) UpdateApp(uri, token string, in *ormapi.RegionApp) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/UpdateApp", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) ShowApp(uri, token string, in *ormapi.RegionApp) ([]edgeproto.App, int, error) {
	out := edgeproto.App{}
	outlist := []edgeproto.App{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowApp", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) AddAppAutoProvPolicy(uri, token string, in *ormapi.RegionAppAutoProvPolicy) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/AddAppAutoProvPolicy", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) RemoveAppAutoProvPolicy(uri, token string, in *ormapi.RegionAppAutoProvPolicy) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/RemoveAppAutoProvPolicy", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

type AppApiClient interface {
	CreateApp(uri, token string, in *ormapi.RegionApp) (*edgeproto.Result, int, error)
	DeleteApp(uri, token string, in *ormapi.RegionApp) (*edgeproto.Result, int, error)
	UpdateApp(uri, token string, in *ormapi.RegionApp) (*edgeproto.Result, int, error)
	ShowApp(uri, token string, in *ormapi.RegionApp) ([]edgeproto.App, int, error)
	AddAppAutoProvPolicy(uri, token string, in *ormapi.RegionAppAutoProvPolicy) (*edgeproto.Result, int, error)
	RemoveAppAutoProvPolicy(uri, token string, in *ormapi.RegionAppAutoProvPolicy) (*edgeproto.Result, int, error)
}
