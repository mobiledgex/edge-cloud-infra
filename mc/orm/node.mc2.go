// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: node.proto

package orm

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/labstack/echo"
import "context"
import "io"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/gogo/protobuf/gogoproto"
import _ "github.com/mobiledgex/edge-cloud/protogen"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func ShowNode(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionNode{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region

	err = ShowNodeStream(ctx, rc, &in.Node, func(res *edgeproto.Node) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowNodeStream(ctx context.Context, rc *RegionContext, obj *edgeproto.Node, cb func(res *edgeproto.Node)) error {
	var authz *ShowAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = NewShowAuthz(ctx, rc.region, rc.username, ResourceConfig, ActionView)
		if err == echo.ErrForbidden {
			return nil
		}
		if err != nil {
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
	api := edgeproto.NewNodeApiClient(rc.conn)
	stream, err := api.ShowNode(ctx, obj)
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

func ShowNodeObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Node) ([]edgeproto.Node, error) {
	arr := []edgeproto.Node{}
	err := ShowNodeStream(ctx, rc, obj, func(res *edgeproto.Node) {
		arr = append(arr, *res)
	})
	return arr, err
}
