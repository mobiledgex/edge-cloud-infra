package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud-infra/shepherd/shepherd_common"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/k8smgmt"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var dockerStatsFormat = `"{\"App\":\"{{.Name}}\",\"memory\":{\"raw\":\"{{.MemUsage}}\",\"percent\":\"{{.MemPerc}}\"},\"cpu\":\"{{.CPUPerc}}\",\"io\":{\"network\":\"{{.NetIO}}\",\"block\":\"{{.BlockIO}}\"}}"`
var dockerStatsCmd = "docker stats --no-stream --format " + dockerStatsFormat

// Prerequisite - install small edge-cloud utility on the VM running this docker containers
var resTrackerCmd = "resource-tracker"

type ContainerMem struct {
	Raw     string
	Percent string
}
type ContainerIO struct {
	Network string
	Block   string
}
type ContainerStats struct {
	App    string
	Memory ContainerMem
	Cpu    string
	IO     ContainerIO
}

type DockerStats struct {
	Containers []ContainerStats
}

// Docker Cluster
type DockerClusterStats struct {
	key    edgeproto.ClusterInstKey
	client pc.PlatformClient
	shepherd_common.ClusterMetrics
}

func (c *DockerClusterStats) GetClusterStats() *shepherd_common.ClusterMetrics {
	if err := collectDockerClusterMMetrics(c); err != nil {
		log.DebugLog(log.DebugLevelMetrics, "Could not collect cluster metrics", "Docker cluster", c)
		return nil
	}
	return &c.ClusterMetrics
}

// Currently we are collecting stats for all apps in the cluster in one shot
// Implementing  EDGECLOUD-1183 would allow us to query by label and we can have each app be an individual metric
func (c *DockerClusterStats) GetAppStats() map[shepherd_common.MetricAppInstKey]*shepherd_common.AppMetrics {
	metrics := collectDockerAppMetrics(c)
	if metrics == nil {
		log.DebugLog(log.DebugLevelMetrics, "Could not collect app metrics", "Docker Container", c)
	}
	return metrics
}

// Get the output of the container stats on the platform and format them properly
func (c *DockerClusterStats) GetContainerStats() (*DockerStats, error) {
	resp, err := c.client.Output(dockerStatsCmd)
	if err != nil {
		errstr := fmt.Sprintf("Failed to run <%s>", dockerStatsCmd)
		log.DebugLog(log.DebugLevelMetrics, errstr, "err", err.Error())
		return nil, err
	}
	dockerResp := &DockerStats{}
	stats := strings.Split(resp, "\n")
	for _, c := range stats {
		if c == "" {
			// last string is an empty string
			continue
		}
		containerStat := ContainerStats{}
		if err = json.Unmarshal([]byte(c), &containerStat); err != nil {
			log.DebugLog(log.DebugLevelMetrics, "Failed to marshal stats", "stats", c, "err", err.Error())
			continue
		}
		dockerResp.Containers = append(dockerResp.Containers, containerStat)
	}
	return dockerResp, nil
}

func parsePercentStr(pStr string) (float64, error) {
	i := strings.Index(pStr, "%")
	if i < 0 {
		return 0, fmt.Errorf("Invalid percentage string")
	}
	return strconv.ParseFloat(pStr[:i], 64)
}

// parse data in the format "1.629MiB / 1.952GiB / 12KB / 12B" into [1.629* 1000000, 1.952 * 1000000000, 12*1000 , 12]
func parseComputeUnitsDelim(dataStr string) ([]uint64, error) {
	var items []uint64
	var scale uint64
	// token function to find first letter(K/M/G)
	tokenFunc := func(c rune) bool {
		return unicode.IsLetter(c)
	}

	vals := strings.Split(dataStr, " / ")
	for _, v := range vals {

		i := strings.IndexFunc(v, tokenFunc)
		if i == -1 {
			//if just a number add to array
			if t, err := strconv.ParseUint(v, 10, 64); err == nil {
				items = append(items, t)
			} else {
				log.DebugLog(log.DebugLevelMetrics, "Failed to parse data", "val", v, "err", err)
			}
			continue
		}
		// deal with KiB, MiB, GiB
		switch v[i] {
		case 'B':
			scale = 1
		case 'k':
			fallthrough
		case 'K':
			scale = 1024
		case 'm':
			fallthrough
		case 'M':
			scale = 1024 * 1024
		case 'g':
			fallthrough
		case 'G':
			scale = 1024 * 1024
		default:
			log.DebugLog(log.DebugLevelMetrics, "Unknown Unit string", "units", v[i])
			continue
		}

		if t, err := strconv.ParseFloat(v[0:i], 64); err == nil {
			t *= float64(scale)
			items = append(items, uint64(t))
		} else {
			log.DebugLog(log.DebugLevelMetrics, "Failed to parse data", "val", v[0:i], "err", err)
		}
	}
	return items, nil
}

