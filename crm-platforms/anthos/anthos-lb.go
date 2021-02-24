package anthos

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

type LbInfo struct {
	Name            string
	ExternalIpAddr  string
	InternalIpAddr  string
	LbListenDevName string
}

func (a *AnthosPlatform) GetSharedLBName(ctx context.Context, key *edgeproto.CloudletKey) string {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetSharedLBName", "key", key)
	name := cloudcommon.GetRootLBFQDN(key, a.commonPf.PlatformConfig.AppDNSRoot)
	return name
}

func (a *AnthosPlatform) GetLbNameForCluster(ctx context.Context, clusterInst *edgeproto.ClusterInst) string {
	lbName := a.sharedLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		lbName = cloudcommon.GetDedicatedLBFQDN(a.commonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey, a.commonPf.PlatformConfig.AppDNSRoot)
	}
	return lbName
}

func (a *AnthosPlatform) GetDirForLb(ctx context.Context, name string) string {
	return a.GetConfigDir() + "/lbconfig-" + name
}

func (a *AnthosPlatform) GetLbInfo(ctx context.Context, client ssh.Client, name string) (*LbInfo, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetLbInfo", "name", name)
	lbdir := a.GetDirForLb(ctx, name)
	lfinfoFile := lbdir + "/lbinfo.yml"
	out, err := client.Output("cat " + lfinfoFile)
	if err != nil {
		if !strings.Contains(out, "No such file") {
			return nil, fmt.Errorf(LbInfoDoesNotExist)
		}
	}
	var lbInfo LbInfo
	err = yaml.Unmarshal([]byte(out), &lbInfo)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Unmarshal fail", "out", out, "err", err)
		return nil, fmt.Errorf("Unmarshal failed for lbinfo- %v", err)
	}
	return &lbInfo, nil
}

func (a *AnthosPlatform) SetupLb(ctx context.Context, client ssh.Client, name string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupLb", "name", name)
	lbdir := a.GetDirForLb(ctx, name)
	lfinfoFile := lbdir + "/lbinfo.yml"

	// see if file exists
	out, err := client.Output("ls " + lfinfoFile)
	if err != nil {
		if !strings.Contains(out, "No such file") {
			return fmt.Errorf("Unexpected error listing lbinfo file: %s - %v", out, err)
		}
		// create dir, it may already exist in which case do not overwrite
		log.SpanLog(ctx, log.DebugLevelInfra, "creating directory for LB", "lbdir", lbdir)
		err := pc.CreateDir(ctx, client, lbdir, pc.NoOverwrite)
		if err != nil {
			return fmt.Errorf("Unable to create LB Dir: %s - %v", lbdir, err)
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "New LB, assign free IP")
		dev, externalIp, internalIp, err := a.AssignFreeLbIp(ctx, client)
		if err != nil {
			return err
		}
		lbInfo := LbInfo{
			Name:            name,
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
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "LBInfo file already exists")
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupLb done")
	return nil
}
