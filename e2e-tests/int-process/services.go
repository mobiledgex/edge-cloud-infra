package intprocess

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

const (
	PrometheusContainer    = "cloudletPrometheus"
	PrometheusImagePath    = "prom/prometheus:latest"
	PormtheusRulesPrefix   = "rulefile_"
	CloudletPrometheusPort = "9092"
)

var prometheusConfig = `global:
  evaluation_interval: 15s
rule_files:
- "/tmp/` + PormtheusRulesPrefix + `*"
scrape_configs:
- job_name: envoy_targets
  scrape_interval: 5s
  file_sd_configs:
  - files:
    - '/tmp/prom_targets.json'
`

func getShepherdProc(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (*Shepherd, []process.StartOp, error) {
	opts := []process.StartOp{}

	cloudletKeyStr, err := json.Marshal(cloudlet.Key)
	if err != nil {
		return nil, opts, fmt.Errorf("unable to marshal cloudlet key")
	}

	envVars := make(map[string]string)
	notifyAddr := ""
	tlsCertFile := ""
	vaultAddr := ""
	span := ""
	region := ""
	useVaultCAs := false
	useVaultCerts := false
	if pfConfig != nil {
		// Same vault role-id/secret-id as CRM
		for k, v := range pfConfig.EnvVar {
			envVars[k] = v
		}
		notifyAddr = cloudlet.NotifySrvAddr
		tlsCertFile = pfConfig.TlsCertFile
		vaultAddr = pfConfig.VaultAddr
		span = pfConfig.Span
		region = pfConfig.Region
		useVaultCAs = pfConfig.UseVaultCas
		useVaultCerts = pfConfig.UseVaultCerts
	}

	for envKey, envVal := range cloudlet.EnvVar {
		envVars[envKey] = envVal
	}

	opts = append(opts, process.WithDebug("api,infra,notify,metrics"))

	return &Shepherd{
		NotifyAddrs: notifyAddr,
		CloudletKey: string(cloudletKeyStr),
		Platform:    cloudlet.PlatformType.String(),
		Common: process.Common{
			Hostname: cloudlet.Key.Name,
			EnvVars:  envVars,
		},
		TLS: process.TLSCerts{
			ServerCert: tlsCertFile,
		},
		VaultAddr:     vaultAddr,
		PhysicalName:  cloudlet.PhysicalName,
		Span:          span,
		Region:        region,
		UseVaultCAs:   useVaultCAs,
		UseVaultCerts: useVaultCerts,
	}, opts, nil
}

func GetShepherdCmd(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (string, *map[string]string, error) {
	ShepherdProc, opts, err := getShepherdProc(cloudlet, pfConfig)
	if err != nil {
		return "", nil, err
	}

	return ShepherdProc.String(opts...), &ShepherdProc.Common.EnvVars, nil
}

func StartShepherdService(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (*Shepherd, error) {
	shepherdProc, opts, err := getShepherdProc(cloudlet, pfConfig)
	if err != nil {
		return nil, err
	}

	err = shepherdProc.StartLocal("/tmp/"+cloudlet.Key.Name+".shepherd.log", opts...)
	if err != nil {
		return nil, err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "started "+shepherdProc.GetExeName())

	return shepherdProc, nil
}

func StopShepherdService(ctx context.Context, cloudlet *edgeproto.Cloudlet) error {
	args := ""
	if cloudlet != nil {
		ShepherdProc, _, err := getShepherdProc(cloudlet, nil)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "cannot stop Shepherdserver", "err", err)
			return err
		}
		args = util.EscapeJson(ShepherdProc.LookupArgs())
	}

	// max wait time for process to go down gracefully, after which it is killed forcefully
	maxwait := 1 * time.Second

	c := make(chan string)
	go process.KillProcessesByName("shepherd", maxwait, args, c)

	log.SpanLog(ctx, log.DebugLevelInfra, "stopped Shepherdserver", "msg", <-c)
	return nil
}

// Prometheus config is common for all the types of deployement
func GetCloudletPrometheusConfig() string {
	return prometheusConfig
}
func GetCloudletPrometheusConfigHostFilePath() string {
	return "/tmp/prometheus.yml"
}

// command line options for prometheus container
func GetCloudletPrometheusCmdArgs() []string {
	return []string{
		"--config.file=/etc/prometheus/prometheus.yml",
		"--web.listen-address=:" + CloudletPrometheusPort,
		"--web.enable-lifecycle",
	}
}

// base docker run args
func GetCloudletPrometheusDockerArgs(cloudlet *edgeproto.Cloudlet, cfgFile string) []string {

	// label with a cloudlet name and org
	cloudletName := util.DockerSanitize(cloudlet.Key.Name)
	cloudletOrg := util.DockerSanitize(cloudlet.Key.Organization)

	return []string{
		"-l", "cloudlet=" + cloudletName,
		"-l", "cloudletorg=" + cloudletOrg,
		"-p", CloudletPrometheusPort + ":" + CloudletPrometheusPort, // container interface
		"-v", "/tmp:/tmp",
		"-v", cfgFile + ":/etc/prometheus/prometheus.yml",
	}
}

// Starts prometheus container and connects it to the default ports
func StartCloudletPrometheus(ctx context.Context, cloudlet *edgeproto.Cloudlet) error {
	cfgFile := GetCloudletPrometheusConfigHostFilePath()
	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(GetCloudletPrometheusConfig())
	if err != nil {
		return err
	}

	args := GetCloudletPrometheusDockerArgs(cloudlet, cfgFile)
	cmdOpts := GetCloudletPrometheusCmdArgs()

	// local container specific options
	args = append([]string{"run", "--rm"}, args...)
	// set name and image path
	args = append(args, []string{"--name", PrometheusContainer, PrometheusImagePath}...)
	args = append(args, cmdOpts...)

	_, err = process.StartLocal(PrometheusContainer, "docker", args, nil, "/tmp/cloudlet_prometheus.log")
	if err != nil {
		return err
	}
	return nil
}

func StopCloudletPrometheus(ctx context.Context) error {
	cmd := exec.Command("docker", "kill", PrometheusContainer)
	cmd.Run()
	return nil
}

func CloudletPrometheusExists(ctx context.Context) bool {
	cmd := exec.Command("docker", "logs", PrometheusContainer)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil && strings.Contains(out.String(), "No such container") {
		return false
	}
	return true
}
