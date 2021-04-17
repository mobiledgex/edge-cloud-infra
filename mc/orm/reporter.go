package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

// Create reporter to generate usage reports
func CreateReporter(c echo.Context) error {
	return nil
}

func UpdateReporter(c echo.Context) error {
	return nil
}

func DeleteReporter(c echo.Context) error {
	return nil
}

func ShowReporter(c echo.Context) error {
	return nil
}

func GenerateReport(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	report := ormapi.GenerateReport{}
	if err := c.Bind(&report); err != nil {
		return bindErr(c, err)
	}
	org := report.Org
	if org == "" {
		return c.JSON(http.StatusBadRequest, Msg("Org not specified"))
	}
	// TODO: Add auth for Operator access only
	// TODO: Validate region
	// get org details
	orgCheck, err := orgExists(ctx, org)
	if err != nil {
		return err
	}
	if orgCheck.Type != OrgTypeOperator {
		return c.JSON(http.StatusBadRequest, Msg("Reporter can only be created for Operator org"))
	}

	if !report.StartTime.Before(report.EndTime) {
		return c.JSON(http.StatusBadRequest, Msg("start time must be before end time"))
	}

	if !report.EndTime.Before(time.Now()) {
		return c.JSON(http.StatusBadRequest, Msg("end time must be historical time"))
	}

	/*
		if report.EndTime.Sub(report.StartTime).Hours() < (7 * 24) {
			return c.JSON(http.StatusBadRequest, Msg("time range must be atleast 7 days"))
		}
	*/

	if report.EndTime.Sub(report.StartTime).Hours() > (90 * 24) {
		return c.JSON(http.StatusBadRequest, Msg("time range must not be more than 90 days"))
	}

	return GenerateCloudletReport(c, claims.Username, &report)
}

func GetCloudletSummaryData(ctx context.Context, username string, report *ormapi.GenerateReport) ([][]string, error) {
	rc := &RegionContext{}
	rc.username = username
	rc.region = report.Region
	obj := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: report.Org,
		},
	}
	cloudlets := [][]string{}
	err := ShowCloudletStream(ctx, rc, &obj, func(res *edgeproto.Cloudlet) {
		platformTypeStr := edgeproto.PlatformType_CamelName[int32(res.PlatformType)]
		platformTypeStr = strings.TrimPrefix(platformTypeStr, "PlatformType")
		stateStr := edgeproto.TrackedState_CamelName[int32(res.State)]
		cloudletData := []string{res.Key.Name, platformTypeStr, stateStr, res.ContainerVersion}

		cloudlets = append(cloudlets, cloudletData)
	})
	if err != nil {
		return nil, err
	}
	return cloudlets, nil
}

