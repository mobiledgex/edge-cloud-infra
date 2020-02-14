// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: alert.proto

/*
Package orm is a generated protocol buffer package.

It is generated from these files:
	alert.proto
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
	settings.proto
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
	AppInstClientKey
	AppInstClient
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
	Notice
	OperatorKey
	Operator
	OperatorCode
	OutboundSecurityRule
	PrivacyPolicy
	CloudletRefs
	ClusterRefs
	ResTagTableKey
	ResTagTable
	Result
	Settings
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
	// swagger:route POST /auth/ctrl/ShowAlert Alert ShowAlert
	// Show alerts.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowAlert", ShowAlert)
	// swagger:route POST /auth/ctrl/CreateFlavor Flavor CreateFlavor
	// Create a Flavor.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateFlavor", CreateFlavor)
	// swagger:route POST /auth/ctrl/DeleteFlavor Flavor DeleteFlavor
	// Delete a Flavor.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteFlavor", DeleteFlavor)
	// swagger:route POST /auth/ctrl/UpdateFlavor Flavor UpdateFlavor
	// Update a Flavor.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateFlavor", UpdateFlavor)
	// swagger:route POST /auth/ctrl/ShowFlavor Flavor ShowFlavor
	// Show Flavors.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowFlavor", ShowFlavor)
	// swagger:route POST /auth/ctrl/AddFlavorRes Flavor AddFlavorRes
	// Add Optional Resource.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/AddFlavorRes", AddFlavorRes)
	// swagger:route POST /auth/ctrl/RemoveFlavorRes Flavor RemoveFlavorRes
	// Remove Optional Resource.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/RemoveFlavorRes", RemoveFlavorRes)
	// swagger:route POST /auth/ctrl/CreateApp App CreateApp
	// Create Application.
	//  Creates a definition for an application instance for Cloudlet deployment.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateApp", CreateApp)
	// swagger:route POST /auth/ctrl/DeleteApp App DeleteApp
	// Delete Application.
	//  Deletes a definition of an Application instance. Make sure no other application instances exist with that definition. If they do exist, you must delete those Application instances first.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteApp", DeleteApp)
	// swagger:route POST /auth/ctrl/UpdateApp App UpdateApp
	// Update Application.
	//  Updates the definition of an Application instance.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateApp", UpdateApp)
	// swagger:route POST /auth/ctrl/ShowApp App ShowApp
	// Show Applications.
	//  Lists all Application definitions managed from the Edge Controller. Any fields specified will be used to filter results.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowApp", ShowApp)
	// swagger:route POST /auth/ctrl/CreateOperatorCode OperatorCode CreateOperatorCode
	// Create a code for an Operator.
	//
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateOperatorCode", CreateOperatorCode)
	// swagger:route POST /auth/ctrl/DeleteOperatorCode OperatorCode DeleteOperatorCode
	// Delete a code for an Operator.
	//
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteOperatorCode", DeleteOperatorCode)
	// swagger:route POST /auth/ctrl/ShowOperatorCode OperatorCode ShowOperatorCode
	// Show OperatorCodes.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowOperatorCode", ShowOperatorCode)
	// swagger:route POST /auth/ctrl/CreateResTagTable ResTagTable CreateResTagTable
	// Create TagTable.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateResTagTable", CreateResTagTable)
	// swagger:route POST /auth/ctrl/DeleteResTagTable ResTagTable DeleteResTagTable
	// Delete TagTable.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteResTagTable", DeleteResTagTable)
	// swagger:route POST /auth/ctrl/UpdateResTagTable ResTagTable UpdateResTagTable
	// .
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateResTagTable", UpdateResTagTable)
	// swagger:route POST /auth/ctrl/ShowResTagTable ResTagTable ShowResTagTable
	// show TagTable.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowResTagTable", ShowResTagTable)
	// swagger:route POST /auth/ctrl/AddResTag ResTagTable AddResTag
	// add new tag(s) to TagTable.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/AddResTag", AddResTag)
	// swagger:route POST /auth/ctrl/RemoveResTag ResTagTable RemoveResTag
	// remove existing tag(s) from TagTable.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/RemoveResTag", RemoveResTag)
	// swagger:route POST /auth/ctrl/GetResTagTable ResTagTableKey GetResTagTable
	// Fetch a copy of the TagTable.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/GetResTagTable", GetResTagTable)
	// swagger:route POST /auth/ctrl/CreateCloudlet Cloudlet CreateCloudlet
	// Create Cloudlet.
	//  Sets up Cloudlet services on the Operators compute resources, and integrated as part of MobiledgeX edge resource portfolio. These resources are managed from the Edge Controller.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateCloudlet", CreateCloudlet)
	group.Match([]string{method}, "/ctrl/StreamCloudlet", StreamCloudlet)
	// swagger:route POST /auth/ctrl/DeleteCloudlet Cloudlet DeleteCloudlet
	// Delete Cloudlet.
	//  Removes the Cloudlet services where they are no longer managed from the Edge Controller.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteCloudlet", DeleteCloudlet)
	// swagger:route POST /auth/ctrl/UpdateCloudlet Cloudlet UpdateCloudlet
	// Update Cloudlet.
	//  Updates the Cloudlet configuration and manages the upgrade of Cloudlet services.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateCloudlet", UpdateCloudlet)
	// swagger:route POST /auth/ctrl/ShowCloudlet Cloudlet ShowCloudlet
	// Show Cloudlets.
	//  Lists all the cloudlets managed from Edge Controller.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowCloudlet", ShowCloudlet)
	// swagger:route POST /auth/ctrl/AddCloudletResMapping CloudletResMap AddCloudletResMapping
	// Add Optional Resource tag table.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/AddCloudletResMapping", AddCloudletResMapping)
	// swagger:route POST /auth/ctrl/RemoveCloudletResMapping CloudletResMap RemoveCloudletResMapping
	// Add Optional Resource tag table.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/RemoveCloudletResMapping", RemoveCloudletResMapping)
	// swagger:route POST /auth/ctrl/FindFlavorMatch FlavorMatch FindFlavorMatch
	// Discover if flavor produces a matching platform flavor.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/FindFlavorMatch", FindFlavorMatch)
	// swagger:route POST /auth/ctrl/ShowCloudletInfo CloudletInfo ShowCloudletInfo
	// Show CloudletInfos.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowCloudletInfo", ShowCloudletInfo)
	// swagger:route POST /auth/ctrl/CreateClusterInst ClusterInst CreateClusterInst
	// Create Cluster Instance.
	//  Creates an instance of a Cluster on a Cloudlet, defined by a Cluster Key and a Cloudlet Key. ClusterInst is a collection of compute resources on a Cloudlet on which AppInsts are deployed.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateClusterInst", CreateClusterInst)
	group.Match([]string{method}, "/ctrl/StreamClusterInst", StreamClusterInst)
	// swagger:route POST /auth/ctrl/DeleteClusterInst ClusterInst DeleteClusterInst
	// Delete Cluster Instance.
	//  Deletes an instance of a Cluster deployed on a Cloudlet.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteClusterInst", DeleteClusterInst)
	// swagger:route POST /auth/ctrl/UpdateClusterInst ClusterInst UpdateClusterInst
	// Update Cluster Instance.
	//  Updates an instance of a Cluster deployed on a Cloudlet.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateClusterInst", UpdateClusterInst)
	// swagger:route POST /auth/ctrl/ShowClusterInst ClusterInst ShowClusterInst
	// Show Cluster Instances.
	//  Lists all the cluster instances managed by Edge Controller.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowClusterInst", ShowClusterInst)
	// swagger:route POST /auth/ctrl/CreateAppInst AppInst CreateAppInst
	// Create Application Instance.
	//  Creates an instance of an App on a Cloudlet where it is defined by an App plus a ClusterInst key. Many of the fields here are inherited from the App definition.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateAppInst", CreateAppInst)
	group.Match([]string{method}, "/ctrl/StreamAppInst", StreamAppInst)
	// swagger:route POST /auth/ctrl/DeleteAppInst AppInst DeleteAppInst
	// Delete Application Instance.
	//  Deletes an instance of the App from the Cloudlet.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteAppInst", DeleteAppInst)
	// swagger:route POST /auth/ctrl/RefreshAppInst AppInst RefreshAppInst
	// Refresh Application Instance.
	//  Restarts an App instance with new App settings or image.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/RefreshAppInst", RefreshAppInst)
	// swagger:route POST /auth/ctrl/UpdateAppInst AppInst UpdateAppInst
	// Update Application Instance.
	//  Updates an Application instance and then refreshes it.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateAppInst", UpdateAppInst)
	// swagger:route POST /auth/ctrl/ShowAppInst AppInst ShowAppInst
	// Show Application Instances.
	//  Lists all the Application instances managed by the Edge Controller. Any fields specified will be used to filter results.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowAppInst", ShowAppInst)
	// swagger:route POST /auth/ctrl/ShowAppInstClient AppInstClientKey ShowAppInstClient
	// Show application instance clients.
	//
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowAppInstClient", ShowAppInstClient)
	// swagger:route POST /auth/ctrl/CreateAutoScalePolicy AutoScalePolicy CreateAutoScalePolicy
	// Create an Auto Scale Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateAutoScalePolicy", CreateAutoScalePolicy)
	// swagger:route POST /auth/ctrl/DeleteAutoScalePolicy AutoScalePolicy DeleteAutoScalePolicy
	// Delete an Auto Scale Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteAutoScalePolicy", DeleteAutoScalePolicy)
	// swagger:route POST /auth/ctrl/UpdateAutoScalePolicy AutoScalePolicy UpdateAutoScalePolicy
	// Update an Auto Scale Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateAutoScalePolicy", UpdateAutoScalePolicy)
	// swagger:route POST /auth/ctrl/ShowAutoScalePolicy AutoScalePolicy ShowAutoScalePolicy
	// Show Auto Scale Policies.
	//  Any fields specified will be used to filter results.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowAutoScalePolicy", ShowAutoScalePolicy)
	// swagger:route POST /auth/ctrl/CreateAutoProvPolicy AutoProvPolicy CreateAutoProvPolicy
	// Create an Auto Provisioning Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateAutoProvPolicy", CreateAutoProvPolicy)
	// swagger:route POST /auth/ctrl/DeleteAutoProvPolicy AutoProvPolicy DeleteAutoProvPolicy
	// Delete an Auto Provisioning Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteAutoProvPolicy", DeleteAutoProvPolicy)
	// swagger:route POST /auth/ctrl/UpdateAutoProvPolicy AutoProvPolicy UpdateAutoProvPolicy
	// Update an Auto Provisioning Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateAutoProvPolicy", UpdateAutoProvPolicy)
	// swagger:route POST /auth/ctrl/ShowAutoProvPolicy AutoProvPolicy ShowAutoProvPolicy
	// Show Auto Provisioning Policies.
	//  Any fields specified will be used to filter results.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowAutoProvPolicy", ShowAutoProvPolicy)
	// swagger:route POST /auth/ctrl/AddAutoProvPolicyCloudlet AutoProvPolicyCloudlet AddAutoProvPolicyCloudlet
	// Add a Cloudlet to the Auto Provisioning Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/AddAutoProvPolicyCloudlet", AddAutoProvPolicyCloudlet)
	// swagger:route POST /auth/ctrl/RemoveAutoProvPolicyCloudlet AutoProvPolicyCloudlet RemoveAutoProvPolicyCloudlet
	// Remove a Cloudlet from the Auto Provisioning Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/RemoveAutoProvPolicyCloudlet", RemoveAutoProvPolicyCloudlet)
	// swagger:route POST /auth/ctrl/CreateCloudletPool CloudletPool CreateCloudletPool
	// Create a CloudletPool.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateCloudletPool", CreateCloudletPool)
	// swagger:route POST /auth/ctrl/DeleteCloudletPool CloudletPool DeleteCloudletPool
	// Delete a CloudletPool.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteCloudletPool", DeleteCloudletPool)
	// swagger:route POST /auth/ctrl/ShowCloudletPool CloudletPool ShowCloudletPool
	// Show CloudletPools.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowCloudletPool", ShowCloudletPool)
	// swagger:route POST /auth/ctrl/CreateCloudletPoolMember CloudletPoolMember CreateCloudletPoolMember
	// Add a Cloudlet to a CloudletPool.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreateCloudletPoolMember", CreateCloudletPoolMember)
	// swagger:route POST /auth/ctrl/DeleteCloudletPoolMember CloudletPoolMember DeleteCloudletPoolMember
	// Remove a Cloudlet from a CloudletPool.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeleteCloudletPoolMember", DeleteCloudletPoolMember)
	// swagger:route POST /auth/ctrl/ShowCloudletPoolMember CloudletPoolMember ShowCloudletPoolMember
	// Show the Cloudlet to CloudletPool relationships.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowCloudletPoolMember", ShowCloudletPoolMember)
	// swagger:route POST /auth/ctrl/ShowPoolsForCloudlet CloudletKey ShowPoolsForCloudlet
	// Show CloudletPools that have Cloudlet as a member.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowPoolsForCloudlet", ShowPoolsForCloudlet)
	// swagger:route POST /auth/ctrl/ShowCloudletsForPool CloudletPoolKey ShowCloudletsForPool
	// Show Cloudlets that belong to the Pool.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowCloudletsForPool", ShowCloudletsForPool)
	// swagger:route POST /auth/ctrl/RunCommand ExecRequest RunCommand
	// Run a Command or Shell on a container.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/RunCommand", RunCommand)
	// swagger:route POST /auth/ctrl/RunConsole ExecRequest RunConsole
	// Run console on a VM.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/RunConsole", RunConsole)
	// swagger:route POST /auth/ctrl/ShowLogs ExecRequest ShowLogs
	// View logs for AppInst.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowLogs", ShowLogs)
	// swagger:route POST /auth/ctrl/ShowNode Node ShowNode
	// Show all Nodes connected to all Controllers.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowNode", ShowNode)
	// swagger:route POST /auth/ctrl/CreatePrivacyPolicy PrivacyPolicy CreatePrivacyPolicy
	// Create a Privacy Policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/CreatePrivacyPolicy", CreatePrivacyPolicy)
	// swagger:route POST /auth/ctrl/DeletePrivacyPolicy PrivacyPolicy DeletePrivacyPolicy
	// Delete a Privacy policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/DeletePrivacyPolicy", DeletePrivacyPolicy)
	// swagger:route POST /auth/ctrl/UpdatePrivacyPolicy PrivacyPolicy UpdatePrivacyPolicy
	// Update a Privacy policy.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdatePrivacyPolicy", UpdatePrivacyPolicy)
	// swagger:route POST /auth/ctrl/ShowPrivacyPolicy PrivacyPolicy ShowPrivacyPolicy
	// Show Privacy Policies.
	//  Any fields specified will be used to filter results.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowPrivacyPolicy", ShowPrivacyPolicy)
	// swagger:route POST /auth/ctrl/ShowCloudletRefs CloudletRefs ShowCloudletRefs
	// Show CloudletRefs (debug only).
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowCloudletRefs", ShowCloudletRefs)
	// swagger:route POST /auth/ctrl/ShowClusterRefs ClusterRefs ShowClusterRefs
	// Show ClusterRefs (debug only).
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowClusterRefs", ShowClusterRefs)
	// swagger:route POST /auth/ctrl/UpdateSettings Settings UpdateSettings
	// Update settings.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/UpdateSettings", UpdateSettings)
	// swagger:route POST /auth/ctrl/ResetSettings Settings ResetSettings
	// Reset all settings to their defaults.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ResetSettings", ResetSettings)
	// swagger:route POST /auth/ctrl/ShowSettings Settings ShowSettings
	// Show settings.
	// Security:
	//   Bearer:
	// responses:
	//   200: success
	//   400: badRequest
	//   403: forbidden
	//   404: notFound
	group.Match([]string{method}, "/ctrl/ShowSettings", ShowSettings)
}
