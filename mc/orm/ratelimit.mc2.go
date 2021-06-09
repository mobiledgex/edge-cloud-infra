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

func CreateRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.RateLimitSettings.GetKey().GetTags())
	resp, err := CreateRateLimitSettingsObj(ctx, rc, &in.RateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func CreateRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateRateLimitSettings(); err != nil {
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
	return api.CreateRateLimitSettings(ctx, obj)
}

func UpdateRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.RateLimitSettings.GetKey().GetTags())
	resp, err := UpdateRateLimitSettingsObj(ctx, rc, &in.RateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func UpdateRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateRateLimitSettings(); err != nil {
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
	return api.UpdateRateLimitSettings(ctx, obj)
}

func DeleteRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.RateLimitSettings.GetKey().GetTags())
	resp, err := DeleteRateLimitSettingsObj(ctx, rc, &in.RateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func DeleteRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteRateLimitSettings(); err != nil {
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
	return api.DeleteRateLimitSettings(ctx, obj)
}

func ResetRateLimitSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionRateLimitSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.RateLimitSettings.GetKey().GetTags())
	resp, err := ResetRateLimitSettingsObj(ctx, rc, &in.RateLimitSettings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return setReply(c, resp)
}

func ResetRateLimitSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.RateLimitSettings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForResetRateLimitSettings(); err != nil {
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
	return api.ResetRateLimitSettings(ctx, obj)
}

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
