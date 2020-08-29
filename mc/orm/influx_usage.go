package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var clusterCheckpointFields = []string{
	"\"flavor\"",
	"\"status\"",
	"\"nodecount\"",
	"\"ipaccess\"",
}

var clusterUsageEventFields = []string{
	"\"flavor\"",
	"\"event\"",
	"\"status\"",
	"\"nodecount\"",
	"\"ipaccess\"",
}

type usageTracker struct {
	flavor    string
	time      time.Time
	nodecount int64
	ipaccess  string
}

// TODO: sync this up with controllers checkPointInterval somehow
var checkpointInterval = "3m"

// Get most recent checkpoint with respect to t
func prevCheckpoint(t time.Time) time.Time {
	if checkpointInterval == "MONTH" {
		y, m, _ := t.Date()
		return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	}
	// TODO: whe you move checkpointInterval to command line, error check this
	dur, _ := time.ParseDuration(checkpointInterval)
	return t.Truncate(dur)
}

func GetClusterUsage(event *client.Response, checkpoint *client.Response, start, end time.Time, region string) (*ormapi.AllUsage, error) {
	usageRecords := ormapi.AllUsage{
		Data: make([]ormapi.UsageRecord, 0),
	}
	clusterTracker := make(map[edgeproto.ClusterInstKey]usageTracker)

	// check to see if the influx output is empty or invalid
	emptyEvents, err := checkInfluxOutput(event, EVENT_CLUSTERINST)
	if err != nil {
		return nil, err
	}
	emptyCheckpoints, err := checkInfluxOutput(checkpoint, cloudcommon.ClusterInstCheckpoints)
	if err != nil {
		return nil, err
	}
	if emptyEvents && emptyCheckpoints {
		return &usageRecords, nil
	}

	// grab the checkpoints of clusters that are up
	if !emptyCheckpoints {
		for _, values := range checkpoint.Results[0].Series[0].Values {
			// format [timestamp cluster clusterorg cloudlet cloudletorg flavor status nodecount ipaccess]
			if len(values) != 9 {
				return nil, fmt.Errorf("Error parsing influx response")
			}
			timestamp, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[0]))
			if err != nil {
				return nil, fmt.Errorf("Unable to parse timestamp: %v", err)
			}
			cluster := fmt.Sprintf("%v", values[1])
			clusterorg := fmt.Sprintf("%v", values[2])
			cloudlet := fmt.Sprintf("%v", values[3])
			cloudletorg := fmt.Sprintf("%v", values[4])
			flavor := fmt.Sprintf("%v", values[5])
			status := fmt.Sprintf("%v", values[6])
			nodecount, err := values[7].(json.Number).Int64()
			if err != nil {
				return nil, fmt.Errorf("Error trying to convert nodecount to int: %s", err)
			}
			ipaccess := fmt.Sprintf("%v", values[8])

			if status == cloudcommon.InstanceUp {
				newTracker := edgeproto.ClusterInstKey{
					ClusterKey: edgeproto.ClusterKey{Name: cluster},
					CloudletKey: edgeproto.CloudletKey{
						Organization: cloudletorg,
						Name:         cloudlet,
					},
					Organization: clusterorg,
				}
				clusterTracker[newTracker] = usageTracker{
					flavor:    flavor,
					time:      timestamp,
					nodecount: nodecount,
					ipaccess:  ipaccess,
				}
			}
		}
	}

	// these records are ordered from most recent, so iterate backwards
	if !emptyEvents {
		for i := len(event.Results[0].Series[0].Values) - 1; i >= 0; i-- {
			values := event.Results[0].Series[0].Values[i]
			// value should be of the format [timestamp cluster clusterorg cloudlet cloudletorg flavor event status nodecount ipaccess]
			if len(values) != 10 {
				return nil, fmt.Errorf("Error parsing influx response")
			}
			timestamp, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[0]))
			if err != nil {
				return nil, fmt.Errorf("Unable to parse timestamp: %v", err)
			}

			cluster := fmt.Sprintf("%v", values[1])
			clusterorg := fmt.Sprintf("%v", values[2])
			cloudlet := fmt.Sprintf("%v", values[3])
			cloudletorg := fmt.Sprintf("%v", values[4])
			flavor := fmt.Sprintf("%v", values[5])
			event := fmt.Sprintf("%v", values[6])
			status := fmt.Sprintf("%v", values[7])
			nodecount, err := values[8].(json.Number).Int64()
			if err != nil {
				return nil, fmt.Errorf("Error trying to convert nodecount to int: %s", err)
			}
			ipaccess := fmt.Sprintf("%v", values[9])

			//if the timestamp is before start and its a down, then get rid of it in the cluster tracker
			//otherwise put it in the cluster tracker
			newKey := edgeproto.ClusterInstKey{
				ClusterKey: edgeproto.ClusterKey{Name: cluster},
				CloudletKey: edgeproto.CloudletKey{
					Organization: cloudletorg,
					Name:         cloudlet,
				},
				Organization: clusterorg,
			}
			tracker, ok := clusterTracker[newKey]
			if status == cloudcommon.InstanceUp {
				if !ok {
					newTracker := usageTracker{
						flavor:    flavor,
						time:      timestamp,
						nodecount: nodecount,
						ipaccess:  ipaccess,
					}
					clusterTracker[newKey] = newTracker
				}
			} else if status == cloudcommon.InstanceDown {
				if !timestamp.Before(start) {
					newRecord := ormapi.UsageRecord{
						Region:       region,
						Organization: clusterorg,
						ClusterName:  cluster,
						Cloudlet:     cloudlet,
						CloudletOrg:  cloudletorg,
						EndTime:      timestamp,
						Note:         event,
						Flavor:       flavor,
						NumNodes:     int(nodecount),
						IpAccess:     ipaccess,
					}
					if tracker.time.Before(start) {
						newRecord.StartTime = start
					} else {
						newRecord.StartTime = tracker.time
					}
					newRecord.Duration = newRecord.EndTime.Sub(newRecord.StartTime)

					usageRecords.Data = append(usageRecords.Data, newRecord)
				}
				delete(clusterTracker, newKey)
			} else {
				return nil, fmt.Errorf("Unexpected influx status: %s", status)
			}
		}
	}

	// anything still in the clusterTracker is a currently running clusterinst
	for k, v := range clusterTracker {
		newRecord := ormapi.UsageRecord{
			Region:       region,
			Organization: k.Organization,
			ClusterName:  k.ClusterKey.Name,
			Cloudlet:     k.CloudletKey.Name,
			CloudletOrg:  k.CloudletKey.Organization,
			EndTime:      end,
			Note:         "Running",
			Flavor:       v.flavor,
			NumNodes:     int(v.nodecount),
			IpAccess:     v.ipaccess,
		}
		if v.time.Before(start) {
			newRecord.StartTime = start
		} else {
			newRecord.StartTime = v.time
		}
		newRecord.Duration = newRecord.EndTime.Sub(newRecord.StartTime)
		newRecord.Flavor = v.flavor
		newRecord.IpAccess = v.ipaccess

		usageRecords.Data = append(usageRecords.Data, newRecord)
	}

	return &usageRecords, nil
}

