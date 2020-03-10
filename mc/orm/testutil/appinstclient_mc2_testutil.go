// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinstclient.proto

package testutil

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/googleapis/google/api"
import _ "github.com/mobiledgex/edge-cloud/protogen"
import _ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

func TestShowAppInstClient(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.AppInstClientKey) ([]edgeproto.AppInstClient, int, error) {
	dat := &ormapi.RegionAppInstClientKey{}
	dat.Region = region
	dat.AppInstClientKey = *in
	return mcClient.ShowAppInstClient(uri, token, dat)
}
func TestPermShowAppInstClient(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.AppInstClient, int, error) {
	in := &edgeproto.AppInstClientKey{}
	in.Key.AppKey.Organization = org
	return TestShowAppInstClient(mcClient, uri, token, region, in)
}
