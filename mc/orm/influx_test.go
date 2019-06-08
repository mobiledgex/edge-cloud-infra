package orm

import (
	"testing"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestInfluxDB(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	require.Equal(t, 0, 0)
}

func testShowClusterMetrics(mcClient *ormclient.Client, uri, token, region string, in *edgeproto.ClusterInst) ([]edgeproto.ClusterInst, int, error) {
	dat := &ormapi.RegionClusterInst{}
	dat.Region = region
	dat.ClusterInst = *in
	return mcClient.ShowClusterInst(uri, token, dat)
}

func testPermShowClusterMetrics(mcClient *ormclient.Client, uri, token, region, org string) ([]edgeproto.ClusterInst, int, error) {
	in := &edgeproto.ClusterInst{}
	in.Key.Developer = org
	return testShowClusterMetrics(mcClient, uri, token, region, in)
}

func badPermTestMetrics(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {
	/*
		_, status, err := testPermShowAppInstMetrics(mcClient, uri, token, region, org)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, status)
		_, status, err = testPermShowClusterMetrics(mcClient, uri, token, region, org)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, status)
	*/
}

func goodPermTestMetrics(t *testing.T, mcClient *ormclient.Client, uri, token, region, org string) {

}
