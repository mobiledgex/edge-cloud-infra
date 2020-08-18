package orm

import (
	"fmt"
	"strings"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func GetClusterUsage(eventRecords *client.Response) (*ormapi.AllUsage, error) {
	usageRecords := ormapi.AllUsage{
		Data: make([]ormapi.UsageRecord, 0),
	}
	clusterTracker := make([]*edgeproto.ClusterInst, 0)

	// check to see if the influx output is empty or invalid
	if len(eventRecords.Results) == 0 || len(eventRecords.Results[0].Series) == 0 {
		// empty, no event logs at all
		return &usageRecords, nil
	} else if len(eventRecords.Results) != 1 ||
		len(eventRecords.Results[0].Series) != 1 ||
		len(eventRecords.Results[0].Series[0].Values) == 0 ||
		len(eventRecords.Results[0].Series[0].Values[0]) == 0 ||
		eventRecords.Results[0].Series[0].Name != EVENT_CLUSTERINST {
		// should only be 1 series, the 'dbName' one
		return nil, fmt.Errorf("Error parsing influx, unexpected format")
	}

	for _, values := range eventRecords.Results[0].Series[0].Values {
		// value should be of the format [timestamp cluster clusterorg cloudlet cloudletorg flavor vcpu ram disk other event status]
		// TODO: fix flavors in cluster events log to be a flavor list, and maybe get rid of vcpu,ram,disk,other fields
		if len(values) != 12 {
			return nil, fmt.Errorf("Error parsing influx response")
		}
		timestamp, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", values[0]))
		if err != nil {
			return nil, fmt.Errorf("Unable to parse timestamp: %v", err)
		}
	}

	return nil, nil
}

// Common method to handle both app and cluster metrics
func GetUsageCommon(c echo.Context) error {
	var cmd string

	rc := &InfluxDBContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims
	ctx := GetContext(c)

	if strings.HasSuffix(c.Path(), "usage/app") {
		cmd, err = GetAppEventsCmd(ctx, c, rc)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(c.Path(), "usage/cluster") {
		cmd, err = GetClusterEventsCmd(ctx, c, rc)
		if err != nil {
			return err
		}
	} else {
		return setReply(c, echo.ErrNotFound, nil)
	}

	err = influxStream(ctx, rc, cloudcommon.EventsDbName, cmd, func(res interface{}) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		return WriteError(c, err)
	}
	return nil
}
