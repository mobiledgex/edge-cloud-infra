// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoscalepolicy.proto

package orm

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/labstack/echo"
import "net/http"
import "context"
import "io"
import "github.com/mobiledgex/edge-cloud/log"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import "google.golang.org/grpc/status"
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

func CreateAutoScalePolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AutoScalePolicy.Key.Developer)
	resp, err := CreateAutoScalePolicyObj(ctx, rc, &in.AutoScalePolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func CreateAutoScalePolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AutoScalePolicy) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, obj.Key.Developer,
		ResourceDeveloperPolicy, ActionManage) {
		return nil, echo.ErrForbidden
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
	api := edgeproto.NewAutoScalePolicyApiClient(rc.conn)
	return api.CreateAutoScalePolicy(ctx, obj)
}

func DeleteAutoScalePolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AutoScalePolicy.Key.Developer)
	resp, err := DeleteAutoScalePolicyObj(ctx, rc, &in.AutoScalePolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func DeleteAutoScalePolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AutoScalePolicy) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, obj.Key.Developer,
		ResourceDeveloperPolicy, ActionManage) {
		return nil, echo.ErrForbidden
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
	api := edgeproto.NewAutoScalePolicyApiClient(rc.conn)
	return api.DeleteAutoScalePolicy(ctx, obj)
}

func UpdateAutoScalePolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AutoScalePolicy.Key.Developer)
	resp, err := UpdateAutoScalePolicyObj(ctx, rc, &in.AutoScalePolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func UpdateAutoScalePolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AutoScalePolicy) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, obj.Key.Developer,
		ResourceDeveloperPolicy, ActionManage) {
		return nil, echo.ErrForbidden
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
	api := edgeproto.NewAutoScalePolicyApiClient(rc.conn)
	return api.UpdateAutoScalePolicy(ctx, obj)
}

func ShowAutoScalePolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AutoScalePolicy.Key.Developer)

	err = ShowAutoScalePolicyStream(ctx, rc, &in.AutoScalePolicy, func(res *edgeproto.AutoScalePolicy) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowAutoScalePolicyStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AutoScalePolicy, cb func(res *edgeproto.AutoScalePolicy)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceDeveloperPolicy, ActionView)
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
	api := edgeproto.NewAutoScalePolicyApiClient(rc.conn)
	stream, err := api.ShowAutoScalePolicy(ctx, obj)
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
			if !authz.Ok(res.Key.Developer) {
				continue
			}
		}
		cb(res)
	}
	return nil
}

func ShowAutoScalePolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AutoScalePolicy) ([]edgeproto.AutoScalePolicy, error) {
	arr := []edgeproto.AutoScalePolicy{}
	err := ShowAutoScalePolicyStream(ctx, rc, obj, func(res *edgeproto.AutoScalePolicy) {
		arr = append(arr, *res)
	})
	return arr, err
}
