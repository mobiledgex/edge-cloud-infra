package orm

import (
	"bytes"
	"context"
	fmt "fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var devInfluxClientMetricsDBTemplate *template.Template
var operatorInfluxClientMetricsDBTemplate *template.Template

type influxClientMetricsQueryArgs struct {
	// Query args
	Selector     string
	Measurement  string
	AppInstName  string
	AppVersion   string
	ClusterName  string
	CloudletName string
	CloudletList string
	OrgField     string
	ApiCallerOrg string
	CloudletOrg  string
	ClusterOrg   string
	AppOrg       string
	StartTime    string
	EndTime      string
	Last         int
	// ClientApi metric query args
	Method string
	CellId string
	// ClientAppUsage and ClientCloudletUsage metric query args
	DeviceCarrier   string
	DataNetworkType string
	DeviceModel     string
	DeviceOs        string
	LocationTile    string
}

// ClientApiUsageFields is DME metrics
var ClientApiUsageFields = []string{
	"\"apporg\"",
	"\"app\"",
	"\"ver\"",
	"\"cloudletorg\"",
	"\"cloudlet\"",
}

var ApiFields = []string{
	"\"dmeId\"",
	"\"cellID\"",
	"\"method\"",
	"\"foundCloudlet\"",
	"\"foundOperator\"",
	"\"reqs\"",
	"\"errs\"",
	"\"0s\"",
	"\"5ms\"",
	"\"10ms\"",
	"\"25ms\"",
	"\"50ms\"",
	"\"100ms\"",
}

var ClientAppUsageFields = []string{
	"\"app\"",
	"\"apporg\"",
	"\"ver\"",
	"\"cluster\"",
	"\"clusterorg\"",
	"\"cloudlet\"",
	"\"cloudletorg\"",
}

var ClientCloudletUsageFields = []string{
	"\"cloudlet\"",
	"\"cloudletorg\"",
}

var LatencyFields = []string{
	//"\"0ms\"", // TODO: make sure this is ms
	"\"0s\"",
	"\"5ms\"",
	"\"10ms\"",
	"\"25ms\"",
	"\"50ms\"",
	"\"100ms\"",
	"\"max\"",
	"\"min\"",
	"\"avg\"",
	"\"variance\"",
	"\"stddev\"",
	"\"numsamples\"",
}

var ClientAppUsageLatencyFields = []string{
	"\"locationtile\"",
}

var ClientCloudletUsageLatencyFields = []string{
	"\"locationtile\"",
	"\"devicecarrier\"",
	"\"datanetworktype\"",
}

var DeviceInfoFields = []string{
	"\"deviceos\"",
	"\"devicemodel\"",
	"\"devicecarrier\"",
	"\"numsessions\"",
}

var ClientAppUsageDeviceInfoFields = []string{
	"\"datanetworktype\"",
}

var ClientCloudletUsageDeviceInfoFields = []string{
	"\"locationtile\"",
}

const (
	CLIENT_APIUSAGE      = "dme"
	CLIENT_APPUSAGE      = "clientappusage"
	CLIENT_CLOUDLETUSAGE = "clientcloudletusage"
)

var devInfluxClientMetricsDBT = `SELECT {{.Selector}} from /{{.Measurement}}/` +
	` WHERE "{{.OrgField}}"='{{.ApiCallerOrg}}'` +
	`{{if .AppInstName}} AND "app"='{{.AppInstName}}'{{end}}` +
	`{{if .AppOrg}} AND "apporg"='{{.AppOrg}}'{{end}}` +
	`{{if .ClusterName}} AND "cluster"='{{.ClusterName}}'{{end}}` +
	`{{if .AppVersion}} AND "ver"='{{.AppVersion}}'{{end}}` +
	`{{if .CloudletName}} AND "cloudlet"='{{.CloudletName}}'{{end}}` +
	`{{if .CloudletList}} AND ({{.CloudletList}}){{end}}` +
	`{{if .CloudletOrg}} AND "cloudletorg"='{{.CloudletOrg}}'{{end}}` +
	`{{if .Method}} AND "method"='{{.Method}}'{{end}}` +
	`{{if .CellId}} AND "cellID"='{{.CellId}}'{{end}}` +
	`{{if .DeviceCarrier}} AND "devicecarrier"='{{.DeviceCarrier}}'{{end}}` +
	`{{if .DataNetworkType}} AND "datanetworktype"='{{.DataNetworkType}}'{{end}}` +
	`{{if .DeviceOs}} AND "deviceos"='{{.DeviceOs}}'{{end}}` +
	`{{if .DeviceModel}} AND "devicemodel"='{{.DeviceModel}}'{{end}}` +
	`{{if .LocationTile}} AND "locationtile"='{{.LocationTile}}'{{end}}` +
	`{{if .StartTime}} AND time >= '{{.StartTime}}'{{end}}` +
	`{{if .EndTime}} AND time <= '{{.EndTime}}'{{end}}` +
	` order by time desc{{if ne .Last 0}} limit {{.Last}}{{end}}`

var operatorInfluxClientMetricsDBT = `SELECT {{.Selector}} from /{{.Measurement}}/` +
	` WHERE "cloudletorg"='{{.CloudletOrg}}'` +
	`{{if .CloudletName}} AND "cloudlet"='{{.CloudletName}}'{{end}}` +
	`{{if .DeviceCarrier}} AND "devicecarrier"='{{.DeviceCarrier}}'{{end}}` +
	`{{if .DataNetworkType}} AND "datanetworktype"='{{.DataNetworkType}}'{{end}}` +
	`{{if .DeviceOs}} AND "deviceos"='{{.DeviceOs}}'{{end}}` +
	`{{if .DeviceModel}} AND "devicemodel"='{{.DeviceModel}}'{{end}}` +
	`{{if .LocationTile}} AND "locationtile"='{{.LocationTile}}'{{end}}` +
	`{{if .StartTime}} AND time >= '{{.StartTime}}'{{end}}` +
	`{{if .EndTime}} AND time <= '{{.EndTime}}'{{end}}` +
	` order by time desc{{if ne .Last 0}} limit {{.Last}}{{end}}`

func init() {
	devInfluxClientMetricsDBTemplate = template.Must(template.New("influxquery").Parse(devInfluxClientMetricsDBT))
	operatorInfluxClientMetricsDBTemplate = template.Must(template.New("influxquery").Parse(operatorInfluxClientMetricsDBT))
}

func fillTimeAndGetCmdForClientMetricsQuery(q *influxClientMetricsQueryArgs, tmpl *template.Template, start *time.Time, end *time.Time) string {
	// Figure out the start/end time range for the query
	if !start.IsZero() {
		buf, err := start.MarshalText()
		if err == nil {
			q.StartTime = string(buf)
		}
	}
	if !end.IsZero() {
		buf, err := end.MarshalText()
		if err == nil {
			q.EndTime = string(buf)
		}
	}
	// We set max number of responses we will get from InfluxDB
	if q.Last == 0 {
		q.Last = maxEntriesFromInfluxDb
	}
	// now that we know all the details of the query - build it
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, q); err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to run template", "tmpl", tmpl, "args", q, "error", err)
		return ""
	}
	return buf.String()
}

