// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: network.proto

package ctrlapi

import (
	"context"
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
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

func CreateNetworkStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Network, connObj RegionConn, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewNetworkApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.CreateNetwork(ctx, obj)
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

func DeleteNetworkStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Network, connObj RegionConn, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewNetworkApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.DeleteNetwork(ctx, obj)
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

func UpdateNetworkStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Network, connObj RegionConn, cb func(res *edgeproto.Result) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewNetworkApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.UpdateNetwork(ctx, obj)
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

func ShowNetworkStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Network, connObj RegionConn, authz authzShow, cb func(res *edgeproto.Network) error) error {
	conn, err := connObj.GetRegionConn(ctx, rc.Region)
	if err != nil {
		return err
	}
	api := edgeproto.NewNetworkApiClient(conn)
	log.SpanLog(ctx, log.DebugLevelApi, "start controller api")
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish controller api")
	stream, err := api.ShowNetwork(ctx, obj)
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
				if !authz.Ok(res.Key.Organization) {
					continue
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
