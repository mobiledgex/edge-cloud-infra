// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoscalepolicy.proto

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
import _ "github.com/gogo/protobuf/gogoproto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var _ = edgeproto.GetFields

func badPermCreateAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermCreateAutoScalePolicy(mcClient, uri, token, region, org)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermCreateAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermCreateAutoScalePolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermDeleteAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermDeleteAutoScalePolicy(mcClient, uri, token, region, org)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermDeleteAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermDeleteAutoScalePolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermUpdateAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermUpdateAutoScalePolicy(mcClient, uri, token, region, org)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermUpdateAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermUpdateAutoScalePolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermShowAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermShowAutoScalePolicy(mcClient, uri, token, region, org)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermShowAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermShowAutoScalePolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	badPermCreateAutoScalePolicy(t, mcClient, uri, token, region, org)
	badPermUpdateAutoScalePolicy(t, mcClient, uri, token, region, org)
	badPermDeleteAutoScalePolicy(t, mcClient, uri, token, region, org)
}

func badPermTestShowAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	list, status, err := testutil.TestPermShowAutoScalePolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, showcount int) {
	goodPermCreateAutoScalePolicy(t, mcClient, uri, token, region, org)
	goodPermUpdateAutoScalePolicy(t, mcClient, uri, token, region, org)
	goodPermDeleteAutoScalePolicy(t, mcClient, uri, token, region, org)

	// make sure region check works
	_, status, err := testutil.TestPermCreateAutoScalePolicy(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermUpdateAutoScalePolicy(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermDeleteAutoScalePolicy(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShowAutoScalePolicy(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShowAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, count int) {
	list, status, err := testutil.TestPermShowAutoScalePolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	list, status, err = testutil.TestPermShowAutoScalePolicy(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestAutoScalePolicy(t *testing.T, mcClient *ormclient.Client, uri, token1, token2, region, org1, org2 string, showcount int) {
	badPermTestAutoScalePolicy(t, mcClient, uri, token1, region, org2)
	badPermTestShowAutoScalePolicy(t, mcClient, uri, token1, region, org2)
	badPermTestAutoScalePolicy(t, mcClient, uri, token2, region, org1)
	badPermTestShowAutoScalePolicy(t, mcClient, uri, token2, region, org1)

	goodPermTestAutoScalePolicy(t, mcClient, uri, token1, region, org1, showcount)
	goodPermTestAutoScalePolicy(t, mcClient, uri, token2, region, org2, showcount)
}
