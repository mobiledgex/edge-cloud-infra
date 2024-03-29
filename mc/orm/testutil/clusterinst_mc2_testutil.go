// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

package testutil

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/mctestclient"
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

func TestCreateClusterInst(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.ClusterInst, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionClusterInst{}
	dat.Region = region
	dat.ClusterInst = *in
	for _, fn := range modFuncs {
		fn(&dat.ClusterInst)
	}
	return mcClient.CreateClusterInst(uri, token, dat)
}
func TestPermCreateClusterInst(mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.ClusterInst{}
	if targetCloudlet != nil {
		in.Key.CloudletKey = *targetCloudlet
	}
	in.Key.Organization = org
	return TestCreateClusterInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestDeleteClusterInst(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.ClusterInst, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionClusterInst{}
	dat.Region = region
	dat.ClusterInst = *in
	for _, fn := range modFuncs {
		fn(&dat.ClusterInst)
	}
	return mcClient.DeleteClusterInst(uri, token, dat)
}
func TestPermDeleteClusterInst(mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.ClusterInst{}
	if targetCloudlet != nil {
		in.Key.CloudletKey = *targetCloudlet
	}
	in.Key.Organization = org
	return TestDeleteClusterInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestUpdateClusterInst(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.ClusterInst, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionClusterInst{}
	dat.Region = region
	dat.ClusterInst = *in
	for _, fn := range modFuncs {
		fn(&dat.ClusterInst)
	}
	return mcClient.UpdateClusterInst(uri, token, dat)
}
func TestPermUpdateClusterInst(mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.Result, int, error) {
	in := &edgeproto.ClusterInst{}
	if targetCloudlet != nil {
		in.Key.CloudletKey = *targetCloudlet
		in.Fields = append(in.Fields,
			edgeproto.ClusterInstFieldKeyCloudletKeyName,
			edgeproto.ClusterInstFieldKeyCloudletKeyOrganization,
		)
	}
	in.Key.Organization = org
	in.Fields = append(in.Fields, edgeproto.ClusterInstFieldKeyOrganization)
	return TestUpdateClusterInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestShowClusterInst(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.ClusterInst, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.ClusterInst, int, error) {
	dat := &ormapi.RegionClusterInst{}
	dat.Region = region
	dat.ClusterInst = *in
	for _, fn := range modFuncs {
		fn(&dat.ClusterInst)
	}
	return mcClient.ShowClusterInst(uri, token, dat)
}
func TestPermShowClusterInst(mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.ClusterInst)) ([]edgeproto.ClusterInst, int, error) {
	in := &edgeproto.ClusterInst{}
	in.Key.Organization = org
	return TestShowClusterInst(mcClient, uri, token, region, in, modFuncs...)
}

func TestDeleteIdleReservableClusterInsts(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.IdleReservableClusterInsts, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionIdleReservableClusterInsts{}
	dat.Region = region
	dat.IdleReservableClusterInsts = *in
	for _, fn := range modFuncs {
		fn(&dat.IdleReservableClusterInsts)
	}
	return mcClient.DeleteIdleReservableClusterInsts(uri, token, dat)
}
func TestPermDeleteIdleReservableClusterInsts(mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) (*edgeproto.Result, int, error) {
	in := &edgeproto.IdleReservableClusterInsts{}
	return TestDeleteIdleReservableClusterInsts(mcClient, uri, token, region, in, modFuncs...)
}

func (s *TestClient) CreateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionClusterInst{
		Region:      s.Region,
		ClusterInst: *in,
	}
	out, status, err := s.McClient.CreateClusterInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) DeleteClusterInst(ctx context.Context, in *edgeproto.ClusterInst) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionClusterInst{
		Region:      s.Region,
		ClusterInst: *in,
	}
	out, status, err := s.McClient.DeleteClusterInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) UpdateClusterInst(ctx context.Context, in *edgeproto.ClusterInst) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionClusterInst{
		Region:      s.Region,
		ClusterInst: *in,
	}
	out, status, err := s.McClient.UpdateClusterInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) ShowClusterInst(ctx context.Context, in *edgeproto.ClusterInst) ([]edgeproto.ClusterInst, error) {
	inR := &ormapi.RegionClusterInst{
		Region:      s.Region,
		ClusterInst: *in,
	}
	out, status, err := s.McClient.ShowClusterInst(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) DeleteIdleReservableClusterInsts(ctx context.Context, in *edgeproto.IdleReservableClusterInsts) (*edgeproto.Result, error) {
	inR := &ormapi.RegionIdleReservableClusterInsts{
		Region:                     s.Region,
		IdleReservableClusterInsts: *in,
	}
	out, status, err := s.McClient.DeleteIdleReservableClusterInsts(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) ShowClusterInstInfo(ctx context.Context, in *edgeproto.ClusterInstInfo) ([]edgeproto.ClusterInstInfo, error) {
	return nil, nil
}
