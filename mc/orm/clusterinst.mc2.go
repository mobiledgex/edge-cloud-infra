// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

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

var streamClusterInst = &StreamObj{}

func StreamClusterInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionClusterInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	streamer := streamClusterInst.Get(in.ClusterInst.Key)
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

func CreateClusterInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionClusterInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = CreateClusterInstStream(ctx, rc, &in.ClusterInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamClusterInst.Add(in.ClusterInst.Key, streamer)
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
		streamClusterInst.Remove(in.ClusterInst.Key, streamer)
	}
	return nil
}

func CreateClusterInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst, cb func(res *edgeproto.Result)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authzCreateClusterInst(ctx, rc.region, rc.username, obj,
			ResourceClusterInsts, ActionManage); err != nil {
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
	api := edgeproto.NewClusterInstApiClient(rc.conn)
	stream, err := api.CreateClusterInst(ctx, obj)
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

func CreateClusterInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := CreateClusterInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func DeleteClusterInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionClusterInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = DeleteClusterInstStream(ctx, rc, &in.ClusterInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamClusterInst.Add(in.ClusterInst.Key, streamer)
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
		streamClusterInst.Remove(in.ClusterInst.Key, streamer)
	}
	return nil
}

func DeleteClusterInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst, cb func(res *edgeproto.Result)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceClusterInsts, ActionManage); err != nil {
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
	api := edgeproto.NewClusterInstApiClient(rc.conn)
	stream, err := api.DeleteClusterInst(ctx, obj)
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

func DeleteClusterInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := DeleteClusterInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func UpdateClusterInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionClusterInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	streamer := NewStreamer()
	defer streamer.Stop()
	streamAdded := false

	err = UpdateClusterInstStream(ctx, rc, &in.ClusterInst, func(res *edgeproto.Result) {
		if !streamAdded {
			streamClusterInst.Add(in.ClusterInst.Key, streamer)
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
		streamClusterInst.Remove(in.ClusterInst.Key, streamer)
	}
	return nil
}

func UpdateClusterInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst, cb func(res *edgeproto.Result)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceClusterInsts, ActionManage); err != nil {
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
	api := edgeproto.NewClusterInstApiClient(rc.conn)
	stream, err := api.UpdateClusterInst(ctx, obj)
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

func UpdateClusterInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := UpdateClusterInstStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func ShowClusterInst(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionClusterInst{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	err = ShowClusterInstStream(ctx, rc, &in.ClusterInst, func(res *edgeproto.ClusterInst) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowClusterInstStream(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst, cb func(res *edgeproto.ClusterInst)) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceClusterInsts, ActionView)
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
	api := edgeproto.NewClusterInstApiClient(rc.conn)
	stream, err := api.ShowClusterInst(ctx, obj)
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
		cb(res)
	}
	return nil
}

func ShowClusterInstObj(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterInst) ([]edgeproto.ClusterInst, error) {
	arr := []edgeproto.ClusterInst{}
	err := ShowClusterInstStream(ctx, rc, obj, func(res *edgeproto.ClusterInst) {
		arr = append(arr, *res)
	})
	return arr, err
}
