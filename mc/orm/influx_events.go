package orm

import (
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
)

var FlavorFields = []string{
	"flavor",
	"vcpu",
	"ram",
	"disk",
	"other",
}

var eventLogSelectors = []string{
	"event",
	"status",
}

const (
	EVENT_APPINST     = "appinst"
	EVENT_CLUSTERINST = "clusterinst"
	EVENT_CLOUDLET    = "cloudlet"
)

func getEventFields(eventType string) string {
	var selectors []string
	switch eventType {
	case EVENT_APPINST:
		selectors = AppFields
	case EVENT_CLUSTERINST:
		selectors = append(ClusterFields, FlavorFields...)
	case EVENT_CLOUDLET:
		selectors = CloudletFields
	default:
		return "*"
	}
	return strings.Join(append(selectors, eventLogSelectors...), ",")
}

// Query is a template with a specific set of if/else
func AppInstEventsQuery(obj *ormapi.RegionAppInstEvents) string {
	arg := influxQueryArgs{
		Selector:     getEventFields(EVENT_APPINST),
		Measurement:  EVENT_APPINST,
		AppInstName:  k8smgmt.NormalizeName(obj.AppInst.AppKey.Name),
		DeveloperOrg: obj.AppInst.AppKey.Organization,
		CloudletName: obj.AppInst.ClusterInstKey.CloudletKey.Name,
		ClusterName:  obj.AppInst.ClusterInstKey.ClusterKey.Name,
		OperatorOrg:  obj.AppInst.ClusterInstKey.CloudletKey.Organization,
		Last:         obj.Last,
	}
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Query is a template with a specific set of if/else
func ClusterEventsQuery(obj *ormapi.RegionClusterInstEvents) string {
	arg := influxQueryArgs{
		Selector:     getEventFields(EVENT_CLUSTERINST),
		Measurement:  EVENT_CLUSTERINST,
		CloudletName: obj.ClusterInst.CloudletKey.Name,
		ClusterName:  obj.ClusterInst.ClusterKey.Name,
		OperatorOrg:  obj.ClusterInst.CloudletKey.Organization,
		Last:         obj.Last,
	}
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Query is a template with a specific set of if/else
func CloudletEventsQuery(obj *ormapi.RegionCloudletEvents) string {
	arg := influxQueryArgs{
		Selector:     getEventFields(EVENT_CLOUDLET),
		Measurement:  EVENT_CLOUDLET,
		CloudletName: obj.Cloudlet.Name,
		OperatorOrg:  obj.Cloudlet.Organization,
		Last:         obj.Last,
	}
	return fillTimeAndGetCmd(&arg, operatorInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Common method to handle both app and cluster metrics
func GetEventsCommon(c echo.Context) error {
	var cmd, org string

	rc := &InfluxDBContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims
	ctx := GetContext(c)

	if strings.HasSuffix(c.Path(), "events/app") {
		in := ormapi.RegionAppInstEvents{}
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

		cmd = AppInstEventsQuery(&in)

		// Check the developer against who is logged in
		if !authorized(ctx, rc.claims.Username, org, ResourceAppAnalytics, ActionView) {
			return setReply(c, echo.ErrForbidden, nil)
		}
	} else if strings.HasSuffix(c.Path(), "events/cluster") {
		in := ormapi.RegionClusterInstEvents{}
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

		cmd = ClusterEventsQuery(&in)

		// Check the developer org against who is logged in
		if !authorized(ctx, rc.claims.Username, org, ResourceClusterAnalytics, ActionView) {
			return echo.ErrForbidden
		}
	} else if strings.HasSuffix(c.Path(), "events/cloudlet") {
		in := ormapi.RegionCloudletEvents{}
		success, err := ReadConn(c, &in)
		if !success {
			return err
		}
		// Operator name has to be specified
		if in.Cloudlet.Organization == "" {
			return setReply(c, fmt.Errorf("Cloudlet details must be present"), nil)
		}
		rc.region = in.Region
		org = in.Cloudlet.Organization

		cmd = CloudletEventsQuery(&in)

		// Check the operator against who is logged in
		if !authorized(ctx, rc.claims.Username, org, ResourceCloudletAnalytics, ActionView) {
			return setReply(c, echo.ErrForbidden, nil)
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
