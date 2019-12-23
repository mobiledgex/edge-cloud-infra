// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: refs.proto

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
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func ShowCloudletRefs(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletRefs{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.CloudletRefs.Key.OperatorKey.Name)

	err = ShowCloudletRefsStream(ctx, rc, &in.CloudletRefs, func(res *edgeproto.CloudletRefs) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		return WriteError(c, err)
	}
	return nil
}

func ShowCloudletRefsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletRefs, cb func(res *edgeproto.CloudletRefs)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceCloudlets, ActionView)
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
	api := edgeproto.NewCloudletRefsApiClient(rc.conn)
	stream, err := api.ShowCloudletRefs(ctx, obj)
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
			if !authz.Ok(res.Key.OperatorKey.Name) {
				continue
			}
		}
		cb(res)
	}
	return nil
}

func ShowCloudletRefsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletRefs) ([]edgeproto.CloudletRefs, error) {
	arr := []edgeproto.CloudletRefs{}
	err := ShowCloudletRefsStream(ctx, rc, obj, func(res *edgeproto.CloudletRefs) {
		arr = append(arr, *res)
	})
	return arr, err
}

func ShowClusterRefs(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionClusterRefs{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	rc.region = in.Region

	err = ShowClusterRefsStream(ctx, rc, &in.ClusterRefs, func(res *edgeproto.ClusterRefs) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		return WriteError(c, err)
	}
	return nil
}

func ShowClusterRefsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterRefs, cb func(res *edgeproto.ClusterRefs)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceClusterInsts, ActionView)
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
	api := edgeproto.NewClusterRefsApiClient(rc.conn)
	stream, err := api.ShowClusterRefs(ctx, obj)
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
		cb(res)
	}
	return nil
}

func ShowClusterRefsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterRefs) ([]edgeproto.ClusterRefs, error) {
	arr := []edgeproto.ClusterRefs{}
	err := ShowClusterRefsStream(ctx, rc, obj, func(res *edgeproto.ClusterRefs) {
		arr = append(arr, *res)
	})
	return arr, err
}
