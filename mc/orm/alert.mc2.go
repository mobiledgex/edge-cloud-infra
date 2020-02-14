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
	// Following are `Flavor.fields` values to be used to specify which fields to update:
	// ```
	// FlavorFieldKey = 2
	// FlavorFieldKeyName = 2.1
	// FlavorFieldRam = 3
	// FlavorFieldVcpus = 4
	// FlavorFieldDisk = 5
	// FlavorFieldOptResMap = 6
	// FlavorFieldOptResMapKey = 6.1
	// FlavorFieldOptResMapValue = 6.2
	// ```
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
	// Following are `App.fields` values to be used to specify which fields to update:
	// ```
	// AppFieldKey = 2
	// AppFieldKeyDeveloperKey = 2.1
	// AppFieldKeyDeveloperKeyName = 2.1.2
	// AppFieldKeyName = 2.2
	// AppFieldKeyVersion = 2.3
	// AppFieldImagePath = 4
	// AppFieldImageType = 5
	// AppFieldAccessPorts = 7
	// AppFieldDefaultFlavor = 9
	// AppFieldDefaultFlavorName = 9.1
	// AppFieldAuthPublicKey = 12
	// AppFieldCommand = 13
	// AppFieldAnnotations = 14
	// AppFieldDeployment = 15
	// AppFieldDeploymentManifest = 16
	// AppFieldDeploymentGenerator = 17
	// AppFieldAndroidPackageName = 18
	// AppFieldDelOpt = 20
	// AppFieldConfigs = 21
	// AppFieldConfigsKind = 21.1
	// AppFieldConfigsConfig = 21.2
	// AppFieldScaleWithCluster = 22
	// AppFieldInternalPorts = 23
	// AppFieldRevision = 24
	// AppFieldOfficialFqdn = 25
	// AppFieldMd5Sum = 26
	// AppFieldDefaultSharedVolumeSize = 27
	// AppFieldAutoProvPolicy = 28
	// AppFieldAccessType = 29
	// AppFieldDefaultPrivacyPolicy = 30
	// ```
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
	// Following are `ResTagTable.fields` values to be used to specify which fields to update:
	// ```
	// ResTagTableFieldKey = 2
	// ResTagTableFieldKeyName = 2.1
	// ResTagTableFieldKeyOperatorKey = 2.2
	// ResTagTableFieldKeyOperatorKeyName = 2.2.1
	// ResTagTableFieldTags = 3
	// ResTagTableFieldTagsKey = 3.1
	// ResTagTableFieldTagsValue = 3.2
	// ResTagTableFieldAzone = 4
	// ```
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
	// Following are `Cloudlet.fields` values to be used to specify which fields to update:
	// ```
	// CloudletFieldKey = 2
	// CloudletFieldKeyOperatorKey = 2.1
	// CloudletFieldKeyOperatorKeyName = 2.1.1
	// CloudletFieldKeyName = 2.2
	// CloudletFieldLocation = 5
	// CloudletFieldLocationLatitude = 5.1
	// CloudletFieldLocationLongitude = 5.2
	// CloudletFieldLocationHorizontalAccuracy = 5.3
	// CloudletFieldLocationVerticalAccuracy = 5.4
	// CloudletFieldLocationAltitude = 5.5
	// CloudletFieldLocationCourse = 5.6
	// CloudletFieldLocationSpeed = 5.7
	// CloudletFieldLocationTimestamp = 5.8
	// CloudletFieldLocationTimestampSeconds = 5.8.1
	// CloudletFieldLocationTimestampNanos = 5.8.2
	// CloudletFieldIpSupport = 6
	// CloudletFieldStaticIps = 7
	// CloudletFieldNumDynamicIps = 8
	// CloudletFieldTimeLimits = 9
	// CloudletFieldTimeLimitsCreateClusterInstTimeout = 9.1
	// CloudletFieldTimeLimitsUpdateClusterInstTimeout = 9.2
	// CloudletFieldTimeLimitsDeleteClusterInstTimeout = 9.3
	// CloudletFieldTimeLimitsCreateAppInstTimeout = 9.4
	// CloudletFieldTimeLimitsUpdateAppInstTimeout = 9.5
	// CloudletFieldTimeLimitsDeleteAppInstTimeout = 9.6
	// CloudletFieldErrors = 10
	// CloudletFieldStatus = 11
	// CloudletFieldStatusTaskNumber = 11.1
	// CloudletFieldStatusMaxTasks = 11.2
	// CloudletFieldStatusTaskName = 11.3
	// CloudletFieldStatusStepName = 11.4
	// CloudletFieldState = 12
	// CloudletFieldCrmOverride = 13
	// CloudletFieldDeploymentLocal = 14
	// CloudletFieldPlatformType = 15
	// CloudletFieldNotifySrvAddr = 16
	// CloudletFieldFlavor = 17
	// CloudletFieldFlavorName = 17.1
	// CloudletFieldPhysicalName = 18
	// CloudletFieldEnvVar = 19
	// CloudletFieldEnvVarKey = 19.1
	// CloudletFieldEnvVarValue = 19.2
	// CloudletFieldContainerVersion = 20
	// CloudletFieldConfig = 21
	// CloudletFieldConfigContainerRegistryPath = 21.1
	// CloudletFieldConfigCloudletVmImagePath = 21.2
	// CloudletFieldConfigNotifyCtrlAddrs = 21.3
	// CloudletFieldConfigVaultAddr = 21.4
	// CloudletFieldConfigTlsCertFile = 21.5
	// CloudletFieldConfigEnvVar = 21.6
	// CloudletFieldConfigEnvVarKey = 21.6.1
	// CloudletFieldConfigEnvVarValue = 21.6.2
	// CloudletFieldConfigPlatformTag = 21.8
	// CloudletFieldConfigTestMode = 21.9
	// CloudletFieldConfigSpan = 21.10
	// CloudletFieldConfigCleanupMode = 21.11
	// CloudletFieldConfigRegion = 21.12
	// CloudletFieldResTagMap = 22
	// CloudletFieldResTagMapKey = 22.1
	// CloudletFieldResTagMapValue = 22.2
	// CloudletFieldResTagMapValueName = 22.2.1
	// CloudletFieldResTagMapValueOperatorKey = 22.2.2
	// CloudletFieldResTagMapValueOperatorKeyName = 22.2.2.1
	// CloudletFieldAccessVars = 23
	// CloudletFieldAccessVarsKey = 23.1
	// CloudletFieldAccessVarsValue = 23.2
	// CloudletFieldVmImageVersion = 24
	// CloudletFieldPackageVersion = 25
	// ```
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
	// Following are `ClusterInst.fields` values to be used to specify which fields to update:
	// ```
	// ClusterInstFieldKey = 2
	// ClusterInstFieldKeyClusterKey = 2.1
	// ClusterInstFieldKeyClusterKeyName = 2.1.1
	// ClusterInstFieldKeyCloudletKey = 2.2
	// ClusterInstFieldKeyCloudletKeyOperatorKey = 2.2.1
	// ClusterInstFieldKeyCloudletKeyOperatorKeyName = 2.2.1.1
	// ClusterInstFieldKeyCloudletKeyName = 2.2.2
	// ClusterInstFieldKeyDeveloper = 2.3
	// ClusterInstFieldFlavor = 3
	// ClusterInstFieldFlavorName = 3.1
	// ClusterInstFieldLiveness = 9
	// ClusterInstFieldAuto = 10
	// ClusterInstFieldState = 4
	// ClusterInstFieldErrors = 5
	// ClusterInstFieldCrmOverride = 6
	// ClusterInstFieldIpAccess = 7
	// ClusterInstFieldAllocatedIp = 8
	// ClusterInstFieldNodeFlavor = 11
	// ClusterInstFieldDeployment = 15
	// ClusterInstFieldNumMasters = 13
	// ClusterInstFieldNumNodes = 14
	// ClusterInstFieldStatus = 16
	// ClusterInstFieldStatusTaskNumber = 16.1
	// ClusterInstFieldStatusMaxTasks = 16.2
	// ClusterInstFieldStatusTaskName = 16.3
	// ClusterInstFieldStatusStepName = 16.4
	// ClusterInstFieldExternalVolumeSize = 17
	// ClusterInstFieldAutoScalePolicy = 18
	// ClusterInstFieldAvailabilityZone = 19
	// ClusterInstFieldImageName = 20
	// ClusterInstFieldReservable = 21
	// ClusterInstFieldReservedBy = 22
	// ClusterInstFieldSharedVolumeSize = 23
	// ClusterInstFieldPrivacyPolicy = 24
	// ClusterInstFieldMasterNodeFlavor = 25
	// ```
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
	// Following are `AppInst.fields` values to be used to specify which fields to update:
	// ```
	// AppInstFieldKey = 2
	// AppInstFieldKeyAppKey = 2.1
	// AppInstFieldKeyAppKeyDeveloperKey = 2.1.1
	// AppInstFieldKeyAppKeyDeveloperKeyName = 2.1.1.2
	// AppInstFieldKeyAppKeyName = 2.1.2
	// AppInstFieldKeyAppKeyVersion = 2.1.3
	// AppInstFieldKeyClusterInstKey = 2.4
	// AppInstFieldKeyClusterInstKeyClusterKey = 2.4.1
	// AppInstFieldKeyClusterInstKeyClusterKeyName = 2.4.1.1
	// AppInstFieldKeyClusterInstKeyCloudletKey = 2.4.2
	// AppInstFieldKeyClusterInstKeyCloudletKeyOperatorKey = 2.4.2.1
	// AppInstFieldKeyClusterInstKeyCloudletKeyOperatorKeyName = 2.4.2.1.1
	// AppInstFieldKeyClusterInstKeyCloudletKeyName = 2.4.2.2
	// AppInstFieldKeyClusterInstKeyDeveloper = 2.4.3
	// AppInstFieldCloudletLoc = 3
	// AppInstFieldCloudletLocLatitude = 3.1
	// AppInstFieldCloudletLocLongitude = 3.2
	// AppInstFieldCloudletLocHorizontalAccuracy = 3.3
	// AppInstFieldCloudletLocVerticalAccuracy = 3.4
	// AppInstFieldCloudletLocAltitude = 3.5
	// AppInstFieldCloudletLocCourse = 3.6
	// AppInstFieldCloudletLocSpeed = 3.7
	// AppInstFieldCloudletLocTimestamp = 3.8
	// AppInstFieldCloudletLocTimestampSeconds = 3.8.1
	// AppInstFieldCloudletLocTimestampNanos = 3.8.2
	// AppInstFieldUri = 4
	// AppInstFieldLiveness = 6
	// AppInstFieldMappedPorts = 9
	// AppInstFieldMappedPortsProto = 9.1
	// AppInstFieldMappedPortsInternalPort = 9.2
	// AppInstFieldMappedPortsPublicPort = 9.3
	// AppInstFieldMappedPortsPathPrefix = 9.4
	// AppInstFieldMappedPortsFqdnPrefix = 9.5
	// AppInstFieldMappedPortsEndPort = 9.6
	// AppInstFieldFlavor = 12
	// AppInstFieldFlavorName = 12.1
	// AppInstFieldState = 14
	// AppInstFieldErrors = 15
	// AppInstFieldCrmOverride = 16
	// AppInstFieldRuntimeInfo = 17
	// AppInstFieldRuntimeInfoContainerIds = 17.1
	// AppInstFieldCreatedAt = 21
	// AppInstFieldCreatedAtSeconds = 21.1
	// AppInstFieldCreatedAtNanos = 21.2
	// AppInstFieldAutoClusterIpAccess = 22
	// AppInstFieldStatus = 23
	// AppInstFieldStatusTaskNumber = 23.1
	// AppInstFieldStatusMaxTasks = 23.2
	// AppInstFieldStatusTaskName = 23.3
	// AppInstFieldStatusStepName = 23.4
	// AppInstFieldRevision = 24
	// AppInstFieldForceUpdate = 25
	// AppInstFieldUpdateMultiple = 26
	// AppInstFieldConfigs = 27
	// AppInstFieldConfigsKind = 27.1
	// AppInstFieldConfigsConfig = 27.2
	// AppInstFieldSharedVolumeSize = 28
	// AppInstFieldHealthCheck = 29
	// AppInstFieldPrivacyPolicy = 30
	// ```
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
	// Following are `AutoScalePolicy.fields` values to be used to specify which fields to update:
	// ```
	// AutoScalePolicyFieldKey = 2
	// AutoScalePolicyFieldKeyDeveloper = 2.1
	// AutoScalePolicyFieldKeyName = 2.2
	// AutoScalePolicyFieldMinNodes = 3
	// AutoScalePolicyFieldMaxNodes = 4
	// AutoScalePolicyFieldScaleUpCpuThresh = 5
	// AutoScalePolicyFieldScaleDownCpuThresh = 6
	// AutoScalePolicyFieldTriggerTimeSec = 7
	// ```
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
	// Following are `AutoProvPolicy.fields` values to be used to specify which fields to update:
	// ```
	// AutoProvPolicyFieldKey = 2
	// AutoProvPolicyFieldKeyDeveloper = 2.1
	// AutoProvPolicyFieldKeyName = 2.2
	// AutoProvPolicyFieldDeployClientCount = 3
	// AutoProvPolicyFieldDeployIntervalCount = 4
	// AutoProvPolicyFieldCloudlets = 5
	// AutoProvPolicyFieldCloudletsKey = 5.1
	// AutoProvPolicyFieldCloudletsKeyOperatorKey = 5.1.1
	// AutoProvPolicyFieldCloudletsKeyOperatorKeyName = 5.1.1.1
	// AutoProvPolicyFieldCloudletsKeyName = 5.1.2
	// AutoProvPolicyFieldCloudletsLoc = 5.2
	// AutoProvPolicyFieldCloudletsLocLatitude = 5.2.1
	// AutoProvPolicyFieldCloudletsLocLongitude = 5.2.2
	// AutoProvPolicyFieldCloudletsLocHorizontalAccuracy = 5.2.3
	// AutoProvPolicyFieldCloudletsLocVerticalAccuracy = 5.2.4
	// AutoProvPolicyFieldCloudletsLocAltitude = 5.2.5
	// AutoProvPolicyFieldCloudletsLocCourse = 5.2.6
	// AutoProvPolicyFieldCloudletsLocSpeed = 5.2.7
	// AutoProvPolicyFieldCloudletsLocTimestamp = 5.2.8
	// AutoProvPolicyFieldCloudletsLocTimestampSeconds = 5.2.8.1
	// AutoProvPolicyFieldCloudletsLocTimestampNanos = 5.2.8.2
	// ```
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
	// Following are `PrivacyPolicy.fields` values to be used to specify which fields to update:
	// ```
	// PrivacyPolicyFieldKey = 2
	// PrivacyPolicyFieldKeyDeveloper = 2.1
	// PrivacyPolicyFieldKeyName = 2.2
	// PrivacyPolicyFieldOutboundSecurityRules = 3
	// PrivacyPolicyFieldOutboundSecurityRulesProtocol = 3.1
	// PrivacyPolicyFieldOutboundSecurityRulesPortRangeMin = 3.2
	// PrivacyPolicyFieldOutboundSecurityRulesPortRangeMax = 3.3
	// PrivacyPolicyFieldOutboundSecurityRulesRemoteCidr = 3.4
	// ```
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
	// Following are `Settings.fields` values to be used to specify which fields to update:
	// ```
	// SettingsFieldShepherdMetricsCollectionInterval = 2
	// SettingsFieldShepherdHealthCheckRetries = 3
	// SettingsFieldShepherdHealthCheckInterval = 4
	// SettingsFieldAutoDeployIntervalSec = 5
	// SettingsFieldAutoDeployOffsetSec = 6
	// SettingsFieldAutoDeployMaxIntervals = 7
	// SettingsFieldCreateAppInstTimeout = 8
	// SettingsFieldUpdateAppInstTimeout = 9
	// SettingsFieldDeleteAppInstTimeout = 10
	// SettingsFieldCreateClusterInstTimeout = 11
	// SettingsFieldUpdateClusterInstTimeout = 12
	// SettingsFieldDeleteClusterInstTimeout = 13
	// SettingsFieldMasterNodeFlavor = 14
	// SettingsFieldLoadBalancerMaxPortRange = 15
	// ```
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
