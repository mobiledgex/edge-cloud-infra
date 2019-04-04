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
var cloudletName = flag.String("cloudlet", "local", "Cloudlet Name")
var clusterName = flag.String("cluster", "myclust", "Cluster Name")
var collectInterval = flag.Duration("interval", time.Second*15, "Metrics collection interval")
var tlsCertFile = flag.String("tls", "", "server tls cert file.  Keyfile and CA file mex-ca.crt must be in same directory")
var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var cloudletKeyStr = flag.String("cloudletKey", "", "Json or Yaml formatted cloudletKey for the cloudlet in which this CRM is instantiated; e.g. '{\"operator_key\":{\"name\":\"TMUS\"},\"name\":\"tmocloud1\"}'")

//map keeping track of all the currently running prometheuses
//TODO: figure out exactly what the types need to be
var promMap map[string]edgeproto.ClusterInstKey

var MEXPrometheusAppName = "MEXPrometheusAppName"

// myCloudlet is the information for the cloudlet in which the MEXThresher is instantiated.
// The key for myCloudlet is provided as a configuration - either command line or
// from a file.
var myCloudletKey edgeproto.CloudletKey

type AppInstHandler struct {
}

//TODO: fill out update and delete
func (c *AppInstHandler) Update(in *edgeproto.AppInst, rev int64) {
	if (in.Key.AppKey.Name == MEXPrometheusAppName) && (in.State == edgeproto.TrackedState_Ready) {
		//get the ip and register it in the map

	}
}
func (c *AppInstHandler) Delete(in *edgeproto.AppInst, rev int64) {
}
func (c *AppInstHandler) Prune(keys map[edgeproto.AppInstKey]struct{}) {
}
func (c *AppInstHandler) Flush(notifyId int64) {
}

func initNotifyClient(addrs string, tlsCertFile string) *notify.Client {
	notifyClient := notify.NewClient(strings.Split(addrs, ","), tlsCertFile)
	notifyClient.RegisterRecv(notify.NewAppInstRecv(&AppInstHandler{}))
	log.InfoLog("notify client to", "addrs", addrs)
	return notifyClient
}

func main() {
	flag.Parse()
	//TODO: figure out if we wanna do standalone stuff
	promMap = make(map[string]edgeproto.ClusterInstKey)
	cloudcommon.ParseMyCloudletKey(false, cloudletKeyStr, &myCloudletKey)
	notifyClient := initNotifyClient(*notifyAddrs, *tlsCertFile)
	notifyClient.Start()
	defer notifyClient.Stop()
}
