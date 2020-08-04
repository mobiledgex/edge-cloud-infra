// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: vmpool.proto

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
import _ "github.com/gogo/protobuf/types"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Auto-generated code: DO NOT EDIT

var _ = edgeproto.GetFields

func badPermCreateVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermCreateVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermCreateVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermCreateVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermDeleteVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermDeleteVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermDeleteVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermDeleteVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermUpdateVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermUpdateVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermUpdateVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermUpdateVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermShowVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermShowVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermShowVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	_, status, err := testutil.TestPermShowVMPool(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermAddVMPoolMember(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPoolMember)) {
	_, status, err := testutil.TestPermAddVMPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermAddVMPoolMember(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPoolMember)) {
	_, status, err := testutil.TestPermAddVMPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermRemoveVMPoolMember(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPoolMember)) {
	_, status, err := testutil.TestPermRemoveVMPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermRemoveVMPoolMember(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPoolMember)) {
	_, status, err := testutil.TestPermRemoveVMPoolMember(mcClient, uri, token, region, org, modFuncs...)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

// This tests the user cannot modify the object because the obj belongs to
// an organization that the user does not have permissions for.
func badPermTestVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, modFuncs ...func(*edgeproto.VMPool)) {
	badPermCreateVMPool(t, mcClient, uri, token, region, org, modFuncs...)
	badPermUpdateVMPool(t, mcClient, uri, token, region, org, modFuncs...)
	badPermDeleteVMPool(t, mcClient, uri, token, region, org, modFuncs...)
}

func badPermTestShowVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	// show is allowed but won't show anything
	list, status, err := testutil.TestPermShowVMPool(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(list))
}

// This tests the user can modify the object because the obj belongs to
// an organization that the user has permissions for.
func goodPermTestVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, showcount int, modFuncs ...func(*edgeproto.VMPool)) {
	goodPermCreateVMPool(t, mcClient, uri, token, region, org)
	goodPermUpdateVMPool(t, mcClient, uri, token, region, org)
	goodPermDeleteVMPool(t, mcClient, uri, token, region, org)

	// make sure region check works
	_, status, err := testutil.TestPermCreateVMPool(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermUpdateVMPool(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	_, status, err = testutil.TestPermDeleteVMPool(mcClient, uri, token, "bad region", org, modFuncs...)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)

	goodPermTestShowVMPool(t, mcClient, uri, token, region, org, showcount)
}

func goodPermTestShowVMPool(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string, count int) {
	list, status, err := testutil.TestPermShowVMPool(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, count, len(list))

	// make sure region check works
	list, status, err = testutil.TestPermShowVMPool(mcClient, uri, token, "bad region", org)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, 0, len(list))
}

// Test permissions for user with token1 who should have permissions for
// modifying obj1, and user with token2 who should have permissions for obj2.
// They should not have permissions to modify each other's objects.
func permTestVMPool(t *testing.T, mcClient *ormclient.Client, uri, token1, token2, region, org1, org2 string, showcount int, modFuncs ...func(*edgeproto.VMPool)) {
	badPermTestVMPool(t, mcClient, uri, token1, region, org2, modFuncs...)
	badPermTestShowVMPool(t, mcClient, uri, token1, region, org2)
	badPermTestVMPool(t, mcClient, uri, token2, region, org1, modFuncs...)
	badPermTestShowVMPool(t, mcClient, uri, token2, region, org1)

	goodPermTestVMPool(t, mcClient, uri, token1, region, org1, showcount, modFuncs...)
	goodPermTestVMPool(t, mcClient, uri, token2, region, org2, showcount, modFuncs...)
}
