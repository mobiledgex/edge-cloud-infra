// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudletpool.proto

package orm

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/labstack/echo"
import "net/http"
import "context"
import "io"
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
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := CreateCloudletPoolObj(ctx, rc, &in.CloudletPool)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func CreateCloudletPoolObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceCloudletPools, ActionManage) {
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
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := DeleteCloudletPoolObj(ctx, rc, &in.CloudletPool)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func DeleteCloudletPoolObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPool) (*edgeproto.Result, error) {
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
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceCloudletPools, ActionView)
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
			if !authz.Ok("") {
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

func CreateCloudletPoolMember(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPoolMember{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := CreateCloudletPoolMemberObj(ctx, rc, &in.CloudletPoolMember)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func CreateCloudletPoolMemberObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceCloudletPools, ActionManage) {
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
	api := edgeproto.NewCloudletPoolMemberApiClient(rc.conn)
	return api.CreateCloudletPoolMember(ctx, obj)
}

func DeleteCloudletPoolMember(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPoolMember{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := DeleteCloudletPoolMemberObj(ctx, rc, &in.CloudletPoolMember)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func DeleteCloudletPoolMemberObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceCloudletPools, ActionManage) {
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
	api := edgeproto.NewCloudletPoolMemberApiClient(rc.conn)
	return api.DeleteCloudletPoolMember(ctx, obj)
}

func ShowCloudletPoolMember(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPoolMember{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region

	err = ShowCloudletPoolMemberStream(ctx, rc, &in.CloudletPoolMember, func(res *edgeproto.CloudletPoolMember) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowCloudletPoolMemberStream(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolMember, cb func(res *edgeproto.CloudletPoolMember)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceCloudletPools, ActionView)
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
	api := edgeproto.NewCloudletPoolMemberApiClient(rc.conn)
	stream, err := api.ShowCloudletPoolMember(ctx, obj)
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

func ShowCloudletPoolMemberObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolMember) ([]edgeproto.CloudletPoolMember, error) {
	arr := []edgeproto.CloudletPoolMember{}
	err := ShowCloudletPoolMemberStream(ctx, rc, obj, func(res *edgeproto.CloudletPoolMember) {
		arr = append(arr, *res)
	})
	return arr, err
}

func ShowPoolsForCloudlet(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletKey{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region

	err = ShowPoolsForCloudletStream(ctx, rc, &in.CloudletKey, func(res *edgeproto.CloudletPool) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowPoolsForCloudletStream(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletKey, cb func(res *edgeproto.CloudletPool)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceCloudletPools, ActionView)
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
	api := edgeproto.NewCloudletPoolShowApiClient(rc.conn)
	stream, err := api.ShowPoolsForCloudlet(ctx, obj)
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

func ShowPoolsForCloudletObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletKey) ([]edgeproto.CloudletPool, error) {
	arr := []edgeproto.CloudletPool{}
	err := ShowPoolsForCloudletStream(ctx, rc, obj, func(res *edgeproto.CloudletPool) {
		arr = append(arr, *res)
	})
	return arr, err
}

func ShowCloudletsForPool(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletPoolKey{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region

	err = ShowCloudletsForPoolStream(ctx, rc, &in.CloudletPoolKey, func(res *edgeproto.Cloudlet) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowCloudletsForPoolStream(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolKey, cb func(res *edgeproto.Cloudlet)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceCloudletPools, ActionView)
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
	api := edgeproto.NewCloudletPoolShowApiClient(rc.conn)
	stream, err := api.ShowCloudletsForPool(ctx, obj)
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

func ShowCloudletsForPoolObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletPoolKey) ([]edgeproto.Cloudlet, error) {
	arr := []edgeproto.Cloudlet{}
	err := ShowCloudletsForPoolStream(ctx, rc, obj, func(res *edgeproto.Cloudlet) {
		arr = append(arr, *res)
	})
	return arr, err
}
