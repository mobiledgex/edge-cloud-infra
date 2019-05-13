package thresher

import (
	"flag"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
)

//commandline options
var operatorName = flag.String("operator", "local", "Cloudlet Operator Name")
var cloudletName = flag.String("cloudlet", "localtest", "Cloudlet Name")
var clusterName = flag.String("cluster", "AppCluster", "Cluster Name")
var collectInterval = flag.Duration("interval", time.Second*15, "Metrics collection interval")
var tlsCertFile = flag.String("tls", "", "server9 tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")

var exporterT *template.Template

var MEXMetricsExporterEnvVars = `- name: MEX_CLUSTER_NAME
  valueFrom:
    configMapKeyRef:
      name: mexcluster-info
      key: ClusterName
      optional: true
- name: MEX_CLOUDLET_NAME
  valueFrom:
    configMapKeyRef:
      name: mexcluster-info
      key: CloudletName
      optional: true
- name: MEX_OPERATOR_NAME
  valueFrom:
    configMapKeyRef:
      name: mexcluster-info
      key: OperatorName
      optional: true
`
var MEXMetricsExporterEnvTempl = `- name: MEX_INFLUXDB_ADDR
  value: {{.InfluxDBAddr}}
- name: MEX_INFLUXDB_USER
  value: {{.InfluxDBUser}}
- name: MEX_INFLUXDB_PASS
  value: {{.InfluxDBPass}}
- name: MEX_SCRAPE_INTERVAL
  value: {{.Interval}} `

var MEXMetricsExporterApp = edgeproto.App{
	Key: edgeproto.AppKey{
		Name:    MEXMetricsExporterAppName,
		Version: MEXMetricsExporterAppVer,
		DeveloperKey: edgeproto.DeveloperKey{
			Name: cloudcommon.DeveloperMobiledgeX,
		},
	},
	ImagePath:     "registry.mobiledgex.net:5000/mobiledgex/metrics-exporter:latest",
	ImageType:     edgeproto.ImageType_ImageTypeDocker,
	DefaultFlavor: edgeproto.FlavorKey{Name: *appFlavor},
	DelOpt:        edgeproto.DeleteType_AutoDelete,
}

type exporterData struct {
	InfluxDBAddr string
	InfluxDBUser string
	InfluxDBPass string
	Interval     string
}

// myCloudlet is the information for the cloudlet in which the MEXThresher is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file.
var myCloudletKey edgeproto.CloudletKey

func appInstCreateApi(apiClient edgeproto.AppInstApiClient, appInst edgeproto.AppInst) (*edgeproto.Result, error) {
	ctx := context.TODO()
	stream, err := apiClient.CreateAppInst(ctx, &appInst)
	var res *edgeproto.Result
	if err == nil {
		for {
			res, err = stream.Recv()
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				break
			}
		}
	}
	return res, err
}

// create an thresher appInst
func createMEXMetricsExporterInst(dialOpts grpc.DialOption, instKey edgeproto.ClusterInstKey, app *edgeproto.App) error {
	//update flavor
	app.DefaultFlavor = edgeproto.FlavorKey{Name: *appFlavor}
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()
	apiClient := edgeproto.NewAppInstApiClient(conn)

	appInst := edgeproto.AppInst{
		Key: edgeproto.AppInstKey{
			AppKey:      app.Key,
			CloudletKey: instKey.CloudletKey,
			Id:          1,
		},
		ClusterInstKey: instKey,
	}

	res, err := appInstCreateApi(apiClient, appInst)
	if err != nil {
		// Handle non-fatal errors
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyExists.Error()) {
			log.DebugLog(log.DebugLevelMexos, "appinst already exists", "app", app.String(), "cluster", instKey.String())
			return nil
		}
		if strings.Contains(err.Error(), edgeproto.ErrEdgeApiAppNotFound.Error()) {
			log.DebugLog(log.DebugLevelMexos, "app doesn't exist, create it first", "app", app.String())
			// Create the app
			if err = createAppCommon(dialOpts, app); err == nil {
				if res, err = appInstCreateApi(apiClient, appInst); err == nil {
					log.DebugLog(log.DebugLevelMexos, "create appinst", "appinst", appInst.String(), "result", res.String())
					return nil
				}
			}
		}
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateAppInst failed: %s", errstr)
	}
	log.DebugLog(log.DebugLevelMexos, "create appinst", "appinst", appInst.String(), "result", res.String())
	return nil

}

