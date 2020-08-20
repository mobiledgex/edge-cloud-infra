package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/mobiledgex/edge-cloud-infra/mc/orm/alertmgr"
)

var alertmanagerAddr = flag.String("alertmgrAddr", "0.0.0.0:9093", "Alertmanager address")
var alertmanagerConfigFile = flag.String("configFile", "/etc/prometheus/alertmanager.yml", "Alertmanager config file")
var httpAddr = flag.String("httpAddr", "0.0.0.0:9094", "Http API endpoint")

var SidevarServer *alertmgr.SidecarServer

func main() {
	flag.Parse()

	SidevarServer = alertmgr.NewSidecarServer(*alertmanagerAddr, *alertmanagerConfigFile, *httpAddr)
	err := SidevarServer.Run()
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)

}
