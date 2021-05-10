package orm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/gcs"
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
	NoCloudlet    = ""
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
			reportTime = getNextReportTime(reportTime)
			span.Finish()
			continue
		}
		regions, err := getAllRegions(ctx)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get regions", "err", err)
			continue
		}

		storageClient, err := gcs.NewClient(ctx, serverConfig.vaultConfig, serverConfig.DeploymentTag)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to setup GCS storage client", "err", err)
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
				Timezone:  reporter.Timezone,
			}
			wg.Add(1)
			go func(username string, genReport *ormapi.GenerateReport, wg *sync.WaitGroup) {
				log.SpanLog(ctx, log.DebugLevelInfo, "Generate operator report", "args", genReport)
				tags := map[string]string{"cloudletorg": genReport.Org}
				var output bytes.Buffer
				err = GenerateCloudletReport(ctx, username, regions, genReport, &output)
				if err == nil {
					// Upload PDF report to cloudlet
					filename := ormapi.GetReportFileName(genReport)
					err = storageClient.UploadObject(ctx, filename, &output)
					if err != nil {
						nodeMgr.Event(ctx, "Cloudlet report upload failure", genReport.Org, tags, err)
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to upload cloudlet report to cloudlet", "org", genReport.Org, "err", err)
					}
					// Update next schedule date
					newDate := StripTime(genReport.EndTime.AddDate(0, 0, dayUnit))
					err = updateScheduleDate(ctx, genReport.Org, newDate)
					if err != nil {
						nodeMgr.Event(ctx, "Cloudlet report schedule update failure", genReport.Org, tags, err)
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to update schedule date for reporter", "org", genReport.Org, "err", err)
					}
				} else {
					log.SpanLog(ctx, log.DebugLevelInfo, "failed to generate cloudlet report", "org", genReport.Org, "err", err)
					nodeMgr.Event(ctx, "Cloudlet report generation failure", genReport.Org, tags, err)
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
			log.SpanLog(ctx, log.DebugLevelInfo, "Timedout generating operator reports")
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

	// check if user is authorized to create reporter
	if err := authorized(ctx, claims.Username, reporter.Org, ResourceCloudlets, ActionManage); err != nil {
		return setReply(c, err, nil)
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

	if reporter.Timezone == "" {
		// check if timezone is present as part of user's setting
		// this is set from console UI
		reporter.Timezone, _ = GetUserTimezone(ctx, claims.Username)
	}
	if reporter.Timezone == "" {
		// defaults to UTC
		reporter.Timezone = "UTC"
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
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	// Pull json directly so we can unmarshal twice.
	// First time is to do lookup, second time is to apply
	// modified fields.
	body, err := ioutil.ReadAll(c.Request().Body)
	in := ormapi.Reporter{}
	err = json.Unmarshal(body, &in)
	if err != nil {
		return bindErr(c, err)
	}
	if in.Org == "" {
		return c.JSON(http.StatusBadRequest, Msg("Organization name not specified"))
	}
	lookup := ormapi.Reporter{
		Org: in.Org,
	}
	reporter := ormapi.Reporter{}
	db := loggedDB(ctx)
	res := db.Where(&lookup).First(&reporter)
	if res.RecordNotFound() {
		return c.JSON(http.StatusBadRequest, Msg("Reporter not found"))
	}
	if res.Error != nil {
		return c.JSON(http.StatusInternalServerError, MsgErr(dbErr(res.Error)))
	}

	// check if user is authorized to update reporter
	if err := authorized(ctx, claims.Username, reporter.Org, ResourceCloudlets, ActionManage); err != nil {
		return setReply(c, err, nil)
	}

	oldEmail := reporter.Email
	oldUsername := reporter.Username
	oldSchedule := reporter.Schedule
	oldScheduleDate := reporter.ScheduleDate
	// apply specified fields
	err = json.Unmarshal(body, &reporter)
	if err != nil {
		return bindErr(c, err)
	}
	if reporter.Email != oldEmail {
		// validate email
		if !util.ValidEmail(reporter.Email) {
			return setReply(c, fmt.Errorf("Reporter email is invalid"), nil)
		}
	}

	if reporter.Username != oldUsername {
		return c.JSON(http.StatusBadRequest, Msg("Cannot change username"))
	}

	if reporter.Schedule != oldSchedule {
		// validate report schedule
		if _, ok := edgeproto.ReportSchedule_name[int32(reporter.Schedule)]; !ok {
			return setReply(c, fmt.Errorf("invalid schedule"), nil)
		}
	}

	if reporter.ScheduleDate != oldScheduleDate {
		// Schedule date should only be date with no time value
		reporter.ScheduleDate = StripTime(reporter.ScheduleDate)
	}
	err = db.Save(&reporter).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, MsgErr(dbErr(err)))
	}
	return nil
}

func DeleteReporter(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reporter := ormapi.Reporter{}
	if err := c.Bind(&reporter); err != nil {
		return bindErr(c, err)
	}
	// check if user is authorized to delete reporter
	if err := authorized(ctx, claims.Username, reporter.Org, ResourceCloudlets, ActionManage); err != nil {
		return setReply(c, err, nil)
	}
	db := loggedDB(ctx)
	err = db.Delete(&reporter).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	return c.JSON(http.StatusOK, Msg("reporter deleted"))
}

func ShowReporter(c echo.Context) error {
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	db := loggedDB(ctx)

	filter := ormapi.Reporter{}
	if c.Request().ContentLength > 0 {
		if err := c.Bind(&filter); err != nil {
			return bindErr(c, err)
		}
	}
	authOrgs, err := enforcer.GetAuthorizedOrgs(ctx, claims.Username, ResourceCloudlets, ActionView)
	if err != nil {
		return err
	}
	_, admin := authOrgs[""]
	_, orgFound := authOrgs[filter.Org]
	if filter.Org != "" && !admin && !orgFound {
		// no perms for specified org
		return echo.ErrForbidden
	}

	reporters := []ormapi.Reporter{}
	err = db.Where(&filter.Org).Find(&reporters).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	return c.JSON(http.StatusOK, reporters)
}

func GetUserTimezone(ctx context.Context, username string) (string, error) {
	// check if timezone is present as part of user's setting
	// this is set from console UI
	user := ormapi.User{Name: username}
	db := loggedDB(ctx)
	err := db.Where(&user).First(&user).Error
	if err != nil {
		return "", dbErr(err)
	}
	if user.Metadata != "" {
		metadata := make(map[string]string)
		err = json.Unmarshal([]byte(user.Metadata), &metadata)
		if err != nil {
			return "", fmt.Errorf("Invalid user metadata: %v, %v", user.Metadata, err)
		}
		if timezone, ok := metadata["Timezone"]; ok {
			return timezone, nil
		}
	}
	return "", nil
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
	// auth for Operator access only
	// get org details
	orgCheck, err := orgExists(ctx, org)
	if err != nil {
		return setReply(c, err, nil)
	}
	if orgCheck.Type != OrgTypeOperator {
		return c.JSON(http.StatusBadRequest, Msg("report can only be generated for Operator org"))
	}

	// check if user is authorized to generate cloudlet reports
	if err := authorized(ctx, claims.Username, report.Org, ResourceCloudlets, ActionManage); err != nil {
		return setReply(c, err, nil)
	}

	if report.Timezone == "" {
		// check if timezone is present as part of user's setting
		// this is set from console UI
		report.Timezone, _ = GetUserTimezone(ctx, claims.Username)
	}
	if report.Timezone == "" {
		// defaults to UTC
		report.Timezone = "UTC"
	}

	if !report.StartTime.Before(report.EndTime) {
		return c.JSON(http.StatusBadRequest, Msg("start time must be before end time"))
	}

	if !report.EndTime.Before(time.Now()) {
		return c.JSON(http.StatusBadRequest, Msg("end time must be historical time"))
	}

	if report.EndTime.Sub(report.StartTime).Hours() < (7 * 24) {
		return c.JSON(http.StatusBadRequest, Msg("time range must be atleast 7 days"))
	}

	if report.EndTime.Sub(report.StartTime).Hours() > (31 * 24) {
		return c.JSON(http.StatusBadRequest, Msg("time range must not be more than 31 days"))
	}

	regions, err := getAllRegions(ctx)
	if err != nil {
		return setReply(c, err, nil)
	}

	var output bytes.Buffer
	err = GenerateCloudletReport(ctx, claims.Username, regions, &report, &output)
	if err != nil {
		return setReply(c, err, nil)
	}

	return c.HTMLBlob(http.StatusOK, output.Bytes())
}

func GetCloudletSummaryData(ctx context.Context, username string, report *ormapi.GenerateReport) ([][]string, error) {
	rc := &RegionContext{
		region:   report.Region,
		username: username,
	}
	obj := edgeproto.Cloudlet{
		Key: edgeproto.CloudletKey{
			Organization: report.Org,
		},
	}
	cloudlets := [][]string{}
	cloudletsPresent := make(map[string]struct{})
	err := ShowCloudletStream(ctx, rc, &obj, func(res *edgeproto.Cloudlet) {
		platformTypeStr := edgeproto.PlatformType_CamelName[int32(res.PlatformType)]
		platformTypeStr = strings.TrimPrefix(platformTypeStr, "PlatformType")
		stateStr := edgeproto.TrackedState_CamelName[int32(res.State)]
		cloudletData := []string{res.Key.Name, platformTypeStr, stateStr}

		cloudlets = append(cloudlets, cloudletData)
		cloudletsPresent[res.Key.Name] = struct{}{}
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(cloudlets, func(i, j int) bool {
		return cloudlets[i][0] > cloudlets[j][0]
	})
	return cloudlets, nil
}

func GetCloudletPoolSummaryData(ctx context.Context, username string, report *ormapi.GenerateReport) ([][]string, error) {
	// get pools associated with orgs
	op := ormapi.OrgCloudletPool{
		Region:          report.Region,
		CloudletPoolOrg: report.Org,
	}
	poolAcceptedDevelopers := make(map[string][]string)
	poolPendingDevelopers := make(map[string][]string)
	acceptedOps, err := showCloudletPoolAccessObj(ctx, username, &op, accessTypeGranted)
	if err != nil {
		return nil, err
	}
	pendingOps, err := showCloudletPoolAccessObj(ctx, username, &op, accessTypePending)
	if err != nil {
		return nil, err
	}
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
		region:   report.Region,
		username: username,
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
	sort.Slice(cloudletpools, func(i, j int) bool {
		return cloudletpools[i][0] > cloudletpools[j][0]
	})
	return cloudletpools, nil
}

func GetCloudletResourceUsageData(ctx context.Context, username string, report *ormapi.GenerateReport) (map[string]TimeChartDataMap, error) {
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

	// Check the operator against the username
	if err := authorized(ctx, username, report.Org, ResourceCloudletAnalytics, ActionView); err != nil {
		return nil, err
	}

	chartMap := make(map[string]TimeChartDataMap)
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
					if len(val) < 4 {
						// not enough data
						continue
					}
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
					if _, ok := chartMap[cloudlet]; !ok {
						chartMap[cloudlet] = make(TimeChartDataMap)
					}
					for resIndex := 3; resIndex < len(val); resIndex++ {
						resName := row.Columns[resIndex]
						if _, ok := chartMap[cloudlet][resName]; !ok {
							chartMap[cloudlet][resName] = []TimeChartData{TimeChartData{}}
						}
						clData := chartMap[cloudlet][resName][0]

						if val[resIndex] == nil {
							continue
						}
						resVal, err := val[resIndex].(json.Number).Float64()
						if err != nil {
							log.SpanLog(ctx, log.DebugLevelInfo, "failed to parse resource value", "value", val[resIndex])
							continue
						}

						clData.XValues = append(clData.XValues, time)
						clData.YValues = append(clData.YValues, float64(resVal))
						chartMap[cloudlet][resName][0] = clData
					}
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	return chartMap, nil
}

func GetCloudletFlavorUsageData(ctx context.Context, username string, report *ormapi.GenerateReport) (map[string]PieChartDataMap, error) {
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

	// Check the operator against the username
	if err := authorized(ctx, username, report.Org, ResourceCloudletAnalytics, ActionView); err != nil {
		return nil, err
	}

	flavorMap := make(map[string]PieChartDataMap)
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
					if len(val) < 5 {
						// not enough data
						continue
					}
					cloudlet, ok := val[1].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch cloudlet name", "cloudlet", val[1])
					}
					if val[3] == nil {
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
						flavorMap[cloudlet] = make(PieChartDataMap)
					}
					if curVal, ok := flavorMap[cloudlet][flavor]; ok {
						if countVal <= curVal {
							continue
						}
					}
					flavorMap[cloudlet][flavor] = countVal
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	return flavorMap, nil
}

func GetCloudletEvents(ctx context.Context, username string, timezone *time.Location, report *ormapi.GenerateReport) (map[string][][]string, error) {
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

	events, err := nodeMgr.ShowEvents(ctx, &search)
	if err != nil {
		return nil, err
	}
	eventsData := make(map[string][][]string)
	for _, event := range events {
		cloudlet, ok := event.Mtags["cloudlet"]
		if !ok {
			log.SpanLog(ctx, log.DebugLevelInfo, "missing cloudlet name in event, skipping", "event", event)
			continue
		}
		timestamp := event.Timestamp.Format(TimeFormatDayDateTime)
		if timezone != nil {
			timestamp = event.Timestamp.In(timezone).Format(TimeFormatDayDateTime)
		}
		entry := []string{timestamp, event.Name}
		if _, ok := eventsData[cloudlet]; !ok {
			eventsData[cloudlet] = [][]string{}
		}
		eventsData[cloudlet] = append(eventsData[cloudlet], entry)
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

func GetCloudletAlerts(ctx context.Context, username string, timezone *time.Location, report *ormapi.GenerateReport) (map[string][][]string, error) {
	alertsData := make(map[string][][]string)
	rc := &RegionContext{
		region:   report.Region,
		username: username,
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
		if timezone != nil {
			alertTime = alertTime.In(timezone)
		}
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
		alertTimeStr := alertTime.Format(TimeFormatDayDateTime)
		entry := []string{alertTimeStr, desc, alert.State}
		if _, ok := alertsData[cloudlet]; !ok {
			alertsData[cloudlet] = [][]string{}
		}
		alertsData[cloudlet] = append(alertsData[cloudlet], entry)
	}
	return alertsData, nil
}

func GetCloudletAppUsageData(ctx context.Context, username string, report *ormapi.GenerateReport) (map[string]PieChartDataMap, error) {
	rc := &RegionContext{
		region:   report.Region,
		username: username,
	}
	obj := &edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				CloudletKey: edgeproto.CloudletKey{
					Organization: report.Org,
				},
			},
		},
		State: edgeproto.TrackedState_READY,
	}
	appInsts, err := ShowAppInstObj(ctx, rc, obj)
	if err != nil {
		return nil, err
	}
	appsCount := make(map[string]PieChartDataMap)
	for _, appInst := range appInsts {
		cloudletName := appInst.Key.ClusterInstKey.CloudletKey.Name
		appOrg := appInst.Key.AppKey.Organization
		if _, ok := appsCount[cloudletName]; !ok {
			appsCount[cloudletName] = make(PieChartDataMap)
		}
		if _, ok := appsCount[cloudletName][appOrg]; ok {
			appsCount[cloudletName][appOrg]++
		} else {
			appsCount[cloudletName][appOrg] = 1
		}
	}
	return appsCount, nil
}

func GetAppResourceUsageData(ctx context.Context, username string, report *ormapi.GenerateReport) (map[string]map[string]TimeChartDataMap, error) {
	rc := &InfluxDBContext{}
	dbNames := []string{cloudcommon.DeveloperMetricsDbName}
	in := ormapi.RegionAppInstMetrics{
		Region: report.Region,
		AppInst: edgeproto.AppInstKey{
			ClusterInstKey: edgeproto.VirtualClusterInstKey{
				CloudletKey: edgeproto.CloudletKey{
					Organization: report.Org,
				},
			},
		},
		Selector:  "cpu,mem,disk,network",
		StartTime: report.StartTime,
		EndTime:   report.EndTime,
	}
	rc.region = in.Region
	claims := &UserClaims{Username: username}
	cloudletList, err := checkPermissionsAndGetCloudletList(ctx, claims, in.Region, in.AppInst.AppKey.Organization, ResourceAppAnalytics, in.AppInst.ClusterInstKey.CloudletKey)
	if err != nil {
		return nil, err
	}
	cmd := AppInstMetricsQuery(&in, cloudletList)

	/*
	 * map structure:
	 *     cloudlet:
	 *         apporg:
	 *             cpu: resourcedata
	 *             mem: resourcedata
	 *             disk: resourcedata
	 */
	cpuKey := "cpu"
	memKey := "mem"
	diskKey := "disk"
	sendBytesKey := "sendBytes"
	recvBytesKey := "recvBytes"
	resIndexMap := map[string]int{
		cpuKey:       9,
		memKey:       10,
		diskKey:      11,
		sendBytesKey: 12,
		recvBytesKey: 13,
	}
	appMap := make(map[edgeproto.AppInstKey]map[string]TimeChartData)
	err = influxStream(ctx, rc, dbNames, cmd, func(res interface{}) {
		results, ok := res.([]influxdb.Result)
		if !ok {
			return
		}
		for _, result := range results {
			for _, row := range result.Series {
				/*
					columns[0] -> time
					columns[1] -> appname
					columns[2] -> appvers
					columns[3] -> clustername
					columns[4] -> clusterorg
					columns[5] -> cloudlet
					columns[6] -> cloudletorg
					columns[7] -> apporg
					columns[8] -> pod
					columns[9:] -> cpu,mem,disk,sendBytes,recvBytes
				*/
				for _, val := range row.Values {
					if len(val) < 10 {
						// not enough data
						continue
					}
					timeStr, ok := val[0].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch resource time", "time", val[0])
					}
					time, err := time.Parse("2006-01-02T15:04:05Z", timeStr)
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to parse resource time", "time", timeStr)
						continue
					}
					appname, ok := val[1].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch app name", "app", val[1])
					}
					appvers, ok := val[2].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch app version", "appvers", val[2])
					}
					apporg, ok := val[7].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch app org", "apporg", val[7])
					}
					clustername, ok := val[3].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch cluster name", "clustername", val[3])
					}
					clusterorg, ok := val[4].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch cluster org", "clusterorg", val[4])
					}
					cloudlet, ok := val[5].(string)
					if !ok {
						log.SpanLog(ctx, log.DebugLevelInfo, "failed to fetch cloudlet name", "cloudlet", val[5])
					}
					appInstKey := edgeproto.AppInstKey{
						AppKey: edgeproto.AppKey{
							Name:         appname,
							Organization: apporg,
							Version:      appvers,
						},
						ClusterInstKey: edgeproto.VirtualClusterInstKey{
							ClusterKey: edgeproto.ClusterKey{
								Name: clustername,
							},
							Organization: clusterorg,
							CloudletKey: edgeproto.CloudletKey{
								Name:         cloudlet,
								Organization: report.Org,
							},
						},
					}
					appId := fmt.Sprintf("%s|%s|%s|%s", appname, appvers, clustername, clusterorg)
					if _, ok := appMap[appInstKey]; !ok {
						appMap[appInstKey] = make(map[string]TimeChartData)
						for resName, _ := range resIndexMap {
							appMap[appInstKey][resName] = TimeChartData{Name: appId}
						}
					}
					// parse resource fields
					for resName, resIndex := range resIndexMap {
						if resIndex >= len(val) {
							break
						}
						if val[resIndex] != nil {
							resVal, err := val[resIndex].(json.Number).Float64()
							if err != nil {
								log.SpanLog(ctx, log.DebugLevelInfo, "failed to parse resource value", "resource", resName, "value", val[resIndex])
								continue
							}

							clData := appMap[appInstKey][resName]
							clData.XValues = append(clData.XValues, time)
							clData.YValues = append(clData.YValues, float64(resVal))
							appMap[appInstKey][resName] = clData
						}
					}
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	chartMap := make(map[string]map[string]TimeChartDataMap)

	for appInstKey, resChartData := range appMap {
		cloudlet := appInstKey.ClusterInstKey.CloudletKey.Name
		apporg := appInstKey.AppKey.Organization
		for resName, resData := range resChartData {
			if _, ok := chartMap[cloudlet]; !ok {
				chartMap[cloudlet] = make(map[string]TimeChartDataMap)
			}
			if _, ok := chartMap[cloudlet][apporg]; !ok {
				chartMap[cloudlet][apporg] = make(TimeChartDataMap)
			}
			resTitle := apporg + " " + resName
			if _, ok := chartMap[cloudlet][apporg][resTitle]; !ok {
				chartMap[cloudlet][apporg][resTitle] = []TimeChartData{}
			}
			chartMap[cloudlet][apporg][resTitle] = append(chartMap[cloudlet][apporg][resTitle], resData)
		}
	}
	return chartMap, nil
}

