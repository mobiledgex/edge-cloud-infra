package openstack

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

const MINIMUM_DISK_SIZE uint64 = 20

type Platform struct {
	rootLBName  string
	rootLB      *MEXRootLB
	cloudletKey *edgeproto.CloudletKey
	flavorList  []*edgeproto.FlavorInfo
	config      platform.PlatformConfig
	vaultConfig *vault.Config
	openRCVars  map[string]string
	commonPf    mexos.CommonPlatform
	envVars     map[string]*mexos.PropertyInfo
	authKey     *edgeproto.AuthKeyPair
}

func (s *Platform) GetType() string {
	return "openstack"
}

// GetVMSpecForRootLB gets the VM spec for the rootLB when it is not specified within a cluster. This is
// used for Shared RootLb and for VM app based RootLb
func (s *Platform) GetVMSpecForRootLB() (*vmspec.VMCreationSpec, error) {

	var rootlbFlavor edgeproto.Flavor
	err := s.GetCloudletSharedRootLBFlavor(&rootlbFlavor)
	if err != nil {
		return nil, fmt.Errorf("unable to get Shared RootLB Flavor: %v", err)
	}
	vmspec, err := vmspec.GetVMSpec(s.flavorList, rootlbFlavor)
	if err != nil {
		return nil, fmt.Errorf("unable to find VM spec for Shared RootLB: %v", err)
	}
	if vmspec.AvailabilityZone == "" {
		vmspec.AvailabilityZone = s.GetCloudletComputeAvailabilityZone()
	}
	return vmspec, nil
}

func (s *Platform) Init(ctx context.Context, cloudlet *edgeproto.Cloudlet, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	rootLBName := cloudcommon.GetRootLBName(&cloudlet.Key)
	s.cloudletKey = &cloudlet.Key
	s.config = *platformConfig
	log.SpanLog(ctx,
		log.DebugLevelMexos, "init openstack",
		"rootLB", rootLBName,
		"physicalName", platformConfig.PhysicalName,
		"vaultAddr", platformConfig.VaultAddr)

	updateCallback(edgeproto.UpdateTask, "Initializing Openstack platform")

	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	s.vaultConfig = vaultConfig
	log.SpanLog(ctx, log.DebugLevelMexos, "vault auth", "type", vaultConfig.Auth.Type())

	s.authKey = cloudlet.AuthKey

	updateCallback(edgeproto.UpdateTask, "Fetching Openstack access credentials")
	if err := s.commonPf.InitInfraCommon(ctx, vaultConfig, cloudlet.EnvVar); err != nil {
		return err
	}

	if err := s.InitOpenstackProps(ctx, cloudlet, platformConfig.Region, vaultConfig); err != nil {
		return err
	}

	s.flavorList, _, _, err = s.GetFlavorInfo(ctx)
	if err != nil {
		return err
	}

	// create rootLB
	updateCallback(edgeproto.UpdateTask, "Creating RootLB")
	crmRootLB, cerr := NewRootLB(ctx, rootLBName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "created rootLB", "rootlb", crmRootLB.Name)
	s.rootLB = crmRootLB
	s.rootLBName = rootLBName

	vmspec, err := s.GetVMSpecForRootLB()
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelMexos, "calling SetupRootLB")
	updateCallback(edgeproto.UpdateTask, "Setting up RootLB")
	err = s.SetupRootLB(ctx, rootLBName, vmspec, &cloudlet.Key, platformConfig.CloudletVMImagePath, platformConfig.VMImageVersion, edgeproto.DummyUpdateCallback)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "ok, SetupRootLB")

	// set up L7 load balancer
	client, err := s.GetPlatformClient(ctx, rootLBName)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Setting up Proxy")
	err = proxy.InitL7Proxy(ctx, client, proxy.WithDockerNetwork("host"))
	if err != nil {
		return err
	}
	return nil
}

func (s *Platform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	return s.OSGetLimits(ctx, info)
}

func (s *Platform) GetPlatformClient(ctx context.Context, serverName string) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelMexos, "GetPlatformClient", "serverName", serverName)

	if serverName == "" {
		return nil, fmt.Errorf("cannot GetPlatformClient, serverName is empty")
	}
	if s.GetCloudletExternalNetwork() == "" {
		return nil, fmt.Errorf("GetPlatformClient, missing external network in platform config")
	}
	return s.GetSSHClient(ctx, serverName, s.GetCloudletExternalNetwork(), mexos.SSHUser)
}

func (s *Platform) GetPlatformClientRootLB(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	rootLBName := s.rootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBName = cloudcommon.GetDedicatedLBFQDN(s.cloudletKey, &clusterInst.Key.ClusterKey)
	}
	return s.GetPlatformClient(ctx, rootLBName)
}
