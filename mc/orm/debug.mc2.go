// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: debug.proto

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
	"io"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func EnableDebugLevels(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionDebugRequest{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)

	err = EnableDebugLevelsStream(ctx, rc, &in.DebugRequest, func(res *edgeproto.DebugReply) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func EnableDebugLevelsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest, cb func(res *edgeproto.DebugReply)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForEnableDebugLevels(); err != nil {
		return err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
			return err
		}
	}
	if rc.conn == nil {
		conn, err := connectNotifyRoot(ctx)
		if err != nil {
			return err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewDebugApiClient(rc.conn)
	stream, err := api.EnableDebugLevels(ctx, obj)
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
		cb(res)
	}
	return nil
}

func EnableDebugLevelsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest) ([]edgeproto.DebugReply, error) {
	arr := []edgeproto.DebugReply{}
	err := EnableDebugLevelsStream(ctx, rc, obj, func(res *edgeproto.DebugReply) {
		arr = append(arr, *res)
	})
	return arr, err
}

func DisableDebugLevels(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionDebugRequest{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)

	err = DisableDebugLevelsStream(ctx, rc, &in.DebugRequest, func(res *edgeproto.DebugReply) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func DisableDebugLevelsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest, cb func(res *edgeproto.DebugReply)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDisableDebugLevels(); err != nil {
		return err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
			return err
		}
	}
	if rc.conn == nil {
		conn, err := connectNotifyRoot(ctx)
		if err != nil {
			return err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewDebugApiClient(rc.conn)
	stream, err := api.DisableDebugLevels(ctx, obj)
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
		cb(res)
	}
	return nil
}

func DisableDebugLevelsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest) ([]edgeproto.DebugReply, error) {
	arr := []edgeproto.DebugReply{}
	err := DisableDebugLevelsStream(ctx, rc, obj, func(res *edgeproto.DebugReply) {
		arr = append(arr, *res)
	})
	return arr, err
}

func ShowDebugLevels(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionDebugRequest{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)

	err = ShowDebugLevelsStream(ctx, rc, &in.DebugRequest, func(res *edgeproto.DebugReply) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowDebugLevelsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest, cb func(res *edgeproto.DebugReply)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionView); err != nil {
			return err
		}
	}
	if rc.conn == nil {
		conn, err := connectNotifyRoot(ctx)
		if err != nil {
			return err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewDebugApiClient(rc.conn)
	stream, err := api.ShowDebugLevels(ctx, obj)
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
		cb(res)
	}
	return nil
}

func ShowDebugLevelsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest) ([]edgeproto.DebugReply, error) {
	arr := []edgeproto.DebugReply{}
	err := ShowDebugLevelsStream(ctx, rc, obj, func(res *edgeproto.DebugReply) {
		arr = append(arr, *res)
	})
	return arr, err
}

func RunDebug(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionDebugRequest{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)

	err = RunDebugStream(ctx, rc, &in.DebugRequest, func(res *edgeproto.DebugReply) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func RunDebugStream(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest, cb func(res *edgeproto.DebugReply)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRunDebug(); err != nil {
		return err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceConfig, ActionManage); err != nil {
			return err
		}
	}
	if rc.conn == nil {
		conn, err := connectNotifyRoot(ctx)
		if err != nil {
			return err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}
	api := edgeproto.NewDebugApiClient(rc.conn)
	stream, err := api.RunDebug(ctx, obj)
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
		cb(res)
	}
	return nil
}

func RunDebugObj(ctx context.Context, rc *RegionContext, obj *edgeproto.DebugRequest) ([]edgeproto.DebugReply, error) {
	arr := []edgeproto.DebugReply{}
	err := RunDebugStream(ctx, rc, obj, func(res *edgeproto.DebugReply) {
		arr = append(arr, *res)
	})
	return arr, err
}
