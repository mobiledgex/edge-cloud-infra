// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: settings.proto

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
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func UpdateSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	resp, err := UpdateSettingsObj(ctx, rc, &in.Settings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func UpdateSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Settings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
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
	api := edgeproto.NewSettingsApiClient(rc.conn)
	return api.UpdateSettings(ctx, obj)
}

func ResetSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	resp, err := ResetSettingsObj(ctx, rc, &in.Settings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func ResetSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Settings) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
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
	api := edgeproto.NewSettingsApiClient(rc.conn)
	return api.ResetSettings(ctx, obj)
}

func ShowSettings(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionSettings{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	resp, err := ShowSettingsObj(ctx, rc, &in.Settings)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func ShowSettingsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Settings) (*edgeproto.Settings, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionView); err != nil {
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
	api := edgeproto.NewSettingsApiClient(rc.conn)
	return api.ShowSettings(ctx, obj)
}
