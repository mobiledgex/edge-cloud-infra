package collections

import (
	"context"
	"fmt"
	"strconv"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud-infra/billing/zuora"
	"github.com/mobiledgex/edge-cloud-infra/mc/orm"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	opentracing "github.com/opentracing/opentracing-go"
)

var clusterInstUsageInfluxCmd = `select "org","cloudlet","cloudletorg","cluster","clusterorg","flavor","start","end","uptime" from "clusterinst-usage"` +
	`where time >= '%s' and time < '%s'`

func CollectDailyClusterUsage(ctx context.Context) {
	for {
		span := log.StartSpan(log.DebugLevelInfo, "Cluster usage collection thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
		select {
		case <-time.After(timeTilNextDay()):
			controllers, err := orm.ShowControllerObj(ctx, nil)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get regions to query influx", "err", err)
				return
			}
			regions := make(map[string]bool)
			for _, controller := range controllers {
				regions[controller.Region] = true
			}
			// get usage from every region
			now := time.Now()
			// grab usage from the day before
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
			cmd := fmt.Sprintf(clusterInstUsageInfluxCmd, yesterday.Format(time.RFC3339), today.Format(time.RFC3339))
			for region, _ := range regions {
				// connect to influx and query it
				influx, err := orm.ConnectInfluxDB(ctx, region)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfo, "Unable to connect to influx", "region", region, "err", err)
					continue
				}
				query := influxdb.Query{
					Command:  cmd,
					Database: cloudcommon.EventsDbName,
				}
				resp, err := influx.Query(query)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfo, "InfluxDB query failed",
						"region", region, "query", query, "resp", resp, "err", err)
					continue
				}
				if resp.Error() != nil {
					log.SpanLog(ctx, log.DebugLevelInfo, "InfluxDB query failed",
						"region", region, "query", query, "err", resp.Error())
					continue
				}
				empty, err := checkInfluxQueryOutput(resp.Results, cloudcommon.ClusterInstUsage)
				if empty {
					// no usage records to upload
					continue
				}
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfo, "Invalid influx output", "region", region, "err", err)
					continue
				}
				// call RecordClusterUsage for every entry
				for _, values := range resp.Results[0].Series[0].Values {
					// value should be of the format [timestamp org cloudlet cloudletorg cluster clusterorg flavor start end uptime]
					org := fmt.Sprintf("%v", values[1])
					cloudlet := fmt.Sprintf("%v", values[2])
					cloudletorg := fmt.Sprintf("%v", values[3])
					cluster := fmt.Sprintf("%v", values[4])
					clusterorg := fmt.Sprintf("%v", values[5])
					flavor := fmt.Sprintf("%v", values[6])
					start, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[7]))
					end, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[8]))
					uptime, _ := strconv.ParseFloat(fmt.Sprintf("%v", values[9]), 64)
					key := edgeproto.ClusterInstKey{
						ClusterKey:   edgeproto.ClusterKey{Name: cluster},
						Organization: clusterorg,
						CloudletKey:  edgeproto.CloudletKey{Name: cloudlet, Organization: cloudletorg},
					}
					var accountInfo *zuora.AccountInfo
					accountInfo, err = orm.GetAccountObj(ctx, org)
					if err != nil {
						log.SpanLog(ctx, log.DebugLevelInfo, "Unable to get account info", "org", org, "err", err)
					} else {
						err = zuora.RecordClusterUsage(accountInfo, &key, flavor, start, end, uptime)
						if err != nil {
							log.SpanLog(ctx, log.DebugLevelInfo, "Unable to record Cluster usage in Zuora", "err", err)
						}
					}
				}
			}
		}
		span.Finish()
	}
}

// This one is for demo purposes (to please the wonho)
// func timeTilNextDay() time.Duration {
// 	// make sure to change today and yesterday in CollectDailyClusterUsage if you enable this version
// 	return time.Minute * 3
// }

func timeTilNextDay() time.Duration {
	now := time.Now()
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 30, 0, 0, time.UTC)
	return nextDay.Sub(now)
}

func checkInfluxQueryOutput(result []client.Result, dbName string) (bool, error) {
	empty := false
	var valid error
	if len(result) == 0 || len(result[0].Series) == 0 {
		empty = true
	} else if len(result) != 1 ||
		len(result[0].Series) != 1 ||
		len(result[0].Series[0].Values) == 0 ||
		len(result[0].Series[0].Values[0]) == 0 ||
		result[0].Series[0].Name != dbName {
		// should only be 1 series, the 'dbName' one
		valid = fmt.Errorf("Error parsing influx, unexpected format")
	}
	return empty, valid
}
