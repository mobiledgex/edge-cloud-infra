// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoscalepolicy.proto

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

func CreateAutoScalePolicy(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.AutoScalePolicy.GetKey().GetTags())
	span.SetTag("org", in.AutoScalePolicy.Key.Organization)

	obj := &in.AutoScalePolicy
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateAutoScalePolicy(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceDeveloperPolicy, ActionManage, withRequiresOrg(obj.Key.Organization)); err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	resp, err := ctrlapi.CreateAutoScalePolicyObj(ctx, rc, obj, conn)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func DeleteAutoScalePolicy(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.AutoScalePolicy.GetKey().GetTags())
	span.SetTag("org", in.AutoScalePolicy.Key.Organization)

	obj := &in.AutoScalePolicy
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteAutoScalePolicy(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceDeveloperPolicy, ActionManage); err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	resp, err := ctrlapi.DeleteAutoScalePolicyObj(ctx, rc, obj, conn)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func UpdateAutoScalePolicy(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	dat, err := ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.AutoScalePolicy.GetKey().GetTags())
	span.SetTag("org", in.AutoScalePolicy.Key.Organization)
	err = ormutil.SetRegionObjFields(dat, &in)
	if err != nil {
		return err
	}

	obj := &in.AutoScalePolicy
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateAutoScalePolicy(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceDeveloperPolicy, ActionManage); err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	resp, err := ctrlapi.UpdateAutoScalePolicyObj(ctx, rc, obj, conn)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func ShowAutoScalePolicy(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionAutoScalePolicy{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.AutoScalePolicy.GetKey().GetTags())
	span.SetTag("org", in.AutoScalePolicy.Key.Organization)

	obj := &in.AutoScalePolicy
	var authz *AuthzShow
	if !rc.SkipAuthz {
		authz, err = newShowAuthz(ctx, rc.Region, rc.Username, ResourceDeveloperPolicy, ActionView)
		if err != nil {
			return err
		}
	}
	conn, err := connCache.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}

	cb := func(res *edgeproto.AutoScalePolicy) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlapi.ShowAutoScalePolicyStream(ctx, rc, obj, conn, authz.Ok, cb)
	if err != nil {
		return err
	}
	return nil
}
