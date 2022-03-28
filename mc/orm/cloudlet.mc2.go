// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package orm

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ctrlclient"
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

func CreateGPUDriver(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriver{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.GPUDriver.GetKey().GetTags())
	span.SetTag("org", in.GPUDriver.Key.Organization)

	obj := &in.GPUDriver
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateGPUDriver(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage, withRequiresOrg(obj.Key.Organization)); err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.CreateGPUDriverStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func DeleteGPUDriver(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriver{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.GPUDriver.GetKey().GetTags())
	span.SetTag("org", in.GPUDriver.Key.Organization)

	obj := &in.GPUDriver
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteGPUDriver(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.DeleteGPUDriverStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func UpdateGPUDriver(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriver{}
	dat, err := ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.GPUDriver.GetKey().GetTags())
	span.SetTag("org", in.GPUDriver.Key.Organization)
	err = ormutil.SetRegionObjFields(dat, &in)
	if err != nil {
		return err
	}

	obj := &in.GPUDriver
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateGPUDriver(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.UpdateGPUDriverStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func ShowGPUDriver(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriver{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.GPUDriver.GetKey().GetTags())
	span.SetTag("org", in.GPUDriver.Key.Organization)

	obj := &in.GPUDriver
	var authz ctrlclient.ShowGPUDriverAuthz
	if !rc.SkipAuthz {
		authz, err = newShowGPUDriverAuthz(ctx, rc.Region, rc.Username, ResourceCloudlets, ActionView)
		if err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.GPUDriver) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.ShowGPUDriverStream(ctx, rc, obj, connCache, authz, cb)
	if err != nil {
		return err
	}
	return nil
}

func AddGPUDriverBuild(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriverBuildMember{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.GPUDriverBuildMember.GetKey().GetTags())
	span.SetTag("org", in.GPUDriverBuildMember.Key.Organization)

	obj := &in.GPUDriverBuildMember
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForAddGPUDriverBuild(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.AddGPUDriverBuildStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func RemoveGPUDriverBuild(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriverBuildMember{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.GPUDriverBuildMember.GetKey().GetTags())
	span.SetTag("org", in.GPUDriverBuildMember.Key.Organization)

	obj := &in.GPUDriverBuildMember
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRemoveGPUDriverBuild(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.RemoveGPUDriverBuildStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func GetGPUDriverBuildURL(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriverBuildMember{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.GPUDriverBuildMember.GetKey().GetTags())
	span.SetTag("org", in.GPUDriverBuildMember.Key.Organization)

	obj := &in.GPUDriverBuildMember
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authzGetGPUDriverBuildURL(ctx, rc.Region, rc.Username, obj,
			ResourceCloudlets, ActionView); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GetGPUDriverBuildURLObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func GetGPUDriverLicenseConfig(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionGPUDriverKey{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.GPUDriverKey.Organization)

	obj := &in.GPUDriverKey
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Organization,
			ResourceCloudletAnalytics, ActionView); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GetGPUDriverLicenseConfigObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func CreateCloudlet(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudlet{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())
	span.SetTag("org", in.Cloudlet.Key.Organization)

	obj := &in.Cloudlet
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForCreateCloudlet(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authzCreateCloudlet(ctx, rc.Region, rc.Username, obj,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}
	// Need access to database for federation handling
	rc.Database = database

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.CreateCloudletStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func DeleteCloudlet(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudlet{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())
	span.SetTag("org", in.Cloudlet.Key.Organization)

	obj := &in.Cloudlet
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForDeleteCloudlet(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}
	// Need access to database for federation handling
	rc.Database = database

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.DeleteCloudletStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func UpdateCloudlet(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudlet{}
	dat, err := ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())
	span.SetTag("org", in.Cloudlet.Key.Organization)
	err = ormutil.SetRegionObjFields(dat, &in)
	if err != nil {
		return err
	}

	obj := &in.Cloudlet
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForUpdateCloudlet(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authzUpdateCloudlet(ctx, rc.Region, rc.Username, obj,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}
	// Need access to database for federation handling
	rc.Database = database

	cb := func(res *edgeproto.Result) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.UpdateCloudletStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func ShowCloudlet(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudlet{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.Cloudlet.GetKey().GetTags())

	obj := &in.Cloudlet
	var authz ctrlclient.ShowCloudletAuthz
	if !rc.SkipAuthz {
		authz, err = newShowCloudletAuthz(ctx, rc.Region, rc.Username, ResourceCloudlets, ActionView)
		if err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.Cloudlet) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.ShowCloudletStream(ctx, rc, obj, connCache, authz, cb)
	if err != nil {
		return err
	}
	return nil
}

func GetCloudletManifest(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletKey{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)

	obj := &in.CloudletKey
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForGetCloudletManifest(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GetCloudletManifestObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func GetCloudletProps(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletProps{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletProps.Organization)

	obj := &in.CloudletProps
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Organization,
			ResourceCloudletAnalytics, ActionView); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GetCloudletPropsObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func GetCloudletResourceQuotaProps(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletResourceQuotaProps{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletResourceQuotaProps.Organization)

	obj := &in.CloudletResourceQuotaProps
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Organization,
			ResourceCloudletAnalytics, ActionView); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GetCloudletResourceQuotaPropsObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func GetCloudletResourceUsage(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletResourceUsage{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletResourceUsage.GetKey().GetTags())
	span.SetTag("org", in.CloudletResourceUsage.Key.Organization)

	obj := &in.CloudletResourceUsage
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionView); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GetCloudletResourceUsageObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func AddCloudletResMapping(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletResMap{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletResMap.GetKey().GetTags())
	span.SetTag("org", in.CloudletResMap.Key.Organization)

	obj := &in.CloudletResMap
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForAddCloudletResMapping(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.AddCloudletResMappingObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func RemoveCloudletResMapping(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletResMap{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletResMap.GetKey().GetTags())
	span.SetTag("org", in.CloudletResMap.Key.Organization)

	obj := &in.CloudletResMap
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRemoveCloudletResMapping(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.RemoveCloudletResMappingObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func AddCloudletAllianceOrg(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletAllianceOrg{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletAllianceOrg.GetKey().GetTags())
	span.SetTag("org", in.CloudletAllianceOrg.Key.Organization)

	obj := &in.CloudletAllianceOrg
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForAddCloudletAllianceOrg(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authzAddCloudletAllianceOrg(ctx, rc.Region, rc.Username, obj,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.AddCloudletAllianceOrgObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func RemoveCloudletAllianceOrg(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletAllianceOrg{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletAllianceOrg.GetKey().GetTags())
	span.SetTag("org", in.CloudletAllianceOrg.Key.Organization)

	obj := &in.CloudletAllianceOrg
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRemoveCloudletAllianceOrg(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.RemoveCloudletAllianceOrgObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func FindFlavorMatch(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionFlavorMatch{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.FlavorMatch.GetKey().GetTags())
	span.SetTag("org", in.FlavorMatch.Key.Organization)

	obj := &in.FlavorMatch
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionView); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.FindFlavorMatchObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func ShowFlavorsForCloudlet(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletKey{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)

	obj := &in.CloudletKey
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authzShowFlavorsForCloudlet(ctx, rc.Region, rc.Username, obj,
			ResourceCloudlets, ActionView); err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.FlavorKey) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.ShowFlavorsForCloudletStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func GetOrganizationsOnCloudlet(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletKey{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)

	obj := &in.CloudletKey
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Organization,
			ResourceCloudlets, ActionView); err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.Organization) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.GetOrganizationsOnCloudletStream(ctx, rc, obj, connCache, cb)
	if err != nil {
		return err
	}
	return nil
}

func RevokeAccessKey(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletKey{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)

	obj := &in.CloudletKey
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForRevokeAccessKey(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.RevokeAccessKeyObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func GenerateAccessKey(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletKey{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Organization)

	obj := &in.CloudletKey
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForGenerateAccessKey(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GenerateAccessKeyObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func GetCloudletGPUDriverLicenseConfig(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletKey{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	span.SetTag("org", in.CloudletKey.Key.Organization)

	obj := &in.CloudletKey
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudletAnalytics, ActionView); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.GetCloudletGPUDriverLicenseConfigObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func ShowCloudletInfo(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletInfo{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletInfo.GetKey().GetTags())
	span.SetTag("org", in.CloudletInfo.Key.Organization)

	obj := &in.CloudletInfo
	var authz ctrlclient.ShowCloudletInfoAuthz
	if !rc.SkipAuthz {
		authz, err = newShowCloudletInfoAuthz(ctx, rc.Region, rc.Username, ResourceCloudletAnalytics, ActionView)
		if err != nil {
			return err
		}
	}

	cb := func(res *edgeproto.CloudletInfo) error {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		return WriteStream(c, &payload)
	}
	err = ctrlclient.ShowCloudletInfoStream(ctx, rc, obj, connCache, authz, cb)
	if err != nil {
		return err
	}
	return nil
}

func InjectCloudletInfo(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletInfo{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletInfo.GetKey().GetTags())
	span.SetTag("org", in.CloudletInfo.Key.Organization)

	obj := &in.CloudletInfo
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForInjectCloudletInfo(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.InjectCloudletInfoObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}

func EvictCloudletInfo(c echo.Context) error {
	ctx := ormutil.GetContext(c)
	rc := &ormutil.RegionContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.Username = claims.Username

	in := ormapi.RegionCloudletInfo{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}
	rc.Region = in.Region
	rc.Database = database
	span := log.SpanFromContext(ctx)
	span.SetTag("region", in.Region)
	log.SetTags(span, in.CloudletInfo.GetKey().GetTags())
	span.SetTag("org", in.CloudletInfo.Key.Organization)

	obj := &in.CloudletInfo
	log.SetContextTags(ctx, edgeproto.GetTags(obj))
	if err := obj.IsValidArgsForEvictCloudletInfo(); err != nil {
		return err
	}
	if !rc.SkipAuthz {
		if err := authorized(ctx, rc.Username, obj.Key.Organization,
			ResourceCloudlets, ActionManage); err != nil {
			return err
		}
	}

	resp, err := ctrlclient.EvictCloudletInfoObj(ctx, rc, obj, connCache)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		return err
	}
	return ormutil.SetReply(c, resp)
}
