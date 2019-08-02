package e2esetup

import (
	"fmt"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	influxq "github.com/mobiledgex/edge-cloud/controller/influxq_client"
)

func RunCheckMetrics(actionSubtype string) error {
	switch actionSubtype {
	case "shepherd":
		return CheckInfluxShepherd()
	default:
		return fmt.Errorf("Unsupported action type: " + actionSubtype)
	}

}

func CheckInfluxShepherd() error {
	//start an influx client so we can read from it
	influxAddr := "127.0.0.1:8086" //maybe somehow pull this from "local_multi.yml" so its not hardcoded
	influxAuth := &cloudcommon.InfluxCreds{}
	influxQ := influxq.NewInfluxQ(cloudcommon.DeveloperMetricsDbName, influxAuth.User, influxAuth.Pass)
	err := influxQ.Start(influxAddr, "")
	if err != nil {
		return fmt.Errorf("Failed to start influx queue address %s, %v", influxAddr, err)
	}
	defer influxQ.Stop()

	//check all the measurements
	var result []client.Result
	result, err = influxQ.QueryDB("select \"cpu\" from \"cluster-cpu\" LIMIT 1")
	if err != nil {
		return err
	}
	//figure out how to read result
	fmt.Printf("asdf %v\n", result[0].Series[0].Values[0])

	result, err = influxQ.QueryDB("select \"mem\" from \"cluster-mem\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"disk\" from \"cluster-disk\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"sendBytes\",\"recvBytes\" from \"cluster-network\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"tcpConns\",\"tcpRetrans\" from \"cluster-tcp\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"udpSend\",\"udpRecv\",\"udpRecvErr\" from \"cluster-udp\" LIMIT 1")

	result, err = influxQ.QueryDB("select \"cpu\" from \"appinst-cpu\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"mem\" from \"appinst-mem\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"disk\" from \"appinst-disk\" LIMIT 1")
	result, err = influxQ.QueryDB("select \"sendBytes\",\"recvBytes\" from \"appinst-network\" LIMIT 1")

	return nil
}
