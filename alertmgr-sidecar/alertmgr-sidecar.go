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
var alertmanagerConfigFile = flag.String("configFile", "/var/tmp/alertmanager.yml", "Alertmanager config file")
var httpAddr = flag.String("httpAddr", "0.0.0.0:9094", "Http API endpoint")
var tlsCert = flag.String("tlsCert", "", "server tls cert file.")
var tlsCertKey = flag.String("tlsCertKey", "", "server tls cert key file.")
var clientCert = flag.String("tlsClientCert", "", "client tls cert file")

var SidecarServer *alertmgr.SidecarServer

func getConfigInfo() (*alertmgr.AlertmgrInitInfo, error) {
	initInfo := alertmgr.AlertmgrInitInfo{
		Email:          os.Getenv("ALERTMANAGER_SMTP_EMAIL"),
		User:           os.Getenv("ALERTMANAGER_SMTP_USER"),
		Token:          os.Getenv("ALERTMANAGER_SMTP_TOKEN"),
		Smtp:           os.Getenv("ALERTMANAGER_SMTP_SERVER"),
		Port:           os.Getenv("ALERTMANAGER_SMTP_SERVER_PORT"),
		Tls:            os.Getenv("ALERTMANAGER_SMTP_SERVER_TLS"),
		ResolveTimeout: os.Getenv("ALERTMANAGER_RESOLVE_TIMEOUT"),
	}
	// if smtp server and username are not set, environment is invalid
	if initInfo.Smtp == "" || initInfo.Email == "" {
		return nil, fmt.Errorf("Invalid environment %v\n", initInfo)
	}
	if initInfo.ResolveTimeout == "" {
		// default 5m
		initInfo.ResolveTimeout = "5m"
	}
	if initInfo.Port == "" {
		// default to 587 and TLS
		initInfo.Port = "587"
	}
	if initInfo.Tls == "" {
		// default to true
		initInfo.Tls = "true"
	}
	if initInfo.Tls != "true" && initInfo.Tls != "false" {
		return nil, fmt.Errorf("ALERTMANAGER_SMTP_SERVER_TLS must be either \"true\", or \"false\"")
	}
	return &initInfo, nil
}

func main() {
	flag.Parse()
	log.SetDebugLevelStrs(*debugLevels)
	log.InitTracer("")

	config, err := getConfigInfo()
	if err != nil {
		log.FatalLog("No default init info for alertmgr sidecar server is found", "err", err)
	}

	SidecarServer, err := alertmgr.NewSidecarServer(*alertmanagerAddr, *alertmanagerConfigFile,
		*httpAddr, config, *clientCert, *tlsCert, *tlsCertKey)
	if err != nil {
		log.FatalLog("Unable to init alertmgr sidecar", "err", err)
	}
	err = SidecarServer.Run()
	if err != nil {
		log.FatalLog("Unable to start alertmgr sidecar", "err", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// wait until process in killed/interrupted
	sig := <-sigChan
	fmt.Println(sig)
}
