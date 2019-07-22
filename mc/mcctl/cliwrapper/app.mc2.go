// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

/*
Package cliwrapper is a generated protocol buffer package.

It is generated from these files:
	app.proto
	app_inst.proto
	cloudlet.proto
	cluster.proto
	clusterinst.proto
	common.proto
	controller.proto
	developer.proto
	exec.proto
	flavor.proto
	metric.proto
	node.proto
	notice.proto
	operator.proto
	refs.proto
	result.proto
	version.proto

It has these top-level messages:
	AppKey
	ConfigFile
	App
	AppInstKey
	AppInst
	AppInstRuntime
	AppInstInfo
	AppInstMetrics
	CloudletKey
	OperationTimeLimits
	CloudletInfraCommon
	AzureProperties
	GcpProperties
	OpenStackProperties
	CloudletInfraProperties
	PlatformConfig
	Cloudlet
	FlavorInfo
	CloudletInfo
	CloudletMetrics
	ClusterKey
	ClusterInstKey
	ClusterInst
	ClusterInstInfo
	StatusInfo
	ControllerKey
	Controller
	DeveloperKey
	Developer
	ExecRequest
	FlavorKey
	Flavor
	MetricTag
	MetricVal
	Metric
	NodeKey
	Node
	Notice
	OperatorKey
	Operator
	CloudletRefs
	ClusterRefs
	Result
*/
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
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error) {
	args := []string{"ctrl", "CreateApp"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Revision", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) DeleteApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error) {
	args := []string{"ctrl", "DeleteApp"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Revision", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) UpdateApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error) {
	args := []string{"ctrl", "UpdateApp"}
	out := edgeproto.Result{}
	noconfig := strings.Split("Revision", ",")
	st, err := s.runObjs(uri, token, args, in, &out, withIgnore(noconfig))
	return out, st, err
}

func (s *Client) ShowApp(uri, token string, in *ormapi.RegionApp) ([]edgeproto.App, int, error) {
	args := []string{"ctrl", "ShowApp"}
	outlist := []edgeproto.App{}
	noconfig := strings.Split("Revision", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
