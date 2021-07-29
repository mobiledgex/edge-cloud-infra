package orm

import (
	"context"
	"net/http"
	"testing"

	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/mctestclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/require"
)

func testPermShowClusterMetrics(mcClient *mctestclient.Client, uri, token, region, org, selector string, data *edgeproto.ClusterInstKey) (*ormapi.AllMetrics, int, error) {
	in := &edgeproto.ClusterInstKey{}
	if data != nil {
		in = data
	} else {
		in.ClusterKey.Name = "testcluster"
	}
	in.Organization = org
	dat := &ormapi.RegionClusterInstMetrics{}
	dat.Region = region
	dat.Selector = selector
	dat.ClusterInst = *in
	return mcClient.ShowClusterMetrics(uri, token, dat)
}

func testPermShowAppInstMetrics(mcClient *mctestclient.Client, uri, token, region, org, selector string, data *edgeproto.AppInstKey) (*ormapi.AllMetrics, int, error) {
	in := &edgeproto.AppInstKey{}
	if data != nil {
		in = data
	} else {
		in.ClusterInstKey.ClusterKey.Name = "testcluster"
	}
	in.AppKey.Organization = org
	dat := &ormapi.RegionAppInstMetrics{}
	dat.Region = region
	dat.Selector = selector
	dat.AppInst = *in
	return mcClient.ShowAppMetrics(uri, token, dat)
}

func testPermShowCloudletMetrics(mcClient *mctestclient.Client, uri, token, region, org, selector string, data *edgeproto.CloudletKey) (*ormapi.AllMetrics, int, error) {
	in := &edgeproto.CloudletKey{}
	if data != nil {
		in = data
	} else {
		in.Name = "testcloudlet"
	}
	in.Organization = org
	dat := &ormapi.RegionCloudletMetrics{}
	dat.Region = region
	dat.Selector = selector
	dat.Cloudlet = *in
	return mcClient.ShowCloudletMetrics(uri, token, dat)
}

func testPermShowClientMetrics(mcClient *mctestclient.Client, uri, token, region, org, selector string, data *edgeproto.AppInstKey) (*ormapi.AllMetrics, int, error) {
	in := &edgeproto.AppInstKey{}
	if data != nil {
		in = data
	}
	in.AppKey.Organization = org
	in.ClusterInstKey.ClusterKey.Name = "testcluster"
	dat := &ormapi.RegionClientApiUsageMetrics{}
	dat.Region = region
	dat.Selector = selector
	dat.AppInst = *in
	return mcClient.ShowClientApiUsageMetrics(uri, token, dat)
}

func testPassCheckPermissionsAndGetCloudletList(t *testing.T, ctx context.Context, username, region string, devOrgs []string,
	resource string, cloudletKeys []edgeproto.CloudletKey, expectedCloudlets []string) {

	list, err := checkPermissionsAndGetCloudletList(ctx, username, region, devOrgs, resource, cloudletKeys)
	require.Nil(t, err)
	require.ElementsMatch(t, expectedCloudlets, list)
}

func testFailCheckPermissionsAndGetCloudletList(t *testing.T, ctx context.Context, username, region string, devOrgs []string,
	resource string, cloudletKeys []edgeproto.CloudletKey, errorContains string) {

	list, err := checkPermissionsAndGetCloudletList(ctx, username, region, devOrgs, resource, cloudletKeys)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), errorContains)
	require.Empty(t, list)
}

func badPermTestMetrics(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string) {
	// AppInst Metrics tests
	_, status, err := testPermShowAppInstMetrics(mcClient, uri, token, region, org, "cpu", nil)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	// ClusterInst Metrics tests
	_, status, err = testPermShowClusterMetrics(mcClient, uri, token, region, org, "cpu", nil)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	// Cloudlet Metrics tests
	_, status, err = testPermShowCloudletMetrics(mcClient, uri, token, region, org, "utilization", nil)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	// Client Metrics tests
	_, status, err = testPermShowClientMetrics(mcClient, uri, token, region, org, "api", nil)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermTestMetrics(t *testing.T, mcClient *mctestclient.Client, uri, devToken, operToken, region, devOrg, operOrg string) {
	// AppInst Metrics tests
	list, status, err := testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "cpu", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "mem", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "disk", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "network", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "connections", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	// multiple selector check
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "cpu,mem", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "*", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)

	// bad region check
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, "bad region", devOrg, "cpu", nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	// bad selector check
	list, status, err = testPermShowAppInstMetrics(mcClient, uri, devToken, region, devOrg, "bad selector", nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid appinst selector: bad selector")
	require.Equal(t, http.StatusBadRequest, status)

	// ClusterInst Metrics tests
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, region, devOrg, "cpu", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, region, devOrg, "mem", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, region, devOrg, "disk", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, region, devOrg, "network", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, region, devOrg, "tcp", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, region, devOrg, "udp", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)

	// bad region check
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, "bad region", devOrg, "cpu", nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	// bad selector check
	list, status, err = testPermShowClusterMetrics(mcClient, uri, devToken, region, devOrg, "bad selector", nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid cluster selector: bad selector")
	require.Equal(t, http.StatusBadRequest, status)

	// Cloudlet Metrics tests
	list, status, err = testPermShowCloudletMetrics(mcClient, uri, operToken, region, operOrg, "utilization", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	list, status, err = testPermShowCloudletMetrics(mcClient, uri, operToken, region, operOrg, "network", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)

	// bad region check
	list, status, err = testPermShowCloudletMetrics(mcClient, uri, operToken, "bad region", operOrg, "utilization", nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"bad region\" not found")
	require.Equal(t, http.StatusBadRequest, status)
	// bad selector check
	list, status, err = testPermShowCloudletMetrics(mcClient, uri, operToken, region, operOrg, "bad selector", nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid cloudlet selector: bad selector")
	require.Equal(t, http.StatusBadRequest, status)

	// Client Metrics test
	list, status, err = testPermShowClientMetrics(mcClient, uri, devToken, region, devOrg, "api", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)

	// bad selector check
	list, status, err = testPermShowClientMetrics(mcClient, uri, devToken, region, devOrg, "bad selector", nil)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid dme selector: bad selector")
	require.Equal(t, http.StatusBadRequest, status)

	// invalid input check
	appInst := edgeproto.AppInstKey{
		AppKey: edgeproto.AppKey{
			Name: "drop measurements \\",
		},
	}
	list, status, err = testPermShowClientMetrics(mcClient, uri, devToken, region, devOrg, "cpu", &appInst)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid app passed in")
	require.Equal(t, http.StatusBadRequest, status)
	cloudlet := edgeproto.CloudletKey{
		Name: "select * from api",
	}
	list, status, err = testPermShowCloudletMetrics(mcClient, uri, operToken, region, operOrg, "utilization", &cloudlet)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid cloudlet passed in")
	require.Equal(t, http.StatusBadRequest, status)
	cluster := edgeproto.ClusterInstKey{
		ClusterKey: edgeproto.ClusterKey{
			Name: "\\'\\;drop measurement \"cloudlet-ipusage\"",
		},
	}
	list, status, err = testPermShowClusterMetrics(mcClient, uri, operToken, region, operOrg, "utilization", &cluster)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid cluster passed in")
	require.Equal(t, http.StatusBadRequest, status)
}
