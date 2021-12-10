package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	"github.com/mobiledgex/edge-cloud-infra/promutils"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

func GetAppMetricsCustom(c echo.Context) error {
	var cmd string

	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctx := ormutil.GetContext(c)
	if !strings.HasSuffix(c.Path(), "metrics/app/custom") {
		return fmt.Errorf("Unsupported path")
	}
	in := ormapi.RegionCustomAppMetrics{}
	_, err = ReadConn(c, &in)
	if err != nil {
		return err
	}

	cloudletList, err := checkPermissionsAndGetCloudletList(ctx, claims.Username, in.Region, []string{in.AppInst.AppKey.Organization},
		ResourceAppAnalytics, []edgeproto.CloudletKey{in.AppInst.ClusterInstKey.CloudletKey})
	if err != nil {
		return err
	}

	// validate all the passed in arguments
	if err = util.ValidateNames(in.AppInst.GetTags()); err != nil {
		return err
	}

	if err = validateMetricsCommon(&in.MetricsCommon); err != nil {
		return err
	}
	cmd = getPromAppQuery(&in, cloudletList)
	resp, err := thanosProxy(ctx, in.Region, cmd)
	if err != nil {
		return err
	}
	return ormutil.SetReply(c, resp)
}

// TODO - how to deal with orgs, validation of input - if custom metric, need to make sure no `{`, or `}`
func getPromAppQuery(obj *ormapi.RegionCustomAppMetrics, cloudletList []string) string {
	var labelFilters []string
	if obj.AppInst.AppKey.Name != "" {
		labelFilters = append(labelFilters, `label_mexAppName="`+util.DNSSanitize(obj.AppInst.AppKey.Name)+`"`)
	}
	if obj.AppInst.AppKey.Version != "" {
		labelFilters = append(labelFilters, `label_mexAppVersion="`+util.DNSSanitize(obj.AppInst.AppKey.Version)+`"`)
	}
	labelFilter := "{" + strings.Join(labelFilters, ",") + "}"

	switch obj.Measurement {
	case "cpu":
		return url.QueryEscape(promutils.GetPromQueryWithK8sLabels(labelFilter, promutils.PromQCpuPod))
	case "mem":
		return url.QueryEscape(promutils.GetPromQueryWithK8sLabels(labelFilter, promutils.PromQMemPercentPod))
	case "disk":
		return url.QueryEscape(promutils.GetPromQueryWithK8sLabels(labelFilter, promutils.PromQDiskPercentPod))
	}
	return url.QueryEscape(promutils.GetPromQueryWithK8sLabels(labelFilter, obj.Measurement))
}

// We could use grpc interface instead of http one
func thanosProxy(ctx context.Context, region, query string) (*ormapi.PromResp, error) {
	log.SpanLog(ctx, log.DebugLevelApi, "start Thanos api", "region", region)
	defer log.SpanLog(ctx, log.DebugLevelApi, "finish Thanos api")

	addr, err := GetThanosUrl(ctx, region)
	if err != nil {
		return nil, err
	}

	thanosClient := http.Client{
		Timeout: time.Second * 2,
	}
	url := "http://" + addr + "/api/v1/query?query=" + query
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to build a thanos request - %s", err.Error())
	}

	res, getErr := thanosClient.Do(req)
	if getErr != nil {
		return nil, fmt.Errorf("Unable to run a thanos query - %s", err.Error())
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("Unable to decode result - %s", err.Error())
	}
	log.DebugLog(log.DebugLevelInfo, "XXX Thanos result.", "query", url, "bodyStr", string(body), "body", body)
	var promResponse ormapi.PromResp
	err = json.Unmarshal(body, &promResponse)
	if err != nil {
		log.DebugLog(log.DebugLevelInfo, "Unable to decode byte str", "bodyStr", string(body), "body", body, "err", err)
		return nil, fmt.Errorf("Unable to decode metrics data - %s", err.Error())
	} else {
		log.DebugLog(log.DebugLevelInfo, "PromData", "result", promResponse)
	}
	return &promResponse, nil
}

func GetThanosUrl(ctx context.Context, region string) (string, error) {
	return "127.0.0.1:29090", nil
}
