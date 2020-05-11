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

var CloudletPrometheusContainer = "cloudletPrometheus"

var prometheusConfig = `rule_files:
- "/tmp/prom_rules.yml"
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

// TODO ---
//func getCloudletPrometheusProc(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (*Shepherd, []process.StartOp, error) {
//}

// TODO - get a process
func GetCloudletPrometheusCmd(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (string, *map[string]string, error) {
	envVars := make(map[string]string)
	return "", &envVars, nil
}

// Starts prometheus container and connects it to the default ports
func StartCloudletPrometheus(ctx context.Context, cloudlet *edgeproto.Cloudlet) error {
	cfgFile := "/tmp/prometheus.yml"
	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(prometheusConfig)
	if err != nil {
		return err
	}

	// label with a cloudlet name and org
	cloudletName := util.DockerSanitize(cloudlet.Key.Name)
	cloudletOrg := util.DockerSanitize(cloudlet.Key.Organization)

	args := []string{
		"run", "--rm", "--name", CloudletPrometheusContainer,
		"-l", "cloudlet=" + cloudletName,
		"-l", "cloudletorg=" + cloudletOrg,
		"-p", "9092:9090", // container interface
		"-v", "/tmp:/tmp",
		"-v", cfgFile + ":/etc/prometheus/prometheus.yml",
		"prom/prometheus:latest",
		"--config.file=/etc/prometheus/prometheus.yml",
	}
	_, err = process.StartLocal(CloudletPrometheusContainer, "docker", args, nil, "/tmp/cloudlet_prometheus.log")
	if err != nil {
		return err
	}
	return nil
}

func StopCloudletPrometheus(ctx context.Context, cloudlet *edgeproto.Cloudlet) error {
	// label with a cloudlet name and org
	cloudletName := util.DockerSanitize(cloudlet.Key.Name)
	cloudletOrg := util.DockerSanitize(cloudlet.Key.Organization)

	// kill the container that matches the cloudlet
	cmd := exec.Command("docker", "ps", "-q",
		"--filter", "label=cloudlet="+cloudletName,
		"--filter", "label=cloudletorg="+cloudletOrg,
		"--filter", "name="+CloudletPrometheusContainer)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if string(out) == "" {
		log.SpanLog(ctx, log.DebugLevelInfo, "StopCloudletPormetheus no container existst for cloudlet",
			"cloudlet", cloudlet)
		return nil
	}
	// kill the container we found
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd = exec.Command("docker", "kill", string(out))
	err = cmd.Run()
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfo, "Failed to kill prometheus docker container", "command", cmd, "error", err, "output", output)
		return err
	}
	return nil
}

func StopAllCloudletPrometheus(ctx context.Context) error {
	cmd := exec.Command("docker", "kill", CloudletPrometheusContainer)
	cmd.Run()
	return nil
}

func CloudletPrometheusExists(ctx context.Context) bool {
	cmd := exec.Command("docker", "logs", CloudletPrometheusContainer)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil && strings.Contains(out.String(), "No such container") {
		return false
	}
	return true
}
