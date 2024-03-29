// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cloudletpool.proto

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

func badPermCreateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermCreateCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badCreateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, st, err := testutil.TestPermCreateCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermCreateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermCreateCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionCreateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	out, status, err := testutil.TestPermCreateCloudletPool(mcClient, uri, token, "bad region", org, modFuncs...)
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

func badPermDeleteCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermDeleteCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badDeleteCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, st, err := testutil.TestPermDeleteCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermDeleteCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermDeleteCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionDeleteCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	out, status, err := testutil.TestPermDeleteCloudletPool(mcClient, uri, token, "bad region", org, modFuncs...)
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

func badPermUpdateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermUpdateCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badUpdateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, st, err := testutil.TestPermUpdateCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermUpdateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermUpdateCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionUpdateCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	out, status, err := testutil.TestPermUpdateCloudletPool(mcClient, uri, token, "bad region", org, modFuncs...)
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

func badPermShowCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermShowCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badShowCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, st, err := testutil.TestPermShowCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermShowCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	_, status, err := testutil.TestPermShowCloudletPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionShowCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	out, status, err := testutil.TestPermShowCloudletPool(mcClient, uri, token, "bad region", org, modFuncs...)
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

func badPermAddCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	_, status, err := testutil.TestPermAddCloudletPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badAddCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	_, st, err := testutil.TestPermAddCloudletPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermAddCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	_, status, err := testutil.TestPermAddCloudletPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionAddCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	out, status, err := testutil.TestPermAddCloudletPoolMember(mcClient, uri, token, "bad region", org, modFuncs...)
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

func badPermRemoveCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	_, status, err := testutil.TestPermRemoveCloudletPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badRemoveCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	_, st, err := testutil.TestPermRemoveCloudletPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermRemoveCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	_, status, err := testutil.TestPermRemoveCloudletPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

func badRegionRemoveCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, org string, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	out, status, err := testutil.TestPermRemoveCloudletPoolMember(mcClient, uri, token, "bad region", org, modFuncs...)
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
func badPermTestCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPool)) {
	badPermCreateCloudletPool(t, mcClient, uri, token, region, org, modFuncs...)
	badPermUpdateCloudletPool(t, mcClient, uri, token, region, org, modFuncs...)
	badPermDeleteCloudletPool(t, mcClient, uri, token, region, org, modFuncs...)
}
func badPermTestShowCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	var status int
	var err error
	list0, status, err := testutil.TestPermShowCloudletPool(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list0))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.CloudletPool)) {
	goodPermCreateCloudletPool(t, mcClient, uri, token, region, org, modFuncs...)
	goodPermUpdateCloudletPool(t, mcClient, uri, token, region, org, modFuncs...)
	goodPermDeleteCloudletPool(t, mcClient, uri, token, region, org, modFuncs...)
	goodPermTestShowCloudletPool(t, mcClient, uri, token, region, org, showcount)
	// make sure region check works
	badRegionCreateCloudletPool(t, mcClient, uri, token, org, modFuncs...)
	badRegionUpdateCloudletPool(t, mcClient, uri, token, org, modFuncs...)
	badRegionDeleteCloudletPool(t, mcClient, uri, token, org, modFuncs...)
}
func goodPermTestShowCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, count int) {
	var status int
	var err error
	list0, status, err := testutil.TestPermShowCloudletPool(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list0))

	badRegionShowCloudletPool(t, mcClient, uri, token, org)
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestCloudletPool(t *testing.T, mcClient *mctestclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.CloudletPool)) {
	badPermTestCloudletPool(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestCloudletPool(t, mcClient, uri, token2, region, org1, modFuncs...)
	badPermTestShowCloudletPool(t, mcClient, uri, token1, region, org2)
	badPermTestShowCloudletPool(t, mcClient, uri, token2, region, org1)
	goodPermTestCloudletPool(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestCloudletPool(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	badPermAddCloudletPoolMember(t, mcClient, uri, token, region, org, modFuncs...)
	badPermRemoveCloudletPoolMember(t, mcClient, uri, token, region, org, modFuncs...)
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	goodPermAddCloudletPoolMember(t, mcClient, uri, token, region, org, modFuncs...)
	goodPermRemoveCloudletPoolMember(t, mcClient, uri, token, region, org, modFuncs...)
	// make sure region check works
	badRegionAddCloudletPoolMember(t, mcClient, uri, token, org, modFuncs...)
	badRegionRemoveCloudletPoolMember(t, mcClient, uri, token, org, modFuncs...)
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestCloudletPoolMember(t *testing.T, mcClient *mctestclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.CloudletPoolMember)) {
	badPermTestCloudletPoolMember(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestCloudletPoolMember(t, mcClient, uri, token2, region, org1, modFuncs...)
	goodPermTestCloudletPoolMember(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestCloudletPoolMember(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}