func GetAppStateEvents(ctx context.Context, username string, timezone *time.Location, report *ormapi.GenerateReport) (map[string][][]string, error) {
	search := node.EventSearch{
		Match: node.EventMatch{
			Names:   []string{"AppInst online", "AppInst offline"},
			Types:   []string{node.EventType},
			Regions: []string{report.Region},
			Tags: map[string]string{
				"cloudletorg": report.Org,
			},
		},
		TimeRange: util.TimeRange{
			StartTime: report.StartTime,
			EndTime:   report.EndTime,
		},
	}

	events, err := nodeMgr.ShowEvents(ctx, &search)
	if err != nil {
		return nil, err
	}
	reqdTags := map[string]string{
		"app":        "",
		"apporg":     "",
		"appver":     "",
		"cloudlet":   "",
		"cluster":    "",
		"clusterorg": "",
	}
	eventsData := make(map[string][][]string)
	for _, event := range events {
		found := true
		for tag, _ := range reqdTags {
			tagVal, ok := event.Mtags[tag]
			if !ok {
				log.SpanLog(ctx, log.DebugLevelInfo, "missing tag in event, skipping", "tag", tag, "event", event)
				found = false
				break
			}
			reqdTags[tag] = tagVal
		}
		if !found {
			continue
		}
		timestamp := event.Timestamp.Format(TimeFormatDayDateTime)
		if timezone != nil {
			timestamp = event.Timestamp.In(timezone).Format(TimeFormatDayDateTime)
		}

		entry := []string{
			timestamp,
			reqdTags["app"] + " | " + reqdTags["appver"],         // App Identifier
			reqdTags["apporg"],                                   // Developer Org
			reqdTags["cluster"] + " | " + reqdTags["clusterorg"], // Cluster Identifier
			event.Name, // Event name
		}
		cloudlet := reqdTags["cloudlet"]
		if _, ok := eventsData[cloudlet]; !ok {
			eventsData[cloudlet] = [][]string{}
		}
		eventsData[cloudlet] = append(eventsData[cloudlet], entry)
	}
	return eventsData, nil
}