func GetCloudletPoolSummaryData(ctx context.Context, username string, report *ormapi.GenerateReport) ([][]string, error) {
	// get pools associated with orgs
	db := loggedDB(ctx)
	op := ormapi.OrgCloudletPool{
		Region:          report.Region,
		CloudletPoolOrg: report.Org,
	}
	ops := []ormapi.OrgCloudletPool{}
	err := db.Where(&op).Find(&ops).Error
	if err != nil {
		return nil, err
	}
	poolAcceptedDevelopers := make(map[string][]string)
	poolPendingDevelopers := make(map[string][]string)
	acceptedOps := getAccessGranted(ops)
	pendingOps := getAccessPending(ops)
	for _, op := range acceptedOps {
		if poolDevs, ok := poolAcceptedDevelopers[op.CloudletPool]; ok {
			poolDevs = append(poolDevs, op.Org)
			poolAcceptedDevelopers[op.CloudletPool] = poolDevs
		} else {
			poolAcceptedDevelopers[op.CloudletPool] = []string{op.Org}
		}
	}
	for _, op := range pendingOps {
		if poolDevs, ok := poolPendingDevelopers[op.CloudletPool]; ok {
			poolDevs = append(poolDevs, op.Org)
			poolPendingDevelopers[op.CloudletPool] = poolDevs
		} else {
			poolPendingDevelopers[op.CloudletPool] = []string{op.Org}
		}
	}
	rc := RegionContext{
		region:    report.Region,
		username:  username,
		skipAuthz: true,
	}
	poolKey := edgeproto.CloudletPoolKey{Organization: report.Org}
	poolCloudlets := make(map[string][]string)
	err = ShowCloudletPoolStream(ctx, &rc, &edgeproto.CloudletPool{Key: poolKey}, func(pool *edgeproto.CloudletPool) {
		for _, name := range pool.Cloudlets {
			if cloudlets, ok := poolCloudlets[op.CloudletPool]; ok {
				cloudlets = append(cloudlets, name)
				poolCloudlets[pool.Key.Name] = cloudlets
			} else {
				poolCloudlets[pool.Key.Name] = []string{name}
			}
		}
	})

	// TODO: Remove me
	// Add dummy data
	poolCloudlets["enterprise-pool"] = []string{"cloudlet1", "cloudlet2"}
	poolAcceptedDevelopers["enterprise-pool"] = []string{"user0org", "user1org", "user2org", "user3org"}
	poolPendingDevelopers["enterprise-pool"] = []string{"user4org", "user5org"}
	poolCloudlets["enterprise-pool1"] = []string{"cloudlet1", "cloudlet2", "cloudlet3"}
	poolAcceptedDevelopers["enterprise-pool1"] = []string{"user0org", "user1org", "user2org", "user3org"}
	poolPendingDevelopers["enterprise-pool1"] = []string{"user4org", "user5org"}

	cloudletpools := [][]string{}
	for poolName, poolCloudlets := range poolCloudlets {
		// get accepted developers
		poolAcceptedDevs, _ := poolAcceptedDevelopers[poolName]
		poolPendingDevs, _ := poolPendingDevelopers[poolName]
		entry := []string{
			poolName,
			strings.Join(poolCloudlets, "\n"),
			strings.Join(poolAcceptedDevs, "\n"),
			strings.Join(poolPendingDevs, "\n"),
		}
		cloudletpools = append(cloudletpools, entry)
	}
	return cloudletpools, nil
}

