package orm

import (
	"context"
	fmt "fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mc/mcctl/mctestclient"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/promutils"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/require"
)

var (
	testSumByExpr     = "sum by(app,appver,apporg,cluster,clusterorg,cloudlet,cloudletorg)(envoy_cluster_upstream_cx_active)"
	testPromTimeRange = v1.Range{
		Start: (time.Date(2020, 1, 1, 1, 1, 0, 0, time.UTC)).Add(-1 * FallbackTimeRange),
		End:   time.Date(2020, 1, 1, 1, 1, 0, 0, time.UTC),
		Step:  DefaultAppInstTimeWindow,
	}
	testMetricRequest = ormapi.RegionCustomAppMetrics{}

	testPrometheusResponseVector = `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {
				"__name__": "envoy_cluster_upstream_cx_active",
				"app": "DevOrg SDK Demo",
				"apporg": "DevOrg",
				"appver": "1.0",
				"cloudlet": "localtest",
				"cloudletorg": "mexdev",
				"cluster": "AppCluster",
				"clusterorg": "DevOrg",
				"job": "MobiledgeX Monitoring",
				"port": "5656",
				"region": "local",
				"tenant_id": "default-tenant"
			  },
			  "value": [
				1640213591.808,
				"5"
			  ]
			},
			{
			  "metric": {
				"__name__": "envoy_cluster_upstream_cx_active",
				"app": "Face Detection Demo",
				"apporg": "DevOrg",
				"appver": "1.0",
				"cloudlet": "localtest",
				"cloudletorg": "mexdev",
				"cluster": "AppCluster",
				"clusterorg": "DevOrg",
				"job": "MobiledgeX Monitoring",
				"port": "8008",
				"region": "local",
				"tenant_id": "default-tenant"
			  },
			  "value": [
				1640213591.808,
				"6"
			  ]
			}
		  ]
		}
	  }`
	testAllDataMetricsFromVector = ormapi.AllMetrics{
		Data: []ormapi.MetricData{
			ormapi.MetricData{
				Series: []ormapi.MetricSeries{
					ormapi.MetricSeries{
						Name: "connections",
						Tags: map[string]string{
							"region":      "local",
							"app":         "DevOrg SDK Demo",
							"apporg":      "DevOrg",
							"appver":      "1.0",
							"cluster":     "AppCluster",
							"clusterorg":  "DevOrg",
							"cloudlet":    "localtest",
							"cloudletorg": "mexdev",
							"port":        "5656",
						},
						Values: [][]interface{}{
							[]interface{}{
								(float64)(5),
								(float64)(1640213591000),
							},
						},
					},
					ormapi.MetricSeries{
						Name: "connections",
						Tags: map[string]string{
							"region":      "local",
							"app":         "Face Detection Demo",
							"apporg":      "DevOrg",
							"appver":      "1.0",
							"cluster":     "AppCluster",
							"clusterorg":  "DevOrg",
							"cloudlet":    "localtest",
							"cloudletorg": "mexdev",
							"port":        "8008",
						},
						Values: [][]interface{}{
							[]interface{}{
								(float64)(6),
								(float64)(1640213591000),
							},
						},
					},
				},
			},
		},
	}
	testPrometheusResponseMatrix = `{
		"status": "success",
		"data": {
		  "resultType": "matrix",
		  "result": [
			{
			  "metric": {
				"__name__": "envoy_cluster_upstream_cx_active",
				"app": "DevOrg SDK Demo",
				"apporg": "DevOrg",
				"appver": "1.0",
				"cloudlet": "localtest",
				"cloudletorg": "mexdev",
				"cluster": "AppCluster",
				"clusterorg": "DevOrg",
				"job": "MobiledgeX Monitoring",
				"port": "5656",
				"region": "local",
				"tenant_id": "default-tenant"
			  },
			  "values": [
				[
				  1640209590.781,
				  "1"
				],
				[
				  1640213910.781,
				  "2"
				]
			  ]
			},
			{
			  "metric": {
				"__name__": "envoy_cluster_upstream_cx_active",
				"app": "Face Detection Demo",
				"apporg": "DevOrg",
				"appver": "1.0",
				"cloudlet": "localtest",
				"cloudletorg": "mexdev",
				"cluster": "AppCluster",
				"clusterorg": "DevOrg",
				"job": "MobiledgeX Monitoring",
				"port": "8008",
				"region": "local",
				"tenant_id": "default-tenant"
			  },
			  "values": [
				[
				  1640209590.781,
				  "3"
				],
				[
				  1640213910.781,
				  "4"
				]
			  ]
			}
		  ]
		}
	  }`
	testAllDataMetricsFromMatrix = ormapi.AllMetrics{
		Data: []ormapi.MetricData{
			ormapi.MetricData{
				Series: []ormapi.MetricSeries{
					ormapi.MetricSeries{
						Name: "connections",
						Tags: map[string]string{
							"region":      "local",
							"app":         "DevOrg SDK Demo",
							"apporg":      "DevOrg",
							"appver":      "1.0",
							"cluster":     "AppCluster",
							"clusterorg":  "DevOrg",
							"cloudlet":    "localtest",
							"cloudletorg": "mexdev",
							"port":        "5656",
						},
						Values: [][]interface{}{
							[]interface{}{
								(float64)(1),
								(float64)(1640209590000),
							},
							[]interface{}{
								(float64)(2),
								(float64)(1640213910000),
							},
						},
					},
					ormapi.MetricSeries{
						Name: "connections",
						Tags: map[string]string{
							"region":      "local",
							"app":         "Face Detection Demo",
							"apporg":      "DevOrg",
							"appver":      "1.0",
							"cluster":     "AppCluster",
							"clusterorg":  "DevOrg",
							"cloudlet":    "localtest",
							"cloudletorg": "mexdev",
							"port":        "8008",
						},
						Values: [][]interface{}{
							[]interface{}{
								(float64)(3),
								(float64)(1640209590000),
							},
							[]interface{}{
								(float64)(4),
								(float64)(1640213910000),
							},
						},
					},
				},
			},
		},
	}
)