// Query is a template with a specific set of if/else
func ClusterCheckpointsQuery(obj *ormapi.RegionClusterInstUsage) string {
	arg := influxQueryArgs{
		Selector:     strings.Join(append(ClusterFields, clusterCheckpointFields...), ","),
		Measurement:  cloudcommon.ClusterInstCheckpoints,
		OrgField:     "org",
		ApiCallerOrg: obj.ClusterInst.Organization,
		CloudletName: obj.ClusterInst.CloudletKey.Name,
		ClusterName:  obj.ClusterInst.ClusterKey.Name,
		CloudletOrg:  obj.ClusterInst.CloudletKey.Organization,
	}
	// set endtime to start and back up starttime by a checkpoint interval to hit the most recent
	// checkpoint that occurred before startTime
	checkpointTime := prevCheckpoint(obj.StartTime)
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &checkpointTime, &checkpointTime)
}

func ClusterUsageEventsQuery(obj *ormapi.RegionClusterInstUsage) string {
	arg := influxQueryArgs{
		Selector:     strings.Join(append(ClusterFields, clusterUsageEventFields...), ","),
		Measurement:  EVENT_CLUSTERINST,
		OrgField:     "clusterorg",
		ApiCallerOrg: obj.ClusterInst.Organization,
		CloudletName: obj.ClusterInst.CloudletKey.Name,
		ClusterName:  obj.ClusterInst.ClusterKey.Name,
		CloudletOrg:  obj.ClusterInst.CloudletKey.Organization,
	}
	queryStart := prevCheckpoint(obj.StartTime)
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &queryStart, &obj.EndTime)
}

