package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
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

var (
	reportTrigger chan bool

	ReportTimeout = 1 * time.Hour
)

func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func StripTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func getNextReportTime(now time.Time) time.Time {
	// report once a day at the start of the day 12am
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return nextDay
}

func updateScheduleDate(ctx context.Context, reportOrg string, newDate time.Time) error {
	db := loggedDB(ctx)
	lookup := ormapi.Reporter{
		Org: reportOrg,
	}
	updateReporter := ormapi.Reporter{}
	res := db.Where(&lookup).First(&updateReporter)
	if res.RecordNotFound() {
		// reporter got deleted in meantime, ignore
		return nil
	}
	if res.Error != nil {
		return dbErr(res.Error)
	}
	updateReporter.ScheduleDate = newDate
	err := db.Save(&updateReporter).Error
	if err != nil {
		return dbErr(res.Error)
	}
	return nil
}

func getAllRegions(ctx context.Context) ([]string, error) {
	controllers, err := ShowControllerObj(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to get regions: %v", err)
	}

	regions := []string{}
	for _, controller := range controllers {
		regions = append(regions, controller.Region)
	}
	return regions, nil
}

// Start report generation thread to run every day 12AM.
func GenerateReports() {
	reportTime := getNextReportTime(time.Now())
	for {
		select {
		case <-time.After(reportTime.Sub(time.Now())):
		case <-reportTrigger:
		}
		span := log.StartSpan(log.DebugLevelInfo, "Operator report generation thread")
		ctx := log.ContextWithSpan(context.Background(), span)
		// get list of all reporters
		db := loggedDB(ctx)
		reporters := []ormapi.Reporter{}
		err := db.Find(&reporters).Error
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get list of reporters", "err", err)
			// TODO: Generate alert?
			reportTime = getNextReportTime(reportTime)
			span.Finish()
			continue
		}
		regions, err := getAllRegions(ctx)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get regions", "err", err)
			continue
		}

		wgDone := make(chan bool)
		var wg sync.WaitGroup
		// For each reporter
		for _, reporter := range reporters {
			//   check if report time matches schedule date
			//      * get start & end date from schedule date & schedule interval
			//      * verify it reportTime
			if !DateEqual(reporter.ScheduleDate, time.Now()) {
				// not scheduled to generate report
				continue
			}
			dayUnit := 0
			switch reporter.Schedule {
			case edgeproto.ReportSchedule_EveryWeek:
				dayUnit = 7
			case edgeproto.ReportSchedule_Every15Days:
				dayUnit = 15
			case edgeproto.ReportSchedule_Every30Days:
				dayUnit = 30
			default:
				log.SpanLog(ctx, log.DebugLevelInfo, "Invalid reporter schedule", "schedule", reporter.Schedule)
				continue
			}
			startTime := StripTime(reporter.ScheduleDate.AddDate(0, 0, -dayUnit))
			endTime := StripTime(reporter.ScheduleDate)

			// generate report in a separate thread
			genReport := ormapi.GenerateReport{
				Org:       reporter.Org,
				StartTime: startTime,
				EndTime:   endTime,
			}
			wg.Add(1)
			go func(username string, genReport *ormapi.GenerateReport, wg *sync.WaitGroup) {
				log.SpanLog(ctx, log.DebugLevelInfo, "Generate operator report", "args", genReport)
				err = GenerateCloudletReport(ctx, username, regions, genReport)
				if err == nil {
					// Update next schedule date
					newDate := StripTime(genReport.EndTime.AddDate(0, 0, dayUnit))
					err = updateScheduleDate(ctx, genReport.Org, newDate)
					if err != nil {
						// TODO: Generate alert & retry
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to update schedule date for reporter", "org", genReport.Org, "err", err)
					}
				} else {
					// TODO: Generate alert & retry
					log.SpanLog(ctx, log.DebugLevelInfo, "failed to generate operator report", "org", genReport.Org, "err", err)
				}
				wg.Done()
			}(reporter.Username, &genReport, &wg)
		}
		go func() {
			wg.Wait()
			close(wgDone)
		}()
		// wait for all threads to exit
		select {
		case <-wgDone:
			log.SpanLog(ctx, log.DebugLevelInfo, "Done Generating operator reports")
		case <-time.After(ReportTimeout):
			// TODO: log & generate alert
		}
		reportTime = getNextReportTime(reportTime)
		span.Finish()
	}
}

