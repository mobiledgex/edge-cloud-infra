// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ratelimit.proto

package testutil

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/mctestclient"
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

func TestCreateRateLimitSettings(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.RateLimitSettings, modFuncs ...func(*edgeproto.RateLimitSettings)) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionRateLimitSettings{}
	dat.Region = region
	dat.RateLimitSettings = *in
	for _, fn := range modFuncs {
		fn(&dat.RateLimitSettings)
	}
	return mcClient.CreateRateLimitSettings(uri, token, dat)
}
func TestPermCreateRateLimitSettings(mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.RateLimitSettings)) (*edgeproto.Result, int, error) {
	in := &edgeproto.RateLimitSettings{}
	return TestCreateRateLimitSettings(mcClient, uri, token, region, in, modFuncs...)
}

func TestUpdateRateLimitSettings(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.RateLimitSettings, modFuncs ...func(*edgeproto.RateLimitSettings)) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionRateLimitSettings{}
	dat.Region = region
	dat.RateLimitSettings = *in
	for _, fn := range modFuncs {
		fn(&dat.RateLimitSettings)
	}
	return mcClient.UpdateRateLimitSettings(uri, token, dat)
}
func TestPermUpdateRateLimitSettings(mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.RateLimitSettings)) (*edgeproto.Result, int, error) {
	in := &edgeproto.RateLimitSettings{}
	return TestUpdateRateLimitSettings(mcClient, uri, token, region, in, modFuncs...)
}

func TestDeleteRateLimitSettings(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.RateLimitSettings, modFuncs ...func(*edgeproto.RateLimitSettings)) (*edgeproto.Result, int, error) {
	dat := &ormapi.RegionRateLimitSettings{}
	dat.Region = region
	dat.RateLimitSettings = *in
	for _, fn := range modFuncs {
		fn(&dat.RateLimitSettings)
	}
	return mcClient.DeleteRateLimitSettings(uri, token, dat)
}
func TestPermDeleteRateLimitSettings(mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.RateLimitSettings)) (*edgeproto.Result, int, error) {
	in := &edgeproto.RateLimitSettings{}
	return TestDeleteRateLimitSettings(mcClient, uri, token, region, in, modFuncs...)
}

func TestShowRateLimitSettings(mcClient *mctestclient.Client, uri, token, region string, in *edgeproto.RateLimitSettings, modFuncs ...func(*edgeproto.RateLimitSettings)) ([]edgeproto.RateLimitSettings, int, error) {
	dat := &ormapi.RegionRateLimitSettings{}
	dat.Region = region
	dat.RateLimitSettings = *in
	for _, fn := range modFuncs {
		fn(&dat.RateLimitSettings)
	}
	return mcClient.ShowRateLimitSettings(uri, token, dat)
}
func TestPermShowRateLimitSettings(mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.RateLimitSettings)) ([]edgeproto.RateLimitSettings, int, error) {
	in := &edgeproto.RateLimitSettings{}
	return TestShowRateLimitSettings(mcClient, uri, token, region, in, modFuncs...)
}

func (s *TestClient) CreateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	inR := &ormapi.RegionRateLimitSettings{
		Region:            s.Region,
		RateLimitSettings: *in,
	}
	out, status, err := s.McClient.CreateRateLimitSettings(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) UpdateRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	inR := &ormapi.RegionRateLimitSettings{
		Region:            s.Region,
		RateLimitSettings: *in,
	}
	out, status, err := s.McClient.UpdateRateLimitSettings(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) DeleteRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	inR := &ormapi.RegionRateLimitSettings{
		Region:            s.Region,
		RateLimitSettings: *in,
	}
	out, status, err := s.McClient.DeleteRateLimitSettings(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}

func (s *TestClient) ShowRateLimitSettings(ctx context.Context, in *edgeproto.RateLimitSettings) ([]edgeproto.RateLimitSettings, error) {
	inR := &ormapi.RegionRateLimitSettings{
		Region:            s.Region,
		RateLimitSettings: *in,
	}
	out, status, err := s.McClient.ShowRateLimitSettings(s.Uri, s.Token, inR)
	if err == nil && status != 200 {
		err = fmt.Errorf("status: %d\n", status)
	}
	return out, err
}