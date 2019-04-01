package openstack

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform/pc"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type Platform struct {
	rootLBName string
	rootLB     *mexos.MEXRootLB
}

func (s *Platform) GetType() string {
	return "openstack"
}

func (s *Platform) Init(key *edgeproto.CloudletKey) error {
	rootLBName := cloudcommon.GetRootLBFQDN(key)
	log.DebugLog(log.DebugLevelMexos, "init openstack", "rootLB", rootLBName)

	if err := mexos.InitInfraCommon(); err != nil {
		return err
	}
	if err := mexos.InitOpenstackProps(); err != nil {
		return err
	}
	mexos.CloudletInfraCommon.NetworkScheme = "priv-subnet,mex-k8s-net-1,10.101.X.0/24"

	// create rootLB
	crmRootLB, cerr := mexos.NewRootLB(rootLBName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	log.DebugLog(log.DebugLevelMexos, "created rootLB", "rootlb", crmRootLB.Name)
	s.rootLB = crmRootLB
	s.rootLBName = rootLBName

	log.DebugLog(log.DebugLevelMexos, "calling RunMEXAgentCloudletKey", "cloudletkeystr", key.GetKeyString())
	err := mexos.RunMEXAgentCloudletKey(rootLBName, key.GetKeyString())
	if err != nil {
		return err
	}
	log.DebugLog(log.DebugLevelMexos, "ok, RunMEXAgentCloudletKey with cloudlet key")
	return nil
}

func (s *Platform) GatherCloudletInfo(info *edgeproto.CloudletInfo) error {
	return mexos.OSGetLimits(info)
}

func (s *Platform) GetPlatformClient(rootLBName string) (pc.PlatformClient, error) {
	if rootLBName == "" {
		return nil, fmt.Errorf("cannot validate kubernetes parameters, rootLB is empty")
	}
	if mexos.GetCloudletExternalNetwork() == "" {
		return nil, fmt.Errorf("validate kubernetes parameters, missing external network in platform config")
	}
	return mexos.GetSSHClient(rootLBName, mexos.GetCloudletExternalNetwork(), mexos.SSHUser)
}