func checkInfluxOutput(resp *client.Response, measurement string) (bool, error) {
	// check to see if the influx output is empty or invalid
	if len(resp.Results) == 0 || len(resp.Results[0].Series) == 0 {
		// empty, no event logs at all
		return true, nil
	} else if len(resp.Results) != 1 ||
		len(resp.Results[0].Series) != 1 ||
		len(resp.Results[0].Series[0].Values) == 0 ||
		len(resp.Results[0].Series[0].Values[0]) == 0 ||
		resp.Results[0].Series[0].Name != measurement {
		// should only be 1 series, the 'measurement' one
		return false, fmt.Errorf("Error parsing influx, unexpected format")
	}
	return false, nil
}

func GetEventAndCheckpoint(ctx context.Context, rc *InfluxDBContext, eventCmd, checkpointCmd string) (*client.Response, *client.Response, error) {
	var eventResponse, checkpointResponse *client.Response
	err := influxStream(ctx, rc, cloudcommon.EventsDbName, eventCmd, func(res interface{}) {
		resp, ok := res.([]client.Result)
		if ok {
			eventResponse = &client.Response{Results: resp}
		}
	})
	if err != nil {
		return nil, nil, err
	}
	err = influxStream(ctx, rc, cloudcommon.EventsDbName, checkpointCmd, func(res interface{}) {
		resp, ok := res.([]client.Result)
		if ok {
			checkpointResponse = &client.Response{Results: resp}
		}
	})
	if err != nil {
		return nil, nil, err
	}
	if eventResponse == nil {
		return nil, nil, fmt.Errorf("unable to get event log")
	} else if checkpointResponse == nil {
		return nil, nil, fmt.Errorf("unable to get checkpoint log")
	} else {
		return eventResponse, checkpointResponse, nil
	}
}

// Common method to handle both app and cluster metrics
func GetUsageCommon(c echo.Context) error {
	var checkpointCmd, eventCmd, org string
	var usage *ormapi.AllUsage
	rc := &InfluxDBContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims
	ctx := GetContext(c)

	if strings.HasSuffix(c.Path(), "usage/app") {
		in := ormapi.RegionAppInstEvents{} // TODO: change this to RegionAppInstUsage{}
		success, err := ReadConn(c, &in)
		if !success {
			return err
		}
		// Developer name has to be specified
		if in.AppInst.AppKey.Organization == "" {
			return setReply(c, fmt.Errorf("App details must be present"), nil)
		}
		rc.region = in.Region
		org = in.AppInst.AppKey.Organization

		eventCmd = AppInstEventsQuery(&in)
		checkpointCmd = AppInstEventsQuery(&in) // TODO change this to checkpoint when we write one

		// Check the developer against who is logged in
		if err := authorized(ctx, rc.claims.Username, org, ResourceAppAnalytics, ActionView); err != nil {
			return setReply(c, err, nil)
		}

		eventResp, checkResp, err := GetEventAndCheckpoint(ctx, rc, eventCmd, checkpointCmd)
		usage, err = GetClusterUsage(eventResp, checkResp, in.StartTime, in.EndTime, in.Region) // TODO: change this to app when we write one
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(c.Path(), "usage/cluster") {
		in := ormapi.RegionClusterInstUsage{}
		success, err := ReadConn(c, &in)
		if !success {
			return err
		}
		// Developer org name has to be specified
		if in.ClusterInst.Organization == "" {
			return setReply(c, fmt.Errorf("Cluster details must be present"), nil)
		}
		rc.region = in.Region
		org = in.ClusterInst.Organization

		eventCmd = ClusterUsageEventsQuery(&in)
		checkpointCmd = ClusterCheckpointsQuery(&in)

		// Check the developer org against who is logged in
		if err := authorized(ctx, rc.claims.Username, org, ResourceClusterAnalytics, ActionView); err != nil {
			return err
		}

		eventResp, checkResp, err := GetEventAndCheckpoint(ctx, rc, eventCmd, checkpointCmd)
		if err != nil {
			return err
		}
		usage, err = GetClusterUsage(eventResp, checkResp, in.StartTime, in.EndTime, in.Region)
		if err != nil {
			return err
		}
	} else {
		return setReply(c, echo.ErrNotFound, nil)
	}

	// calculate usage
	payload := ormapi.StreamPayload{}
	payload.Data = &usage.Data
	WriteStream(c, &payload)

	return nil
}