// This is a helper function to get docker container statistics.
// Currntly it runs "docker stats" on rootLB and gets coarse stats from the resource utilization
// If a more detailed reource usage is needed /containers/(id)/stats API endpoint should be used
// To get to the API endpoint on a rootLB netcat can be used:
//   $ echo -e "GET /containers/mobiledgexsdkdemo/stats?stream=0 HTTP/1.0\r\n" | nc -q -1 -U /var/run/docker.sock | grep "^{" | jq
func collectDockerAppMetrics(p *DockerClusterStats) map[shepherd_common.MetricAppInstKey]*shepherd_common.AppMetrics {
	appStatsMap := make(map[shepherd_common.MetricAppInstKey]*shepherd_common.AppMetrics)

	stats, err := p.GetContainerStats()
	if err != nil {
		log.DebugLog(log.DebugLevelMetrics, "Failed to collect App stats for docker cluster", "err", err)
		return nil
	}

	appKey := shepherd_common.MetricAppInstKey{
		ClusterInstKey: p.key,
	}

	// We scraped it at the same time, so same timestamp for everything
	ts, _ := types.TimestampProto(time.Now())
	for _, containerStats := range stats.Containers {
		// EDGECLOUD-1183  - once done we should not normalize the name
		appKey.Pod = k8smgmt.NormalizeName(containerStats.App)
		stat, found := appStatsMap[appKey]
		if !found {
			stat = &shepherd_common.AppMetrics{}
			appStatsMap[appKey] = stat
		}
		stat.CpuTS, stat.MemTS, stat.DiskTS, stat.NetSentTS, stat.NetRecvTS = ts, ts, ts, ts, ts
		// cpu is in the form "0.00%" - remove the % at the end and cast to float
		stat.Cpu, err = parsePercentStr(containerStats.Cpu)
		if err != nil {
			log.DebugLog(log.DebugLevelMetrics, "Failed to parse CPU usage", "App", appKey, "stats", containerStats, "err", err)
		}

		memData, err := parseComputeUnitsDelim(containerStats.Memory.Raw)
		if err != nil {
			log.DebugLog(log.DebugLevelMetrics, "Failed to parse Mem usage", "App", appKey, "stats", containerStats, "err", err)
		} else {
			stat.Mem = memData[0]
		}
		// Disk usage is unsupported
		stat.Disk = 0
		netIO, err := parseComputeUnitsDelim(containerStats.IO.Network)
		if err != nil {
			log.DebugLog(log.DebugLevelMetrics, "Failed to parse Network usage", "App", appKey, "stats", containerStats, "err", err)
		} else {
			if len(netIO) > 1 {
				stat.NetSent = netIO[1]
				stat.NetRecv = netIO[0]
			} else {
				log.DebugLog(log.DebugLevelMetrics, "Failed to parse network data", "netio", netIO)
			}
		}
	}

	return appStatsMap
}

func collectDockerClusterMMetrics(p *DockerClusterStats) error {
	// VM stats from Openstack might be a better idea going forward, but for now use a simple script to scrape the metrics on the RootLB
	resp, err := p.client.Output(resTrackerCmd)
	if err != nil {
		log.DebugLog(log.DebugLevelMetrics, "Failed to run", "cmd", resTrackerCmd, "err", err.Error())
		return err
	}
	if err = json.Unmarshal([]byte(resp), &p.ClusterMetrics); err != nil {
		log.DebugLog(log.DebugLevelMetrics, "Failed to marshal machine metrics", "stats", resp, "err", err.Error())
		return err
	}
	// set timestamps to current time
	p.CpuTS, _ = types.TimestampProto(time.Now())
	p.MemTS, p.DiskTS, p.NetSentTS, p.NetRecvTS, p.TcpConnsTS, p.TcpRetransTS, p.UdpSentTS, p.UdpRecvTS, p.UdpRecvErrTS = p.CpuTS, p.CpuTS, p.CpuTS, p.CpuTS, p.CpuTS, p.CpuTS, p.CpuTS, p.CpuTS, p.CpuTS
	return nil
}
