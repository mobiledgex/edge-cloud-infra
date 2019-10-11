// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: alert.proto

/*
Package cliwrapper is a generated protocol buffer package.

It is generated from these files:
	alert.proto
	app.proto
	app_inst.proto
	cloudlet.proto
	cloudletpool.proto
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
	Alert
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
	CloudletPoolKey
	CloudletPool
	CloudletPoolMember
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
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func (s *Client) ShowAlert(uri, token string, in *ormapi.RegionAlert) ([]edgeproto.Alert, int, error) {
	args := []string{"region", "ShowAlert"}
	outlist := []edgeproto.Alert{}
	noconfig := strings.Split("", ",")
	ops := []runOp{
		withIgnore(noconfig),
	}
	st, err := s.runObjs(uri, token, args, in, &outlist, ops...)
	return outlist, st, err
}
