// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: app_inst.proto

package orm

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/labstack/echo"
import "context"
import "io"
import "github.com/mobiledgex/edge-cloud/log"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var streamAppInst = &StreamObj{}

func StreamAppInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AppInst.Key.AppKey.DeveloperKey.Name)

	streamer := streamAppInst.Get(in.AppInst.Key)
	if streamer != nil {
		payload := ormapi.StreamPayload{}
		streamCh := streamer.Subscribe()
		serverClosed := make(chan bool)
		go func() {
			for streamMsg := range streamCh {
				switch out := streamMsg.(type) {
				case string:
					payload.Data = &edgeproto.Result{Message: out}
					WriteStream(c, &payload)
				case error:
					WriteError(c, out)
				default:
					WriteError(c, fmt.Errorf("Unsupported message type received: %v", streamMsg))
				}
			}
			CloseConn(c)
			serverClosed <- true
		}()
		// Wait for client/server to close
		// * Server closure is set via above serverClosed flag
		// * Client closure is sent from client via a message
		WaitForConnClose(c, serverClosed)
		streamer.Unsubscribe(streamCh)
	} else {
		WriteError(c, fmt.Errorf("Key doesn't exist"))
		CloseConn(c)
	}
	return nil
}

func CreateAppInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AppInst.Key.AppKey.DeveloperKey.Name)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = CreateAppInstStream(ctx, rc, &in.AppInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamAppInst.Add(in.AppInst.Key, streamer)
			streamAdded = true
		}
		payload := ormapi.StreamPayload{}
		payload.Data = res
		streamer.Publish(res.Message)
		WriteStream(c, &payload)
	})
	if err != nil {
		streamer.Publish(err)
		WriteError(c, err)
	}
	if streamAdded {
		streamAppInst.Remove(in.AppInst.Key, streamer)
	}
	return nil
}

func CreateAppInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst, cb func(res *edgeproto.Result)) error {
	if !rc.skipAuthz {
		if err := authzCreateAppInst(ctx, rc.region, rc.username, obj,
			ResourceAppInsts, ActionManage); err != nil {
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
	api := edgeproto.NewAppInstApiClient(rc.conn)
	stream, err := api.CreateAppInst(ctx, obj)
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

func CreateAppInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := CreateAppInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func DeleteAppInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AppInst.Key.AppKey.DeveloperKey.Name)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = DeleteAppInstStream(ctx, rc, &in.AppInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamAppInst.Add(in.AppInst.Key, streamer)
			streamAdded = true
		}
		payload := ormapi.StreamPayload{}
		payload.Data = res
		streamer.Publish(res.Message)
		WriteStream(c, &payload)
	})
	if err != nil {
		streamer.Publish(err)
		WriteError(c, err)
	}
	if streamAdded {
		streamAppInst.Remove(in.AppInst.Key, streamer)
	}
	return nil
}

func DeleteAppInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst, cb func(res *edgeproto.Result)) error {
	if !rc.skipAuthz && !authorized(ctx, rc.username, obj.Key.AppKey.DeveloperKey.Name,
		ResourceAppInsts, ActionManage) {
		return echo.ErrForbidden
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
	api := edgeproto.NewAppInstApiClient(rc.conn)
	stream, err := api.DeleteAppInst(ctx, obj)
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

func DeleteAppInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := DeleteAppInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func RefreshAppInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AppInst.Key.AppKey.DeveloperKey.Name)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = RefreshAppInstStream(ctx, rc, &in.AppInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamAppInst.Add(in.AppInst.Key, streamer)
			streamAdded = true
		}
		payload := ormapi.StreamPayload{}
		payload.Data = res
		streamer.Publish(res.Message)
		WriteStream(c, &payload)
	})
	if err != nil {
		streamer.Publish(err)
		WriteError(c, err)
	}
	if streamAdded {
		streamAppInst.Remove(in.AppInst.Key, streamer)
	}
	return nil
}

func RefreshAppInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst, cb func(res *edgeproto.Result)) error {
	if !rc.skipAuthz && !authorized(ctx, rc.username, obj.Key.AppKey.DeveloperKey.Name,
		ResourceAppInsts, ActionManage) {
		return echo.ErrForbidden
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
	api := edgeproto.NewAppInstApiClient(rc.conn)
	stream, err := api.RefreshAppInst(ctx, obj)
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

func RefreshAppInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := RefreshAppInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func UpdateAppInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AppInst.Key.AppKey.DeveloperKey.Name)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = UpdateAppInstStream(ctx, rc, &in.AppInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamAppInst.Add(in.AppInst.Key, streamer)
			streamAdded = true
		}
		payload := ormapi.StreamPayload{}
		payload.Data = res
		streamer.Publish(res.Message)
		WriteStream(c, &payload)
	})
	if err != nil {
		streamer.Publish(err)
		WriteError(c, err)
	}
	if streamAdded {
		streamAppInst.Remove(in.AppInst.Key, streamer)
	}
	return nil
}

func UpdateAppInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst, cb func(res *edgeproto.Result)) error {
	if !rc.skipAuthz && !authorized(ctx, rc.username, obj.Key.AppKey.DeveloperKey.Name,
		ResourceAppInsts, ActionManage) {
		return echo.ErrForbidden
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
	api := edgeproto.NewAppInstApiClient(rc.conn)
	stream, err := api.UpdateAppInst(ctx, obj)
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

func UpdateAppInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := UpdateAppInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func ShowAppInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AppInst.Key.AppKey.DeveloperKey.Name)

	err = ShowAppInstStream(ctx, rc, &in.AppInst, func(res *edgeproto.AppInst) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowAppInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst, cb func(res *edgeproto.AppInst)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceAppInsts, ActionView)
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
	api := edgeproto.NewAppInstApiClient(rc.conn)
	stream, err := api.ShowAppInst(ctx, obj)
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
			if !authz.Ok(res.Key.AppKey.DeveloperKey.Name) {
				continue
			}
		}
		cb(res)
	}
	return nil
}

func ShowAppInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst) ([]edgeproto.AppInst, error) {
	arr := []edgeproto.AppInst{}
	err := ShowAppInstStream(ctx, rc, obj, func(res *edgeproto.AppInst) {
		arr = append(arr, *res)
	})
	return arr, err
}

func SetAppInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.AppInst.Key.AppKey.DeveloperKey.Name)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = SetAppInstStream(ctx, rc, &in.AppInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamAppInst.Add(in.AppInst.Key, streamer)
			streamAdded = true
		}
		payload := ormapi.StreamPayload{}
		payload.Data = res
		streamer.Publish(res.Message)
		WriteStream(c, &payload)
	})
	if err != nil {
		streamer.Publish(err)
		WriteError(c, err)
	}
	if streamAdded {
		streamAppInst.Remove(in.AppInst.Key, streamer)
	}
	return nil
}

func SetAppInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst, cb func(res *edgeproto.Result)) error {
	if !rc.skipAuthz && !authorized(ctx, rc.username, obj.Key.AppKey.DeveloperKey.Name,
		ResourceAppInsts, ActionManage) {
		return echo.ErrForbidden
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
	api := edgeproto.NewAppInstApiClient(rc.conn)
	stream, err := api.SetAppInst(ctx, obj)
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

func SetAppInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := SetAppInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}