func GenerateCloudletReport(ctx context.Context, username string, regions []string, report *ormapi.GenerateReport, pdfOut *bytes.Buffer) error {
	// fetch logo path
	logoPath := serverConfig.StaticDir + "/MobiledgeX_Logo.png"
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		return fmt.Errorf("Missing logo")
	}
	pdfReport, err := NewReport(report)
	if err != nil {
		return err
	}
	for _, region := range regions {
		log.SpanLog(ctx, log.DebugLevelInfo, "Generate operator report for region", "region", region)
		// start new page for every region
		report.Region = region
		pdfReport.AddPage()

		pdfReport.AddReportTitle(logoPath)
		pdfReport.AddHeader(report, logoPath, NoCloudlet)
		pdfReport.AddFooter()
		pdfReport.AddOperatorInfo(report)
		pdfReport.AddHorizontalLine()

		// Step-1: Gather all data
		// -------------------------
		// Get cloudlet summary
		cloudlets_summary, err := GetCloudletSummaryData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet summary: %v", err)
		}

		// Get list of cloudletpools
		cloudletpools, err := GetCloudletPoolSummaryData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet pool summary: %v", err)
		}

		cloudlets := make(map[string]struct{})
		// Get cloudlet resource usage metrics
		resourceUsageCharts, err := GetCloudletResourceUsageData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet resource usage data: %v", err)
		}
		for cloudletName, _ := range resourceUsageCharts {
			cloudlets[cloudletName] = struct{}{}
		}

		// Get top flavors used per Cloudlet
		flavorData, err := GetCloudletFlavorUsageData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet flavor usage data: %v", err)
		}
		for cloudletName, _ := range flavorData {
			cloudlets[cloudletName] = struct{}{}
		}

		// Get cloudlet events
		eventsData, err := GetCloudletEvents(ctx, username, pdfReport.timezone, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet events: %v", err)
		}
		for cloudletName, _ := range eventsData {
			cloudlets[cloudletName] = struct{}{}
		}

		// Get cloudlet alerts
		alertsData, err := GetCloudletAlerts(ctx, username, pdfReport.timezone, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet alerts: %v", err)
		}
		for cloudletName, _ := range alertsData {
			cloudlets[cloudletName] = struct{}{}
		}

		// Get app count by developer on cloudlet
		appCountData, err := GetCloudletAppUsageData(ctx, username, report)
		if err != nil {
			return fmt.Errorf("failed to get cloudlet app count data: %v", err)
		}
		for cloudletName, _ := range appCountData {
			cloudlets[cloudletName] = struct{}{}
		}

		// Get app state events
		appEventsData, err := GetAppStateEvents(ctx, username, pdfReport.timezone, report)
		if err != nil {
			return fmt.Errorf("failed to get app events: %v", err)
		}
		for cloudletName, _ := range appEventsData {
			cloudlets[cloudletName] = struct{}{}
		}

		// Step-2: Render data
		// -------------------------
		// Get list of cloudlets
		header := []string{"Name", "Platform Type", "Last Known State"}
		columnsWidth := []float64{60, 30, 35}
		pdfReport.AddTable("Cloudlets", header, cloudlets_summary, columnsWidth)

		// Get list of cloudletpools
		header = []string{"Name", "Associated Cloudlets", "Accepted Developers", "Pending Developers"}
		columnsWidth = []float64{30, 60, 50, 50}
		pdfReport.AddTable("CloudletPools", header, cloudletpools, columnsWidth)

		// Sort cloudlet by name
		cloudletNames := []string{}
		for k := range cloudlets {
			cloudletNames = append(cloudletNames, k)
		}
		sort.Strings(cloudletNames)

		// Show per cloudlet reports
		for _, cloudletName := range cloudletNames {
			// Start new page
			pdfReport.AddHeader(report, logoPath, cloudletName)
			pdfReport.AddPage()

			pdfReport.AddPageTitle(cloudletName)

			// Get cloudlet resource usage metrics
			if data, ok := resourceUsageCharts[cloudletName]; ok {
				err = pdfReport.AddResourceTimeCharts(cloudletName, data, ChartSpec{FillColor: true})
				if err != nil {
					return err
				}
			}

			// Get top flavors used per Cloudlet
			if data, ok := flavorData[cloudletName]; ok {
				err = pdfReport.AddPieChart(cloudletName, "Flavors Used", data)
				if err != nil {
					return err
				}
			}
			// Get cloudlet events
			if data, ok := eventsData[cloudletName]; ok {
				header = []string{"Timestamp", "Description"}
				columnsWidth = []float64{40, 100}
				pdfReport.AddTable("Cloudlet Events", header, data, columnsWidth)
			}

			// Get cloudlet alerts
			if data, ok := alertsData[cloudletName]; ok {
				header = []string{"Timestamp", "Description", "State"}
				columnsWidth = []float64{40, 100, 30}
				pdfReport.AddTable("Cloudlet Alerts", header, data, columnsWidth)
			}

			// Get app count by developer on cloudlet
			if data, ok := appCountData[cloudletName]; ok {
				err = pdfReport.AddPieChart(cloudletName, "App Count By Developer", data)
				if err != nil {
					return err
				}
			}

			// Get app state events
			if data, ok := appEventsData[cloudletName]; ok {
				header = []string{"Timestamp", "App Info", "Developer", "Cluster Info", "State"}
				columnsWidth = []float64{30, 40, 40, 40, 30}
				pdfReport.AddTable("Developer App State", header, data, columnsWidth)
			}
		}

		pdfReport.AddHeader(report, logoPath, NoCloudlet)

		if err = pdfReport.Err(); err != nil {
			return fmt.Errorf("failed to create PDF report: %s\n", err.Error())
		}
		err = pdfReport.Output(pdfOut)
		if err != nil {
			return fmt.Errorf("cannot get PDF output: %v", err)
		}
	}
	return nil
}

