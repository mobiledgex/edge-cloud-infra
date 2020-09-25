// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operatorcode.proto

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

func CreateOperatorCode(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionOperatorCode{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.OperatorCode.Organization)
	resp, err := CreateOperatorCodeObj(ctx, rc, &in.OperatorCode)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func CreateOperatorCodeObj(ctx context.Context, rc *RegionContext, obj *edgeproto.OperatorCode) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Organization,
			ResourceCloudlets, ActionManage, withRequiresOrg(obj.Organization)); err != nil {
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
	api := edgeproto.NewOperatorCodeApiClient(rc.conn)
	return api.CreateOperatorCode(ctx, obj)
}

func DeleteOperatorCode(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionOperatorCode{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.OperatorCode.Organization)
	resp, err := DeleteOperatorCodeObj(ctx, rc, &in.OperatorCode)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func DeleteOperatorCodeObj(ctx context.Context, rc *RegionContext, obj *edgeproto.OperatorCode) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Organization,
			ResourceCloudlets, ActionManage); err != nil {
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
	api := edgeproto.NewOperatorCodeApiClient(rc.conn)
	return api.DeleteOperatorCode(ctx, obj)
}

func ShowOperatorCode(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionOperatorCode{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.OperatorCode.Organization)

	err = ShowOperatorCodeStream(ctx, rc, &in.OperatorCode, func(res *edgeproto.OperatorCode) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowOperatorCodeStream(ctx context.Context, rc *RegionContext, obj *edgeproto.OperatorCode, cb func(res *edgeproto.OperatorCode)) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceCloudlets, ActionView)
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
	api := edgeproto.NewOperatorCodeApiClient(rc.conn)
	stream, err := api.ShowOperatorCode(ctx, obj)
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
			if !authz.Ok(res.Organization) {
				continue
			}
		}
		cb(res)
	}
	return nil
}

func ShowOperatorCodeObj(ctx context.Context, rc *RegionContext, obj *edgeproto.OperatorCode) ([]edgeproto.OperatorCode, error) {
	arr := []edgeproto.OperatorCode{}
	err := ShowOperatorCodeStream(ctx, rc, obj, func(res *edgeproto.OperatorCode) {
		arr = append(arr, *res)
	})
	return arr, err
}
