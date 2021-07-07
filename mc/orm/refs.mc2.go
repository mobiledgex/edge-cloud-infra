// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: refs.proto

package orm

import (
	"context"
	fmt "fmt"
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

func ShowCloudletRefs(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletRefs{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletRefs.GetKey().GetTags())
	span.SetTag("org", in.CloudletRefs.Key.Organization)

	err = ShowCloudletRefsStream(ctx, rc, &in.CloudletRefs, func(res *edgeproto.CloudletRefs) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

func ShowCloudletRefsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletRefs, cb func(res *edgeproto.CloudletRefs) error) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceCloudlets, ActionView)
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

func ShowCloudletRefsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletRefs) ([]edgeproto.CloudletRefs, error) {
	arr := []edgeproto.CloudletRefs{}
	err := ShowCloudletRefsStream(ctx, rc, obj, func(res *edgeproto.CloudletRefs) error {
		arr = append(arr, *res)
		return nil
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
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterRefs.GetKey().GetTags())

	err = ShowClusterRefsStream(ctx, rc, &in.ClusterRefs, func(res *edgeproto.ClusterRefs) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

func ShowClusterRefsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterRefs, cb func(res *edgeproto.ClusterRefs) error) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceClusterInsts, ActionView)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func ShowClusterRefsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.ClusterRefs) ([]edgeproto.ClusterRefs, error) {
	arr := []edgeproto.ClusterRefs{}
	err := ShowClusterRefsStream(ctx, rc, obj, func(res *edgeproto.ClusterRefs) error {
		arr = append(arr, *res)
		return nil
	})
	return arr, err
}

func ShowAppInstRefs(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionAppInstRefs{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.AppInstRefs.GetKey().GetTags())
	span.SetTag("org", in.AppInstRefs.Key.Organization)

	err = ShowAppInstRefsStream(ctx, rc, &in.AppInstRefs, func(res *edgeproto.AppInstRefs) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	})
	if err != nil {
		return err
	}
	return nil
}

func ShowAppInstRefsStream(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInstRefs, cb func(res *edgeproto.AppInstRefs) error) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceAppInsts, ActionView)
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
	api := edgeproto.NewAppInstRefsApiClient(rc.conn)
	stream, err := api.ShowAppInstRefs(ctx, obj)
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

func ShowAppInstRefsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.AppInstRefs) ([]edgeproto.AppInstRefs, error) {
	arr := []edgeproto.AppInstRefs{}
	err := ShowAppInstRefsStream(ctx, rc, obj, func(res *edgeproto.AppInstRefs) error {
		arr = append(arr, *res)
		return nil
	})
	return arr, err
}
