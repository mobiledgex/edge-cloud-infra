// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ratelimit.proto

package orm

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
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

func ShowRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionRateLimitSettings{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.RateLimitSettings.GetKey().GetTags())

	err = ShowRateLimitSettingsStream(ctx, rc, &in.RateLimitSettings, func(res *edgeproto.RateLimitSettings) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

func ShowRateLimitSettingsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.RateLimitSettings, cb func(res *edgeproto.RateLimitSettings) error) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceConfig, ActionView)
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	stream, err := api.ShowRateLimitSettings(ctx, obj)
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
			if !authz.Ok("") {
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

func ShowRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.RateLimitSettings) ([]edgeproto.RateLimitSettings, error) {
	arr := []edgeproto.RateLimitSettings{}
	err := ShowRateLimitSettingsStream(ctx, rc, obj, func(res *edgeproto.RateLimitSettings) error {
		arr = append(arr, *res)
		return nil
	})
	return arr, err
}

func CreateFlowRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlowRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.FlowRateLimitSettings.GetKey().GetTags())
	resp, err := CreateFlowRateLimitSettingsObj(ctx, rc, &in.FlowRateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func CreateFlowRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateFlowRateLimitSettings(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	return api.CreateFlowRateLimitSettings(ctx, obj)
}

func UpdateFlowRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlowRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.FlowRateLimitSettings.GetKey().GetTags())
	resp, err := UpdateFlowRateLimitSettingsObj(ctx, rc, &in.FlowRateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func UpdateFlowRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateFlowRateLimitSettings(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	return api.UpdateFlowRateLimitSettings(ctx, obj)
}

func DeleteFlowRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlowRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.FlowRateLimitSettings.GetKey().GetTags())
	resp, err := DeleteFlowRateLimitSettingsObj(ctx, rc, &in.FlowRateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func DeleteFlowRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.FlowRateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteFlowRateLimitSettings(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	return api.DeleteFlowRateLimitSettings(ctx, obj)
}

func ShowFlowRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlowRateLimitSettings{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.FlowRateLimitSettings.GetKey().GetTags())

	err = ShowFlowRateLimitSettingsStream(ctx, rc, &in.FlowRateLimitSettings, func(res *edgeproto.FlowRateLimitSettings) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

func ShowFlowRateLimitSettingsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.FlowRateLimitSettings, cb func(res *edgeproto.FlowRateLimitSettings) error) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceConfig, ActionView)
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	stream, err := api.ShowFlowRateLimitSettings(ctx, obj)
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
			if !authz.Ok("") {
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

func ShowFlowRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.FlowRateLimitSettings) ([]edgeproto.FlowRateLimitSettings, error) {
	arr := []edgeproto.FlowRateLimitSettings{}
	err := ShowFlowRateLimitSettingsStream(ctx, rc, obj, func(res *edgeproto.FlowRateLimitSettings) error {
		arr = append(arr, *res)
		return nil
	})
	return arr, err
}

func CreateMaxReqsRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionMaxReqsRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.MaxReqsRateLimitSettings.GetKey().GetTags())
	resp, err := CreateMaxReqsRateLimitSettingsObj(ctx, rc, &in.MaxReqsRateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func CreateMaxReqsRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateMaxReqsRateLimitSettings(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	return api.CreateMaxReqsRateLimitSettings(ctx, obj)
}

func UpdateMaxReqsRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionMaxReqsRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.MaxReqsRateLimitSettings.GetKey().GetTags())
	resp, err := UpdateMaxReqsRateLimitSettingsObj(ctx, rc, &in.MaxReqsRateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func UpdateMaxReqsRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateMaxReqsRateLimitSettings(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	return api.UpdateMaxReqsRateLimitSettings(ctx, obj)
}

func DeleteMaxReqsRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionMaxReqsRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.MaxReqsRateLimitSettings.GetKey().GetTags())
	resp, err := DeleteMaxReqsRateLimitSettingsObj(ctx, rc, &in.MaxReqsRateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func DeleteMaxReqsRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.MaxReqsRateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteMaxReqsRateLimitSettings(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	return api.DeleteMaxReqsRateLimitSettings(ctx, obj)
}

func ShowMaxReqsRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionMaxReqsRateLimitSettings{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.MaxReqsRateLimitSettings.GetKey().GetTags())

	err = ShowMaxReqsRateLimitSettingsStream(ctx, rc, &in.MaxReqsRateLimitSettings, func(res *edgeproto.MaxReqsRateLimitSettings) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

func ShowMaxReqsRateLimitSettingsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.MaxReqsRateLimitSettings, cb func(res *edgeproto.MaxReqsRateLimitSettings) error) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceConfig, ActionView)
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
	api := edgeproto.NewRateLimitSettingsApiClient(rc.conn)
	stream, err := api.ShowMaxReqsRateLimitSettings(ctx, obj)
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
			if !authz.Ok("") {
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

func ShowMaxReqsRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.MaxReqsRateLimitSettings) ([]edgeproto.MaxReqsRateLimitSettings, error) {
	arr := []edgeproto.MaxReqsRateLimitSettings{}
	err := ShowMaxReqsRateLimitSettingsStream(ctx, rc, obj, func(res *edgeproto.MaxReqsRateLimitSettings) error {
		arr = append(arr, *res)
		return nil
	})
	return arr, err
}
