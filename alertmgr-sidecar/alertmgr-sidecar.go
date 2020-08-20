package main

import (
	"flag"
	"fmt"

	"os"
	"os/signal"

	"github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr"
	"github.com/mobiledgex/edge-cloud/log"
)

var debugLevels = flag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
var alertmanagerAddr = flag.String("alertmgrAddr", "0.0.0.0:9093", "Alertmanager address")
var alertmanagerConfigFile = flag.String("configFile", "/tmp/alertmanager.yml", "Alertmanager config file")
var httpAddr = flag.String("httpAddr", "0.0.0.0:9094", "Http API endpoint")

var SidevarServer *alertmgr.SidecarServer

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer("")

	SidevarServer = alertmgr.NewSidecarServer(*alertmanagerAddr, *alertmanagerConfigFile, *httpAddr)
	err := SidevarServer.Run()
	if err != nil {
		log.FatalLog("Unable to start alertmgr sidecar", "err", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)

}
