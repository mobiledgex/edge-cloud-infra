package orm

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
)

var retryMax = 3
var retryPercentage = 0.05 // this number is a percentage, so that the retryInterval is based off of the collectionInterval

var InfluxMinimumTimestamp, _ = time.Parse(time.RFC3339, "1677-09-21T00:13:44Z")

func CollectBillingUsage(ctx context.Context, collectInterval time.Duration) {
	retryInterval := 5 * time.Minute
	prevCollectTime := InfluxMinimumTimestamp
	nextCollectTime := nextCollectTime(time.Now(), collectInterval)
	if collectInterval.Seconds() > float64(0) {
		retryInterval = time.Duration(retryPercentage * float64(usageInterval))
	}
	for {
		select {
		case <-time.After(nextCollectTime.Sub(time.Now())):
			span := log.StartSpan(log.DebugLevelInfo, "Billing usage collection thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
			controllers, err := ShowControllerObj(ctx, nil)
			if err != nil {
				retryCount := 0
				for retryCount < retryMax {
					log.SpanLog(ctx, log.DebugLevelInfo, fmt.Sprintf("Unable to get regions to query influx, retrying in %v", retryInterval), "err", err)
					time.Sleep(retryInterval)
					controllers, err = ShowControllerObj(ctx, nil)
					if err == nil {
						break
					}
					retryCount = retryCount + 1
				}
				if err != nil {
					unsuccessfulCollects = unsuccessfulCollects + 1
					log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get regions to query influx, waiting until next collection period", "err", err)
					nextCollectTime = nextCollectTime(nextCollectTime, collectInterval)
					span.Finish()
					continue
				}
			}

			regions := make(map[string]bool)
			for _, controller := range controllers {
				regions[controller.Region] = true
			}
			// get usage from every region
			for region, _ := range regions {
				recordRegionUsage(ctx, region, prevCollectTime, nextCollectTime)
			}
			prevCollectTime = nextCollectTime
			nextCollectTime = nextCollectTime(nextCollectTime, collectInterval)
			span.Finish()
		}
	}
}

func recordRegionUsage(ctx context.Context, region string, start, end time.Time) {
	rc := InfluxDBContext{region: region}
	appIn := ormapi.RegionAppInstUsage{
		Region:    region,
		StartTime: start,
		EndTime:   end,
		VmOnly:    true,
	}
	eventCmd := AppInstUsageEventsQuery(&appIn)
	checkpointCmd := AppInstCheckpointsQuery(&appIn)
	eventResp, checkResp, err := GetEventAndCheckpoint(ctx, rc, eventCmd, checkpointCmd)
	if err != nil {
		log.SpanLog(ctx, "Error gathering app usage for billing", "region", region, "err", err)
		return
	}
	appUsage, err = GetAppUsage(eventResp, checkResp, appIn.StartTime, appIn.EndTime, appIn.Region)
	if err != nil {
		log.SpanLog(ctx, "Error parsing app usage for billing", "region", region, "err", err)
		return
	}
	recordAppUsages(ctx, appUsage)

	clusterIn := ormapi.RegionClusterInstUsage{
		Region:    region,
		StartTime: start,
		EndTime:   end,
	}
	eventCmd = ClusterUsageEventsQuery(&clusterIn)
	checkpointCmd = ClusterCheckpointsQuery(&clusterIn)
	eventResp, checkResp, err := GetEventAndCheckpoint(ctx, rc, eventCmd, checkpointCmd)
	if err != nil {
		log.SpanLog(ctx, "Error gathering cluster usage for billing", "region", region, "err", err)
		return
	}
	clusterUsage, err = GetClusterUsage(eventResp, checkResp, clusterIn.StartTime, clusterIn.EndTime, clusterIn.Region)
	if err != nil {
		log.SpanLog(ctx, "Error parsing cluster usage for billing", "region", region, "err", err)
		return
	}
	recordClusterUsages(ctx, clusterUsage)
}

func recordAppUsages(ctx context.Context, usage *ormapi.MetricData) {
	orgTracker := make(map[string][]billing.UsageRecord)
	if len(usage.Series[0]) == 0 {
		// techincally if GetAppUsage doesnt fail, this should be impossible, but check anyway so we dont crash if it did happen
		log.SpanLog(ctx, log.DebugLevelInfo, "Invalid app usage")
	}
	for _, value := range usage.Series[0].Values {
		// ordering is from appInstDataColumns
		newAppInst := edgeproto.AppInstKey{
			AppKey: edgeproto.AppKey{
				Name:         value[1],
				Organization: value[2],
				Version:      value[3],
			},
			ClusterInstKey: edgeproto.ClusterInstKey{
				Organization: value[5],
				ClusterKey:   edgeproto.ClusterKey{Name: value[4]},
				CloudletKey: edgeproto.CloudletKey{
					Name:         value[6],
					Organization: value[7],
				},
			},
		}
		startTime := time.Parse(time.RFC3339, value[10])
		endTime := time.Parse(time.RFC3339, value[11])
		newRecord := billing.UsageRecord{
			FlavorName: value[8],
			NodeCount:  1,
			AppInst:    &newAppInst,
			StartTime:  startTime,
			EndTime:    endTime,
		}
		records, _ := orgTracker[newAppInst.AppKey.Organization]
		orgTracker[newAppInst.AppKey.Organization] = append(records, newRecord)
	}
	for org, record := range orgTracker {
		var accountInfo *billing.AccountInfo
		accountInfo, err := orm.GetAccountObj(ctx, org)
		if er != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get account info", "org", org, "err", err)
		} else {
			err = serverConfig.BillingService.RecordUsage(accountInfo, record)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Unable to record app usage", "err", err)
			}
		}
	}
}

