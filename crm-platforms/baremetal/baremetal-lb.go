package baremetal

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
	yaml "gopkg.in/yaml.v2"
)

var LbInfoDoesNotExist string = "LB info does not exist"
var LbConfigDir = "lbconfig"

type LbInfo struct {
	Name            string
	ExternalIpAddr  string
	InternalIpAddr  string
	LbListenDevName string
}

func (b *BareMetalPlatform) GetSharedLBName(ctx context.Context, key *edgeproto.CloudletKey) string {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSharedLBName", "key", key)
	name := cloudcommon.GetRootLBFQDN(key, b.commonPf.PlatformConfig.AppDNSRoot)
	return name
}

func (b *BareMetalPlatform) GetLbNameForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	lbName := b.sharedLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName = cloudcommon.GetDedicatedLBFQDN(b.commonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey, b.commonPf.PlatformConfig.AppDNSRoot)
	}
	return lbName
}

func (b *BareMetalPlatform) getLbConfigFile(ctx context.Context, lbname string) string {
	return LbConfigDir + "/" + lbname + "-lbconfig.yml"
}

func (b *BareMetalPlatform) GetLbInfo(ctx context.Context, client ssh.Client, lbname string) (*LbInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetLbInfo", "name", lbname)
	lfinfoFile := b.getLbConfigFile(ctx, lbname)
	out, err := client.Output("cat " + lfinfoFile)
	if err != nil {
		if strings.Contains(out, "No such file") {
			return nil, fmt.Errorf("LbInfoDoesNotExist")
		}
		return nil, fmt.Errorf("error getting lbinfo: %v", err)
	}
	var lbInfo LbInfo
	err = yaml.Unmarshal([]byte(out), &lbInfo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Unmarshal fail", "out", out, "err", err)
		return nil, fmt.Errorf("Unmarshal failed for lbinfo- %v", err)
	}
	return &lbInfo, nil
}

func (b *BareMetalPlatform) DeleteLbInfo(ctx context.Context, client ssh.Client, lbname string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteLbInfo", "lbname", lbname)
	lfinfoFile := b.getLbConfigFile(ctx, lbname)
	out, err := client.Output("rm -f " + lfinfoFile)
	if err != nil {
		if !strings.Contains(out, "No such file") {
			return fmt.Errorf("Error deleting lbinfo")
		}
	}
	return nil
}

func (b *BareMetalPlatform) SetupLb(ctx context.Context, client ssh.Client, lbname string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupLb", "lbname", lbname)
	lfinfoFile := b.getLbConfigFile(ctx, lbname)

	// see if file exists
	out, err := client.Output("ls " + lfinfoFile)
	if err != nil {
		if !strings.Contains(out, "No such file") {
			return fmt.Errorf("Unexpected error listing lbinfo file: %s - %v", out, err)
		}
		// create dir, it may aleady exist in which case do not overwrite
		log.SpanLog(ctx, log.DebugLevelInfra, "creating directory for LB", "LbConfigDir", LbConfigDir)
		err := pc.CreateDir(ctx, client, LbConfigDir, pc.NoOverwrite)
		if err != nil {
			return fmt.Errorf("Unable to create LB Dir: %s - %v", LbConfigDir, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "New LB, assign free IP")
		dev, externalIp, internalIp, err := b.AssignFreeLbIp(ctx, client)
		if err != nil {
			return err
		}
		lbInfo := LbInfo{
			Name:            lbname,
			ExternalIpAddr:  externalIp,
			InternalIpAddr:  internalIp,
			LbListenDevName: dev,
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "creating lbinfo file", "lbdir", lfinfoFile)
		lbYaml, err := yaml.Marshal(&lbInfo)
		if err != nil {
			return fmt.Errorf("Unable to marshal LB info - %v", err)
		}
		err = pc.WriteFile(client, lfinfoFile, string(lbYaml), "lbinfo", pc.NoSudo)
		if err != nil {
			return fmt.Errorf("Unable to create LB info file %v", err)
		}
		if err = b.commonPf.ActivateFQDNA(ctx, lbname, externalIp); err != nil {
			return err
		}
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "LBInfo file already exists")
	}
	return nil
}

func (b *BareMetalPlatform) GetRootLBFlavor(ctx context.Context) (*edgeproto.Flavor, error) {
	return &edgeproto.Flavor{
		Vcpus: uint64(0),
		Ram:   uint64(0),
		Disk:  uint64(0),
	}, nil
}
