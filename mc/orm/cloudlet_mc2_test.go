// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudlet.proto

package orm

import edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
import "net/http"
import "testing"
import "github.com/stretchr/testify/require"
import "github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
import "github.com/mobiledgex/edge-cloud-infra/mc/orm/testutil"
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

var _ = edgeproto.GetFields

func badPermCreateCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermCreateCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermCreateCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermCreateCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermDeleteCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermDeleteCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermDeleteCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermDeleteCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermUpdateCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermUpdateCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermUpdateCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermUpdateCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermShowCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermShowCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermShowCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermShowCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermGetCloudletManifest(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermGetCloudletManifest(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermGetCloudletManifest(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	_, status, err := testutil.TestPermGetCloudletManifest(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermGetCloudletProps(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletProps)) {
	_, status, err := testutil.TestPermGetCloudletProps(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermGetCloudletProps(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletProps)) {
	_, status, err := testutil.TestPermGetCloudletProps(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermAddCloudletResMapping(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletResMap)) {
	_, status, err := testutil.TestPermAddCloudletResMapping(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermAddCloudletResMapping(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletResMap)) {
	_, status, err := testutil.TestPermAddCloudletResMapping(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermRemoveCloudletResMapping(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletResMap)) {
	_, status, err := testutil.TestPermRemoveCloudletResMapping(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermRemoveCloudletResMapping(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletResMap)) {
	_, status, err := testutil.TestPermRemoveCloudletResMapping(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermFindFlavorMatch(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.FlavorMatch)) {
	_, status, err := testutil.TestPermFindFlavorMatch(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermFindFlavorMatch(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.FlavorMatch)) {
	_, status, err := testutil.TestPermFindFlavorMatch(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Cloudlet)) {
	badPermCreateCloudlet(t, mcClient, uri, token, region, org, modFuncs...)
	badPermUpdateCloudlet(t, mcClient, uri, token, region, org, modFuncs...)
	badPermDeleteCloudlet(t, mcClient, uri, token, region, org, modFuncs...)
}

func badPermTestShowCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	list, status, err := testutil.TestPermShowCloudlet(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.Cloudlet)) {
	goodPermCreateCloudlet(t, mcClient, uri, token, region, org)
	goodPermUpdateCloudlet(t, mcClient, uri, token, region, org)
	goodPermDeleteCloudlet(t, mcClient, uri, token, region, org)

	// make sure region check works
	_, status, err := testutil.TestPermCreateCloudlet(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermUpdateCloudlet(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermDeleteCloudlet(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShowCloudlet(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShowCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, count int) {
	list, status, err := testutil.TestPermShowCloudlet(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	list, status, err = testutil.TestPermShowCloudlet(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestCloudlet(t *testing.T, mcClient *ormclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.Cloudlet)) {
	badPermTestCloudlet(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestShowCloudlet(t, mcClient, uri, token1, region, org2)
	badPermTestCloudlet(t, mcClient, uri, token2, region, org1, modFuncs...)
	badPermTestShowCloudlet(t, mcClient, uri, token2, region, org1)

	goodPermTestCloudlet(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestCloudlet(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}

var _ = edgeproto.GetFields

func badPermShowCloudletInfo(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletInfo)) {
	_, status, err := testutil.TestPermShowCloudletInfo(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermShowCloudletInfo(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletInfo)) {
	_, status, err := testutil.TestPermShowCloudletInfo(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermInjectCloudletInfo(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletInfo)) {
	_, status, err := testutil.TestPermInjectCloudletInfo(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermInjectCloudletInfo(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletInfo)) {
	_, status, err := testutil.TestPermInjectCloudletInfo(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermEvictCloudletInfo(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletInfo)) {
	_, status, err := testutil.TestPermEvictCloudletInfo(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermEvictCloudletInfo(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletInfo)) {
	_, status, err := testutil.TestPermEvictCloudletInfo(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}