func ClientApiUsageMetricsQuery(obj *ormapi.RegionClientApiUsageMetrics, cloudletList []string) string {
	arg := influxClientMetricsQueryArgs{
		Selector:     getFields(obj.Selector, CLIENT_APIUSAGE),
		Measurement:  getMeasurementString(obj.Selector, CLIENT_APIUSAGE),
		AppInstName:  obj.AppInst.AppKey.Name,
		AppVersion:   obj.AppInst.AppKey.Version,
		ApiCallerOrg: obj.AppInst.AppKey.Organization,
		ClusterOrg:   obj.AppInst.ClusterInstKey.Organization,
		CloudletList: generateCloudletList(cloudletList),
		ClusterName:  obj.AppInst.ClusterInstKey.ClusterKey.Name,
		Method:       obj.Method,
		Last:         obj.Last,
	}
	if obj.AppInst.AppKey.Organization != "" {
		arg.OrgField = "apporg"
		arg.ApiCallerOrg = obj.AppInst.AppKey.Organization
		arg.CloudletOrg = obj.AppInst.ClusterInstKey.CloudletKey.Organization
	} else {
		arg.OrgField = "cloudletorg"
		arg.ApiCallerOrg = obj.AppInst.ClusterInstKey.CloudletKey.Organization
		arg.AppOrg = obj.AppInst.AppKey.Organization
	}
	if obj.CellId != 0 {
		arg.CellId = strconv.FormatUint(uint64(obj.CellId), 10)
	}
	return fillTimeAndGetCmdForClientMetricsQuery(&arg, devInfluxClientMetricsDBTemplate, &obj.StartTime, &obj.EndTime)
}

