package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type NginxScrapePoint struct {
	App       string
	Cluster   string
	Dev       string
	Client    pc.PlatformClient
	Container string //this is specific to k8s, will need to accomodate for docker as well (when we get there)
}

type NginxMetrics struct {
	ActiveConn  uint64
	Accepts     uint64
	HandledConn uint64
	Requests    uint64
	Reading     uint64
	Writing     uint64
	Waiting     uint64
	Ts          *types.Timestamp
}

var nginxAddChan chan NginxScrapePoint
var nginxRemoveChan chan string

func InitNginxScraper() {
	nginxAddChan = make(chan NginxScrapePoint)
	nginxRemoveChan = make(chan string)
	go NginxScraper()
}

func CollectNginxStats(appInst *edgeproto.AppInst) {
	// add/remove from the list of nginx endpoints to hit
	if appInst.State == edgeproto.TrackedState_READY {
		new := NginxScrapePoint{
			App:       appInst.Key.AppKey.Name,
			Cluster:   appInst.Key.ClusterInstKey.ClusterKey.Name,
			Dev:       appInst.Key.AppKey.DeveloperKey.Name,
			Container: k8smgmt.NormalizeName(appInst.Key.AppKey.Name),
		}

		clusterInst := edgeproto.ClusterInst{}
		found := ClusterInstCache.Get(&appInst.Key.ClusterInstKey, &clusterInst)
		if !found {
			log.DebugLog(log.DebugLevelMetrics, "Unable to find clusterInst for "+appInst.Key.AppKey.Name)
			return
		}
		var err error
		new.Client, err = pf.GetPlatformClient(&clusterInst)
		if err != nil {
			// If we cannot get a platform client no point in trying to get metrics
			log.DebugLog(log.DebugLevelMetrics, "Failed to acquire platform client", "cluster", clusterInst.Key, "error", err)
			return
		}
		//this can block for a bit so run it in a separate thread
		go func() {
			nginxAddChan <- new
		}()
	} else {
		//if the app is anything other than ready, stop tracking it
		go func() {
			nginxRemoveChan <- appInst.Key.AppKey.Name
		}()
	}
}

func NginxScraper() {
	nginxMap := make(map[string]NginxScrapePoint)
	for true {
		//check if there are any new apps we need to start/stop scraping for
		select {
		case new := <-nginxAddChan:
			nginxMap[new.App] = new
		case old := <-nginxRemoveChan:
			delete(nginxMap, old)
		case <-time.After(*collectInterval):
			for _, v := range nginxMap {
				span := log.StartSpan(log.DebugLevelSampled, "send-metric")
				span.SetTag("operator", cloudletKey.OperatorKey.Name)
				span.SetTag("cloudlet", cloudletKey.Name)
				span.SetTag("cluster", v.Cluster)
				ctx := log.ContextWithSpan(context.Background(), span)

				metrics, err := QueryNginx(v)
				if err != nil {
					log.DebugLog(log.DebugLevelMetrics, "Error retrieving nginx metrics", "clusterInst", v.Cluster, "error", err.Error())
				} else {
					//send to crm->controller->influx
					influxData := newNginxMetric(v, metrics)
					fmt.Printf("ASDF sending %+v\n", influxData)
					MetricSender.Update(ctx, influxData)
				}
				span.Finish()
			}
		}
	}

}

func QueryNginx(scrapePoint NginxScrapePoint) (*NginxMetrics, error) {
	//build the query
	request := fmt.Sprintf("docker exec -it %s curl http://127.0.0.1:8080/nginx_metrics", scrapePoint.Container)
	resp, err := scrapePoint.Client.Output(request)
	if err != nil {
		errstr := fmt.Sprintf("Failed to run <%s>", request)
		log.DebugLog(log.DebugLevelMetrics, errstr, "err", err.Error())
		return nil, err
	}
	//if this is the first time, or the container got restarted, install curl
	if strings.Contains(resp, "executable file not found") {
		log.DebugLog(log.DebugLevelMexos, "Installing curl onto docker container for metrics collection")
		installer := fmt.Sprintf("docker exec -it %s apt-get update; docker exec -it %s apt-get --assume-yes install curl", scrapePoint.Container, scrapePoint.Container)
		resp, err = scrapePoint.Client.Output(installer)
		if err != nil {
			return nil, fmt.Errorf("can't install curl on nginx container %s, %s, %v", *name, resp, err)
		}
		//now retry curling
		resp, err = scrapePoint.Client.Output(request)
		if err != nil {
			errstr := fmt.Sprintf("Failed to run <%s>", request)
			log.DebugLog(log.DebugLevelMetrics, errstr, "err", err.Error())
			return nil, err
		}
	}
	metrics := &NginxMetrics{}
	err = parseNginxResp(resp, metrics)
	if err != nil {
		return nil, fmt.Errorf("Error parsing response: %v", err)
	}
	return metrics, nil
}

func parseNginxResp(resp string, metrics *NginxMetrics) error {
	trimmedResp := OutputTrim(resp)
	lines := strings.Split(trimmedResp, "\n")

	var err error
	//fix this hardcoding later
	if len(lines) != 4 {
		return fmt.Errorf("Unexpected output format")
	}

	//first line has active connections
	stats := strings.Split(lines[0], " ")
	if metrics.ActiveConn, err = strconv.ParseUint(stats[2], 10, 64); err != nil {
		return err
	}
	//third line for accepts, handled, and requests
	stats = strings.Split(lines[2], " ")
	if metrics.Accepts, err = strconv.ParseUint(stats[1], 10, 64); err != nil {
		return err
	}
	if metrics.HandledConn, err = strconv.ParseUint(stats[2], 10, 64); err != nil {
		return err
	}
	if metrics.Requests, err = strconv.ParseUint(stats[3], 10, 64); err != nil {
		return err
	}
	//last line has reading, writing, waiting
	stats = strings.Split(lines[3], " ")
	if metrics.Reading, err = strconv.ParseUint(stats[1], 10, 64); err != nil {
		return err
	}
	if metrics.Writing, err = strconv.ParseUint(stats[3], 10, 64); err != nil {
		return err
	}
	if metrics.Waiting, err = strconv.ParseUint(stats[5], 10, 64); err != nil {
		return err
	}
	metrics.Ts, _ = types.TimestampProto(time.Now())
	return nil
}

func newNginxMetric(scrapePoint NginxScrapePoint, data *NginxMetrics) *edgeproto.Metric {
	metric := edgeproto.Metric{}
	metric.Name = "appinst-nginx"
	metric.Timestamp = *data.Ts
	metric.AddTag("operator", cloudletKey.OperatorKey.Name)
	metric.AddTag("cloudlet", cloudletKey.Name)
	metric.AddTag("cluster", scrapePoint.Cluster)
	metric.AddTag("dev", scrapePoint.Dev)
	metric.AddTag("app", scrapePoint.App)

	metric.AddIntVal("active", data.ActiveConn)
	metric.AddIntVal("accepts", data.Accepts)
	metric.AddIntVal("handled", data.HandledConn)
	metric.AddIntVal("requests", data.Requests)
	metric.AddIntVal("reading", data.Reading)
	metric.AddIntVal("writing", data.Writing)
	metric.AddIntVal("waiting", data.Waiting)
	return &metric
}
