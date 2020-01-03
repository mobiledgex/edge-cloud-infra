// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flavor.proto

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

func CreateFlavor(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlavor{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := CreateFlavorObj(ctx, rc, &in.Flavor)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func CreateFlavorObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Flavor) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceFlavors, ActionManage) {
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
	api := edgeproto.NewFlavorApiClient(rc.conn)
	return api.CreateFlavor(ctx, obj)
}

func DeleteFlavor(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlavor{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := DeleteFlavorObj(ctx, rc, &in.Flavor)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func DeleteFlavorObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Flavor) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceFlavors, ActionManage) {
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
	api := edgeproto.NewFlavorApiClient(rc.conn)
	return api.DeleteFlavor(ctx, obj)
}

func UpdateFlavor(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlavor{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := UpdateFlavorObj(ctx, rc, &in.Flavor)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func UpdateFlavorObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Flavor) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceFlavors, ActionManage) {
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
	api := edgeproto.NewFlavorApiClient(rc.conn)
	return api.UpdateFlavor(ctx, obj)
}

func ShowFlavor(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlavor{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region

	err = ShowFlavorStream(ctx, rc, &in.Flavor, func(res *edgeproto.Flavor) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowFlavorStream(ctx context.Context, rc *RegionContext, obj *edgeproto.Flavor, cb func(res *edgeproto.Flavor)) error {
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
	api := edgeproto.NewFlavorApiClient(rc.conn)
	stream, err := api.ShowFlavor(ctx, obj)
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

func ShowFlavorObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Flavor) ([]edgeproto.Flavor, error) {
	arr := []edgeproto.Flavor{}
	err := ShowFlavorStream(ctx, rc, obj, func(res *edgeproto.Flavor) {
		arr = append(arr, *res)
	})
	return arr, err
}

func AddFlavorRes(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlavor{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := AddFlavorResObj(ctx, rc, &in.Flavor)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func AddFlavorResObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Flavor) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceFlavors, ActionManage) {
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
	api := edgeproto.NewFlavorApiClient(rc.conn)
	return api.AddFlavorRes(ctx, obj)
}

func RemoveFlavorRes(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlavor{}
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	rc.region = in.Region
	resp, err := RemoveFlavorResObj(ctx, rc, &in.Flavor)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func RemoveFlavorResObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Flavor) (*edgeproto.Result, error) {
	if !rc.skipAuthz && !authorized(ctx, rc.username, "",
		ResourceFlavors, ActionManage) {
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
	api := edgeproto.NewFlavorApiClient(rc.conn)
	return api.RemoveFlavorRes(ctx, obj)
}
