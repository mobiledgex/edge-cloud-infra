package vmlayer

import (
	"context"
	"fmt"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"

	ssh "github.com/mobiledgex/golang-ssh"
)

// VMProvider is an interface that platforms implement to perform the details of interfacing with the orchestration layer
type VMProvider interface {
	NameSanitize(string) string
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
	CreateVMs(ctx context.Context, vmGroupOrchestrationParams *VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error
	UpdateVMs(ctx context.Context, vmGroupOrchestrationParams *VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error
	DeleteVMs(ctx context.Context, vmGroupName string) error
}

// VMPlatformProvider embeds Platform and VMProvider
type VMPlatformProvider interface {
	platform.Platform
	VMProvider
}

// VMPlatform contains the needed by all VM based platforms
type VMPlatform struct {
	sharedRootLBName string
	sharedRootLB     *MEXRootLB
	vmProvider       VMPlatformProvider
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

func (v *VMPlatform) InitVMProvider(ctx context.Context, provider VMPlatformProvider, updateCallback edgeproto.CacheUpdateCallback) error {
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
	client, err := v.GetSSHClientForServer(ctx, v.sharedRootLBName, v.GetCloudletExternalNetwork())
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
