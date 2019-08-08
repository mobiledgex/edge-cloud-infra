package e2esetup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
)

type InfluxResp struct {
	Results []struct {
		Series []struct {
			Columns []string        `json:"columns"`
			Name    string          `json:"name"`
			Values  [][]interface{} `json:"values"`
		} `json:"series"`
		StatementID int `json:"statement_id"`
	} `json:"results"`
}

var influxQuery = "http://127.0.0.1:8086/query?db=" + cloudcommon.DeveloperMetricsDbName + "&q="
var querySeparator = "%3B" //";"

//queries for measurement selection
var selectMeasurements = []string{
	"select%20%22cpu%22%20from%20%22cluster-cpu%22",                                        //select "cpu" from "cluster-cpu"
	"select%20%22mem%22%20from%20%22cluster-mem%22",                                        //select "mem" from "cluster-mem"
	"select%20%22disk%22%20from%20%22cluster-disk%22",                                      //select "disk" from "cluster-disk"
	"select%20%22sendBytes%22%2C%22recvBytes%22%20from%20%22cluster-network%22",            //select "sendBytes","recvBytes" from "cluster-network"
	"select%20%22tcpConns%22%2C%22tcpRetrans%22%20from%20%22cluster-tcp%22",                //select "tcpConns","tcpRetrans" from "cluster-tcp"
	"select%20%22udpSend%22%2C%22udpRecv%22%2C%22udpRecvErr%22%20from%20%22cluster-udp%22", //select "udpSend","udpRecv","udpRecvErr" from "cluster-udp"
	"select%20%22cpu%22%20from%20%22appinst-cpu%22",                                        //select "cpu" from "appinst-cpu"
	"select%20%22mem%22%20from%20%22appinst-mem%22",                                        //select "mem" from "appinst-mem"
	"select%20%22disk%22%20from%20%22appinst-disk%22",                                      //select "disk" from "appinst-disk"
	"select%20%22sendBytes%22%2C%22recvBytes%22%20from%20%22appinst-network%22",            //select "sendBytes","recvBytes" from "appinst-network"
}

//queries for deleting measurements
var dropMeasurements = []string{
	"drop%20measurement%20%22cluster-cpu%22",
	"drop%20measurement%20%22cluster-mem%22",
	"drop%20measurement%20%22cluster-disk%22",
	"drop%20measurement%20%22cluster-network%22",
	"drop%20measurement%20%22cluster-tcp%22",
	"drop%20measurement%20%22cluster-udp%22",
	"drop%20measurement%20%22appinst-cpu%22",
	"drop%20measurement%20%22appinst-mem%22",
	"drop%20measurement%20%22appinst-disk%22",
	"drop%20measurement%20%22appinst-network%22",
}

func RunCheckMetrics(actionSubtype string) error {
	switch actionSubtype {
	case "shepherd":
		return CheckInfluxShepherd()
	default:
		return fmt.Errorf("Unsupported action type: " + actionSubtype)
	}

}

func CheckInfluxShepherd() error {
	//give shepherd time to collect and push metrics
	time.Sleep(5 * time.Second)
	metric, err := getInfluxMeasurements()
	if err != nil {
		return err
	}

	if err = checkInfluxMeasurements(metric); err != nil {
		return err
	}

	//wipe influx for the next iteration
	clearInfluxMeasurements()

	return nil
}

func getInfluxMeasurements() (*InfluxResp, error) {
	metric := InfluxResp{}
	query := strings.Join(selectMeasurements, querySeparator)
	resp, err := http.Get(influxQuery + query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &metric); err != nil {
		return nil, err
	}
	return &metric, nil
}

func clearInfluxMeasurements() {
	query := strings.Join(dropMeasurements, querySeparator)
	http.Post(influxQuery+query, "application/x-www-form-urlencoded", nil)
	return
}

func checkInfluxMeasurements(metric *InfluxResp) error {
	//for some reason all values are seen as floats, TODO: figure out why

	//make sure all the metrics are there and got pulled
	if len(metric.Results) != len(selectMeasurements) {
		return fmt.Errorf("Shepherd metrics incomplete")
	}

	//cpu from cluster-cpu
	if len(metric.Results[0].Series) > 0 || metric.Results[0].Series[0].Values[0][1] != float64(10.01) {
		return fmt.Errorf("Influx cluster-cpu measurements inconsistent")
	}

	//mem from cluster-mem
	if len(metric.Results[1].Series) > 0 || metric.Results[1].Series[0].Values[0][1] != float64(99.99) {
		return fmt.Errorf("Influx cluster-mem measurements inconsistent")
	}

	//disk from cluster-disk
	if len(metric.Results[2].Series) > 0 || metric.Results[2].Series[0].Values[0][1] != float64(50.0) {
		return fmt.Errorf("Influx cluster-disk measurements inconsistent")
	}

	//cluster-network
	//sendBytes
	if len(metric.Results[3].Series) > 0 || metric.Results[3].Series[0].Values[0][1] != float64(11111) {
		return fmt.Errorf("Influx cluster-network measurements inconsistent")
	}
	//recvBytes
	if metric.Results[3].Series[0].Values[0][2] != float64(22222) {
		return fmt.Errorf("Influx cluster-network measurements inconsistent")
	}

	//cluster-tcp
	//tcpConns
	if len(metric.Results[4].Series) > 0 || metric.Results[4].Series[0].Values[0][1] != float64(33333) {
		return fmt.Errorf("Influx cluster-tcp measurements inconsistent")
	}
	//tcpRetrans
	if metric.Results[4].Series[0].Values[0][2] != float64(44444) {
		return fmt.Errorf("Influx cluster-tcp measurements inconsistent")
	}

	//cluster-udp
	//udpSend
	if len(metric.Results[5].Series) > 0 || metric.Results[5].Series[0].Values[0][1] != float64(55555) {
		return fmt.Errorf("Influx cluster-udp measurements inconsistent")
	}
	//udpRecv
	if metric.Results[5].Series[0].Values[0][2] != float64(66666) {
		return fmt.Errorf("Influx cluster-udp measurements inconsistent")
	}
	//udpRecvErr
	if metric.Results[5].Series[0].Values[0][3] != float64(77777) {
		return fmt.Errorf("Influx cluster-udp measurements inconsistent")
	}

	//cpu from appinst-cpu
	if len(metric.Results[6].Series) > 0 || metric.Results[6].Series[0].Values[0][1] != float64(5.0) {
		return fmt.Errorf("Influx appinst-cpu measurements inconsistent")
	}

	//mem from appinst-mem
	if len(metric.Results[7].Series) > 0 || metric.Results[7].Series[0].Values[0][1] != float64(100000000) {
		return fmt.Errorf("Influx appinst-mem measurements inconsistent")
	}

	//disk from appinst-disk
	if len(metric.Results[8].Series) > 0 || metric.Results[8].Series[0].Values[0][1] != float64(200000000) {
		return fmt.Errorf("Influx appinst-disk measurements inconsistent")
	}

	//appinst-network
	//sendBytes
	if len(metric.Results[9].Series) > 0 || metric.Results[9].Series[0].Values[0][1] != float64(111111) {
		return fmt.Errorf("Influx appinst-network measurements inconsistent")
	}
	//recvBytes
	if metric.Results[9].Series[0].Values[0][2] != float64(222222) {
		return fmt.Errorf("Influx appinst-network measurements inconsistent")
	}
	return nil
}
