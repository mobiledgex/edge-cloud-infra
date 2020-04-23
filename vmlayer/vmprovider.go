package vmlayer

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"

	ssh "github.com/mobiledgex/golang-ssh"
)

// VMProvider is an interface that platforms implement to perform the details of interfacing with the orchestration layer
type VMProvider interface {
	NameSanitize(string) string
	GetType() string
	AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error)
	AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error
	GetServerDetail(ctx context.Context, serverName string) (*ServerDetail, error)
	GetIPFromServerName(ctx context.Context, networkName, serverName string) (*ServerIP, error)
	GetClusterMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error)
	AttachPortToServer(ctx context.Context, serverName, portName string) error
	DetachPortFromServer(ctx context.Context, serverName, portName string) error
	AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error
	NetworkSetupForRootLB(ctx context.Context, client ssh.Client, rootLBName string) error
	WhitelistSecurityRules(ctx context.Context, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error
	RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error
	GetResourceID(ctx context.Context, resourceType ResourceType, resourceName string) (string, error)
	VerifyApiEndpoint(ctx context.Context, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error
	GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error)
	CreateVMs(ctx context.Context, vmGroupOrchestrationParams *VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error
	UpdateVMs(ctx context.Context, vmGroupOrchestrationParams *VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error
	DeleteVMs(ctx context.Context, vmGroupName string) error
}

// VMPlatform contains the needed by all VM based platforms
type VMPlatform struct {
	sharedRootLBName string
	sharedRootLB     *MEXRootLB
	vmProvider       VMProvider
	FlavorList       []*edgeproto.FlavorInfo
	CommonPf         infracommon.CommonPlatform
}

// ResourceType is not exhaustive list, currently only ResourceTypeSecurityGroup is needed
type ResourceType string

const (
	ResourceTypeVM            ResourceType = "VM"
	ResourceTypeSubnet        ResourceType = "Subnet"
	ResourceTypeSecurityGroup ResourceType = "SecGrp"
)

type StringSanitizer func(value string) string

// VMPlatform embeds Platform and VMProvider

func (v *VMPlatform) InitVMProvider(ctx context.Context, provider VMProvider, updateCallback edgeproto.CacheUpdateCallback) error {
	updateCallback(edgeproto.UpdateTask, "InitVMProvider")

	v.vmProvider = provider
	v.sharedRootLBName = v.GetRootLBName(v.CommonPf.PlatformConfig.CloudletKey)

	// create rootLB
	crmRootLB, cerr := v.NewRootLB(ctx, v.sharedRootLBName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	v.sharedRootLB = crmRootLB
	log.SpanLog(ctx, log.DebugLevelInfra, "created shared rootLB", "name", v.sharedRootLBName)

	v.CreateRootLB(ctx, crmRootLB, v.CommonPf.PlatformConfig.CloudletKey, v.CommonPf.PlatformConfig.CloudletVMImagePath, v.CommonPf.PlatformConfig.VMImageVersion, updateCallback)

	log.SpanLog(ctx, log.DebugLevelInfra, "calling SetupRootLB")
	updateCallback(edgeproto.UpdateTask, "Setting up RootLB")
	err := v.SetupRootLB(ctx, v.sharedRootLBName, v.CommonPf.PlatformConfig.CloudletKey, edgeproto.DummyUpdateCallback)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "ok, SetupRootLB")

	// set up L7 load balancer
	client, err := v.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: v.sharedRootLBName})
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

func (v *VMPlatform) GetClusterPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	rootLBName := v.sharedRootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBName = cloudcommon.GetDedicatedLBFQDN(v.CommonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey)
	}
	return v.GetNodePlatformClient(ctx, &edgeproto.CloudletMgmtNode{Name: rootLBName})
}

func (v *VMPlatform) GetNodePlatformClient(ctx context.Context, node *edgeproto.CloudletMgmtNode) (ssh.Client, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetNodePlatformClient", "node", node)

	if node == nil || node.Name == "" {
		return nil, fmt.Errorf("cannot GetNodePlatformClient, as node details are empty")
	}
	if v.GetCloudletExternalNetwork() == "" {
		return nil, fmt.Errorf("GetNodePlatformClient, missing external network in platform config")
	}
	return v.GetSSHClientForServer(ctx, node.Name, v.GetCloudletExternalNetwork())
}

func (v *VMPlatform) ListCloudletMgmtNodes(ctx context.Context, clusterInsts []edgeproto.ClusterInst) ([]edgeproto.CloudletMgmtNode, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "ListCloudletMgmtNodes", "clusterInsts", clusterInsts)
	mgmt_nodes := []edgeproto.CloudletMgmtNode{
		edgeproto.CloudletMgmtNode{
			Type: "platformvm",
			Name: v.GetPlatformVMName(v.CommonPf.PlatformConfig.CloudletKey),
		},
		edgeproto.CloudletMgmtNode{
			Type: "sharedrootlb",
			Name: v.sharedRootLBName,
		},
	}
	for _, clusterInst := range clusterInsts {
		if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
			mgmt_nodes = append(mgmt_nodes, edgeproto.CloudletMgmtNode{
				Type: "dedicatedrootlb",
				Name: cloudcommon.GetDedicatedLBFQDN(v.CommonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey),
			})
		}
	}
	return mgmt_nodes, nil
}
