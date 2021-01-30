// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

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

func CreateCloudlet(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudlet{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())
	span.SetTag("org", in.Cloudlet.Key.Organization)

	err = CreateCloudletStream(ctx, rc, &in.Cloudlet, func(res *edgeproto.Result) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func CreateCloudletStream(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Result)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateCloudlet(); err != nil {
		return err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudlets, ActionManage, withRequiresOrg(obj.Key.Organization)); err != nil {
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	stream, err := api.CreateCloudlet(ctx, obj)
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

func CreateCloudletObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := CreateCloudletStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func DeleteCloudlet(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudlet{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())
	span.SetTag("org", in.Cloudlet.Key.Organization)

	err = DeleteCloudletStream(ctx, rc, &in.Cloudlet, func(res *edgeproto.Result) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func DeleteCloudletStream(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Result)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteCloudlet(); err != nil {
		return err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	stream, err := api.DeleteCloudlet(ctx, obj)
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

func DeleteCloudletObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := DeleteCloudletStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func UpdateCloudlet(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudlet{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())
	span.SetTag("org", in.Cloudlet.Key.Organization)

	err = UpdateCloudletStream(ctx, rc, &in.Cloudlet, func(res *edgeproto.Result) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func UpdateCloudletStream(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Result)) error {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateCloudlet(); err != nil {
		return err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	stream, err := api.UpdateCloudlet(ctx, obj)
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

func UpdateCloudletObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet) ([]edgeproto.Result, error) {
	arr := []edgeproto.Result{}
	err := UpdateCloudletStream(ctx, rc, obj, func(res *edgeproto.Result) {
		arr = append(arr, *res)
	})
	return arr, err
}

func ShowCloudlet(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudlet{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())

	err = ShowCloudletStream(ctx, rc, &in.Cloudlet, func(res *edgeproto.Cloudlet) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

type ShowCloudletAuthz interface {
	Ok(obj *edgeproto.Cloudlet) bool
}

func ShowCloudletStream(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Cloudlet)) error {
	var authz ShowCloudletAuthz
	var err error
	if !rc.skipAuthz {
		authz, err = newShowCloudletAuthz(ctx, rc.region, rc.username, ResourceCloudlets, ActionView)
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	stream, err := api.ShowCloudlet(ctx, obj)
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
			if !authz.Ok(res) {
				continue
			}
		}
		cb(res)
	}
	return nil
}

func ShowCloudletObj(ctx context.Context, rc *RegionContext, obj *edgeproto.Cloudlet) ([]edgeproto.Cloudlet, error) {
	arr := []edgeproto.Cloudlet{}
	err := ShowCloudletStream(ctx, rc, obj, func(res *edgeproto.Cloudlet) {
		arr = append(arr, *res)
	})
	return arr, err
}

func GetCloudletManifest(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletKey{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)
	resp, err := GetCloudletManifestObj(ctx, rc, &in.CloudletKey)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func GetCloudletManifestObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletKey) (*edgeproto.CloudletManifest, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForGetCloudletManifest(); err != nil {
		return nil, err
	}
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.GetCloudletManifest(ctx, obj)
}

func GetCloudletProps(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletProps{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	resp, err := GetCloudletPropsObj(ctx, rc, &in.CloudletProps)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func GetCloudletPropsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletProps) (*edgeproto.CloudletProps, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceCloudlets, ActionView); err != nil {
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.GetCloudletProps(ctx, obj)
}

func GetCloudletResourceQuotaProps(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletResourceQuotaProps{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	resp, err := GetCloudletResourceQuotaPropsObj(ctx, rc, &in.CloudletResourceQuotaProps)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func GetCloudletResourceQuotaPropsObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletResourceQuotaProps) (*edgeproto.CloudletResourceQuotaProps, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, "",
			ResourceCloudlets, ActionView); err != nil {
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.GetCloudletResourceQuotaProps(ctx, obj)
}

func GetCloudletResourceUsage(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletResourceUsage{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletResourceUsage.GetKey().GetTags())
	span.SetTag("org", in.CloudletResourceUsage.Key.Organization)
	resp, err := GetCloudletResourceUsageObj(ctx, rc, &in.CloudletResourceUsage)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func GetCloudletResourceUsageObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletResourceUsage) (*edgeproto.InfraResourcesSnapshot, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForGetCloudletResourceUsage(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.GetCloudletResourceUsage(ctx, obj)
}

func SyncCloudletInfraResources(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletKey{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)
	resp, err := SyncCloudletInfraResourcesObj(ctx, rc, &in.CloudletKey)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func SyncCloudletInfraResourcesObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletKey) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForSyncCloudletInfraResources(); err != nil {
		return nil, err
	}
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.SyncCloudletInfraResources(ctx, obj)
}

func AddCloudletResMapping(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletResMap{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletResMap.GetKey().GetTags())
	span.SetTag("org", in.CloudletResMap.Key.Organization)
	resp, err := AddCloudletResMappingObj(ctx, rc, &in.CloudletResMap)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func AddCloudletResMappingObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForAddCloudletResMapping(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.AddCloudletResMapping(ctx, obj)
}

func RemoveCloudletResMapping(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletResMap{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletResMap.GetKey().GetTags())
	span.SetTag("org", in.CloudletResMap.Key.Organization)
	resp, err := RemoveCloudletResMappingObj(ctx, rc, &in.CloudletResMap)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func RemoveCloudletResMappingObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletResMap) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRemoveCloudletResMapping(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.RemoveCloudletResMapping(ctx, obj)
}

func FindFlavorMatch(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionFlavorMatch{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.FlavorMatch.GetKey().GetTags())
	span.SetTag("org", in.FlavorMatch.Key.Organization)
	resp, err := FindFlavorMatchObj(ctx, rc, &in.FlavorMatch)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func FindFlavorMatchObj(ctx context.Context, rc *RegionContext, obj *edgeproto.FlavorMatch) (*edgeproto.FlavorMatch, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
			ResourceCloudlets, ActionView); err != nil {
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.FindFlavorMatch(ctx, obj)
}

func RevokeAccessKey(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletKey{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)
	resp, err := RevokeAccessKeyObj(ctx, rc, &in.CloudletKey)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func RevokeAccessKeyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletKey) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRevokeAccessKey(); err != nil {
		return nil, err
	}
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.RevokeAccessKey(ctx, obj)
}

func GenerateAccessKey(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletKey{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)
	resp, err := GenerateAccessKeyObj(ctx, rc, &in.CloudletKey)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func GenerateAccessKeyObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletKey) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForGenerateAccessKey(); err != nil {
		return nil, err
	}
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
	api := edgeproto.NewCloudletApiClient(rc.conn)
	return api.GenerateAccessKey(ctx, obj)
}

func ShowCloudletInfo(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletInfo{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	defer CloseConn(c)
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletInfo.GetKey().GetTags())
	span.SetTag("org", in.CloudletInfo.Key.Organization)

	err = ShowCloudletInfoStream(ctx, rc, &in.CloudletInfo, func(res *edgeproto.CloudletInfo) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		WriteError(c, err)
	}
	return nil
}

func ShowCloudletInfoStream(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletInfo, cb func(res *edgeproto.CloudletInfo)) error {
	var authz *AuthzShow
	var err error
	if !rc.skipAuthz {
		authz, err = newShowAuthz(ctx, rc.region, rc.username, ResourceCloudletAnalytics, ActionView)
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
	api := edgeproto.NewCloudletInfoApiClient(rc.conn)
	stream, err := api.ShowCloudletInfo(ctx, obj)
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

func ShowCloudletInfoObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletInfo) ([]edgeproto.CloudletInfo, error) {
	arr := []edgeproto.CloudletInfo{}
	err := ShowCloudletInfoStream(ctx, rc, obj, func(res *edgeproto.CloudletInfo) {
		arr = append(arr, *res)
	})
	return arr, err
}

func InjectCloudletInfo(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletInfo{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletInfo.GetKey().GetTags())
	span.SetTag("org", in.CloudletInfo.Key.Organization)
	resp, err := InjectCloudletInfoObj(ctx, rc, &in.CloudletInfo)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func InjectCloudletInfoObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForInjectCloudletInfo(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
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
	api := edgeproto.NewCloudletInfoApiClient(rc.conn)
	return api.InjectCloudletInfo(ctx, obj)
}

func EvictCloudletInfo(c echo.Context) error {
	ctx := GetContext(c)
	rc := &RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.username = claims.Username

	in := ormapi.RegionCloudletInfo{}
	if err := c.Bind(&in); err != nil {
		return bindErr(c, err)
	}
	rc.region = in.Region
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletInfo.GetKey().GetTags())
	span.SetTag("org", in.CloudletInfo.Key.Organization)
	resp, err := EvictCloudletInfoObj(ctx, rc, &in.CloudletInfo)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
	}
	return setReply(c, err, resp)
}

func EvictCloudletInfoObj(ctx context.Context, rc *RegionContext, obj *edgeproto.CloudletInfo) (*edgeproto.Result, error) {
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForEvictCloudletInfo(); err != nil {
		return nil, err
	}
	if !rc.skipAuthz {
		if err := authorized(ctx, rc.username, obj.Key.Organization,
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
	api := edgeproto.NewCloudletInfoApiClient(rc.conn)
	return api.EvictCloudletInfo(ctx, obj)
}
