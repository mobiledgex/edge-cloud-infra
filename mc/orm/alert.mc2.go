// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: alert.proto

/*
Package orm is a generated protocol buffer package.

It is generated from these files:
	alert.proto
	app.proto
	app_inst.proto
	autoprovpolicy.proto
	autoscalepolicy.proto
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
	privacypolicy.proto
	refs.proto
	restagtable.proto
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
	AutoProvPolicy
	AutoProvCloudlet
	AutoProvCount
	AutoProvCounts
	AutoProvPolicyCloudlet
	PolicyKey
	AutoScalePolicy
	CloudletKey
	OperationTimeLimits
	CloudletInfraCommon
	AzureProperties
	GcpProperties
	OpenStackProperties
	CloudletInfraProperties
	PlatformConfig
	CloudletResMap
	Cloudlet
	FlavorMatch
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
	OutboundSecurityRule
	PrivacyPolicy
	OperatorCode
	CloudletRefs
	ClusterRefs
	ResTagTableKey
	ResTagTable
	Result
*/
package orm

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/labstack/echo"
import "context"
import "io"
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

func ShowAlert(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAlert{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region

	err = ShowAlertStream(ctx, rc, &in.Alert, func(res *edgeproto.Alert) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

type ShowAlertAuthz interface {
	Ok(obj *edgeproto.Alert) bool
}

func ShowAlertStream(ctx context.Context, rc *RegionContext, obj *edgeproto.Alert, cb func(res *edgeproto.Alert)) error {
	var authz ShowAlertAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAlertAuthz(ctx, rc.region, rc.username, ResourceAlert, ActionView)
		if err == echo.ErrForbidden {
			return nil
		}
		if err != nil {
			return err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAlertApiClient(rc.conn)
	stream, err := api.ShowAlert(ctx, obj)
	if err != nil {
		return err
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return err
		}
		if !rc.skipAuthz {
			if !authz.Ok(res) {
				continue
			}
		}
		cb(res)
	}
	return nil
}

func ShowAlertObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Alert) ([]edgeproto.Alert, error) {
	arr := []edgeproto.Alert{}
	err := ShowAlertStream(ctx, rc, obj, func(res *edgeproto.Alert) {
		arr = append(arr, *res)
	})
	return arr, err
}

func addControllerApis(method string, group *echo.Group) {
	group.Match([]string{method}, "/ctrl/ShowAlert", ShowAlert)
	group.Match([]string{method}, "/ctrl/CreateFlavor", CreateFlavor)
	group.Match([]string{method}, "/ctrl/DeleteFlavor", DeleteFlavor)
	group.Match([]string{method}, "/ctrl/UpdateFlavor", UpdateFlavor)
	group.Match([]string{method}, "/ctrl/ShowFlavor", ShowFlavor)
	group.Match([]string{method}, "/ctrl/AddFlavorRes", AddFlavorRes)
	group.Match([]string{method}, "/ctrl/RemoveFlavorRes", RemoveFlavorRes)
	group.Match([]string{method}, "/ctrl/CreateApp", CreateApp)
	group.Match([]string{method}, "/ctrl/DeleteApp", DeleteApp)
	group.Match([]string{method}, "/ctrl/UpdateApp", UpdateApp)
	group.Match([]string{method}, "/ctrl/ShowApp", ShowApp)
	group.Match([]string{method}, "/ctrl/CreateOperatorCode", CreateOperatorCode)
	group.Match([]string{method}, "/ctrl/DeleteOperatorCode", DeleteOperatorCode)
	group.Match([]string{method}, "/ctrl/ShowOperatorCode", ShowOperatorCode)
	group.Match([]string{method}, "/ctrl/CreateResTagTable", CreateResTagTable)
	group.Match([]string{method}, "/ctrl/DeleteResTagTable", DeleteResTagTable)
	group.Match([]string{method}, "/ctrl/UpdateResTagTable", UpdateResTagTable)
	group.Match([]string{method}, "/ctrl/ShowResTagTable", ShowResTagTable)
	group.Match([]string{method}, "/ctrl/AddResTag", AddResTag)
	group.Match([]string{method}, "/ctrl/RemoveResTag", RemoveResTag)
	group.Match([]string{method}, "/ctrl/GetResTagTable", GetResTagTable)
	group.Match([]string{method}, "/ctrl/CreateCloudlet", CreateCloudlet)
	group.Match([]string{method}, "/ctrl/StreamCloudlet", StreamCloudlet)
	group.Match([]string{method}, "/ctrl/DeleteCloudlet", DeleteCloudlet)
	group.Match([]string{method}, "/ctrl/UpdateCloudlet", UpdateCloudlet)
	group.Match([]string{method}, "/ctrl/ShowCloudlet", ShowCloudlet)
	group.Match([]string{method}, "/ctrl/AddCloudletResMapping", AddCloudletResMapping)
	group.Match([]string{method}, "/ctrl/RemoveCloudletResMapping", RemoveCloudletResMapping)
	group.Match([]string{method}, "/ctrl/FindFlavorMatch", FindFlavorMatch)
	group.Match([]string{method}, "/ctrl/ShowCloudletInfo", ShowCloudletInfo)
	group.Match([]string{method}, "/ctrl/CreateClusterInst", CreateClusterInst)
	group.Match([]string{method}, "/ctrl/StreamClusterInst", StreamClusterInst)
	group.Match([]string{method}, "/ctrl/DeleteClusterInst", DeleteClusterInst)
	group.Match([]string{method}, "/ctrl/UpdateClusterInst", UpdateClusterInst)
	group.Match([]string{method}, "/ctrl/ShowClusterInst", ShowClusterInst)
	group.Match([]string{method}, "/ctrl/CreateAppInst", CreateAppInst)
	group.Match([]string{method}, "/ctrl/StreamAppInst", StreamAppInst)
	group.Match([]string{method}, "/ctrl/DeleteAppInst", DeleteAppInst)
	group.Match([]string{method}, "/ctrl/RefreshAppInst", RefreshAppInst)
	group.Match([]string{method}, "/ctrl/UpdateAppInst", UpdateAppInst)
	group.Match([]string{method}, "/ctrl/ShowAppInst", ShowAppInst)
	group.Match([]string{method}, "/ctrl/CreateAutoScalePolicy", CreateAutoScalePolicy)
	group.Match([]string{method}, "/ctrl/DeleteAutoScalePolicy", DeleteAutoScalePolicy)
	group.Match([]string{method}, "/ctrl/UpdateAutoScalePolicy", UpdateAutoScalePolicy)
	group.Match([]string{method}, "/ctrl/ShowAutoScalePolicy", ShowAutoScalePolicy)
	group.Match([]string{method}, "/ctrl/CreateAutoProvPolicy", CreateAutoProvPolicy)
	group.Match([]string{method}, "/ctrl/DeleteAutoProvPolicy", DeleteAutoProvPolicy)
	group.Match([]string{method}, "/ctrl/UpdateAutoProvPolicy", UpdateAutoProvPolicy)
	group.Match([]string{method}, "/ctrl/ShowAutoProvPolicy", ShowAutoProvPolicy)
	group.Match([]string{method}, "/ctrl/AddAutoProvPolicyCloudlet", AddAutoProvPolicyCloudlet)
	group.Match([]string{method}, "/ctrl/RemoveAutoProvPolicyCloudlet", RemoveAutoProvPolicyCloudlet)
	group.Match([]string{method}, "/ctrl/CreateCloudletPool", CreateCloudletPool)
	group.Match([]string{method}, "/ctrl/DeleteCloudletPool", DeleteCloudletPool)
	group.Match([]string{method}, "/ctrl/ShowCloudletPool", ShowCloudletPool)
	group.Match([]string{method}, "/ctrl/CreateCloudletPoolMember", CreateCloudletPoolMember)
	group.Match([]string{method}, "/ctrl/DeleteCloudletPoolMember", DeleteCloudletPoolMember)
	group.Match([]string{method}, "/ctrl/ShowCloudletPoolMember", ShowCloudletPoolMember)
	group.Match([]string{method}, "/ctrl/ShowPoolsForCloudlet", ShowPoolsForCloudlet)
	group.Match([]string{method}, "/ctrl/ShowCloudletsForPool", ShowCloudletsForPool)
	group.Match([]string{method}, "/ctrl/RunCommand", RunCommand)
	group.Match([]string{method}, "/ctrl/ShowNode", ShowNode)
	group.Match([]string{method}, "/ctrl/CreatePrivacyPolicy", CreatePrivacyPolicy)
	group.Match([]string{method}, "/ctrl/DeletePrivacyPolicy", DeletePrivacyPolicy)
	group.Match([]string{method}, "/ctrl/UpdatePrivacyPolicy", UpdatePrivacyPolicy)
	group.Match([]string{method}, "/ctrl/ShowPrivacyPolicy", ShowPrivacyPolicy)
	group.Match([]string{method}, "/ctrl/ShowCloudletRefs", ShowCloudletRefs)
	group.Match([]string{method}, "/ctrl/ShowClusterRefs", ShowClusterRefs)
}