func GetCloudletResourceUsageData(ctx context.Context, username string, report *ormapi.GenerateReport) (map[string][]TimeChartData, error) {
	rc := &InfluxDBContext{}
	dbNames := []string{cloudcommon.CloudletResourceUsageDbName}
	in := ormapi.RegionCloudletMetrics{
		Region: report.Region,
		Cloudlet: edgeproto.CloudletKey{
			Organization: report.Org,
		},
		Selector:  "resourceusage",
		StartTime: report.StartTime,
		EndTime:   report.EndTime,
	}
	rc.region = in.Region
	cmd := CloudletUsageMetricsQuery(&in)

	chartMap := make(map[string]map[string]TimeChartData)
	err := influxStream(ctx, rc, dbNames, cmd, func(res interface{}) {
		results, ok := res.([]influxdb.Result)
		if !ok {
			return
		}
		for _, result := range results {
			for _, row := range result.Series {
				/*
					columns[0]  -> time
					columns[1]  -> cloudlet
					columns[2]  -> cloudletorg
					columns[3:] -> resources
				*/
				for _, val := range row.Values {
					timeStr, ok := val[0].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch resource time", "time", val[0])
					}
					time, err := time.Parse("2006-01-02T15:04:05Z", timeStr)
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to parse resource time", "time", timeStr)
						continue
					}
					cloudlet, ok := val[1].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch cloudlet name", "cloudlet", val[1])
					}
					for resIndex := 3; resIndex < len(val); resIndex++ {
						resName := row.Columns[resIndex]
						if _, ok := chartMap[resName]; !ok {
							chartMap[resName] = make(map[string]TimeChartData)
						}
						if _, ok := chartMap[resName][cloudlet]; !ok {
							chartMap[resName][cloudlet] = TimeChartData{Name: cloudlet}
						}
						clData, _ := chartMap[resName][cloudlet]

						resVal, err := val[resIndex].(json.Number).Float64()
						if err != nil {
							log.SpanLog(ctx, log.DebugLevelInfo, "failed to parse resource value", "value", val[resIndex])
							continue
						}

						clData.XValues = append(clData.XValues, time)
						clData.YValues = append(clData.YValues, float64(resVal))
						chartMap[resName][cloudlet] = clData
					}
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	chartDataOut := make(map[string][]TimeChartData)
	for resName, resData := range chartMap {
		for _, cData := range resData {
			chartDataOut[resName] = append(chartDataOut[resName], cData)
		}
	}
	return chartDataOut, nil
}

func GetCloudletFlavorUsageData(ctx context.Context, username string, report *ormapi.GenerateReport) ([][]string, error) {
	rc := &InfluxDBContext{}
	dbNames := []string{cloudcommon.CloudletResourceUsageDbName}
	in := ormapi.RegionCloudletMetrics{
		Region: report.Region,
		Cloudlet: edgeproto.CloudletKey{
			Organization: report.Org,
		},
		Selector:  "flavorusage",
		StartTime: report.StartTime,
		EndTime:   report.EndTime,
	}
	rc.region = in.Region
	cmd := CloudletUsageMetricsQuery(&in)

	flavorMap := make(map[string]map[string]float64)
	err := influxStream(ctx, rc, dbNames, cmd, func(res interface{}) {
		results, ok := res.([]influxdb.Result)
		if !ok {
			return
		}
		for _, result := range results {
			for _, row := range result.Series {
				/*
					columns[0] -> time
					columns[1] -> cloudlet
					columns[2] -> cloudletorg
					columns[3] -> count
					columns[4] -> flavor
				*/
				for _, val := range row.Values {
					cloudlet, ok := val[1].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch cloudlet name", "cloudlet", val[1])
					}
					countVal, err := val[3].(json.Number).Float64()
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to parse flavor count", "value", val[3])
						continue
					}
					flavor, ok := val[4].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch flavor name", "flavor", val[4])
					}
					if _, ok := flavorMap[cloudlet]; !ok {
						flavorMap[cloudlet] = make(map[string]float64)
					}
					flavorMap[cloudlet][flavor] = countVal
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	flavorOut := [][]string{}
	for cloudlet, flavorCount := range flavorMap {
		flavors := []string{}
		for flavor, _ := range flavorCount {
			flavors = append(flavors, flavor)
		}
		sort.Slice(flavors, func(i, j int) bool {
			return flavorCount[flavors[i]] > flavorCount[flavors[j]]
		})
		count := 0
		topFlavors := []string{}
		for _, flavor := range flavors {
			if count >= 5 {
				break
			}
			topFlavors = append(topFlavors, fmt.Sprintf("%s (%0.0f)", flavor, flavorCount[flavor]))
			count++
		}
		flavorOut = append(flavorOut, []string{cloudlet, strings.Join(topFlavors, "\n")})
	}
	return flavorOut, nil
}

func GetCloudletEvents(ctx context.Context, username string, report *ormapi.GenerateReport) ([][]string, error) {
	search := node.EventSearch{
		Match: node.EventMatch{
			Orgs:    []string{report.Org},
			Types:   []string{node.EventType},
			Regions: []string{report.Region},
		},
		TimeRange: util.TimeRange{
			StartTime: report.StartTime,
			EndTime:   report.EndTime,
		},
	}
	if err := search.TimeRange.Resolve(48 * time.Hour); err != nil {
		return nil, err
	}

	events, err := nodeMgr.ShowEvents(ctx, &search)
	if err != nil {
		return nil, err
	}
	eventsData := [][]string{}
	for _, event := range events {
		cloudlet, ok := event.Mtags["cloudlet"]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfo, "missing cloudlet name in event, skipping", "event", event)
		}
		timestamp := event.Timestamp.Format("Mon Jan 2 15:04:05")
		entry := []string{timestamp, cloudlet, event.Name}
		eventsData = append(eventsData, entry)
	}
	return eventsData, nil
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func GetCloudletAlerts(ctx context.Context, username string, report *ormapi.GenerateReport) ([][]string, error) {
	alertsData := [][]string{}
	rc := &RegionContext{
		region:    report.Region,
		username:  username,
		skipAuthz: true,
	}
	obj := &edgeproto.Alert{
		Labels: map[string]string{
			edgeproto.CloudletKeyTagOrganization: report.Org,
			cloudcommon.AlertScopeTypeTag:        cloudcommon.AlertScopeCloudlet,
		},
	}
	alerts, err := ShowAlertObj(ctx, rc, obj)
	if err != nil {
		return nil, err
	}
	for _, alert := range alerts {
		alertTime := cloudcommon.TimestampToTime(alert.ActiveAt)
		if !inTimeSpan(report.StartTime, report.EndTime, alertTime) {
			continue
		}
		cloudlet, ok := alert.Labels[edgeproto.CloudletKeyTagName]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfo, "missing cloudlet name in alert labels, skipping", "labels", alert.Labels)
		}
		desc, ok := alert.Annotations[cloudcommon.AlertAnnotationDescription]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfo, "missing description in alert annotations, skipping", "annotations", alert.Annotations)
		}
		alertTimeStr := alertTime.Format("Mon Jan 2 15:04:05")
		entry := []string{alertTimeStr, cloudlet, desc, alert.State}
		alertsData = append(alertsData, entry)
	}
	return alertsData, nil
}

