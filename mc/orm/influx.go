package orm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"google.golang.org/grpc/status"
)

var influxDBTemplate *template.Template

// 100 values at a time
var queryChunkSize = 100

type InfluxDBContext struct {
	region string
	claims *UserClaims
	conn   influxdb.Client
}

type InfluxDBVaultData struct {
	username string
	password string
}

type influxQueryArgs struct {
	Selector     string
	Measurement  string
	AppInstName  string
	ClusterName  string
	CloudletName string
	OperatorName string
	StartTime    string
	EndTime      string
}

var influDBT = `SELECT {{if .Selector}}"{{.Selector}}"{{else}}*{{end}} from "{{.Measurement}}"` +
	` WHERE "cluster"='{{.ClusterName}}'` +
	`{{if .AppInstName}} AND "app"=~/{{.AppInstName}}/{{end}}` +
	`{{if .CloudletName}} AND "cloudlet"='{{.CloudletName}}'{{end}}` +
	`{{if .OperatorName}} AND "operator"='{{.OperatorName}}'{{end}}` +
	`{{if .StartTime}} AND time > '{{.StartTime}}'{{end}}` +
	`{{if .EndTime}} AND time < '{{.EndTime}}'{{end}}`

func init() {
	influxDBTemplate = template.Must(template.New("influxquery").Parse(influDBT))
}

func connectInfluxDB(region string) (influxdb.Client, error) {
	addr, err := getInfluxDBAddrForRegion(region)
	if err != nil {
		fmt.Printf("Failed to get addr for region: %s\n", region)
		return nil, err
	}
	creds, err := getInfluxDBCreds(region)
	if err != nil {
		// defeault to empty login
		creds = &InfluxDBVaultData{
			username: "",
			password: "",
		}
	}
	fmt.Printf("Connecting to %s\n", addr)
	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     "http://" + addr,
		Username: creds.username,
		Password: creds.password,
	})
	if err != nil {
		fmt.Printf("Failed to connect to %s err: %v\n", addr, err)
		return nil, err
	}
	return client, nil

}

func getInfluxDBAddrForRegion(region string) (string, error) {
	ctrl, err := getControllerObj(region)
	if err != nil {
		return "", err
	}
	return ctrl.InfluxDB, nil
}

// Get influxDB login credentials from the vault
func getInfluxDBCreds(region string) (*InfluxDBVaultData, error) {
	data, err := mexos.GetVaultData(cloudcommon.InfluxDBVaultPath + "region/influxdb.json")
	if err != nil {
		return nil, err
	}
	influxData := &InfluxDBVaultData{}
	err = mapstructure.Decode(data, influxData)
	if err != nil {
		return nil, err
	}
	return influxData, nil
}

// Query is a template with a specific set of if/else
func AppInstMetricsQuery(obj *ormapi.RegionAppInstMetrics, measurement string) string {
	appInst := obj.AppInst
	arg := influxQueryArgs{
		Selector:     obj.Selector,
		Measurement:  measurement,
		AppInstName:  appInst.Key.AppKey.Name,
		CloudletName: appInst.Key.ClusterInstKey.CloudletKey.Name,
		ClusterName:  appInst.Key.ClusterInstKey.ClusterKey.Name,
		OperatorName: appInst.Key.ClusterInstKey.CloudletKey.OperatorKey.Name,
	}

	// Figure out the start/end time range for the query
	if !obj.StartTime.IsZero() {
		buf, err := obj.StartTime.MarshalText()
		if err == nil {
			arg.StartTime = string(buf)
		}
	}
	if !obj.EndTime.IsZero() {
		buf, err := obj.EndTime.MarshalText()
		if err == nil {
			arg.EndTime = string(buf)
		}
	}
	// now that we know all the details of the query - build it
	buf := bytes.Buffer{}
	if err := influxDBTemplate.Execute(&buf, &arg); err != nil {
		return ""
	}
	return buf.String()
}

// Query is a template with a specific set of if/else
func ClusterMetricsQuery(obj *ormapi.RegionClusterInstMetrics, measurement string) string {
	cluster := obj.ClusterInst
	arg := influxQueryArgs{
		Selector:     obj.Selector,
		Measurement:  measurement,
		CloudletName: cluster.Key.CloudletKey.Name,
		ClusterName:  cluster.Key.ClusterKey.Name,
		OperatorName: cluster.Key.CloudletKey.OperatorKey.Name,
	}

	// Figure out the start/end time range for the query
	if !obj.StartTime.IsZero() {
		buf, err := obj.StartTime.MarshalText()
		if err == nil {
			arg.StartTime = string(buf)
		}
	}
	if !obj.EndTime.IsZero() {
		buf, err := obj.EndTime.MarshalText()
		if err == nil {
			arg.EndTime = string(buf)
		}
	}
	// now that we know all the details of the query - build it
	buf := bytes.Buffer{}
	if err := influxDBTemplate.Execute(&buf, &arg); err != nil {
		return ""
	}
	return buf.String()
}

func metricsStream(rc *InfluxDBContext, dbQuery string, cb func(Data interface{})) error {
	if rc.conn == nil {
		conn, err := connectInfluxDB(rc.region)
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
	fmt.Printf("Got command: [%s]\n", dbQuery)
	fmt.Printf("Running query: %v\n", query)
	resp, err := rc.conn.Query(query)
	if err != nil {
		fmt.Printf("Coudlnt'run query %v\n", err)
		return err
	}
	if resp.Error() != nil {
		fmt.Printf("Got back and error: %v\n", resp)
		return resp.Error()
	}
	fmt.Printf("No error!!! %v\n", resp)
	cb(resp.Results)
	return nil
}

// Common method to handle both app and cluster metrics
func GetMetricsCommon(c echo.Context) error {
	var cmd, user string

	rc := &InfluxDBContext{}
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	rc.claims = claims

	if strings.HasSuffix(c.Path(), "metrics/app") {
		in := ormapi.RegionAppInstMetrics{}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, Msg("Invalid GET data"))
		}
		rc.region = in.Region
		user = in.AppInst.Key.AppKey.DeveloperKey.Name
		cmd = AppInstMetricsQuery(&in, cloudcommon.DeveloperAppMetrics)
	} else if strings.HasSuffix(c.Path(), "metrics/cluster") {
		in := ormapi.RegionClusterInstMetrics{}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, Msg("Invalid GET data"))
		}
		rc.region = in.Region
		user = in.ClusterInst.Key.Developer
		cmd = ClusterMetricsQuery(&in, cloudcommon.DeveloperClusterMetric)
	} else {
		return echo.ErrNotFound
	}

	// Check the developer against who is logged in
	if !enforcer.Enforce(rc.claims.Username, user, ResourceAppAnalytics, ActionView) {
		return echo.ErrForbidden
	}

	wroteHeader := false
	err = metricsStream(rc, cmd, func(res interface{}) {
		if !wroteHeader {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c.Response().WriteHeader(http.StatusOK)
			wroteHeader = true
		}
		payload := ormapi.StreamPayload{}
		payload.Data = res
		json.NewEncoder(c.Response()).Encode(payload)
		c.Response().Flush()
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = fmt.Errorf("%s", st.Message())
		}
		if !wroteHeader {
			return setReply(c, err, nil)
		}
		res := ormapi.Result{}
		res.Message = err.Error()
		res.Code = http.StatusBadRequest
		payload := ormapi.StreamPayload{Result: &res}
		json.NewEncoder(c.Response()).Encode(payload)
	}
	return nil
}
