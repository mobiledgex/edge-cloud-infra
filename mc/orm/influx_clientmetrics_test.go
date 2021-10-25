package orm

import (
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/require"
)

var (
	testApiUsageQuerySampled = `SELECT last("reqs") AS "reqs",last("errs") AS "errs",last("cellID") AS "cellID",last("foundCloudlet") AS "foundCloudlet",last("foundOperator") AS "foundOperator" from "dme-api" WHERE ` +
		`"apporg"='testOrg1' AND "app"='testApp1' AND "ver"='1.0' AND "method"='RegisterClient' AND "foundCloudlet"='testCloudlet1' ` +
		`AND time >= '2019-12-31T13:01:00Z' AND time <= '2020-01-01T01:01:00Z' ` +
		`group by time(12h0m0s),"apporg","app","ver","cloudletorg","cloudlet","dmeId","method" order by time desc ` +
		`limit 1`
	testApiUsageQuerNonSampledLast = `SELECT "reqs","errs","cellID","foundCloudlet","foundOperator" from "dme-api" ` +
		`WHERE "apporg"='testOrg1' AND "app"='testApp1' AND "ver"='1.0' AND "method"='RegisterClient' AND "foundCloudlet"='testCloudlet1' ` +
		`group by "apporg","app","ver","cloudletorg","cloudlet","dmeId","method" order by time desc ` +
		`limit 1`
	testRegionClientApiUsage = ormapi.RegionClientApiUsageMetrics{
		Region: "test",
		AppInst: edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Organization: "testOrg1",
				Name:         "testApp1",
				Version:      "1.0",
			},
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				CloudletKey: edgeproto.CloudletKey{
					Name: "testCloudlet1",
				},
			},
		},
		Method:   "RegisterClient",
		Selector: "api",
	}
)

func TestClientApiUsageMetricsQuery(t *testing.T) {
	maxEntriesFromInfluxDb = 100
	// sampled query
	testRegionClientApiUsage.EndTime = time.Date(2020, 1, 1, 1, 1, 0, 0, time.UTC)
	testRegionClientApiUsage.NumSamples = 1
	err := validateMetricsCommon(&testRegionClientApiUsage.MetricsCommon)
	require.Nil(t, err)
	require.Equal(t, testApiUsageQuerySampled, ClientApiUsageMetricsQuery(&testRegionClientApiUsage, []string{}, nil))

	// non-sampled query
	testRegionClientApiUsage.EndTime = time.Time{}
	testRegionClientApiUsage.StartTime = time.Time{}
	testRegionClientApiUsage.NumSamples = 0
	testRegionClientApiUsage.Limit = 1
	err = validateMetricsCommon(&testRegionClientApiUsage.MetricsCommon)
	require.Nil(t, err)
	require.Equal(t, testApiUsageQuerNonSampledLast, ClientApiUsageMetricsQuery(&testRegionClientApiUsage, []string{}, nil))

}
