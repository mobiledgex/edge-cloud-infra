// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package cliwrapper

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
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	args := []string{"cloudlet", "create"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,Status,Config,NotifySrvAddr,ChefClientKey,State,Errors,CrmAccessPublicKey,CrmAccessKeyUpgradeRequired,CreatedAt,UpdatedAt,TrustPolicyState,HostController,ResTagMap", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) DeleteCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	args := []string{"cloudlet", "delete"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,Status,Config,NotifySrvAddr,ChefClientKey,State,Errors,CrmAccessPublicKey,CrmAccessKeyUpgradeRequired,CreatedAt,UpdatedAt,TrustPolicyState,HostController,ResTagMap", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) UpdateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	args := []string{"cloudlet", "update"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,Status,Config,NotifySrvAddr,ChefClientKey,State,Errors,CrmAccessPublicKey,CrmAccessKeyUpgradeRequired,CreatedAt,UpdatedAt,TrustPolicyState,HostController,PlatformType,DeploymentLocal,Flavor,PhysicalName,ContainerVersion,ResTagMap,VmImageVersion,Deployment,InfraApiAccess,InfraConfig,OverridePolicyContainerVersion,VmPool,ResTagMap", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Cloudlet, int, error) {
	args := []string{"cloudlet", "show"}
	outlist := []edgeproto.Cloudlet{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,Status,Config,NotifySrvAddr,ChefClientKey,State,Errors,CrmAccessPublicKey,CrmAccessKeyUpgradeRequired,CreatedAt,UpdatedAt,TrustPolicyState,HostController,ResTagMap", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) GetCloudletManifest(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.CloudletManifest, int, error) {
	args := []string{"cloudlet", "getmanifest"}
	out := edgeproto.CloudletManifest{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) GetCloudletProps(uri, token string, in *ormapi.RegionCloudletProps) (*edgeproto.CloudletProps, int, error) {
	args := []string{"cloudlet", "getprops"}
	out := edgeproto.CloudletProps{}
	noconfig := strings.Split("Properties", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) GetCloudletResourceQuotaProps(uri, token string, in *ormapi.RegionCloudletResourceQuotaProps) (*edgeproto.CloudletResourceQuotaProps, int, error) {
	args := []string{"cloudlet", "getresourcequotaprops"}
	out := edgeproto.CloudletResourceQuotaProps{}
	noconfig := strings.Split("Properties", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) GetCloudletResourceUsage(uri, token string, in *ormapi.RegionCloudletResourceUsage) (*edgeproto.CloudletResourceUsage, int, error) {
	args := []string{"cloudlet", "getresourceusage"}
	out := edgeproto.CloudletResourceUsage{}
	noconfig := strings.Split("Info", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) AddCloudletResMapping(uri, token string, in *ormapi.RegionCloudletResMap) (*edgeproto.Result, int, error) {
	args := []string{"cloudlet", "addresmapping"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) RemoveCloudletResMapping(uri, token string, in *ormapi.RegionCloudletResMap) (*edgeproto.Result, int, error) {
	args := []string{"cloudlet", "removeresmapping"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) FindFlavorMatch(uri, token string, in *ormapi.RegionFlavorMatch) (*edgeproto.FlavorMatch, int, error) {
	args := []string{"cloudlet", "findflavormatch"}
	out := edgeproto.FlavorMatch{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) RevokeAccessKey(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.Result, int, error) {
	args := []string{"cloudlet", "revokeaccesskey"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) GenerateAccessKey(uri, token string, in *ormapi.RegionCloudletKey) (*edgeproto.Result, int, error) {
	args := []string{"cloudlet", "generateaccesskey"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) ShowCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) ([]edgeproto.CloudletInfo, int, error) {
	args := []string{"cloudletinfo", "show"}
	outlist := []edgeproto.CloudletInfo{}
	noconfig := strings.Split("Resources", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) InjectCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) (*edgeproto.Result, int, error) {
	args := []string{"cloudletinfo", "inject"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Resources", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}

func (s *Client) EvictCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) (*edgeproto.Result, int, error) {
	args := []string{"cloudletinfo", "evict"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Resources", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	if err != nil {
		return nil, st, err
	}
	return &out, st, err
}
