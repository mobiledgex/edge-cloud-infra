// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package cliwrapper

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "strings"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/protoc-gen-cmd/protocmd"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreatePlatform(uri, token string, in *ormapi.RegionPlatform) (edgeproto.Result, int, error) {
	args := []string{"ctrl", "CreatePlatform"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) DeletePlatform(uri, token string, in *ormapi.RegionPlatform) (edgeproto.Result, int, error) {
	args := []string{"ctrl", "DeletePlatform"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) UpdatePlatform(uri, token string, in *ormapi.RegionPlatform) (edgeproto.Result, int, error) {
	args := []string{"ctrl", "UpdatePlatform"}
	out := edgeproto.Result{}
	noconfig := strings.Split("", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) ShowPlatform(uri, token string, in *ormapi.RegionPlatform) ([]edgeproto.Platform, int, error) {
	args := []string{"ctrl", "ShowPlatform"}
	outlist := []edgeproto.Platform{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) CreateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	args := []string{"ctrl", "CreateCloudlet"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,TimeLimits,VaultAddr,TlsCertFile,CrmRoleId,CrmSecretId,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) DeleteCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	args := []string{"ctrl", "DeleteCloudlet"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,TimeLimits,VaultAddr,TlsCertFile,CrmRoleId,CrmSecretId,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
		withStreamOutIncremental(),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) UpdateCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Result, int, error) {
	args := []string{"ctrl", "UpdateCloudlet"}
	outlist := []edgeproto.Result{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,TimeLimits,VaultAddr,TlsCertFile,CrmRoleId,CrmSecretId,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowCloudlet(uri, token string, in *ormapi.RegionCloudlet) ([]edgeproto.Cloudlet, int, error) {
	args := []string{"ctrl", "ShowCloudlet"}
	outlist := []edgeproto.Cloudlet{}
	noconfig := strings.Split("Location.HorizontalAccuracy,Location.VerticalAccuracy,Location.Course,Location.Speed,Location.Timestamp,TimeLimits,VaultAddr,TlsCertFile,CrmRoleId,CrmSecretId,Status", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}

func (s *Client) ShowCloudletInfo(uri, token string, in *ormapi.RegionCloudletInfo) ([]edgeproto.CloudletInfo, int, error) {
	args := []string{"ctrl", "ShowCloudletInfo"}
	outlist := []edgeproto.CloudletInfo{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
