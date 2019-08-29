package mexdind

import (
	"context"
	"fmt"
	"os"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
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
	ctx           context.Context
	generic       dind.Platform
	config        platform.PlatformConfig
	NetworkScheme string
}

func (s *Platform) SetContext(ctx context.Context) {
	s.ctx = ctx
	s.generic.SetContext(ctx)
}

func (s *Platform) GetType() string {
	return "mexdind"
}

func (s *Platform) Init(platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	err := s.generic.Init(platformConfig, updateCallback)
	s.config = *platformConfig
	if err != nil {
		return err
	}

	// Set the test Mode based on what is in PlatformConfig
	mexos.SetTestMode(platformConfig.TestMode)

	if err := mexos.InitInfraCommon(s.ctx, platformConfig.VaultAddr); err != nil {
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

	fqdn := cloudcommon.GetRootLBFQDN(platformConfig.CloudletKey)
	ipaddr, err := s.GetDINDServiceIP()
	if err != nil {
		return fmt.Errorf("init cannot get service ip, %s", err.Error())
	}
	if mexos.GetCloudletNetworkScheme() == cloudcommon.NetworkSchemePublicIP {
		if err := mexos.ActivateFQDNA(s.ctx, fqdn, ipaddr); err != nil {
			log.SpanLog(s.ctx, log.DebugLevelMexos, "error in ActivateFQDNA", "err", err)
			return err
		}
	}
	log.SpanLog(s.ctx, log.DebugLevelMexos, "done init mexdind")
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	return s.generic.GatherCloudletInfo(info)
}

func (s *Platform) GetPlatformClient(clusterInst *edgeproto.ClusterInst) (pc.PlatformClient, error) {
	return s.generic.GetPlatformClient(clusterInst)
}
