// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stream.proto

package testutil

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func TestStreamAppInst(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInstKey, modFuncs ...func(*edgeproto.AppInstKey)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionAppInstKey{}
	dat.Region = region
	dat.AppInstKey = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInstKey)
	}
	return mcClient.StreamAppInst(uri, token, dat)
}
func TestPermStreamAppInst(mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AppInstKey)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.AppInstKey{}
	in.AppKey.Organization = org
	return TestStreamAppInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestStreamClusterInst(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.ClusterInstKey, modFuncs ...func(*edgeproto.ClusterInstKey)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionClusterInstKey{}
	dat.Region = region
	dat.ClusterInstKey = *in
	for _, fn := range modFuncs {
		fn(&dat.ClusterInstKey)
	}
	return mcClient.StreamClusterInst(uri, token, dat)
}
func TestPermStreamClusterInst(mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.ClusterInstKey)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.ClusterInstKey{}
	in.Organization = org
	return TestStreamClusterInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestStreamCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.CloudletKey)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudletKey{}
	dat.Region = region
	dat.CloudletKey = *in
	for _, fn := range modFuncs {
		fn(&dat.CloudletKey)
	}
	return mcClient.StreamCloudlet(uri, token, dat)
}
func TestPermStreamCloudlet(mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletKey)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.CloudletKey{}
	in.Organization = org
	return TestStreamCloudlet(mcClient, uri, token, region, in, modFuncs...)
}

func (s *TestClient) StreamAppInst(ctx context.Context, in *edgeproto.AppInstKey) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionAppInstKey{
		Region:     s.Region,
		AppInstKey: *in,
	}
	out, status, err := s.McClient.StreamAppInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) StreamLocalMsgs(ctx context.Context, in *edgeproto.AppInstKey) ([]edgeproto.Result, error) {
	return nil, nil
}

func (s *TestClient) StreamCloudlet(ctx context.Context, in *edgeproto.CloudletKey) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionCloudletKey{
		Region:      s.Region,
		CloudletKey: *in,
	}
	out, status, err := s.McClient.StreamCloudlet(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) StreamClusterInst(ctx context.Context, in *edgeproto.ClusterInstKey) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionClusterInstKey{
		Region:         s.Region,
		ClusterInstKey: *in,
	}
	out, status, err := s.McClient.StreamClusterInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}
