// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

package orm

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ctrlapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func CreateClusterInst(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionClusterInst{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	obj := &in.ClusterInst
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateClusterInst(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authzCreateClusterInst(ctx, rc.Region, rc.Username, obj,
			ResourceClusterInsts, ActionManage); err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlapi.CreateClusterInstStream(ctx, rc, obj, conn, cb)
	if err != nil {
		return err
	}
	return nil
}

func DeleteClusterInst(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionClusterInst{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	obj := &in.ClusterInst
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteClusterInst(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceClusterInsts, ActionManage); err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlapi.DeleteClusterInstStream(ctx, rc, obj, conn, cb)
	if err != nil {
		return err
	}
	return nil
}

func UpdateClusterInst(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionClusterInst{}
	dat, err := ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)
	err = ormutil.SetRegionObjFields(dat, &in)
	if err != nil {
		return err
	}

	obj := &in.ClusterInst
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateClusterInst(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceClusterInsts, ActionManage); err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlapi.UpdateClusterInstStream(ctx, rc, obj, conn, cb)
	if err != nil {
		return err
	}
	return nil
}

type ShowClusterInstAuthz interface {
	Ok(obj *edgeproto.ClusterInst) (bool, bool)
	Filter(obj *edgeproto.ClusterInst)
}

func ShowClusterInst(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionClusterInst{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.ClusterInst.GetKey().GetTags())
	span.SetTag("org", in.ClusterInst.Key.Organization)

	obj := &in.ClusterInst
	var authz ShowClusterInstAuthz
	if !rc.SkipAuthz {
		authz, err = newShowClusterInstAuthz(ctx, rc.Region, rc.Username, ResourceClusterInsts, ActionView)
		if err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	cb := func(res *edgeproto.ClusterInst) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlapi.ShowClusterInstStream(ctx, rc, obj, conn, authz.Ok, authz.Filter, cb)
	if err != nil {
		return err
	}
	return nil
}

func DeleteIdleReservableClusterInsts(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionIdleReservableClusterInsts{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)

	obj := &in.IdleReservableClusterInsts
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteIdleReservableClusterInsts(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, "",
			ResourceClusterInsts, ActionManage); err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	resp, err := ctrlapi.DeleteIdleReservableClusterInstsObj(ctx, rc, obj, conn)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}
