// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package testutil

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "context"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func TestCreateCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.CreateCloudlet(uri, token, dat)
}
func TestPermCreateCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Result, int, error) {
	in := &edgeproto.Cloudlet{}
	in.Key.Organization = org
	return TestCreateCloudlet(mcClient, uri, token, region, in)
}

func TestDeleteCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.DeleteCloudlet(uri, token, dat)
}
func TestPermDeleteCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Result, int, error) {
	in := &edgeproto.Cloudlet{}
	in.Key.Organization = org
	return TestDeleteCloudlet(mcClient, uri, token, region, in)
}

func TestUpdateCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.UpdateCloudlet(uri, token, dat)
}
func TestPermUpdateCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Result, int, error) {
	in := &edgeproto.Cloudlet{}
	in.Key.Organization = org
	return TestUpdateCloudlet(mcClient, uri, token, region, in)
}

func TestShowCloudlet(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) ([]edgeproto.Cloudlet, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.ShowCloudlet(uri, token, dat)
}
func TestPermShowCloudlet(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.Cloudlet, int, error) {
	in := &edgeproto.Cloudlet{}
	return TestShowCloudlet(mcClient, uri, token, region, in)
}

func TestGetCloudletManifest(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.Cloudlet) (*edgeproto.CloudletManifest, int, error) {
	dat := &ormapi.RegionCloudlet{}
	dat.Region = region
	dat.Cloudlet = *in
	return mcClient.GetCloudletManifest(uri, token, dat)
}
func TestPermGetCloudletManifest(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.CloudletManifest, int, error) {
	in := &edgeproto.Cloudlet{}
	in.Key.Organization = org
	return TestGetCloudletManifest(mcClient, uri, token, region, in)
}

func TestAddCloudletResMapping(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.CloudletResMap) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudletResMap{}
	dat.Region = region
	dat.CloudletResMap = *in
	return mcClient.AddCloudletResMapping(uri, token, dat)
}
func TestPermAddCloudletResMapping(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.Result, int, error) {
	in := &edgeproto.CloudletResMap{}
	in.Key.Organization = org
	return TestAddCloudletResMapping(mcClient, uri, token, region, in)
}

func TestRemoveCloudletResMapping(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.CloudletResMap) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionCloudletResMap{}
	dat.Region = region
	dat.CloudletResMap = *in
	return mcClient.RemoveCloudletResMapping(uri, token, dat)
}
func TestPermRemoveCloudletResMapping(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.Result, int, error) {
	in := &edgeproto.CloudletResMap{}
	in.Key.Organization = org
	return TestRemoveCloudletResMapping(mcClient, uri, token, region, in)
}

func TestFindFlavorMatch(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, int, error) {
	dat := &ormapi.RegionFlavorMatch{}
	dat.Region = region
	dat.FlavorMatch = *in
	return mcClient.FindFlavorMatch(uri, token, dat)
}
func TestPermFindFlavorMatch(mcClient *ormclient.Client, uri, token, region, org string) (*edgeproto.FlavorMatch, int, error) {
	in := &edgeproto.FlavorMatch{}
	in.Key.Organization = org
	return TestFindFlavorMatch(mcClient, uri, token, region, in)
}

func (s *TestClient) CreateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionCloudlet{
		Region:   s.Region,
		Cloudlet: *in,
	}
	out, status, err := s.McClient.CreateCloudlet(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) DeleteCloudlet(ctx context.Context, in *edgeproto.Cloudlet) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionCloudlet{
		Region:   s.Region,
		Cloudlet: *in,
	}
	out, status, err := s.McClient.DeleteCloudlet(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) UpdateCloudlet(ctx context.Context, in *edgeproto.Cloudlet) ([]edgeproto.Result, error) {
	inR := &ormapi.RegionCloudlet{
		Region:   s.Region,
		Cloudlet: *in,
	}
	out, status, err := s.McClient.UpdateCloudlet(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) ShowCloudlet(ctx context.Context, in *edgeproto.Cloudlet) ([]edgeproto.Cloudlet, error) {
	inR := &ormapi.RegionCloudlet{
		Region:   s.Region,
		Cloudlet: *in,
	}
	out, status, err := s.McClient.ShowCloudlet(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) GetCloudletManifest(ctx context.Context, in *edgeproto.Cloudlet) (*edgeproto.CloudletManifest, error) {
	inR := &ormapi.RegionCloudlet{
		Region:   s.Region,
		Cloudlet: *in,
	}
	out, status, err := s.McClient.GetCloudletManifest(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) AddCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	inR := &ormapi.RegionCloudletResMap{
		Region:         s.Region,
		CloudletResMap: *in,
	}
	out, status, err := s.McClient.AddCloudletResMapping(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) RemoveCloudletResMapping(ctx context.Context, in *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	inR := &ormapi.RegionCloudletResMap{
		Region:         s.Region,
		CloudletResMap: *in,
	}
	out, status, err := s.McClient.RemoveCloudletResMapping(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) FindFlavorMatch(ctx context.Context, in *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, error) {
	inR := &ormapi.RegionFlavorMatch{
		Region:      s.Region,
		FlavorMatch: *in,
	}
	out, status, err := s.McClient.FindFlavorMatch(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func TestShowCloudletInfo(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.CloudletInfo) ([]edgeproto.CloudletInfo, int, error) {
	dat := &ormapi.RegionCloudletInfo{}
	dat.Region = region
	dat.CloudletInfo = *in
	return mcClient.ShowCloudletInfo(uri, token, dat)
}
func TestPermShowCloudletInfo(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.CloudletInfo, int, error) {
	in := &edgeproto.CloudletInfo{}
	in.Key.Organization = org
	return TestShowCloudletInfo(mcClient, uri, token, region, in)
}

func (s *TestClient) ShowCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) ([]edgeproto.CloudletInfo, error) {
	inR := &ormapi.RegionCloudletInfo{
		Region:       s.Region,
		CloudletInfo: *in,
	}
	out, status, err := s.McClient.ShowCloudletInfo(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) InjectCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return nil, nil
}

func (s *TestClient) EvictCloudletInfo(ctx context.Context, in *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	return nil, nil
}

func (s *TestClient) ShowCloudletMetrics(ctx context.Context, in *edgeproto.CloudletMetrics) ([]edgeproto.CloudletMetrics, error) {
	return nil, nil
}
