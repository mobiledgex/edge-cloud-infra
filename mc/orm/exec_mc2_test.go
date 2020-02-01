// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: exec.proto

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

func badPermRunCommand(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermRunCommand(mcClient, uri, token, region, org)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermRunCommand(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermRunCommand(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}

var _ = edgeproto.GetFields

func badPermViewLogs(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermViewLogs(mcClient, uri, token, region, org)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermViewLogs(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	_, status, err := testutil.TestPermViewLogs(mcClient, uri, token, region, org)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
}
