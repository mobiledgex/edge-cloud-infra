// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: stream.proto

package orm

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/orm/testutil"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	_ "github.com/mobiledgex/edge-cloud/protogen"
	"github.com/stretchr/testify/require"
	math "math"
	"net/http"
	"testing"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var _ = edgeproto.GetFields

func badPermStreamAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AppInstKey)) {
	_, status, err := testutil.TestPermStreamAppInst(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermStreamAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AppInstKey)) {
	_, status, err := testutil.TestPermStreamAppInst(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermStreamClusterInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.ClusterInstKey)) {
	_, status, err := testutil.TestPermStreamClusterInst(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermStreamClusterInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.ClusterInstKey)) {
	_, status, err := testutil.TestPermStreamClusterInst(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermStreamCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletKey)) {
	_, status, err := testutil.TestPermStreamCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermStreamCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletKey)) {
	_, status, err := testutil.TestPermStreamCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}
