package vmlayer

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"

	ssh "github.com/mobiledgex/golang-ssh"
)

// VMProvider is an interface that platforms implement to perform the details of interfacing with the orchestration layer
type VMProvider interface {
	NameSanitize(string) string
	SetVMProperties(vmProperties *VMProperties)
	InitProvider(ctx context.Context) error
	GetFlavorList(ctx context.Context) ([]*edgeproto.FlavorInfo, error)
	AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error)
	AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error
	GetServerDetail(ctx context.Context, serverName string) (*ServerDetail, error)
	GetConsoleUrl(ctx context.Context, serverName string) (string, error)
	GetIPFromServerName(ctx context.Context, networkName, serverName string) (*ServerIP, error)
	GetClusterMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error)
	AttachPortToServer(ctx context.Context, serverName, portName string) error
	DetachPortFromServer(ctx context.Context, serverName, portName string) error
	AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error
	WhitelistSecurityRules(ctx context.Context, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error
	RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error
	GetResourceID(ctx context.Context, resourceType ResourceType, resourceName string) (string, error)
	InitApiAccessProperties(ctx context.Context, key *edgeproto.CloudletKey, region, physicalName string, vaultConfig *vault.Config, vars map[string]string) error
	VerifyApiEndpoint(ctx context.Context, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error
	SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error
	SetPowerState(ctx context.Context, serverName, serverAction string) error
	GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error
	GetCloudletManifest(ctx context.Context, name string, VMGroupOrchestrationParams *VMGroupOrchestrationParams) (string, error)
	GetRouterDetail(ctx context.Context, routerName string) (*RouterDetail, error)
	CreateVMs(ctx context.Context, vmGroupOrchestrationParams *VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error
	UpdateVMs(ctx context.Context, vmGroupOrchestrationParams *VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error
	DeleteVMs(ctx context.Context, vmGroupName string) error
}

// VMPlatform contains the needed by all VM based platforms
type VMPlatform struct {
	Type         string
	VMProvider   VMProvider
	VMProperties VMProperties
	FlavorList   []*edgeproto.FlavorInfo
}

// ResourceType is not exhaustive list, currently only ResourceTypeSecurityGroup is needed
type ResourceType string

const (
	ResourceTypeVM            ResourceType = "VM"
	ResourceTypeSubnet        ResourceType = "Subnet"
	ResourceTypeSecurityGroup ResourceType = "SecGrp"
)

const (
	VMProviderOpenstack string = "openstack"
	VMProviderVSphere   string = "vsphere"
)

type StringSanitizer func(value string) string

// VMPlatform embeds Platform and VMProvider

func (v *VMPlatform) GetType() string {
	return v.Type
}

func (v *VMPlatform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	rootLBName := v.VMProperties.sharedRootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBName = cloudcommon.GetDedicatedLBFQDN(v.VMProperties.CommonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey)
	}
	return v.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: rootLBName})
}

func (v *VMPlatform) GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNodePlatformClient", "node", node)

	if node == nil || node.Name == "" {
		return nil, fmt.Errorf("cannot GetNodePlatformClient, as node details are empty")
	}
	if v.VMProperties.GetCloudletExternalNetwork() == "" {
		return nil, fmt.Errorf("GetNodePlatformClient, missing external network in platform config")
	}
	return v.GetSSHClientForServer(ctx, node.Name, v.VMProperties.GetCloudletExternalNetwork())
}

func (v *VMPlatform) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst) ([]edgeproto.CloudletMgmtNode, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "ListCloudletMgmtNodes", "clusterInsts", clusterInsts)
	mgmt_nodes := []edgeproto.CloudletMgmtNode{
		edgeproto.CloudletMgmtNode{
			Type: "platformvm",
			Name: v.GetPlatformVMName(v.VMProperties.CommonPf.PlatformConfig.CloudletKey),
		},
		edgeproto.CloudletMgmtNode{
			Type: "sharedrootlb",
			Name: v.VMProperties.sharedRootLBName,
		},
	}
	for _, clusterInst := range clusterInsts {
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			mgmt_nodes = append(mgmt_nodes, edgeproto.CloudletMgmtNode{
				Type: "dedicatedrootlb",
				Name: cloudcommon.GetDedicatedLBFQDN(v.VMProperties.CommonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey),
			})
		}
	}
	return mgmt_nodes, nil
}

func (v *VMPlatform) InitProps(ctx context.Context, platformConfig *platform.PlatformConfig, vaultConfig *vault.Config) error {
	err := v.VMProperties.CommonPf.InitInfraCommon(ctx, platformConfig, VMProviderProps, vaultConfig)
	if err != nil {
		return err
	}
	v.VMProvider.SetVMProperties(&v.VMProperties)
	return nil
}

func (v *VMPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx,
		log.DebugLevelInfra, "Init VMPlatform",
		"physicalName", platformConfig.PhysicalName,
		"vaultAddr", platformConfig.VaultAddr,
		"type",
		v.Type)

	updateCallback(edgeproto.UpdateTask, "Initializing VM platform type: "+v.Type)

	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "vault auth", "type", vaultConfig.Auth.Type())

	if err := v.InitProps(ctx, platformConfig, vaultConfig); err != nil {
		return err
	}

	updateCallback(edgeproto.UpdateTask, "Fetching API Access access credentials")
	if err := v.VMProvider.InitApiAccessProperties(ctx, platformConfig.CloudletKey, platformConfig.Region, platformConfig.PhysicalName, vaultConfig, platformConfig.EnvVars); err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "doing init provider")
	if err := v.VMProvider.InitProvider(ctx); err != nil {
		return err
	}
	v.FlavorList, err = v.VMProvider.GetFlavorList(ctx)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "got flavor list", "flavorList", v.FlavorList)

	v.VMProperties.sharedRootLBName = v.GetRootLBName(v.VMProperties.CommonPf.PlatformConfig.CloudletKey)

	// create rootLB
	crmRootLB, cerr := v.NewRootLB(ctx, v.VMProperties.sharedRootLBName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	v.VMProperties.sharedRootLB = crmRootLB
	log.SpanLog(ctx, log.DebugLevelInfra, "created shared rootLB", "name", v.VMProperties.sharedRootLBName)

	v.CreateRootLB(ctx, crmRootLB, v.VMProperties.CommonPf.PlatformConfig.CloudletKey, v.VMProperties.CommonPf.PlatformConfig.CloudletVMImagePath, v.VMProperties.CommonPf.PlatformConfig.VMImageVersion, updateCallback)

	log.SpanLog(ctx, log.DebugLevelInfra, "calling SetupRootLB")
	updateCallback(edgeproto.UpdateTask, "Setting up RootLB")
	err = v.SetupRootLB(ctx, v.VMProperties.sharedRootLBName, v.VMProperties.CommonPf.PlatformConfig.CloudletKey, updateCallback)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "ok, SetupRootLB")

	// set up L7 load balancer
	client, err := v.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: v.VMProperties.sharedRootLBName})
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