func scrapeIntervalInSeconds(scrapeInterval time.Duration) string {
	var secs = int(scrapeInterval.Seconds()) //round it to the second
	var scrapeStr = strconv.Itoa(secs) + "s"
	return scrapeStr
}

func fillAppConfigs(app *edgeproto.App) error {
	var scrapeStr = scrapeIntervalInSeconds(*scrapeInterval)
	switch app.Key.Name {
	case MEXMetricsExporterAppName:
		ex := exporterData{
			InfluxDBAddr: *influxDBAddr,
			InfluxDBUser: *influxDBUser,
			InfluxDBPass: *influxDBPass,
			Interval:     scrapeStr,
		}
		buf := bytes.Buffer{}
		err := exporterT.Execute(&buf, &ex)
		if err != nil {
			return err
		}
		paramConf := edgeproto.ConfigFile{
			Kind:   k8smgmt.AppConfigEnvYaml,
			Config: buf.String(),
		}
		envConf := edgeproto.ConfigFile{
			Kind:   k8smgmt.AppConfigEnvYaml,
			Config: MEXMetricsExporterEnvVars,
		}

		app.Configs = []*edgeproto.ConfigFile{&paramConf, &envConf}
	case MEXPrometheusAppName:
		ex := exporterData{
			Interval: scrapeStr,
		}
		buf := bytes.Buffer{}
		err := prometheusT.Execute(&buf, &ex)
		if err != nil {
			return err
		}
		// Now add this yaml to the prometheus AppYamls
		config := edgeproto.ConfigFile{
			Kind:   k8smgmt.AppConfigHelmYaml,
			Config: buf.String(),
		}
		app.Configs = []*edgeproto.ConfigFile{&config}
		app.AccessPorts = *externalPorts
	default:
		return fmt.Errorf("Unrecognized app %s", app.Key.Name)
	}
	return nil
}

func createAppCommon(dialOpts grpc.DialOption, app *edgeproto.App) error {
	conn, err := grpc.Dial(*ctrlAddr, dialOpts, grpc.WithBlock(), grpc.WithWaitForHandshake())
	if err != nil {
		return fmt.Errorf("Connect to server %s failed: %s", *ctrlAddr, err.Error())
	}
	defer conn.Close()

	// add app customizations
	if err = fillAppConfigs(app); err != nil {
		return err
	}
	apiClient := edgeproto.NewAppApiClient(conn)
	ctx := context.TODO()
	res, err := apiClient.CreateApp(ctx, app)
	if err != nil {
		// Handle non-fatal errors
		if strings.Contains(err.Error(), objstore.ErrKVStoreKeyExists.Error()) {
			log.DebugLog(log.DebugLevelMexos, "app already exists", "app", app.String())
			return nil
		}
		errstr := err.Error()
		st, ok := status.FromError(err)
		if ok {
			errstr = st.Message()
		}
		return fmt.Errorf("CreateApp failed: %s", errstr)
	}
	log.DebugLog(log.DebugLevelMexos, "create app", "app", app.String(), "result", res.String())
	return nil
}


TODO: add a cluster to the command line args so i have somewhere to start the metrics app 
func main() {
	flag.Parse()
	//TODO: figure out if we wanna do standalone stuff
	promMap = make(map[string]edgeproto.ClusterInstKey)
	cloudcommon.ParseMyCloudletKey(false, cloudletKeyStr, &myCloudletKey)
	//start metrics exporter
	TODO: fix the cluster part of this, it needs to start in rootlb vm
	if err = createMEXMetricsExporterInst(dialOpts, in.Key, &MEXMetricsExporterApp); err != nil {
		log.DebugLog(log.DebugLevelMexos, "Metrics-exporter inst create failed", "cluster", in.Key.ClusterKey.Name,
			"error", err.Error())
	notifyClient := initNotifyClient(*notifyAddrs, *tlsCertFile)
	notifyClient.Start()
	defer notifyClient.Stop()
}