func ClientAppUsageMetricsQuery(ctx context.Context, idc *InfluxDBContext, obj *ormapi.RegionClientAppUsageMetrics, cloudletList []string) (cmd string, db string) {
	basemeasurement := ""
	switch obj.Selector {
	case "latency":
		basemeasurement = cloudcommon.LatencyMetric
	case "deviceinfo":
		basemeasurement = cloudcommon.DeviceMetric
	}
	measurement := basemeasurement
	// Get time definition for specified start and end times
	definition := getTimeDefinitionDuration(obj, 0)
	if definition != 0 {
		measurement = getClientMetricsMeasurementString(ctx, idc, basemeasurement, definition)
	}
	arg := influxClientMetricsQueryArgs{
		Selector:        getFields(obj.Selector, CLIENT_APPUSAGE),
		Measurement:     measurement,
		AppInstName:     obj.AppInst.AppKey.Name,
		AppVersion:      obj.AppInst.AppKey.Version,
		ApiCallerOrg:    obj.AppInst.AppKey.Organization,
		ClusterOrg:      obj.AppInst.ClusterInstKey.Organization,
		CloudletList:    generateCloudletList(cloudletList),
		ClusterName:     obj.AppInst.ClusterInstKey.ClusterKey.Name,
		DeviceCarrier:   obj.DeviceCarrier,
		DataNetworkType: obj.DataNetworkType,
		DeviceOs:        obj.DeviceOs,
		DeviceModel:     obj.DeviceModel,
		LocationTile:    obj.LocationTile,
		Last:            obj.Last,
	}
	if obj.AppInst.AppKey.Organization != "" {
		arg.OrgField = "apporg"
		arg.ApiCallerOrg = obj.AppInst.AppKey.Organization
		arg.CloudletOrg = obj.AppInst.ClusterInstKey.CloudletKey.Organization
	} else {
		arg.OrgField = "cloudletorg"
		arg.ApiCallerOrg = obj.AppInst.ClusterInstKey.CloudletKey.Organization
		arg.AppOrg = obj.AppInst.AppKey.Organization
	}

	if measurement == basemeasurement {
		db = cloudcommon.EdgeEventsMetricsDbName
	} else {
		db = cloudcommon.DownsampledMetricsDbName
	}
	return fillTimeAndGetCmdForClientMetricsQuery(&arg, devInfluxClientMetricsDBTemplate, &obj.StartTime, &obj.EndTime), db
}

func ClientCloudletUsageMetricsQuery(ctx context.Context, idc *InfluxDBContext, obj *ormapi.RegionClientCloudletUsageMetrics) (cmd string, db string) {
	basemeasurement := ""
	switch obj.Selector {
	case "latency":
		basemeasurement = cloudcommon.LatencyMetric
	case "deviceinfo":
		basemeasurement = cloudcommon.DeviceMetric
	}
	measurement := basemeasurement
	// Get time definition for specified start and end times
	definition := getTimeDefinitionDuration(obj, 0)
	if definition != 0 {
		measurement = getClientMetricsMeasurementString(ctx, idc, basemeasurement, definition)
	}
	arg := influxClientMetricsQueryArgs{
		Selector:        getFields(obj.Selector, CLIENT_CLOUDLETUSAGE),
		Measurement:     measurement,
		CloudletName:    obj.Cloudlet.Name,
		CloudletOrg:     obj.Cloudlet.Organization,
		DeviceCarrier:   obj.DeviceCarrier,
		DataNetworkType: obj.DataNetworkType,
		DeviceOs:        obj.DeviceOs,
		DeviceModel:     obj.DeviceModel,
		LocationTile:    obj.LocationTile,
		Last:            obj.Last,
	}

	if measurement == basemeasurement {
		db = cloudcommon.EdgeEventsMetricsDbName
	} else {
		db = cloudcommon.DownsampledMetricsDbName
	}
	return fillTimeAndGetCmdForClientMetricsQuery(&arg, operatorInfluxClientMetricsDBTemplate, &obj.StartTime, &obj.EndTime), db
}

