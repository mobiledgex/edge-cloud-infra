// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flavor.proto

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

func badPermCreateFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermCreateFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermCreateFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermCreateFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermDeleteFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermDeleteFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermDeleteFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermDeleteFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermUpdateFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermUpdateFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermUpdateFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermUpdateFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermShowFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermShowFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermShowFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermShowFlavor(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermAddFlavorRes(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermAddFlavorRes(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermAddFlavorRes(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermAddFlavorRes(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermRemoveFlavorRes(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermRemoveFlavorRes(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermRemoveFlavorRes(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	_, status, err := testutil.TestPermRemoveFlavorRes(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.Flavor)) {
	badPermCreateFlavor(t, mcClient, uri, token, region, org, modFuncs...)
	badPermUpdateFlavor(t, mcClient, uri, token, region, org, modFuncs...)
	badPermDeleteFlavor(t, mcClient, uri, token, region, org, modFuncs...)
}

func badPermTestShowFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	list, status, err := testutil.TestPermShowFlavor(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.Flavor)) {
	goodPermCreateFlavor(t, mcClient, uri, token, region, org)
	goodPermUpdateFlavor(t, mcClient, uri, token, region, org)
	goodPermDeleteFlavor(t, mcClient, uri, token, region, org)

	// make sure region check works
	_, status, err := testutil.TestPermCreateFlavor(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermUpdateFlavor(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermDeleteFlavor(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShowFlavor(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShowFlavor(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, count int) {
	list, status, err := testutil.TestPermShowFlavor(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	list, status, err = testutil.TestPermShowFlavor(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestFlavor(t *testing.T, mcClient *ormclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.Flavor)) {
	badPermTestFlavor(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestShowFlavor(t, mcClient, uri, token1, region, org2)
	badPermTestFlavor(t, mcClient, uri, token2, region, org1, modFuncs...)
	badPermTestShowFlavor(t, mcClient, uri, token2, region, org1)

	goodPermTestFlavor(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestFlavor(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}
