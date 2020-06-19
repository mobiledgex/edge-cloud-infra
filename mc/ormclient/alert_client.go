// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: alert.proto

/*
Package ormclient is a generated protocol buffer package.

It is generated from these files:
	alert.proto
	alldata.proto
	app.proto
	appinst.proto
	appinstclient.proto
	autoprovpolicy.proto
	autoscalepolicy.proto
	cloudlet.proto
	cloudletpool.proto
	cluster.proto
	clusterinst.proto
	common.proto
	controller.proto
	debug.proto
	device.proto
	exec.proto
	flavor.proto
	metric.proto
	node.proto
	notice.proto
	operatorcode.proto
	org.proto
	privacypolicy.proto
	refs.proto
	restagtable.proto
	result.proto
	settings.proto
	version.proto

It has these top-level messages:
	Alert
	AllData
	AppKey
	ConfigFile
	App
	AppAutoProvPolicy
	AppInstKey
	AppInst
	AppInstRuntime
	AppInstInfo
	AppInstMetrics
	AppInstClientKey
	AppInstClient
	AutoProvPolicy
	AutoProvCloudlet
	AutoProvCount
	AutoProvCounts
	AutoProvPolicyCloudlet
	AutoProvInfo
	PolicyKey
	AutoScalePolicy
	CloudletKey
	OperationTimeLimits
	PlatformConfig
	CloudletResMap
	InfraConfig
	Cloudlet
	FlavorMatch
	CloudletManifest
	FlavorInfo
	OSAZone
	OSImage
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
	DebugRequest
	DebugReply
	DebugData
	DeviceReport
	DeviceKey
	Device
	CloudletMgmtNode
	RunCmd
	RunVMConsole
	ShowLog
	ExecRequest
	FlavorKey
	Flavor
	MetricTag
	MetricVal
	Metric
	NodeKey
	Node
	NodeData
	Notice
	OperatorCode
	Organization
	OrganizationData
	OutboundSecurityRule
	PrivacyPolicy
	CloudletRefs
	ClusterRefs
	AppInstRefs
	ResTagTableKey
	ResTagTable
	Result
	Settings
*/
package ormclient

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
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
	out := edgeproto.Alert{}
	outlist := []edgeproto.Alert{}
	status, err := s.PostJsonStreamOut(uri+"/auth/ctrl/ShowAlert", token, in, &out, func() {
		outlist = append(outlist, out)
	})
	return outlist, status, err
}

type AlertApiClient interface {
	ShowAlert(uri, token string, in *ormapi.RegionAlert) ([]edgeproto.Alert, int, error)
}
