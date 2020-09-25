package orm

import (
	"context"
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
		OrgField:     "apporg",
		ApiCallerOrg: obj.AppInst.AppKey.Organization,
		CloudletName: obj.AppInst.ClusterInstKey.CloudletKey.Name,
		ClusterName:  obj.AppInst.ClusterInstKey.ClusterKey.Name,
		CloudletOrg:  obj.AppInst.ClusterInstKey.CloudletKey.Organization,
		Last:         obj.Last,
	}
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Query is a template with a specific set of if/else
func ClusterEventsQuery(obj *ormapi.RegionClusterInstEvents) string {
	arg := influxQueryArgs{
		Selector:     getEventFields(EVENT_CLUSTERINST),
		Measurement:  EVENT_CLUSTERINST,
		OrgField:     "clusterorg",
		ApiCallerOrg: obj.ClusterInst.Organization,
		CloudletName: obj.ClusterInst.CloudletKey.Name,
		ClusterName:  obj.ClusterInst.ClusterKey.Name,
		CloudletOrg:  obj.ClusterInst.CloudletKey.Organization,
		Last:         obj.Last,
	}
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Query is a template with a specific set of if/else
func CloudletEventsQuery(obj *ormapi.RegionCloudletEvents) string {
	arg := influxQueryArgs{
		Selector:     getEventFields(EVENT_CLOUDLET),
		Measurement:  EVENT_CLOUDLET,
		OrgField:     "cloudletorg",
		ApiCallerOrg: obj.Cloudlet.Organization,
		CloudletName: obj.Cloudlet.Name,
		CloudletOrg:  obj.Cloudlet.Organization,
		Last:         obj.Last,
	}
	return fillTimeAndGetCmd(&arg, operatorInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

func GetAppEventsCmd(ctx context.Context, c echo.Context, rc *InfluxDBContext) (string, error) {
	in := ormapi.RegionAppInstEvents{}
	success, err := ReadConn(c, &in)
	if !success {
		return "", err
	}
	// Developer name has to be specified
	if in.AppInst.AppKey.Organization == "" {
		return "", setReply(c, fmt.Errorf("App details must be present"), nil)
	}
	rc.region = in.Region
	org := in.AppInst.AppKey.Organization

	cmd := AppInstEventsQuery(&in)

	// Check the developer against who is logged in
	if err := authorized(ctx, rc.claims.Username, org, ResourceAppAnalytics, ActionView); err != nil {
		return "", setReply(c, err, nil)
	}
	return cmd, nil
}

func GetClusterEventsCmd(ctx context.Context, c echo.Context, rc *InfluxDBContext) (string, error) {
	in := ormapi.RegionClusterInstEvents{}
	success, err := ReadConn(c, &in)
	if !success {
		return "", err
	}
	// Developer org name has to be specified
	if in.ClusterInst.Organization == "" {
		return "", setReply(c, fmt.Errorf("Cluster details must be present"), nil)
	}
	rc.region = in.Region
	org := in.ClusterInst.Organization

	cmd := ClusterEventsQuery(&in)

	// Check the developer org against who is logged in
	if err := authorized(ctx, rc.claims.Username, org, ResourceClusterAnalytics, ActionView); err != nil {
		return "", err
	}
	return cmd, nil
}

func GetCloudletEventsCmd(ctx context.Context, c echo.Context, rc *InfluxDBContext) (string, error) {
	in := ormapi.RegionCloudletEvents{}
	success, err := ReadConn(c, &in)
	if !success {
		return "", err
	}
	// Operator name has to be specified
	if in.Cloudlet.Organization == "" {
		return "", setReply(c, fmt.Errorf("Cloudlet details must be present"), nil)
	}
	rc.region = in.Region
	org := in.Cloudlet.Organization

	cmd := CloudletEventsQuery(&in)

	// Check the operator against who is logged in
	if err := authorized(ctx, rc.claims.Username, org, ResourceCloudletAnalytics, ActionView); err != nil {
		return "", setReply(c, err, nil)
	}
	return cmd, nil
}

// Common method to handle both app and cluster metrics
func GetEventsCommon(c echo.Context) error {
	var cmd string

	rc := &InfluxDBContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims
	ctx := GetContext(c)

	if strings.HasSuffix(c.Path(), "events/app") {
		cmd, err = GetAppEventsCmd(ctx, c, rc)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(c.Path(), "events/cluster") {
		cmd, err = GetClusterEventsCmd(ctx, c, rc)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(c.Path(), "events/cloudlet") {
		cmd, err = GetCloudletEventsCmd(ctx, c, rc)
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
