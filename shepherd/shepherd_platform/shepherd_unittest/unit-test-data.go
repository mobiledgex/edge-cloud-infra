package shepherd_unittest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Sample output of formatted 'docker stats' cmd
var testDockerResult = `{
	"App": "DockerApp1",
	"memory": {
	  "raw": "11.11MiB / 1.111GiB",
	  "percent": "1.11%"
	},
	"cpu": "1.11%",
	"io": {
	  "network": "111B / 1KB",
	  "block": "1B / 1B"
	}
  }`

// Example output of resource-tracker
var testDockerClusterResult = `{
	"Cpu": 10.10101010,
	"Mem": 11.111111,
	"Disk": 12.12121212,
	"NetSent": 1313131313,
	"NetRecv": 1414141414,
	"TcpConns": 1515,
	"TcpRetrans": 16,
	"UdpSent": 1717,
	"UdpRecv": 1818,
	"UdpRecvErr": 19
  }`

func GetUTData(command string) (string, error) {
	str := ""
	// docker stats unit test
	if strings.Contains(command, "docker stats ") {
		// take the json with line breaks and compact it, as that's what the command expects
		str = testDockerResult
	} else if strings.Contains(command, "resource-tracker") {
		str = testDockerClusterResult
	}
	if str != "" {
		buf := new(bytes.Buffer)
		if err := json.Compact(buf, []byte(str)); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return "", fmt.Errorf("No UT Data found")
}
