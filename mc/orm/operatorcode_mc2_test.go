// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: operatorcode.proto

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

func badPermCreateOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.OperatorCode)) {
	_, status, err := testutil.TestPermCreateOperatorCode(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermCreateOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.OperatorCode)) {
	_, status, err := testutil.TestPermCreateOperatorCode(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermDeleteOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.OperatorCode)) {
	_, status, err := testutil.TestPermDeleteOperatorCode(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermDeleteOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.OperatorCode)) {
	_, status, err := testutil.TestPermDeleteOperatorCode(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermShowOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.OperatorCode)) {
	_, status, err := testutil.TestPermShowOperatorCode(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermShowOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.OperatorCode)) {
	_, status, err := testutil.TestPermShowOperatorCode(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.OperatorCode)) {
	badPermCreateOperatorCode(t, mcClient, uri, token, region, org, modFuncs...)
	badPermDeleteOperatorCode(t, mcClient, uri, token, region, org, modFuncs...)
}

func badPermTestShowOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	list, status, err := testutil.TestPermShowOperatorCode(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.OperatorCode)) {
	goodPermCreateOperatorCode(t, mcClient, uri, token, region, org)
	goodPermDeleteOperatorCode(t, mcClient, uri, token, region, org)

	// make sure region check works
	_, status, err := testutil.TestPermCreateOperatorCode(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermDeleteOperatorCode(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShowOperatorCode(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShowOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, count int) {
	list, status, err := testutil.TestPermShowOperatorCode(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	list, status, err = testutil.TestPermShowOperatorCode(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestOperatorCode(t *testing.T, mcClient *ormclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.OperatorCode)) {
	badPermTestOperatorCode(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestShowOperatorCode(t, mcClient, uri, token1, region, org2)
	badPermTestOperatorCode(t, mcClient, uri, token2, region, org1, modFuncs...)
	badPermTestShowOperatorCode(t, mcClient, uri, token2, region, org1)

	goodPermTestOperatorCode(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestOperatorCode(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}
