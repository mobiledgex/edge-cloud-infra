// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudletpool.proto

package orm

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
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

func CreateCloudletPool(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPool{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	if err := in.CloudletPool.IsValidArgsForCreateCloudletPool(); err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletPool.GetKey().GetTags())
	span.SetTag("org", in.CloudletPool.Key.Organization)
	resp, err := CreateCloudletPoolObj(ctx, rc, &in.CloudletPool)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func CreateCloudletPoolObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudletPools, ActionManage, withRequiresOrg(obj.Key.Organization)); err != nil {
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
	api := edgeproto.NewCloudletPoolApiClient(rc.conn)
	return api.CreateCloudletPool(ctx, obj)
}

func DeleteCloudletPool(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPool{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	if err := in.CloudletPool.IsValidArgsForDeleteCloudletPool(); err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletPool.GetKey().GetTags())
	span.SetTag("org", in.CloudletPool.Key.Organization)
	resp, err := DeleteCloudletPoolObj(ctx, rc, &in.CloudletPool)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func DeleteCloudletPoolObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authzDeleteCloudletPool(ctx, rc.region, rc.username, obj,
			ResourceCloudletPools, ActionManage); err != nil {
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
	api := edgeproto.NewCloudletPoolApiClient(rc.conn)
	return api.DeleteCloudletPool(ctx, obj)
}

func UpdateCloudletPool(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPool{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	if err := in.CloudletPool.IsValidArgsForUpdateCloudletPool(); err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletPool.GetKey().GetTags())
	span.SetTag("org", in.CloudletPool.Key.Organization)
	resp, err := UpdateCloudletPoolObj(ctx, rc, &in.CloudletPool)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func UpdateCloudletPoolObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudletPools, ActionManage); err != nil {
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
	api := edgeproto.NewCloudletPoolApiClient(rc.conn)
	return api.UpdateCloudletPool(ctx, obj)
}

func ShowCloudletPool(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPool{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletPool.GetKey().GetTags())
	span.SetTag("org", in.CloudletPool.Key.Organization)

	err = ShowCloudletPoolStream(ctx, rc, &in.CloudletPool, func(res *edgeproto.CloudletPool) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowCloudletPoolStream(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPool, cb func(res *edgeproto.CloudletPool)) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceCloudletPools, ActionView)
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
	api := edgeproto.NewCloudletPoolApiClient(rc.conn)
	stream, err := api.ShowCloudletPool(ctx, obj)
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

func ShowCloudletPoolObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPool) ([]edgeproto.CloudletPool, error) {
	arr := []edgeproto.CloudletPool{}
	err := ShowCloudletPoolStream(ctx, rc, obj, func(res *edgeproto.CloudletPool) {
		arr = append(arr, *res)
	})
	return arr, err
}

func AddCloudletPoolMember(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPoolMember{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	if err := in.CloudletPoolMember.IsValidArgsForAddCloudletPoolMember(); err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletPoolMember.GetKey().GetTags())
	span.SetTag("org", in.CloudletPoolMember.Key.Organization)
	resp, err := AddCloudletPoolMemberObj(ctx, rc, &in.CloudletPoolMember)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func AddCloudletPoolMemberObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudletPools, ActionManage); err != nil {
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
	api := edgeproto.NewCloudletPoolApiClient(rc.conn)
	return api.AddCloudletPoolMember(ctx, obj)
}

func RemoveCloudletPoolMember(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPoolMember{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	if err := in.CloudletPoolMember.IsValidArgsForRemoveCloudletPoolMember(); err != nil {
		return err
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletPoolMember.GetKey().GetTags())
	span.SetTag("org", in.CloudletPoolMember.Key.Organization)
	resp, err := RemoveCloudletPoolMemberObj(ctx, rc, &in.CloudletPoolMember)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func RemoveCloudletPoolMemberObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudletPools, ActionManage); err != nil {
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
	api := edgeproto.NewCloudletPoolApiClient(rc.conn)
	return api.RemoveCloudletPoolMember(ctx, obj)
}
