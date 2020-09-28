// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: appinst.proto

package orm

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/orm/testutil"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	_ "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
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

func badPermCreateAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermCreateAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermCreateAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermCreateAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermDeleteAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermDeleteAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermDeleteAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermDeleteAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermRefreshAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermRefreshAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermRefreshAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermRefreshAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermUpdateAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermUpdateAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermUpdateAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermUpdateAppInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermShowAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermShowAppInst(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermShowAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AppInst)) {
	_, status, err := testutil.TestPermShowAppInst(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.AppInst)) {
	badPermCreateAppInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	badPermUpdateAppInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	badPermDeleteAppInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
}

func badPermTestShowAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	list, status, err := testutil.TestPermShowAppInst(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, showcount int, modFuncs ...func(*edgeproto.AppInst)) {
	goodPermCreateAppInst(t, mcClient, uri, token, region, org, targetCloudlet)
	goodPermUpdateAppInst(t, mcClient, uri, token, region, org, targetCloudlet)
	goodPermDeleteAppInst(t, mcClient, uri, token, region, org, targetCloudlet)

	// make sure region check works
	_, status, err := testutil.TestPermCreateAppInst(mcClient, uri, token, "bad region", org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermUpdateAppInst(mcClient, uri, token, "bad region", org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermDeleteAppInst(mcClient, uri, token, "bad region", org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShowAppInst(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShowAppInst(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, count int) {
	list, status, err := testutil.TestPermShowAppInst(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	list, status, err = testutil.TestPermShowAppInst(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestAppInst(t *testing.T, mcClient *ormclient.Client, uri, token1, token2, region, org1, org2 string, targetCloudlet *edgeproto.CloudletKey, showcount int, modFuncs ...func(*edgeproto.AppInst)) {
	badPermTestAppInst(t, mcClient, uri, token1, region, org2, targetCloudlet, modFuncs...)
	badPermTestShowAppInst(t, mcClient, uri, token1, region, org2)
	badPermTestAppInst(t, mcClient, uri, token2, region, org1, targetCloudlet, modFuncs...)
	badPermTestShowAppInst(t, mcClient, uri, token2, region, org1)

	goodPermTestAppInst(t, mcClient, uri, token1, region, org1, targetCloudlet, showcount, modFuncs...)
	goodPermTestAppInst(t, mcClient, uri, token2, region, org2, targetCloudlet, showcount, modFuncs...)
}