// http server returning dummy prometheus results
func StartUnitTestThanosQueryResponder() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		emptyMsg := `{"status": "success","data": {"resultType": "matrix","result": []},"warnings": []}`
		if r.URL == nil {
			fmt.Fprintln(w, emptyMsg)
			return
		}
		// return dummy data, either vector, or matrix type
		if strings.Contains(r.URL.Path, "query_range") {
			fmt.Fprintf(w, testPrometheusResponseMatrix)
			return
		} else if strings.Contains(r.URL.Path, "query") {
			fmt.Fprintf(w, testPrometheusResponseVector)
			return
		}
		fmt.Fprintln(w, emptyMsg)
	}))
}
func TestWrapExpressionWithAggrFunc(t *testing.T) {
	// test sum aggregation
	testStr := wrapExpressionWithAggrFunc(promutils.PromQConnections, "sum")
	require.Equal(t, testSumByExpr, testStr)
}

func TestGetPromTimeRange(t *testing.T) {
	// empty metrics should return nil
	timeRange := getPromTimeRange(&testMetricRequest, nil)
	require.Nil(t, timeRange)

	// set end time - start time and step should be derived accordingly
	testMetricRequest.MetricsCommon.EndTime = time.Date(2020, 1, 1, 1, 1, 0, 0, time.UTC)
	timeRange = getPromTimeRange(&testMetricRequest, nil)
	require.Equal(t, testPromTimeRange, *timeRange)
}

func TestValidateAppMetricArgs(t *testing.T) {
	// test setup
	log.SetDebugLevel(log.DebugLevelApi | log.DebugLevelNotify | log.DebugLevelInfra)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	// nill check
	err := validateAppMetricArgs(ctx, "testuser", nil)
	require.NotNil(t, err, "nil args should trigger an error")
	require.Contains(t, err.Error(), "Invalid Region App metrics object")
	// Note - other validations are done as part of goodPermTestCustomMetrics
}