func InitReporter() {
	reportTrigger = make(chan bool, 1)
}

func triggerReporter() {
	select {
	case reportTrigger <- true:
	default:
	}
}

// Create reporter to generate usage reports
func CreateReporter(c echo.Context) error {
	// TODO: add user validation
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reporter := ormapi.Reporter{}
	if err := c.Bind(&reporter); err != nil {
		return bindErr(c, err)
	}
	// sanity check
	if reporter.Org == "" {
		return setReply(c, fmt.Errorf("Org name has to be specified"), nil)
	}
	// get org details
	orgCheck, err := orgExists(ctx, reporter.Org)
	if err != nil {
		return err
	}
	if orgCheck.Type != OrgTypeOperator {
		return c.JSON(http.StatusBadRequest, Msg("Reporter can only be created for Operator org"))
	}
	// if an email is not specified send to an email on file
	if reporter.Email == "" {
		reporter.Email = claims.Email
	} else {
		// validate email
		if !util.ValidEmail(reporter.Email) {
			return setReply(c, fmt.Errorf("Reporter email is invalid"), nil)
		}
	}
	// validate report schedule
	if _, ok := edgeproto.ReportSchedule_name[int32(reporter.Schedule)]; !ok {
		return setReply(c, fmt.Errorf("invalid schedule"), nil)
	}
	// ScheduleDate defaults to now
	if reporter.ScheduleDate.IsZero() {
		reporter.ScheduleDate = time.Now()
	}

	// Schedule date should only be date with no time value
	reporter.ScheduleDate = StripTime(reporter.ScheduleDate)
	reporter.Username = claims.Username

	// store in db
	db := loggedDB(ctx)
	err = db.Create(&reporter).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"organizations_pkey") {
			return fmt.Errorf("Reporter is already created for Organization with name %s", reporter.Org)
		}
		return dbErr(err)
	}
	// trigger report generation if schedule date is today as it may have passed our internal report schedule
	if DateEqual(reporter.ScheduleDate, time.Now()) {
		triggerReporter()
	}
	return nil
}

func UpdateReporter(c echo.Context) error {
	// TODO: add code
	// TODO: add user validation
	return c.JSON(http.StatusOK, Msg("reporter updated"))
}

func DeleteReporter(c echo.Context) error {
	// TODO: add user validation
	ctx := GetContext(c)
	reporter := ormapi.Reporter{}
	if err := c.Bind(&reporter); err != nil {
		return bindErr(c, err)
	}
	db := loggedDB(ctx)
	err := db.Delete(&reporter).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	return c.JSON(http.StatusOK, Msg("reporter deleted"))
}

