// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app.proto

package orm

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"google.golang.org/grpc/status"
	"io"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func CreateApp(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionApp{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.App.GetKey().GetTags())
	span.SetTag("org", in.App.Key.Organization)
	resp, err := CreateAppObj(ctx, rc, &in.App)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func CreateAppObj(ctx context.Context, rc *RegionContext, obj *edgeproto.App) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateApp(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authzCreateApp(ctx, rc.region, rc.username, obj,
			ResourceApps, ActionManage); err != nil {
			return nil, err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return nil, err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAppApiClient(rc.conn)
	return api.CreateApp(ctx, obj)
}

func DeleteApp(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionApp{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.App.GetKey().GetTags())
	span.SetTag("org", in.App.Key.Organization)
	resp, err := DeleteAppObj(ctx, rc, &in.App)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func DeleteAppObj(ctx context.Context, rc *RegionContext, obj *edgeproto.App) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteApp(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceApps, ActionManage); err != nil {
			return nil, err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return nil, err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAppApiClient(rc.conn)
	return api.DeleteApp(ctx, obj)
}

func UpdateApp(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionApp{}
	dat, err := ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.App.GetKey().GetTags())
	span.SetTag("org", in.App.Key.Organization)
	err = ormutil.SetRegionObjFields(dat, &in)
	if err != nil {
		return err
	}
	resp, err := UpdateAppObj(ctx, rc, &in.App)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func UpdateAppObj(ctx context.Context, rc *RegionContext, obj *edgeproto.App) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateApp(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authzUpdateApp(ctx, rc.region, rc.username, obj,
			ResourceApps, ActionManage); err != nil {
			return nil, err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return nil, err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAppApiClient(rc.conn)
	return api.UpdateApp(ctx, obj)
}

func ShowApp(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionApp{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.App.GetKey().GetTags())
	span.SetTag("org", in.App.Key.Organization)

	err = ShowAppStream(ctx, rc, &in.App, func(res *edgeproto.App) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

func ShowAppStream(ctx context.Context, rc *RegionContext, obj *edgeproto.App, cb func(res *edgeproto.App) error) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceApps, ActionView)
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
	api := edgeproto.NewAppApiClient(rc.conn)
	stream, err := api.ShowApp(ctx, obj)
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
			if !authz.Ok(res.Key.Organization) {
				continue
			}
		}
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func ShowAppObj(ctx context.Context, rc *RegionContext, obj *edgeproto.App) ([]edgeproto.App, error) {
	arr := []edgeproto.App{}
	err := ShowAppStream(ctx, rc, obj, func(res *edgeproto.App) error {
		arr = append(arr, *res)
		return nil
	})
	return arr, err
}

func AddAppAutoProvPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppAutoProvPolicy{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.AppAutoProvPolicy.AppKey.Organization)
	resp, err := AddAppAutoProvPolicyObj(ctx, rc, &in.AppAutoProvPolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func AddAppAutoProvPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForAddAppAutoProvPolicy(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.AppKey.Organization,
			ResourceApps, ActionManage); err != nil {
			return nil, err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return nil, err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAppApiClient(rc.conn)
	return api.AddAppAutoProvPolicy(ctx, obj)
}

func RemoveAppAutoProvPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppAutoProvPolicy{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.AppAutoProvPolicy.AppKey.Organization)
	resp, err := RemoveAppAutoProvPolicyObj(ctx, rc, &in.AppAutoProvPolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func RemoveAppAutoProvPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppAutoProvPolicy) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRemoveAppAutoProvPolicy(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.AppKey.Organization,
			ResourceApps, ActionManage); err != nil {
			return nil, err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return nil, err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAppApiClient(rc.conn)
	return api.RemoveAppAutoProvPolicy(ctx, obj)
}

func AddAppAlertPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppAlertPolicy{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.AppAlertPolicy.AppKey.Organization)
	resp, err := AddAppAlertPolicyObj(ctx, rc, &in.AppAlertPolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func AddAppAlertPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppAlertPolicy) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForAddAppAlertPolicy(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.AppKey.Organization,
			ResourceApps, ActionManage); err != nil {
			return nil, err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return nil, err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAppApiClient(rc.conn)
	return api.AddAppAlertPolicy(ctx, obj)
}

func RemoveAppAlertPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppAlertPolicy{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.AppAlertPolicy.AppKey.Organization)
	resp, err := RemoveAppAlertPolicyObj(ctx, rc, &in.AppAlertPolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func RemoveAppAlertPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppAlertPolicy) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRemoveAppAlertPolicy(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.AppKey.Organization,
			ResourceApps, ActionManage); err != nil {
			return nil, err
		}
	}
	if rc.conn == nil {
		conn, err := connectController(ctx, rc.region)
		if err != nil {
			return nil, err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewAppApiClient(rc.conn)
	return api.RemoveAppAlertPolicy(ctx, obj)
}

func ShowCloudletsForAppDeployment(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionDeploymentCloudletRequest{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)

	err = ShowCloudletsForAppDeploymentStream(ctx, rc, &in.DeploymentCloudletRequest, func(res *edgeproto.CloudletKey) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

type ShowCloudletsForAppDeploymentAuthz interface {
	Ok(obj *edgeproto.CloudletKey) (bool, bool)
	Filter(obj *edgeproto.CloudletKey)
}

func ShowCloudletsForAppDeploymentStream(ctx context.Context, rc *RegionContext, obj *edgeproto.DeploymentCloudletRequest, cb func(res *edgeproto.CloudletKey) error) error {
	var authz ShowCloudletsForAppDeploymentAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = newShowCloudletsForAppDeploymentAuthz(ctx, rc.region, rc.username, ResourceCloudlets, ActionView)
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
	api := edgeproto.NewAppApiClient(rc.conn)
	stream, err := api.ShowCloudletsForAppDeployment(ctx, obj)
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
			authzOk, filterOutput := authz.Ok(res)
			if !authzOk {
				continue
			}
			if filterOutput {
				authz.Filter(res)
			}
		}
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func ShowCloudletsForAppDeploymentObj(ctx context.Context, rc *RegionContext, obj *edgeproto.DeploymentCloudletRequest) ([]edgeproto.CloudletKey, error) {
	arr := []edgeproto.CloudletKey{}
	err := ShowCloudletsForAppDeploymentStream(ctx, rc, obj, func(res *edgeproto.CloudletKey) error {
		arr = append(arr, *res)
		return nil
	})
	return arr, err
}
