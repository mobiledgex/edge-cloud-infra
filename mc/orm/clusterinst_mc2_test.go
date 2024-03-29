// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: clusterinst.proto

package orm

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/mctestclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/orm/testutil"
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

func badPermCreateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermCreateClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badCreateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, st, err := testutil.TestPermCreateClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermCreateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermCreateClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionCreateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	out, status, err := testutil.TestPermCreateClusterInst(mcClient, uri, token, "bad region", org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	if err.Error() == "Forbidden" {
		require.Equal(t, http.StatusForbidden, status)
	} else {
		require.Contains(t, err.Error(), "\"bad region\" not found")
		require.Equal(t, http.StatusBadRequest, status)
	}
	_ = out
}

var _ = edgeproto.GetFields

func badPermDeleteClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermDeleteClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badDeleteClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, st, err := testutil.TestPermDeleteClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermDeleteClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermDeleteClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionDeleteClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	out, status, err := testutil.TestPermDeleteClusterInst(mcClient, uri, token, "bad region", org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	if err.Error() == "Forbidden" {
		require.Equal(t, http.StatusForbidden, status)
	} else {
		require.Contains(t, err.Error(), "\"bad region\" not found")
		require.Equal(t, http.StatusBadRequest, status)
	}
	_ = out
}

var _ = edgeproto.GetFields

func badPermUpdateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermUpdateClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badUpdateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, st, err := testutil.TestPermUpdateClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermUpdateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermUpdateClusterInst(mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionUpdateClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	out, status, err := testutil.TestPermUpdateClusterInst(mcClient, uri, token, "bad region", org, targetCloudlet, modFuncs...)
	require.NotNil(t, err)
	if err.Error() == "Forbidden" {
		require.Equal(t, http.StatusForbidden, status)
	} else {
		require.Contains(t, err.Error(), "\"bad region\" not found")
		require.Equal(t, http.StatusBadRequest, status)
	}
	_ = out
}

var _ = edgeproto.GetFields

func badPermShowClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermShowClusterInst(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badShowClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, st, err := testutil.TestPermShowClusterInst(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermShowClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.ClusterInst)) {
	_, status, err := testutil.TestPermShowClusterInst(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionShowClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.ClusterInst)) {
	out, status, err := testutil.TestPermShowClusterInst(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	if err.Error() == "Forbidden" {
		require.Equal(t, http.StatusForbidden, status)
	} else {
		require.Contains(t, err.Error(), "\"bad region\" not found")
		require.Equal(t, http.StatusBadRequest, status)
	}
	require.Equal(t, 0, len(out))
}

var _ = edgeproto.GetFields

func badPermDeleteIdleReservableClusterInsts(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) {
	_, status, err := testutil.TestPermDeleteIdleReservableClusterInsts(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badDeleteIdleReservableClusterInsts(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) {
	_, st, err := testutil.TestPermDeleteIdleReservableClusterInsts(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermDeleteIdleReservableClusterInsts(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) {
	_, status, err := testutil.TestPermDeleteIdleReservableClusterInsts(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionDeleteIdleReservableClusterInsts(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) {
	out, status, err := testutil.TestPermDeleteIdleReservableClusterInsts(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	if err.Error() == "Forbidden" {
		require.Equal(t, http.StatusForbidden, status)
	} else {
		require.Contains(t, err.Error(), "\"bad region\" not found")
		require.Equal(t, http.StatusBadRequest, status)
	}
	_ = out
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, modFuncs ...func(*edgeproto.ClusterInst)) {
	badPermCreateClusterInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	badPermUpdateClusterInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	badPermDeleteClusterInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
}
func badPermTestShowClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	var status int
	var err error
	list0, status, err := testutil.TestPermShowClusterInst(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list0))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, targetCloudlet *edgeproto.CloudletKey, showcount int, modFuncs ...func(*edgeproto.ClusterInst)) {
	goodPermCreateClusterInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	goodPermUpdateClusterInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	goodPermDeleteClusterInst(t, mcClient, uri, token, region, org, targetCloudlet, modFuncs...)
	goodPermTestShowClusterInst(t, mcClient, uri, token, region, org, showcount)
	// make sure region check works
	badRegionCreateClusterInst(t, mcClient, uri, token, org, targetCloudlet, modFuncs...)
	badRegionUpdateClusterInst(t, mcClient, uri, token, org, targetCloudlet, modFuncs...)
	badRegionDeleteClusterInst(t, mcClient, uri, token, org, targetCloudlet, modFuncs...)
}
func goodPermTestShowClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, count int) {
	var status int
	var err error
	list0, status, err := testutil.TestPermShowClusterInst(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list0))

	badRegionShowClusterInst(t, mcClient, uri, token, org)
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestClusterInst(t *testing.T, mcClient *mctestclient.Client, uri, token1, token2, region, org1, org2 string, targetCloudlet *edgeproto.CloudletKey, showcount int, modFuncs ...func(*edgeproto.ClusterInst)) {
	badPermTestClusterInst(t, mcClient, uri, token1, region, org2, targetCloudlet, modFuncs...)
	badPermTestClusterInst(t, mcClient, uri, token2, region, org1, targetCloudlet, modFuncs...)
	badPermTestShowClusterInst(t, mcClient, uri, token1, region, org2)
	badPermTestShowClusterInst(t, mcClient, uri, token2, region, org1)
	goodPermTestClusterInst(t, mcClient, uri, token1, region, org1, targetCloudlet, showcount, modFuncs...)
	goodPermTestClusterInst(t, mcClient, uri, token2, region, org2, targetCloudlet, showcount, modFuncs...)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestIdleReservableClusterInsts(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) {
	badPermDeleteIdleReservableClusterInsts(t, mcClient, uri, token, region, org, modFuncs...)
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestIdleReservableClusterInsts(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) {
	goodPermDeleteIdleReservableClusterInsts(t, mcClient, uri, token, region, org, modFuncs...)
	// make sure region check works
	badRegionDeleteIdleReservableClusterInsts(t, mcClient, uri, token, org, modFuncs...)
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestIdleReservableClusterInsts(t *testing.T, mcClient *mctestclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.IdleReservableClusterInsts)) {
	badPermTestIdleReservableClusterInsts(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestIdleReservableClusterInsts(t, mcClient, uri, token2, region, org1, modFuncs...)
	goodPermTestIdleReservableClusterInsts(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestIdleReservableClusterInsts(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}