func GenerateCloudletReport(c echo.Context, username string, report *ormapi.GenerateReport) error {
	ctx := GetContext(c)

	pdf := NewReport()
	AddPageTitle(pdf)
	AddHeader(pdf, report)
	AddFooter(pdf)
	AddOperatorInfo(pdf, report)
	AddHorizontalLine(pdf)

	// Get list of cloudlets
	header := []string{"Name", "Platform Type", "Last Known State", "Version"}
	columnWidth := float64(40)
	cloudlets, err := GetCloudletSummaryData(ctx, username, report)
	if err != nil {
		return setReply(c, err, nil)
	}
	AddTable(pdf, "Cloudlets", header, cloudlets, columnWidth)

	// Get list of cloudletpools
	header = []string{"Name", "Associated Cloudlets", "Accepted Developers", "Pending Developers"}
	columnWidth = float64(40)
	cloudletpools, err := GetCloudletPoolSummaryData(ctx, username, report)
	if err != nil {
		return setReply(c, err, nil)
	}
	AddTable(pdf, "CloudletPools", header, cloudletpools, columnWidth)

	// Get top 5 flavors used per Cloudlet
	header = []string{"Cloudlet", "Top 5 Flavors Used"}
	columnWidth = float64(50)
	pdf.Ln(10)
	flavorData, err := GetCloudletFlavorUsageData(ctx, username, report)
	if err != nil {
		return setReply(c, err, nil)
	}
	AddTable(pdf, "Flavor Usage", header, flavorData, columnWidth)

	// Start new page
	pdf.AddPage()

	// Get cloudlet resource usage metrics
	resourceUsageCharts, err := GetCloudletResourceUsageData(ctx, username, report)
	if err != nil {
		return setReply(c, err, nil)
	}
	err = AddTimeCharts(pdf, "Resource Usage", resourceUsageCharts)
	if err != nil {
		return setReply(c, err, nil)
	}

	// Start new page
	pdf.AddPage()

	// Get cloudlet events
	header = []string{"Timestamp", "Cloudlet", "Description"}
	columnWidth = float64(50)
	eventsData, err := GetCloudletEvents(ctx, username, report)
	if err != nil {
		return setReply(c, err, nil)
	}
	AddTable(pdf, "Cloudlet Events", header, eventsData, columnWidth)

	// Get cloudlet alerts
	header = []string{"Timestamp", "Cloudlet", "Description", "State"}
	columnWidth = float64(40)
	alertsData, err := GetCloudletAlerts(ctx, username, report)
	if err != nil {
		return setReply(c, err, nil)
	}
	AddTable(pdf, "Cloudlet Alerts", header, alertsData, columnWidth)

	if pdf.Err() {
		return setReply(c, fmt.Errorf("failed to create PDF report: %s\n", pdf.Error()), nil)
	}
	err = pdf.OutputFileAndClose("/tmp/hello.pdf")
	if err != nil {
		return setReply(c, fmt.Errorf("cannot save PDF: %s", err), nil)
	}
	return nil
}