/*
 * Get the correct measurement string for already downsampled data for specified start and end time
 * For example, if the duration between Start and End times is 1.5 hr and we have continuous queries that aggregate
 * hourly, daily, and weekly, this function will return the hourly measurement.
 * If the duration is 25 hours, this function will return the daily measurement
 */
func getClientMetricsMeasurementString(ctx context.Context, idc *InfluxDBContext, baseMeasurement string, definition time.Duration) string {
	// Grab settings for specified region
	in := &edgeproto.Settings{}
	rc := &RegionContext{
		username: idc.claims.Username,
		region:   idc.region,
	}
	settings, err := ShowSettingsObj(ctx, rc, in)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable to get continuous query settings - using raw data", "error", err)
		return baseMeasurement
	}

	// Find Continuous Query interval that is closest to definition but less than (ie. finer granularity) than definition
	var optimalInterval time.Duration = 0
	var minDiff time.Duration = 0
	for _, cqs := range settings.EdgeEventsMetricsContinuousQueriesCollectionIntervals {
		diff := time.Duration(cqs.Interval) - definition
		if diff < 0 {
			absMin := diff * -1
			if absMin < minDiff || minDiff == 0 {
				minDiff = absMin
				optimalInterval = time.Duration(cqs.Interval)
			}
		}
	}

	if optimalInterval == 0 {
		log.SpanLog(ctx, log.DebugLevelMetrics, "Unable find interval with finer granularity than time definition - using raw data", "definition", definition)
		return baseMeasurement
	}
	return cloudcommon.CreateInfluxMeasurementName(baseMeasurement, optimalInterval)
}

// TODO: HANDLE selector == "*"
// Make sure correct optional fields are provided for ClientAppUsage
// eg. DeviceOS is not allowed for latency selector/metric
func validateClientAppUsageMetricReq(req *ormapi.RegionClientAppUsageMetrics, selector string) error {
	switch selector {
	case "latency":
		if req.DeviceOs != "" {
			return fmt.Errorf("DeviceOS not allowed for appinst latency metric")
		}
		if req.DeviceModel != "" {
			return fmt.Errorf("DeviceModel not allowed for appinst latency metric")
		}
		if req.DeviceCarrier != "" {
			return fmt.Errorf("DeviceCarrier not allowed for appinst latency metric")
		}
		if req.DataNetworkType != "" {
			return fmt.Errorf("DataNetworkType not allowed for appinst latency metric")
		}
	case "deviceinfo":
		if req.LocationTile != "" {
			return fmt.Errorf("LocationTile not allowed for appinst deviceinfo metric")
		}
	case "custom":
		return fmt.Errorf("Custom stat not implemented yet")
	default:
		return fmt.Errorf("Provided selector \"%s\" is not valid. Must provide only one of \"%s\"", selector, strings.Join(ormapi.ClientAppUsageSelectors, "\", \""))
	}
	return nil
}

// Make sure correct optional fields are provided for ClientCloudletUsage
// eg. DeviceOS is not allowed for latency selector/metric
func validateClientCloudletUsageMetricReq(req *ormapi.RegionClientCloudletUsageMetrics, selector string) error {
	switch selector {
	case "latency":
		if req.DeviceOs != "" {
			return fmt.Errorf("DeviceOS not allowed for cloudlet latency metric")
		}
		if req.DeviceModel != "" {
			return fmt.Errorf("DeviceModel not allowed for cloudlet latency metric")
		}
	case "deviceinfo":
		if req.DataNetworkType != "" {
			return fmt.Errorf("DataNetworkType not allowed for cloudlet deviceinfo metric")
		}
	default:
		return fmt.Errorf("Provided selector \"%s\" is not valid. Must provide only one of \"%s\"", selector, strings.Join(ormapi.ClientCloudletUsageSelectors, "\", \""))
	}
	return nil
}
