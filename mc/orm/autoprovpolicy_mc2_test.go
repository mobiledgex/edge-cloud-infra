// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: autoprovpolicy.proto

package orm

import (
	fmt "fmt"
	_ "github.com/gogo/googleapis/google/api"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
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

func badPermCreateAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermCreateAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badCreateAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, st, err := testutil.TestPermCreateAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermCreateAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermCreateAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermDeleteAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermDeleteAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badDeleteAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, st, err := testutil.TestPermDeleteAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermDeleteAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermDeleteAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermUpdateAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermUpdateAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badUpdateAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, st, err := testutil.TestPermUpdateAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermUpdateAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermUpdateAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermShowAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermShowAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badShowAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, st, err := testutil.TestPermShowAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermShowAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	_, status, err := testutil.TestPermShowAutoProvPolicy(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermAddAutoProvPolicyCloudlet(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicyCloudlet)) {
	_, status, err := testutil.TestPermAddAutoProvPolicyCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badAddAutoProvPolicyCloudlet(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.AutoProvPolicyCloudlet)) {
	_, st, err := testutil.TestPermAddAutoProvPolicyCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermAddAutoProvPolicyCloudlet(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicyCloudlet)) {
	_, status, err := testutil.TestPermAddAutoProvPolicyCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermRemoveAutoProvPolicyCloudlet(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicyCloudlet)) {
	_, status, err := testutil.TestPermRemoveAutoProvPolicyCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Forbidden")
	require.Equal(t, http.StatusForbidden, status)
}

func badRemoveAutoProvPolicyCloudlet(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, status int, modFuncs ...func(*edgeproto.AutoProvPolicyCloudlet)) {
	_, st, err := testutil.TestPermRemoveAutoProvPolicyCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, status, st)
}

func goodPermRemoveAutoProvPolicyCloudlet(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicyCloudlet)) {
	_, status, err := testutil.TestPermRemoveAutoProvPolicyCloudlet(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	badPermCreateAutoProvPolicy(t, mcClient, uri, token, region, org, modFuncs...)
	badPermUpdateAutoProvPolicy(t, mcClient, uri, token, region, org, modFuncs...)
	badPermDeleteAutoProvPolicy(t, mcClient, uri, token, region, org, modFuncs...)
}

func badPermTestShowAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	list, status, err := testutil.TestPermShowAutoProvPolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	goodPermCreateAutoProvPolicy(t, mcClient, uri, token, region, org)
	goodPermUpdateAutoProvPolicy(t, mcClient, uri, token, region, org)
	goodPermDeleteAutoProvPolicy(t, mcClient, uri, token, region, org)

	// make sure region check works
	_, status, err := testutil.TestPermCreateAutoProvPolicy(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermUpdateAutoProvPolicy(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermDeleteAutoProvPolicy(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShowAutoProvPolicy(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShowAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string, count int) {
	list, status, err := testutil.TestPermShowAutoProvPolicy(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	list, status, err = testutil.TestPermShowAutoProvPolicy(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestAutoProvPolicy(t *testing.T, mcClient *mctestclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.AutoProvPolicy)) {
	badPermTestAutoProvPolicy(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestShowAutoProvPolicy(t, mcClient, uri, token1, region, org2)
	badPermTestAutoProvPolicy(t, mcClient, uri, token2, region, org1, modFuncs...)
	badPermTestShowAutoProvPolicy(t, mcClient, uri, token2, region, org1)

	goodPermTestAutoProvPolicy(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestAutoProvPolicy(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}
