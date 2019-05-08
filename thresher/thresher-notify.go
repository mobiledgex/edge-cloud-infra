package thresher

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
	//maybe not care about trackedstate ready? dont want to skip it just bc its still creating
	//if we dont care about it being rady we can just make sure its ready before reaching out to collect metrics every time
	if (in.Key.AppKey.Name == MEXPrometheusAppName) && (in.State == edgeproto.TrackedState_Ready) {
		//get the ip and register it in the map
		promMap[in.Uri] = in
	}
}
func (c *AppInstHandler) Delete(in *edgeproto.AppInst, rev int64) {
	if in.Key.AppKey.Name == MEXPrometheusAppName {
		delete(promMap, in.Uri)
	}
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
