package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mobiledgex/edge-cloud-infra/version"
	pfutils "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/utils"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/notify"
	"github.com/mobiledgex/edge-cloud/tls"
)

var notifyAddrs = flag.String("notifyAddrs", "127.0.0.1:50001", "Comma separated list of controller notify listener addresses")
var hostname = flag.String("hostname", "", "Unique hostname")
var debugLevels = flag.String("d", "", fmt.Sprintf("Comma separated list of %v", log.DebugLevelStrings))
var region = flag.String("region", "local", "region name")

var sigChan chan os.Signal
var nodeMgr node.NodeMgr
var mainStarted chan struct{}
var notifyClient *notify.Client

func main() {
	nodeMgr.InitFlags()
	nodeMgr.AccessKeyClient.InitFlags()
	flag.Parse()

	log.SetDebugLevelStrs(*debugLevels)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx, span, err := nodeMgr.Init(node.NodeTypeFRM, node.CertIssuerRegional,
		node.WithName(*hostname),
		node.WithRegion(*region),
	)
	if err != nil {
		log.FatalLog(err.Error())
	}
	nodeMgr.UpdateNodeProps(ctx, version.InfraBuildProps("Infra"))
	defer nodeMgr.Finish()

	// Load platform implementation.
	platform, err := pfutils.GetPlatform(ctx,
		edgeproto.PlatformType_PLATFORM_TYPE_FEDERATION.String(),
		nodeMgr.UpdateNodeProps)
	if err != nil {
		span.Finish()
		log.FatalLog(err.Error())
	}

	controllerData := NewControllerData(platform, &nodeMgr)

	// ctrl notify
	addrs := strings.Split(*notifyAddrs, ",")
	notifyClientTls, err := nodeMgr.InternalPki.GetClientTlsConfig(ctx,
		nodeMgr.CommonName(),
		node.CertIssuerRegional,
		[]node.MatchCA{node.SameRegionalMatchCA()})
	if err != nil {
		log.FatalLog(err.Error())
	}
	dialOption := tls.GetGrpcDialOption(notifyClientTls)
	notifyClient = notify.NewClient(nodeMgr.Name(), addrs, dialOption,
		notify.ClientUnaryInterceptors(nodeMgr.AccessKeyClient.UnaryAddAccessKey),
		notify.ClientStreamInterceptors(nodeMgr.AccessKeyClient.StreamAddAccessKey),
	)
	notifyClient.SetFilterByCloudletKey()
	notifyClient.SetFilterByFederatedCloudlet()
	InitClientNotify(notifyClient, &nodeMgr, controllerData)
	notifyClient.Start()
	defer notifyClient.Stop()

	span.Finish()
	if mainStarted != nil {
		// for unit testing to detect when main is ready
		close(mainStarted)
	}

	sig := <-sigChan
	fmt.Println(sig)
}
