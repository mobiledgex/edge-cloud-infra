package orm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"google.golang.org/grpc/status"
)

var devInfluxDBTemplate *template.Template
var operatorInfluxDBTemplate *template.Template

// 100 values at a time
var queryChunkSize = 100

type InfluxDBContext struct {
	region string
	claims *UserClaims
	conn   influxdb.Client
}

type influxQueryArgs struct {
	Selector      string
	Measurement   string
	AppInstName   string
	ClusterName   string
	DeveloperName string
	CloudletName  string
	OperatorName  string
	Method        string
	CellId        string
	StartTime     string
	EndTime       string
	Last          int
}

var AppSelectors = []string{
	"cpu",
	"mem",
	"disk",
	"network",
	"connections",
}

var ClusterSelectors = []string{
	"cpu",
	"mem",
	"disk",
	"network",
	"tcp",
	"udp",
}

var CloudletSelectors = []string{
	"network",
	"utilization",
	"ipusage",
}

var ClientSelectors = []string{
	"api",
}

const (
	APPINST  = "appinst"
	CLUSTER  = "cluster"
	CLOUDLET = "cloudlet"
	CLIENT   = "dme"
)

var devInfluDBT = `SELECT {{.Selector}} from "{{.Measurement}}"` +
	` WHERE "dev"='{{.DeveloperName}}'` +
	`{{if .AppInstName}} AND "app"=~/{{.AppInstName}}/{{end}}` +
	`{{if .ClusterName}} AND "cluster"='{{.ClusterName}}'{{end}}` +
	`{{if .CloudletName}} AND "cloudlet"='{{.CloudletName}}'{{end}}` +
	`{{if .OperatorName}} AND "operator"='{{.OperatorName}}'{{end}}` +
	`{{if .Method}} AND "method"='{{.Method}}'{{end}}` +
	`{{if .CellId}} AND "cellID"='{{.CellId}}'{{end}}` +
	`{{if .StartTime}} AND time >= '{{.StartTime}}'{{end}}` +
	`{{if .EndTime}} AND time < '{{.EndTime}}'{{end}}` +
	`order by time desc{{if ne .Last 0}} limit {{.Last}}{{end}}`

var operatorInfluDBT = `SELECT {{.Selector}} from "{{.Measurement}}"` +
	` WHERE "operator"='{{.OperatorName}}'` +
	`{{if .CloudletName}} AND "cloudlet"='{{.CloudletName}}'{{end}}` +
	`{{if .StartTime}} AND time >= '{{.StartTime}}'{{end}}` +
	`{{if .EndTime}} AND time < '{{.EndTime}}'{{end}}` +
	`order by time desc{{if ne .Last 0}} limit {{.Last}}{{end}}`

func init() {
	devInfluxDBTemplate = template.Must(template.New("influxquery").Parse(devInfluDBT))
	operatorInfluxDBTemplate = template.Must(template.New("influxquery").Parse(operatorInfluDBT))
}

func connectInfluxDB(ctx context.Context, region string) (influxdb.Client, error) {
	addr, err := getInfluxDBAddrForRegion(ctx, region)
	if err != nil {
		return nil, err
	}
	creds, err := cloudcommon.GetInfluxDataAuth(serverConfig.vaultConfig, region)
	if err != nil {
		return nil, fmt.Errorf("get influxDB auth failed, %v", err)
	}
	if creds == nil {
		// defeault to empty auth
		creds = &cloudcommon.InfluxCreds{}
	}
	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     addr,
		Username: creds.User,
		Password: creds.Pass,
	})
	log.SpanLog(ctx, log.DebugLevelMetrics, "connecting to influxdb",
		"addr", addr, "err", err)
	if err != nil {
		return nil, err
	}
	return client, nil

}

func getInfluxDBAddrForRegion(ctx context.Context, region string) (string, error) {
	ctrl, err := getControllerObj(ctx, region)
	if err != nil {
		return "", err
	}
	return ctrl.InfluxDB, nil
}