func ShowReport(c echo.Context) error {
	// get list of generated reports from cloud
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reportQuery := ormapi.DownloadReport{}
	if err := c.Bind(&reportQuery); err != nil {
		return bindErr(c, err)
	}
	// sanity check
	if reportQuery.Org == "" {
		return setReply(c, fmt.Errorf("Org name has to be specified"), nil)
	}
	// check if user is authorized to view reports
	if err := authorized(ctx, claims.Username, reportQuery.Org, ResourceCloudlets, ActionManage); err != nil {
		return setReply(c, err, nil)
	}

	storageClient, err := gcs.NewClient(ctx, serverConfig.vaultConfig, serverConfig.DeploymentTag)
	if err != nil {
		return setReply(c, fmt.Errorf("Unable to setup GCS client: %v", err), nil)
	}
	objs, err := storageClient.ListObjects(ctx)
	if err != nil {
		return setReply(c, fmt.Errorf("Unable to get reports from GCS: %v", err), nil)
	}
	pattern := ormapi.GetReportFileNameRE()
	out := []string{}
	for _, obj := range objs {
		allStrs := regexp.MustCompile(pattern).Split(obj, -1)
		if len(allStrs) < 2 {
			continue
		}
		fileOrgPrefix := allStrs[0]
		if fileOrgPrefix != reportQuery.Org &&
			fileOrgPrefix != strings.ToLower(reportQuery.Org) {
			continue
		}
		out = append(out, obj)
	}
	return c.JSON(http.StatusOK, out)
}

