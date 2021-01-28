package orm

import (
	"fmt"
	"text/template"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/util"
)

var appInstGroupQueryTemplate *template.Template

// select mean(cpu) from \"appinst-cpu\" where (apporg='DevOrg') and time >=now() -20m group by time(2m), app fill(previous)"
var AppInstGroupQueryT = `SELECT {{.Selector}} FROM "{{.Measurement}}"` +
	` WHERE ({{.QueryFilter}})` +
	`{{if .StartTime}} AND time >= '{{.StartTime}}'{{end}}` +
	`{{if .EndTime}} AND time <= '{{.EndTime}}'{{end}}` +
	` group by time({{.TimeDefinition}}), app fill(previous)`

func init() {
	appInstGroupQueryTemplate = template.Must(template.New("influxquery").Parse(AppInstGroupQueryT))
}

// Common method to handle both app and cluster metrics
func GetAppMetrics(c echo.Context) error {
	var cmd, org string

	rc := &InfluxDBContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims
	ctx := GetContext(c)
	// Get the current config
	config, err := getConfig(ctx)
	if err == nil {
		maxEntriesFromInfluxDb = config.MaxMetricsDataPoints
	}
	in := ormapi.RegionAppInstMetricsV2{}
	success, err := ReadConn(c, &in)
	if !success {
		return err
	}
	// At least on AppInst org has to be specified
	if len(in.AppInsts) == 0 {
		return setReply(c, fmt.Errorf("At least one app org has to be specified"), nil)
	}
	rc.region = in.Region
	// TODO - don't allow *
	if err = validateSelectorString(in.Selector, APPINST); err != nil {
		return setReply(c, err, nil)
	}
	for _, app := range in.AppInsts {
		org = app.AppKey.Organization
		// Developer name has to be specified
		if org == "" {
			return setReply(c, fmt.Errorf("App org must be present details must be present"), nil)
		}
		// Check the developer against who is logged in
		if err := authorized(ctx, rc.claims.Username, org, ResourceAppAnalytics, ActionView); err != nil {
			return setReply(c, err, nil)
		}

	}
	cmd = GetAppInstsGroupQuery(&in)

	err = influxStream(ctx, rc, cloudcommon.DeveloperMetricsDbName, cmd, func(res interface{}) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		WriteStream(c, &payload)
	})
	if err != nil {
		return WriteError(c, err)
	}
	return nil
}

func GetAppInstsGroupQuery(apps *ormapi.RegionAppInstMetricsV2) string {
	args := influxQueryArgs{
		Selector:       getSelectorForMeasurement(apps.Selector, apps.Function),
		Measurement:    getMeasurementString(apps.Selector, APPINST),
		QueryFilter:    getAppInstQueryFilter(apps),
		TimeDefinition: "10s", // TODO - calculate this
	}
	return fillTimeAndGetCmd(&args, appInstGroupQueryTemplate, &apps.StartTime, &apps.EndTime)
}

// Combine appInst definitions into a filter string in influxDB
// Example: app1/v1.0/appOrg1/cluster1/cloudlet1,app2/v1.1/appOrg2/cluster2/cloudlet1
// string: ("apporg"='appOrg1' AND "app"='app1' AND "ver"='v10' AND "cluster"='cluster1' AND "cloudlet"='cloudlet1') OR
//           ("apporg"='appOrg2' AND "app"='app2' AND "ver"='v11' AND "cluster"='cluster2' AND "cloudlet"='cloudlet1')
func getAppInstQueryFilter(apps *ormapi.RegionAppInstMetricsV2) string {
	filterStr := ``
	for ii, app := range apps.AppInsts {
		filterStr += `("apporg"='` + app.AppKey.Organization + `'`
		if app.AppKey.Name != "" {
			filterStr += ` AND "app"='` + util.DNSSanitize(app.AppKey.Name) + `'`
		}
		if app.AppKey.Version != "" {
			filterStr += ` AND "ver"='` + util.DNSSanitize(app.AppKey.Version) + `'`
		}
		if app.ClusterInstKey.Organization != "" {
			filterStr += ` AND "clusterorg"='` + app.ClusterInstKey.Organization + `'`
		}
		if app.ClusterInstKey.ClusterKey.Name != "" {
			filterStr += ` AND "cluster"='` + app.ClusterInstKey.ClusterKey.Name + `'`
		}
		if app.ClusterInstKey.CloudletKey.Name != "" {
			filterStr += ` AND "cloudlet"='` + app.ClusterInstKey.CloudletKey.Name + `'`
		}
		if app.ClusterInstKey.CloudletKey.Organization != "" {
			filterStr += ` AND "cloudletorg"='` + app.ClusterInstKey.CloudletKey.Organization + `'`
		}

		filterStr += `)`
		// last element
		if len(apps.AppInsts) != ii+1 {
			filterStr += ` OR `
		}
	}
	return filterStr
}

func getSelectorForMeasurement(selector, function string) string {
	switch selector {
	case "cpu":
		return "mean(cpu)"
	case "mem":
		return "max(mem)"
	default:
		return "error"
	}
	//TODO - other than cpu/mem
}
