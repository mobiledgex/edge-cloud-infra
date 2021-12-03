package k8sbm

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

var LbInfoDoesNotExist string = "LB info does not exist"
var LbConfigDir = "lbconfig"

type LbInfo struct {
	Name            string
	ExternalIpAddr  string
	LbListenDevName string
}

// GetSharedLBName returns the "dedicated" FQDN of the default cluster
func (k *K8sBareMetalPlatform) GetSharedLBName(ctx context.Context, cloudletKey *edgeproto.CloudletKey) string {
	return cloudcommon.GetDedicatedLBFQDN(cloudletKey, &k.GetDefaultCluster(cloudletKey).Key.ClusterKey, k.commonPf.PlatformConfig.AppDNSRoot)
}

func (k *K8sBareMetalPlatform) GetLbName(ctx context.Context, appInst *edgeproto.AppInst) string {
	lbName := k.sharedLBName
	if appInst.DedicatedIp {
		return appInst.Uri
	}
	return lbName
}

func (k *K8sBareMetalPlatform) SetupLb(ctx context.Context, client ssh.Client, lbname string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupLb", "lbname", lbname)

	if true {
		// create dir, it may aleady exist in which case do not overwrite
		log.SpanLog(ctx, log.DebugLevelInfra, "creating directory for LB", "LbConfigDir", LbConfigDir)
		err := pc.CreateDir(ctx, client, LbConfigDir, pc.NoOverwrite)
		if err != nil {
			return fmt.Errorf("Unable to create LB Dir: %s - %v", LbConfigDir, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "New LB, assign free IP")
		externalIp, err := k.AssignFreeLbIp(ctx, lbname, client)
		if err != nil {
			return err
		}
		if err = k.commonPf.ActivateFQDNA(ctx, lbname, externalIp); err != nil {
			return err
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "lb ip already exists")
	}
	return nil
}

func (k *K8sBareMetalPlatform) GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error) {
	return &edgeproto.Flavor{
		Vcpus: uint64(0),
		Ram:   uint64(0),
		Disk:  uint64(0),
	}, nil
}
