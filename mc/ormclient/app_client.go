// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

/*
Package ormclient is a generated protocol buffer package.

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
package ormclient

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) CreateApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/CreateApp", token, in, &out)
	return out, status, err
}

func (s *Client) DeleteApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/DeleteApp", token, in, &out)
	return out, status, err
}

func (s *Client) UpdateApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error) {
	out := edgeproto.Result{}
	status, err := s.PostJson(uri+"/auth/ctrl/UpdateApp", token, in, &out)
	return out, status, err
}

func (s *Client) ShowApp(uri, token string, in *ormapi.RegionApp) ([]edgeproto.App, int, error) {
	out := edgeproto.App{}
	outlist := []edgeproto.App{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowApp", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

type AppApiClient interface {
	CreateApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error)
	DeleteApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error)
	UpdateApp(uri, token string, in *ormapi.RegionApp) (edgeproto.Result, int, error)
	ShowApp(uri, token string, in *ormapi.RegionApp) ([]edgeproto.App, int, error)
}
