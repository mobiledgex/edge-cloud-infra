// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinst.proto

package testutil

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
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

func TestCreateAppInst(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInst, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionAppInst{}
	dat.Region = region
	dat.AppInst = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInst)
	}
	return mcClient.CreateAppInst(uri, token, dat)
}
func TestPermCreateAppInst(mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.AppInst{}
	if targetCloudlet != nil {
		in.Key.ClusterInstKey.CloudletKey = *targetCloudlet
	}
	in.Key.AppKey.Organization = org
	return TestCreateAppInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestDeleteAppInst(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInst, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionAppInst{}
	dat.Region = region
	dat.AppInst = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInst)
	}
	return mcClient.DeleteAppInst(uri, token, dat)
}
func TestPermDeleteAppInst(mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.AppInst{}
	if targetCloudlet != nil {
		in.Key.ClusterInstKey.CloudletKey = *targetCloudlet
	}
	in.Key.AppKey.Organization = org
	return TestDeleteAppInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestRefreshAppInst(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInst, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionAppInst{}
	dat.Region = region
	dat.AppInst = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInst)
	}
	return mcClient.RefreshAppInst(uri, token, dat)
}
func TestPermRefreshAppInst(mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.AppInst{}
	if targetCloudlet != nil {
		in.Key.ClusterInstKey.CloudletKey = *targetCloudlet
	}
	in.Key.AppKey.Organization = org
	return TestRefreshAppInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestUpdateAppInst(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInst, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionAppInst{}
	dat.Region = region
	dat.AppInst = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInst)
	}
	return mcClient.UpdateAppInst(uri, token, dat)
}
func TestPermUpdateAppInst(mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.AppInst{}
	if targetCloudlet != nil {
		in.Key.ClusterInstKey.CloudletKey = *targetCloudlet
	}
	in.Key.AppKey.Organization = org
	return TestUpdateAppInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestShowAppInst(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInst, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.AppInst, int, error) {
	dat := &ormapi.RegionAppInst{}
	dat.Region = region
	dat.AppInst = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInst)
	}
	return mcClient.ShowAppInst(uri, token, dat)
}
func TestPermShowAppInst(mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AppInst)) ([]edgeproto.AppInst, int, error) {
	in := &edgeproto.AppInst{}
	in.Key.AppKey.Organization = org
	return TestShowAppInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestRequestAppInstLatency(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInst, modFuncs ...func(*edgeproto.AppInst)) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionAppInst{}
	dat.Region = region
	dat.AppInst = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInst)
	}
	return mcClient.RequestAppInstLatency(uri, token, dat)
}
func TestPermRequestAppInstLatency(mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) (*edgeproto.Result, int, error) {
	in := &edgeproto.AppInst{}
	if targetCloudlet != nil {
		in.Key.ClusterInstKey.CloudletKey = *targetCloudlet
	}
	in.Key.AppKey.Organization = org
	return TestRequestAppInstLatency(mcClient, uri, token, region, in, modFuncs...)
}

func TestDisplayAppInstLatency(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInst, modFuncs ...func(*edgeproto.AppInst)) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionAppInst{}
	dat.Region = region
	dat.AppInst = *in
	for _, fn := range modFuncs {
		fn(&dat.AppInst)
	}
	return mcClient.DisplayAppInstLatency(uri, token, dat)
}
func TestPermDisplayAppInstLatency(mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) (*edgeproto.Result, int, error) {
	in := &edgeproto.AppInst{}
	if targetCloudlet != nil {
		in.Key.ClusterInstKey.CloudletKey = *targetCloudlet
	}
	in.Key.AppKey.Organization = org
	return TestDisplayAppInstLatency(mcClient, uri, token, region, in, modFuncs...)
}

func (s *TestClient) CreateAppInst(ctx context.Context, in *edgeproto.AppInst) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionAppInst{
		Region:  s.Region,
		AppInst: *in,
	}
	out, status, err := s.McClient.CreateAppInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) DeleteAppInst(ctx context.Context, in *edgeproto.AppInst) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionAppInst{
		Region:  s.Region,
		AppInst: *in,
	}
	out, status, err := s.McClient.DeleteAppInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) RefreshAppInst(ctx context.Context, in *edgeproto.AppInst) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionAppInst{
		Region:  s.Region,
		AppInst: *in,
	}
	out, status, err := s.McClient.RefreshAppInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) UpdateAppInst(ctx context.Context, in *edgeproto.AppInst) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionAppInst{
		Region:  s.Region,
		AppInst: *in,
	}
	out, status, err := s.McClient.UpdateAppInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) ShowAppInst(ctx context.Context, in *edgeproto.AppInst) ([]edgeproto.AppInst, error) {
	inR := &ormapi.RegionAppInst{
		Region:  s.Region,
		AppInst: *in,
	}
	out, status, err := s.McClient.ShowAppInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) RequestAppInstLatency(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	inR := &ormapi.RegionAppInst{
		Region:  s.Region,
		AppInst: *in,
	}
	out, status, err := s.McClient.RequestAppInstLatency(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) DisplayAppInstLatency(ctx context.Context, in *edgeproto.AppInst) (*edgeproto.Result, error) {
	inR := &ormapi.RegionAppInst{
		Region:  s.Region,
		AppInst: *in,
	}
	out, status, err := s.McClient.DisplayAppInstLatency(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) ShowAppInstInfo(ctx context.Context, in *edgeproto.AppInstInfo) ([]edgeproto.AppInstInfo, error) {
	return nil, nil
}

func (s *TestClient) ShowAppInstMetrics(ctx context.Context, in *edgeproto.AppInstMetrics) ([]edgeproto.AppInstMetrics, error) {
	return nil, nil
}