func ShowReporter(c echo.Context) error {
	// TODO: add validation + filter
	ctx := GetContext(c)
	db := loggedDB(ctx)
	lookup := ormapi.Reporter{}
	reporters := []ormapi.Reporter{}
	err := db.Where(&lookup).Find(&reporters).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	return c.JSON(http.StatusOK, reporters)
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
		return c.JSON(http.StatusBadRequest, Msg("org not specified"))
	}
	// TODO: Add auth for Operator access only
	// get org details
	orgCheck, err := orgExists(ctx, org)
	if err != nil {
		return err
	}
	if orgCheck.Type != OrgTypeOperator {
		return c.JSON(http.StatusBadRequest, Msg("report can only be generated for Operator org"))
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

	if report.EndTime.Sub(report.StartTime).Hours() > (31 * 24) {
		return c.JSON(http.StatusBadRequest, Msg("time range must not be more than 31 days"))
	}

	regions, err := getAllRegions(ctx)
	if err != nil {
		return setReply(c, err, nil)
	}

	err = GenerateCloudletReport(ctx, claims.Username, regions, &report)
	if err != nil {
		return setReply(c, err, nil)
	}
	return nil
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

	/*
		// TODO: Remove me
		// Add dummy data
		poolCloudlets["enterprise-pool"] = []string{"cloudlet1", "cloudlet2"}
		poolAcceptedDevelopers["enterprise-pool"] = []string{"user0org", "user1org", "user2org", "user3org"}
		poolPendingDevelopers["enterprise-pool"] = []string{"user4org", "user5org"}
		poolCloudlets["enterprise-pool1"] = []string{"cloudlet1", "cloudlet2", "cloudlet3"}
		poolAcceptedDevelopers["enterprise-pool1"] = []string{"user0org", "user1org", "user2org", "user3org"}
		poolPendingDevelopers["enterprise-pool1"] = []string{"user4org", "user5org"}
	*/

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

						if val[resIndex] == nil {
							log.SpanLog(ctx, log.DebugLevelInfo, "resource value is nil", "resname", resName, "value", val[resIndex])
							continue
						}
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
					if val[3] == nil {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch flavor value as it is nil", "cloudlet", val[1])
						continue
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

func getReportFileName(report *ormapi.GenerateReport) string {
	// File name should be of this format: "<orgname>_<startdate>_<enddate>_report.pdf"
	startDate := report.StartTime.Format("20060102") // YYYYMMDD
	endDate := report.EndTime.Format("20060102")
	return report.Org + "_" + startDate + "_" + endDate + "_report.pdf"
}

func GenerateCloudletReport(ctx context.Context, username string, regions []string, report *ormapi.GenerateReport) error {
	pdf := NewReport()
	for _, region := range regions {
		log.SpanLog(ctx, log.DebugLevelInfo, "Generate operator report for region", "region", region)
		// start new page for every region
		report.Region = region
		pdf.AddPage()

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
			return fmt.Errorf("failed to get cloudlet summary: %v", err)
		}
		AddTable(pdf, "Cloudlets", header, cloudlets, columnWidth)

		// Get list of cloudletpools
		header = []string{"Name", "Associated Cloudlets", "Accepted Developers", "Pending Developers"}
		columnWidth = float64(40)
		cloudletpools, err := GetCloudletPoolSummaryData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet pool summary: %v", err)
		}
		AddTable(pdf, "CloudletPools", header, cloudletpools, columnWidth)

		// Get top 5 flavors used per Cloudlet
		header = []string{"Cloudlet", "Top 5 Flavors Used"}
		columnWidth = float64(50)
		pdf.Ln(10)
		flavorData, err := GetCloudletFlavorUsageData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet flavor usage data: %v", err)
		}
		AddTable(pdf, "Flavor Usage", header, flavorData, columnWidth)

		// Start new page
		pdf.AddPage()

		// Get cloudlet resource usage metrics
		resourceUsageCharts, err := GetCloudletResourceUsageData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet resource usage data: %v", err)
		}
		err = AddTimeCharts(pdf, "Resource Usage", resourceUsageCharts)
		if err != nil {
			return err
		}

		// Start new page
		pdf.AddPage()

		// Get cloudlet events
		header = []string{"Timestamp", "Cloudlet", "Description"}
		columnWidth = float64(50)
		eventsData, err := GetCloudletEvents(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet events: %v", err)
		}
		AddTable(pdf, "Cloudlet Events", header, eventsData, columnWidth)

		// Get cloudlet alerts
		header = []string{"Timestamp", "Cloudlet", "Description", "State"}
		columnWidth = float64(40)
		alertsData, err := GetCloudletAlerts(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet alerts: %v", err)
		}
		AddTable(pdf, "Cloudlet Alerts", header, alertsData, columnWidth)

		if pdf.Err() {
			return fmt.Errorf("failed to create PDF report: %s\n", pdf.Error())
		}
		// TODO: Store in temporary location & then upload it to cloud
		filename := getReportFileName(report)
		err = pdf.OutputFileAndClose("/tmp/" + filename)
		if err != nil {
			return fmt.Errorf("cannot save PDF: %s", err)
		}
	}
	return nil
}

func ShowReport(c echo.Context) error {
	// TODO: get list of generated reports from cloudlet & show their downloadable links here
	return nil
}