func recordClusterUsages(ctx context.Context, usage *ormapi.MetricData) {
	orgTracker := make(map[string][]billing.UsageRecord)
	if len(usage.Series[0]) == 0 {
		// techincally if GetClusterUsage doesnt fail, this should be impossible, but check anyway so we dont crash if it did happen
		log.SpanLog(ctx, log.DebugLevelInfo, "Invalid cluster usage")
	}
	for _, value := range usage.Series[0].Values {
		// ordering is from clusterInstDataColumns
		newClusterInst := edgeproto.ClusterInstKey{
			Organization: value[2],
			ClusterKey:   edgeproto.ClusterKey{Name: value[1]},
			CloudletKey: edgeproto.CloudletKey{
				Name:         value[3],
				Organization: value[4],
			},
		}
		startTime := time.Parse(time.RFC3339, value[8])
		endTime := time.Parse(time.RFC3339, value[9])
		nodeCount, _ := strconv.Atoi(value[6])
		newRecord := billing.UsageRecord{
			FlavorName:  value[5],
			NodeCount:   nodeCount,
			ClusterInst: &newClusterInst,
			StartTime:   startTime,
			EndTime:     endTime,
		}
		records, _ := orgTracker[newClusterInst.Organization]
		orgTracker[newClusterInst.Organization] = append(records, newRecord)
	}
	for org, record := range orgTracker {
		var accountInfo *billing.AccountInfo
		accountInfo, err := orm.GetAccountObj(ctx, org)
		if er != nil {
			log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get account info", "org", org, "err", err)
		} else {
			err = serverConfig.BillingService.RecordUsage(accountInfo, record)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Unable to record cluster usage", "err", err)
			}
		}
	}
}

func nextCollectTime(now time.Time, collectInterval time.Duration) time.Time {
	if collectInterval.Seconds() > float64(0) { // if positive, use it
		return now.Truncate(collectInterval).Add(collectInterval) // truncate it so the times are nice
	}
	// if the collection interval specified is less than 0, the default is to collect once a day
	// default is to collect once a day at the start of the day 12am
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return nextDay
}
