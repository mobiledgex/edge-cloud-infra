// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package ormclient

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
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	outlist := []edgeproto.Result{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/CreateCloudlet", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) DeleteCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	outlist := []edgeproto.Result{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/DeleteCloudlet", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) UpdateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	outlist := []edgeproto.Result{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/UpdateCloudlet", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) ShowCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Cloudlet, int, error) {
	out := edgeproto.Cloudlet{}
	outlist := []edgeproto.Cloudlet{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowCloudlet", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) GetCloudletManifest(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.CloudletManifest, int, error) {
	out := edgeproto.CloudletManifest{}
	status, err := s.PostJson(uri+"/auth/ctrl/GetCloudletManifest", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) GetCloudletProps(uri, token string, in *ormapi.RegionCloudletProps) (*edgeproto.CloudletProps, int, error) {
	out := edgeproto.CloudletProps{}
	status, err := s.PostJson(uri+"/auth/ctrl/GetCloudletProps", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) GetCloudletResourceQuotaProps(uri, token string, in *ormapi.RegionCloudletResourceQuotaProps) (*edgeproto.CloudletResourceQuotaProps, int, error) {
	out := edgeproto.CloudletResourceQuotaProps{}
	status, err := s.PostJson(uri+"/auth/ctrl/GetCloudletResourceQuotaProps", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) GetCloudletResourceUsage(uri, token string, in *ormapi.RegionCloudletResourceUsage) (*edgeproto.CloudletResourceUsage, int, error) {
	out := edgeproto.CloudletResourceUsage{}
	status, err := s.PostJson(uri+"/auth/ctrl/GetCloudletResourceUsage", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) AddCloudletResMapping(uri, token string, in *ormapi.RegionCloudletResMap) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/AddCloudletResMapping", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) RemoveCloudletResMapping(uri, token string, in *ormapi.RegionCloudletResMap) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/RemoveCloudletResMapping", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) FindFlavorMatch(uri, token string, in *ormapi.RegionFlavorMatch) (*edgeproto.FlavorMatch, int, error) {
	out := edgeproto.FlavorMatch{}
	status, err := s.PostJson(uri+"/auth/ctrl/FindFlavorMatch", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) RevokeAccessKey(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/RevokeAccessKey", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) GenerateAccessKey(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/GenerateAccessKey", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

type CloudletApiClient interface {
	CreateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error)
	DeleteCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error)
	UpdateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error)
	ShowCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Cloudlet, int, error)
	GetCloudletManifest(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.CloudletManifest, int, error)
	GetCloudletProps(uri, token string, in *ormapi.RegionCloudletProps) (*edgeproto.CloudletProps, int, error)
	GetCloudletResourceQuotaProps(uri, token string, in *ormapi.RegionCloudletResourceQuotaProps) (*edgeproto.CloudletResourceQuotaProps, int, error)
	GetCloudletResourceUsage(uri, token string, in *ormapi.RegionCloudletResourceUsage) (*edgeproto.CloudletResourceUsage, int, error)
	AddCloudletResMapping(uri, token string, in *ormapi.RegionCloudletResMap) (*edgeproto.Result, int, error)
	RemoveCloudletResMapping(uri, token string, in *ormapi.RegionCloudletResMap) (*edgeproto.Result, int, error)
	FindFlavorMatch(uri, token string, in *ormapi.RegionFlavorMatch) (*edgeproto.FlavorMatch, int, error)
	RevokeAccessKey(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.Result, int, error)
	GenerateAccessKey(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.Result, int, error)
}

func (s *Client) ShowCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) ([]edgeproto.CloudletInfo, int, error) {
	out := edgeproto.CloudletInfo{}
	outlist := []edgeproto.CloudletInfo{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowCloudletInfo", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) InjectCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/InjectCloudletInfo", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

func (s *Client) EvictCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) (*edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/EvictCloudletInfo", token, in, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, err
}

type CloudletInfoApiClient interface {
	ShowCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) ([]edgeproto.CloudletInfo, int, error)
	InjectCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) (*edgeproto.Result, int, error)
	EvictCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) (*edgeproto.Result, int, error)
}
