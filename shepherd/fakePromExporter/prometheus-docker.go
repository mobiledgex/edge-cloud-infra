package fakepromexporter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
)

// tools for starting and stopping prometheus in docker containers made to scrape the fakeExporter endpoint for e2e testing

var imageName = "e2e-prom"
var promPort = cloudcommon.PrometheusPort
var dockerRun *exec.Cmd
var dockerError = "docker: Error response from daemon"

func StartPromContainer(ctx context.Context) error {
	// if the image doesnt exist, build it
	if !imageFound(imageName) {
		log.SpanLog(ctx, log.DebugLevelMexos, "Prometheus image not found, building it myself...")
		directory := os.Getenv("GOPATH") + "/src/github.com/mobiledgex/edge-cloud-infra/shepherd/fakePromExporter"
		builder := exec.Command("docker", "build", "-t", imageName, directory)
		err := builder.Run()
		if err != nil {
			return fmt.Errorf("Failed to build docker image for e2e prometheus: %v", err)
		}
	}

	dockerRun = exec.Command("docker", "run", "--rm", "-p", fmt.Sprintf("%d:%d", promPort, promPort), "--name", imageName, imageName)
	// see if the command docker command failed due to e2e-prom already being started by another shepherd
	stderr, err := dockerRun.StderrPipe()
	if err != nil {
		return fmt.Errorf("Could not pipe stdout of docker run command: %v", err)
	}
	err = dockerRun.Start()
	if err != nil {
		return fmt.Errorf("Failed to start docker: %v", err)
	}
	buf := make([]byte, len(dockerError))
	stderr.Read(buf)
	fmt.Printf("readstr: %s\n", string(buf))
	if strings.Contains(string(buf), dockerError) {
		dockerRun.Wait()
		return fmt.Errorf("Failed to run Prometheus container")
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "Prometheus container started")
	return nil
}

func imageFound(name string) bool {
	listCmd := exec.Command("docker", "images")
	output, err := listCmd.Output()
	if err != nil {
		return false
	}
	imageList := strings.Split(string(output), "\n")
	for _, row := range imageList {
		if imageName == strings.SplitN(row, " ", 2)[0] {
			return true
		}
	}
	return false
}

func StopPromContainer() error {
	err := exec.Command("docker", "stop", imageName).Run()
	// dockerRun.Wait()
	return err
}