func testPermShowAppCustomMetrics(mcClient *mctestclient.Client, uri, token, region, org, measurement string, data *ormapi.RegionCustomAppMetrics) (*ormapi.AllMetrics, int, error) {
	dat := &ormapi.RegionCustomAppMetrics{}
	if data != nil {
		dat = data
	} else {
		in := edgeproto.AppInstKey{}
		in.ClusterInstKey.ClusterKey.Name = "testcluster"
		in.AppKey.Organization = org
		dat.Region = region
		dat.Measurement = measurement
		dat.AppInst = in
	}
	return mcClient.ShowAppV2Metrics(uri, token, dat)
}

func badPermTestCustomMetrics(t *testing.T, mcClient *mctestclient.Client, uri, token, region, org string) {
	// bad perm test
	_, status, err := testPermShowAppCustomMetrics(mcClient, uri, token, region, org, "connections", nil)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)
}

func goodPermTestCustomMetrics(t *testing.T, mcClient *mctestclient.Client, uri, devToken, operToken, region, devOrg, operOrg string) {
	// basic good perm test
	list, status, err := testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, "connections", nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	// validate dummy vector data we got back - testing parsePrometheusVector
	require.Equal(t, testAllDataMetricsFromVector, *list, "single data point metrics parsing")

	// if running a free-form metric, require admin permission
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, promutils.PromQConnections, nil)
	require.NotNil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// check validations
	// no measurement specified
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, "", nil)
	require.NotNil(t, err)
	require.Equal(t, "Measurement is required", err.Error())

	// no region
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, "", devOrg, "connections", nil)
	require.NotNil(t, err)
	require.Equal(t, "Region is required", err.Error())

	// check valid timestamps
	arg := ormapi.RegionCustomAppMetrics{}
	arg.MetricsCommon.EndTime = time.Date(2020, 1, 1, 1, 1, 0, 0, time.UTC)
	arg.AppInst.AppKey.Organization = devOrg
	arg.Region = region
	arg.Measurement = "connections"
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, "connections", &arg)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
	// validate dummy matrix data we got back - testing parsePrometheusMatrix
	require.Equal(t, testAllDataMetricsFromMatrix, *list, "single data point metrics parsing")

	// validate aggr function
	arg.Measurement = "connections"
	arg.AggrFunction = "mean"
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, "connections", &arg)
	require.NotNil(t, err)
	require.Equal(t, "Only \"sum\" aggregation function is supported", err.Error())
	arg.AggrFunction = "sum"
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, "connections", &arg)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)

	// invalid port #
	arg.Port = "abc"
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, "connections", &arg)
	require.NotNil(t, err)
	require.Equal(t, "Port number must be an interger", err.Error())
	arg.Port = "234"
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, devToken, region, devOrg, "connections", &arg)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)
}

func adminPermTestCustomMetrics(t *testing.T, mcClient *mctestclient.Client, uri, adminToken, region, adminOrg string) {
	// if running a free-form metric, require admin permission
	list, status, err := testPermShowAppCustomMetrics(mcClient, uri, adminToken, region, adminOrg, promutils.PromQConnections, nil)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, list)

	// No port is allowed to be specified
	arg := ormapi.RegionCustomAppMetrics{}
	arg.AppInst.AppKey.Organization = adminOrg
	arg.Region = region
	arg.Measurement = promutils.PromQConnections
	arg.Port = "123"
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, adminToken, region, adminOrg, "", &arg)
	require.NotNil(t, err)
	require.Equal(t, "Only \"connections\" measurement supports specifying port", err.Error())

	// No aggr function is supported for free-form requests
	arg.Port = ""
	arg.AggrFunction = "sum"
	list, status, err = testPermShowAppCustomMetrics(mcClient, uri, adminToken, region, adminOrg, "", &arg)
	require.NotNil(t, err)
	require.Equal(t, "Only \"connections\" measurement supports aggregate function", err.Error())
}
