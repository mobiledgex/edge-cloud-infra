// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: trustpolicy.proto

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

func (s *Client) CreateTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	outlist := []edgeproto.Result{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/CreateTrustPolicy", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) DeleteTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	outlist := []edgeproto.Result{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/DeleteTrustPolicy", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) UpdateTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	outlist := []edgeproto.Result{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/UpdateTrustPolicy", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

func (s *Client) ShowTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.TrustPolicy, int, error) {
	out := edgeproto.TrustPolicy{}
	outlist := []edgeproto.TrustPolicy{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowTrustPolicy", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

type TrustPolicyApiClient interface {
	CreateTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.Result, int, error)
	DeleteTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.Result, int, error)
	UpdateTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.Result, int, error)
	ShowTrustPolicy(uri, token string, in *ormapi.RegionTrustPolicy) ([]edgeproto.TrustPolicy, int, error)
}
