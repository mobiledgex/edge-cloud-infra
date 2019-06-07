package mexdind

import (
	"fmt"
	"os"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/dind"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// mexdind wraps the generic dind implementation with
// mex-specific behavior, such as setting up DNS and
// registry.mobiledgex.net access secrets.

type Platform struct {
	generic       dind.Platform
	NetworkScheme string
	clusterCache  *edgeproto.ClusterInstInfoCache
}

func (s *Platform) GetType() string {
	return "mexdind"
}

func (s *Platform) Init(key *edgeproto.CloudletKey, physicalName, vaultAddr string, clusterCache *edgeproto.ClusterInstInfoCache) error {
	err := s.generic.Init(key, physicalName, vaultAddr, clusterCache)
	if err != nil {
		return err
	}
	s.clusterCache = clusterCache
	if err := mexos.InitInfraCommon(vaultAddr); err != nil {
		return err
	}

	s.NetworkScheme = os.Getenv("MEX_NETWORK_SCHEME")
	if s.NetworkScheme == "" {
		s.NetworkScheme = cloudcommon.NetworkSchemePrivateIP
	}
	if s.NetworkScheme != cloudcommon.NetworkSchemePrivateIP &&
		s.NetworkScheme != cloudcommon.NetworkSchemePublicIP {
		return fmt.Errorf("Unsupported network scheme for DIND: %s", s.NetworkScheme)
	}
	mexos.CloudletInfraCommon.NetworkScheme = s.NetworkScheme

	fqdn := cloudcommon.GetRootLBFQDN(key)
	ipaddr, err := s.GetDINDServiceIP()
	if err != nil {
		return fmt.Errorf("init cannot get service ip, %s", err.Error())
	}
	if mexos.GetCloudletNetworkScheme() == cloudcommon.NetworkSchemePublicIP {
		if err := mexos.ActivateFQDNA(fqdn, ipaddr); err != nil {
			log.DebugLog(log.DebugLevelMexos, "error in ActivateFQDNA", "err", err)
			return err
		}
	}
	log.DebugLog(log.DebugLevelMexos, "done init mexdind")
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	return s.generic.GatherCloudletInfo(info)
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return s.generic.GetPlatformClient(clusterInst)
}