func fillTimeAndGetCmd(q *influxQueryArgs, tmpl *template.Template, start *time.Time, end *time.Time) string {
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
	// now that we know all the details of the query - build it
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, q); err != nil {
		log.DebugLog(log.DebugLevelApi, "Failed to run template", "tmpl", tmpl, "args", q, "error", err)
		return ""
	}
	return buf.String()
}

func ClientMetricsQuery(obj *ormapi.RegionClientMetrics) string {
	arg := influxQueryArgs{
		Selector:      "*",
		Measurement:   getMeasurementString(obj.Selector, CLIENT),
		AppInstName:   k8smgmt.NormalizeName(obj.AppInst.AppKey.Name),
		DeveloperName: obj.AppInst.AppKey.DeveloperKey.Name,
		CloudletName:  obj.AppInst.ClusterInstKey.CloudletKey.Name,
		ClusterName:   obj.AppInst.ClusterInstKey.ClusterKey.Name,
		OperatorName:  obj.AppInst.ClusterInstKey.CloudletKey.OperatorKey.Name,
		Method:        obj.Method,
		Last:          obj.Last,
	}
	if obj.CellId != 0 {
		arg.CellId = strconv.FormatUint(uint64(obj.CellId), 10)
	}
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Query is a template with a specific set of if/else
func AppInstMetricsQuery(obj *ormapi.RegionAppInstMetrics) string {
	arg := influxQueryArgs{
		Selector:      "*",
		Measurement:   getMeasurementString(obj.Selector, APPINST),
		AppInstName:   k8smgmt.NormalizeName(obj.AppInst.AppKey.Name),
		DeveloperName: obj.AppInst.AppKey.DeveloperKey.Name,
		CloudletName:  obj.AppInst.ClusterInstKey.CloudletKey.Name,
		ClusterName:   obj.AppInst.ClusterInstKey.ClusterKey.Name,
		OperatorName:  obj.AppInst.ClusterInstKey.CloudletKey.OperatorKey.Name,
		Last:          obj.Last,
	}
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Query is a template with a specific set of if/else
func ClusterMetricsQuery(obj *ormapi.RegionClusterInstMetrics) string {
	arg := influxQueryArgs{
		Selector:      "*",
		Measurement:   getMeasurementString(obj.Selector, CLUSTER),
		CloudletName:  obj.ClusterInst.CloudletKey.Name,
		ClusterName:   obj.ClusterInst.ClusterKey.Name,
		DeveloperName: obj.ClusterInst.Developer,
		OperatorName:  obj.ClusterInst.CloudletKey.OperatorKey.Name,
		Last:          obj.Last,
	}
	return fillTimeAndGetCmd(&arg, devInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// Query is a template with a specific set of if/else
func CloudletMetricsQuery(obj *ormapi.RegionCloudletMetrics) string {
	arg := influxQueryArgs{
		Selector:     "*",
		Measurement:  getMeasurementString(obj.Selector, CLOUDLET),
		CloudletName: obj.Cloudlet.Name,
		OperatorName: obj.Cloudlet.OperatorKey.Name,
		Last:         obj.Last,
	}
	return fillTimeAndGetCmd(&arg, operatorInfluxDBTemplate, &obj.StartTime, &obj.EndTime)
}

// TODO: This function should be a streaming fucntion, but currently client library for influxDB
// doesn't implement it in a way could really be using it
func metricsStream(ctx context.Context, rc *InfluxDBContext, dbQuery string, cb func(Data interface{})) error {
	if rc.conn == nil {
		conn, err := connectInfluxDB(ctx, rc.region)
		if err != nil {
			return err
		}
		rc.conn = conn
		defer func() {
			rc.conn.Close()
			rc.conn = nil
		}()
	}

	query := influxdb.Query{
		Command:   dbQuery,
		Database:  cloudcommon.DeveloperMetricsDbName,
		Chunked:   false, // TODO - add chunking. Client lib doesn't support chunk response processing yet
		ChunkSize: queryChunkSize,
	}
	resp, err := rc.conn.Query(query)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMetrics, "InfluxDB query failed",
			"query", query, "resp", resp, "err", err)
		// We return a different error, as we don't want to expose a URL-encoded query to influxDB
		return fmt.Errorf("Connection to InfluxDB failed")
	}
	if resp.Error() != nil {
		return resp.Error()
	}
	cb(resp.Results)
	return nil
}

func Contains(slice []string, elem string) bool {
	for _, val := range slice {
		if val == elem {
			return true
		}
	}
	return false
}

// Function validates the selector passed, we support several selectors: cpu, mem, disk, net
func validateSelectorString(selector, metricType string) error {
	var validSelectors []string
	switch metricType {
	case APPINST:
		validSelectors = AppSelectors
	case CLUSTER:
		validSelectors = ClusterSelectors
	case CLOUDLET:
		validSelectors = CloudletSelectors
	case CLIENT:
		validSelectors = ClientSelectors
	default:
		return fmt.Errorf("Invalid metric type %s", metricType)
	}
	if selector == "*" {
		return nil
	}
	selectors := strings.Split(selector, ",")
	for _, s := range selectors {
		if !Contains(validSelectors, s) {
			return fmt.Errorf("Invalid %s selector: %s", metricType, s)
		}
	}
	return nil
}

func getMeasurementString(selector, measurementType string) string {
	var measurements []string
	switch measurementType {
	case "appinst":
		measurements = AppSelectors
	case "cluster":
		measurements = ClusterSelectors
	case "cloudlet":
		measurements = CloudletSelectors
	case "client":
		measurements = ClientSelectors
	}
	if selector != "*" {
		measurements = strings.Split(selector, ",")
	}
	prefix := measurementType + "-"
	return prefix + strings.Join(measurements, "\",\""+prefix)
}

// Common method to handle both app and cluster metrics
func GetMetricsCommon(c echo.Context) error {
	var errStr, cmd, org string
	var err error

	rc := &InfluxDBContext{}
	var ws *websocket.Conn
	if strings.HasPrefix(c.Request().URL.Path, "/ws") {
		ws, err = websocketConnect(c)
		if err != nil {
			return err
		}
		if ws == nil {
			return nil
		}
		defer ws.Close()
	}

	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims
	ctx := GetContext(c)

	if strings.HasSuffix(c.Path(), "metrics/app") {
		in := ormapi.RegionAppInstMetrics{}
		if ws != nil {
			err = ws.ReadJSON(&in)
		} else {
			err = c.Bind(&in)
		}
		if err != nil {
			errStr = checkForTimeError(fmt.Sprintf("Invalid data: %s", err.Error()))
			return setReply(c, ws, fmt.Errorf(errStr), nil)
		}
		// Developer name has to be specified
		if in.AppInst.AppKey.DeveloperKey.Name == "" {
			return setReply(c, ws, fmt.Errorf("App details must be present"), nil)
		}
		rc.region = in.Region
		org = in.AppInst.AppKey.DeveloperKey.Name
		if err = validateSelectorString(in.Selector, APPINST); err != nil {
			return setReply(c, ws, err, nil)
		}
		cmd = AppInstMetricsQuery(&in)

		// Check the developer against who is logged in
		if !authorized(ctx, rc.claims.Username, org, ResourceAppAnalytics, ActionView) {
			return setReply(c, ws, echo.ErrForbidden, nil)
		}
	} else if strings.HasSuffix(c.Path(), "metrics/cluster") {
		in := ormapi.RegionClusterInstMetrics{}
		if ws != nil {
			err = ws.ReadJSON(&in)
		} else {
			err = c.Bind(&in)
		}
		if err != nil {
			errStr = checkForTimeError(fmt.Sprintf("Invalid data: %s", err.Error()))
			return setReply(c, ws, fmt.Errorf(errStr), nil)
		}
		// Developer name has to be specified
		if in.ClusterInst.Developer == "" {
			return setReply(c, ws, fmt.Errorf("Cluster details must be present"), nil)
		}
		rc.region = in.Region
		org = in.ClusterInst.Developer
		if err = validateSelectorString(in.Selector, CLUSTER); err != nil {
			return setReply(c, ws, err, nil)
		}
		cmd = ClusterMetricsQuery(&in)

		// Check the developer against who is logged in
		if !authorized(ctx, rc.claims.Username, org, ResourceClusterAnalytics, ActionView) {
			return echo.ErrForbidden
		}
	} else if strings.HasSuffix(c.Path(), "metrics/cloudlet") {
		in := ormapi.RegionCloudletMetrics{}
		if ws != nil {
			err = ws.ReadJSON(&in)
		} else {
			err = c.Bind(&in)
		}
		if err != nil {
			errStr = checkForTimeError(fmt.Sprintf("Invalid data: %s", err.Error()))
			return setReply(c, ws, fmt.Errorf(errStr), nil)
		}
		// Operator name has to be specified
		if in.Cloudlet.OperatorKey.Name == "" {
			return setReply(c, ws, fmt.Errorf("Cloudlet details must be present"), nil)
		}
		rc.region = in.Region
		org = in.Cloudlet.OperatorKey.Name
		if err = validateSelectorString(in.Selector, CLOUDLET); err != nil {
			return setReply(c, ws, err, nil)
		}
		cmd = CloudletMetricsQuery(&in)

		// Check the operator against who is logged in
		if !authorized(ctx, rc.claims.Username, org, ResourceCloudletAnalytics, ActionView) {
			return setReply(c, ws, echo.ErrForbidden, nil)
		}
	} else if strings.HasSuffix(c.Path(), "metrics/client") {
		in := ormapi.RegionClientMetrics{}
		if ws != nil {
			err = ws.ReadJSON(&in)
		} else {
			err = c.Bind(&in)
		}
		if err != nil {
			errStr = checkForTimeError(fmt.Sprintf("Invalid data: %s", err.Error()))
			return setReply(c, ws, fmt.Errorf(errStr), nil)
		}
		// Developer name has to be specified
		if in.AppInst.AppKey.DeveloperKey.Name == "" {
			return setReply(c, ws, fmt.Errorf("App details must be present"), nil)
		}
		rc.region = in.Region
		org = in.AppInst.AppKey.DeveloperKey.Name
		if err = validateSelectorString(in.Selector, CLIENT); err != nil {
			return setReply(c, ws, err, nil)
		}
		cmd = ClientMetricsQuery(&in)
		// Check the developer against who is logged in
		// Should the operators logged in be allowed to see the API usage of the apps on their cloudlets?
		if !authorized(ctx, rc.claims.Username, org, ResourceAppAnalytics, ActionView) {
			return setReply(c, ws, echo.ErrForbidden, nil)
		}
	} else {
		return setReply(c, ws, echo.ErrNotFound, nil)
	}

	wroteHeader := false
	err = metricsStream(ctx, rc, cmd, func(res interface{}) {
		payload := ormapi.StreamPayload{}
		payload.Data = res
		if ws != nil {
			ws.WriteJSON(payload)
		} else {
			if !wroteHeader {
				c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				c.Response().WriteHeader(http.StatusOK)
				wroteHeader = true
			}
			json.NewEncoder(c.Response()).Encode(payload)
			c.Response().Flush()
		}
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		if !wroteHeader {
			return setReply(c, ws, err, nil)
		}
		res := ormapi.Result{}
		res.Message = err.Error()
		res.Code = http.StatusBadRequest
		payload := ormapi.StreamPayload{Result: &res}
		if ws != nil {
			ws.WriteJSON(payload)
		} else {
			json.NewEncoder(c.Response()).Encode(payload)
		}
	}
	return nil
}

func checkForTimeError(errStr string) string {
	// special case for errors regarding time format
	// golang's reference time is "2006-01-02T15:04:05Z07:00" (123456 in the posix date command), which is confusing
	refTime := "2006-01-02T15:04:05Z07:00"
	if strings.Contains(errStr, refTime) {
		return fmt.Sprintf("%s into RFC3339 format failed. Example: \"%s\"", strings.Split(errStr, " as")[0], refTime)
	}
	return errStr
}
