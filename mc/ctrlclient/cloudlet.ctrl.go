// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package ctrlclient

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"io"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func CreateGPUDriverStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriver, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.CreateGPUDriver(ctx, obj)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteGPUDriverStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriver, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.DeleteGPUDriver(ctx, obj)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateGPUDriverStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriver, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.UpdateGPUDriver(ctx, obj)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

type ShowGPUDriverAuthz interface {
	Ok(obj *edgeproto.GPUDriver) (bool, bool)
	Filter(obj *edgeproto.GPUDriver)
}

func ShowGPUDriverStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriver, connObj ClientConnMgr, authz ShowGPUDriverAuthz, cb func(res *edgeproto.GPUDriver) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.ShowGPUDriver(ctx, obj)
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
		if !rc.SkipAuthz {
			if authz != nil {
				authzOk, filterOutput := authz.Ok(res)
				if !authzOk {
					continue
				}
				if filterOutput {
					authz.Filter(res)
				}
			}
		}
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddGPUDriverBuildStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriverBuildMember, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.AddGPUDriverBuild(ctx, obj)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveGPUDriverBuildStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriverBuildMember, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.RemoveGPUDriverBuild(ctx, obj)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetGPUDriverBuildURLObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriverBuildMember, connObj ClientConnMgr) (*edgeproto.GPUDriverBuildURL, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GetGPUDriverBuildURL(ctx, obj)
}

func GetGPUDriverLicenseConfigObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.GPUDriverKey, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewGPUDriverApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GetGPUDriverLicenseConfig(ctx, obj)
}

func CreateCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, connObj ClientConnMgr, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

type ShowCloudletAuthz interface {
	Ok(obj *edgeproto.Cloudlet) (bool, bool)
	Filter(obj *edgeproto.Cloudlet)
}

func ShowCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, connObj ClientConnMgr, authz ShowCloudletAuthz, cb func(res *edgeproto.Cloudlet) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
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
		if !rc.SkipAuthz {
			if authz != nil {
				authzOk, filterOutput := authz.Ok(res)
				if !authzOk {
					continue
				}
				if filterOutput {
					authz.Filter(res)
				}
			}
		}
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCloudletManifestObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletKey, connObj ClientConnMgr) (*edgeproto.CloudletManifest, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GetCloudletManifest(ctx, obj)
}

func GetCloudletPropsObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletProps, connObj ClientConnMgr) (*edgeproto.CloudletProps, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GetCloudletProps(ctx, obj)
}

func GetCloudletResourceQuotaPropsObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletResourceQuotaProps, connObj ClientConnMgr) (*edgeproto.CloudletResourceQuotaProps, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GetCloudletResourceQuotaProps(ctx, obj)
}

func GetCloudletResourceUsageObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletResourceUsage, connObj ClientConnMgr) (*edgeproto.CloudletResourceUsage, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GetCloudletResourceUsage(ctx, obj)
}

func AddCloudletResMappingObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletResMap, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.AddCloudletResMapping(ctx, obj)
}

func RemoveCloudletResMappingObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletResMap, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.RemoveCloudletResMapping(ctx, obj)
}

func AddCloudletAllianceOrgObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletAllianceOrg, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.AddCloudletAllianceOrg(ctx, obj)
}

func RemoveCloudletAllianceOrgObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletAllianceOrg, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.RemoveCloudletAllianceOrg(ctx, obj)
}

func FindFlavorMatchObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.FlavorMatch, connObj ClientConnMgr) (*edgeproto.FlavorMatch, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.FindFlavorMatch(ctx, obj)
}

func ShowFlavorsForCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletKey, connObj ClientConnMgr, cb func(res *edgeproto.FlavorKey) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.ShowFlavorsForCloudlet(ctx, obj)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetOrganizationsOnCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletKey, connObj ClientConnMgr, cb func(res *edgeproto.Organization) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.GetOrganizationsOnCloudlet(ctx, obj)
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
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func RevokeAccessKeyObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletKey, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.RevokeAccessKey(ctx, obj)
}

func GenerateAccessKeyObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletKey, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GenerateAccessKey(ctx, obj)
}

func GetCloudletGPUDriverLicenseConfigObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletKey, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.GetCloudletGPUDriverLicenseConfig(ctx, obj)
}

type ShowCloudletInfoAuthz interface {
	Ok(obj *edgeproto.CloudletInfo) (bool, bool)
	Filter(obj *edgeproto.CloudletInfo)
}

func ShowCloudletInfoStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletInfo, connObj ClientConnMgr, authz ShowCloudletInfoAuthz, cb func(res *edgeproto.CloudletInfo) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewCloudletInfoApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
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
		if !rc.SkipAuthz {
			if authz != nil {
				authzOk, filterOutput := authz.Ok(res)
				if !authzOk {
					continue
				}
				if filterOutput {
					authz.Filter(res)
				}
			}
		}
		err = cb(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func InjectCloudletInfoObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletInfo, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletInfoApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.InjectCloudletInfo(ctx, obj)
}

func EvictCloudletInfoObj(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.CloudletInfo, connObj ClientConnMgr) (*edgeproto.Result, error) {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return nil, err
	}
	api := edgeproto.NewCloudletInfoApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	return api.EvictCloudletInfo(ctx, obj)
}