func DownloadReport(c echo.Context) error {
	// download report from cloud with given filename
	ctx := GetContext(c)
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	reportQuery := ormapi.DownloadReport{}
	if err := c.Bind(&reportQuery); err != nil {
		return bindErr(c, err)
	}
	// sanity check
	if reportQuery.Org == "" {
		return setReply(c, fmt.Errorf("Org name has to be specified"), nil)
	}
	if reportQuery.Filename == "" {
		return setReply(c, fmt.Errorf("Report filename has to be specified"), nil)
	}
	pattern := ormapi.GetReportFileNameRE()
	allStrs := regexp.MustCompile(pattern).Split(reportQuery.Filename, -1)
	if len(allStrs) < 2 {
		return setReply(c, fmt.Errorf("Unable to get org name from filename: %s", reportQuery.Filename), nil)
	}
	fileOrgPrefix := allStrs[0]
	if fileOrgPrefix != reportQuery.Org &&
		fileOrgPrefix != strings.ToLower(reportQuery.Org) {
		return setReply(c, fmt.Errorf("Only org %s related reports can be accessed", reportQuery.Org), nil)
	}
	// check if user is authorized to view reports
	if err := authorized(ctx, claims.Username, reportQuery.Org, ResourceCloudlets, ActionManage); err != nil {
		return setReply(c, err, nil)
	}

	storageClient, err := gcs.NewClient(ctx, serverConfig.vaultConfig, serverConfig.DeploymentTag)
	if err != nil {
		return setReply(c, fmt.Errorf("Unable to setup GCS client: %v", err), nil)
	}
	objs, err := storageClient.ListObjects(ctx)
	if err != nil {
		return setReply(c, fmt.Errorf("Unable to get reports from GCS: %v", err), nil)
	}
	found := false
	for _, obj := range objs {
		if reportQuery.Filename == obj {
			found = true
			break
		}
	}
	if !found {
		return setReply(c, fmt.Errorf("Report with name %s does not exist", reportQuery.Filename), nil)
	}
	outBytes, err := storageClient.DownloadObject(ctx, reportQuery.Filename)
	if err != nil {
		return setReply(c, fmt.Errorf("Unable to download report %s: %v", reportQuery.Filename, err), nil)
	}
	return c.HTMLBlob(http.StatusOK, outBytes)
}
