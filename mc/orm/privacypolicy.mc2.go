// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: privacypolicy.proto

package orm

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/labstack/echo"
import "context"
import "io"
import "github.com/mobiledgex/edge-cloud/log"
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

func CreatePrivacyPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionPrivacyPolicy{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.PrivacyPolicy.Key.Organization)
	resp, err := CreatePrivacyPolicyObj(ctx, rc, &in.PrivacyPolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func CreatePrivacyPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.PrivacyPolicy) (*edgeproto.Result, error) {
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceDeveloperPolicy, ActionManage, withRequiresOrg(obj.Key.Organization)); err != nil {
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
	api := edgeproto.NewPrivacyPolicyApiClient(rc.conn)
	return api.CreatePrivacyPolicy(ctx, obj)
}

func DeletePrivacyPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionPrivacyPolicy{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.PrivacyPolicy.Key.Organization)
	resp, err := DeletePrivacyPolicyObj(ctx, rc, &in.PrivacyPolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func DeletePrivacyPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.PrivacyPolicy) (*edgeproto.Result, error) {
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceDeveloperPolicy, ActionManage); err != nil {
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
	api := edgeproto.NewPrivacyPolicyApiClient(rc.conn)
	return api.DeletePrivacyPolicy(ctx, obj)
}

func UpdatePrivacyPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionPrivacyPolicy{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.PrivacyPolicy.Key.Organization)
	resp, err := UpdatePrivacyPolicyObj(ctx, rc, &in.PrivacyPolicy)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func UpdatePrivacyPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.PrivacyPolicy) (*edgeproto.Result, error) {
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceDeveloperPolicy, ActionManage); err != nil {
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
	api := edgeproto.NewPrivacyPolicyApiClient(rc.conn)
	return api.UpdatePrivacyPolicy(ctx, obj)
}

func ShowPrivacyPolicy(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionPrivacyPolicy{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("org", in.PrivacyPolicy.Key.Organization)

	err = ShowPrivacyPolicyStream(ctx, rc, &in.PrivacyPolicy, func(res *edgeproto.PrivacyPolicy) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowPrivacyPolicyStream(ctx context.Context, rc *RegionContext, obj *edgeproto.PrivacyPolicy, cb func(res *edgeproto.PrivacyPolicy)) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceDeveloperPolicy, ActionView)
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
	api := edgeproto.NewPrivacyPolicyApiClient(rc.conn)
	stream, err := api.ShowPrivacyPolicy(ctx, obj)
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

func ShowPrivacyPolicyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.PrivacyPolicy) ([]edgeproto.PrivacyPolicy, error) {
	arr := []edgeproto.PrivacyPolicy{}
	err := ShowPrivacyPolicyStream(ctx, rc, obj, func(res *edgeproto.PrivacyPolicy) {
		arr = append(arr, *res)
	})
	return arr, err
}
