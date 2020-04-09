package openstack

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/proxy"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

type OpenstackPlatform struct {
	openRCVars map[string]string
	commonPf   infracommon.CommonPlatform
}

func (o *OpenstackPlatform) GetType() string {
	return "openstack"
}

func (o *OpenstackPlatform) Init(ctx context.Context, platformConfig *platform.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	sharedRootLbName := o.commonPf.GetRootLBName(platformConfig.CloudletKey)

	log.SpanLog(ctx,
		log.DebugLevelMexos, "init openstack",
		"physicalName", platformConfig.PhysicalName,
		"vaultAddr", platformConfig.VaultAddr)

	updateCallback(edgeproto.UpdateTask, "Initializing Openstack platform")

	vaultConfig, err := vault.BestConfig(platformConfig.VaultAddr)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "vault auth", "type", vaultConfig.Auth.Type())

	updateCallback(edgeproto.UpdateTask, "Fetching Openstack access credentials")
	if err := o.commonPf.InitInfraCommon(ctx, platformConfig, openstackProps, vaultConfig, o); err != nil {
		return err
	}

	if err := o.InitOpenstackProps(ctx, platformConfig.CloudletKey, platformConfig.Region, platformConfig.PhysicalName, vaultConfig, platformConfig.EnvVars); err != nil {
		return err
	}

	o.commonPf.FlavorList, _, _, err = o.GetFlavorInfo(ctx)
	if err != nil {
		return err
	}

	// create rootLB
	updateCallback(edgeproto.UpdateTask, "Creating RootLB")
	crmRootLB, cerr := o.commonPf.NewRootLB(ctx, o.commonPf.SharedRootLBName)
	if cerr != nil {
		return cerr
	}
	if crmRootLB == nil {
		return fmt.Errorf("rootLB is not initialized")
	}
	o.commonPf.SharedRootLBName = sharedRootLbName
	o.commonPf.SharedRootLB = crmRootLB
	log.SpanLog(ctx, log.DebugLevelMexos, "created shared rootLB", "name", crmRootLB.Name)

	vmspec, err := o.commonPf.GetVMSpecForRootLB()
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelMexos, "calling SetupRootLB")
	updateCallback(edgeproto.UpdateTask, "Setting up RootLB")
	err = o.commonPf.SetupRootLB(ctx, o.commonPf.SharedRootLBName, vmspec, platformConfig.CloudletKey, platformConfig.CloudletVMImagePath, platformConfig.VMImageVersion, edgeproto.DummyUpdateCallback)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "ok, SetupRootLB")

	// set up L7 load balancer
	client, err := o.commonPf.GetPlatformClientRootLB(ctx, o.commonPf.SharedRootLBName)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Setting up Proxy")
	err = proxy.InitL7Proxy(ctx, client, proxy.WithDockerNetwork("host"))
	if err != nil {
		return err
	}
	return o.PrepNetwork(ctx)
}

func (o *OpenstackPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	return o.OSGetLimits(ctx, info)
}

// alphanumeric plus -_. first char must be alpha, <= 255 chars.
func (o *OpenstackPlatform) NameSanitize(name string) string {
	r := strings.NewReplacer(
		" ", "",
		"&", "",
		",", "",
		"!", "")
	str := r.Replace(name)
	if str == "" {
		return str
	}
	if !unicode.IsLetter(rune(str[0])) {
		// first character must be alpha
		str = "a" + str
	}
	if len(str) > 255 {
		str = str[:254]
	}
	return str
}

func (o *OpenstackPlatform) GetPlatformClient(ctx context.Context, clusterInst *edgeproto.ClusterInst) (ssh.Client, error) {
	rootLBName := o.commonPf.SharedRootLBName
	if clusterInst.IpAccess == edgeproto.IpAccess_IP_ACCESS_DEDICATED {
		rootLBName = cloudcommon.GetDedicatedLBFQDN(o.commonPf.PlatformConfig.CloudletKey, &clusterInst.Key.ClusterKey)
	}
	return o.commonPf.GetPlatformClientRootLB(ctx, rootLBName)
}

func (o *OpenstackPlatform) DeleteResources(ctx context.Context, resourceGroupName string) error {
	return o.HeatDeleteStack(ctx, resourceGroupName)
}

func (o *OpenstackPlatform) GetServerDetail(ctx context.Context, serverName string) (*infracommon.ServerDetail, error) {
	var sd infracommon.ServerDetail

	osd, err := o.GetOpenstackServerDetails(ctx, serverName)
	if err != nil {
		return nil, err
	}
	sd.Name = osd.Name
	sd.ID = osd.ID
	sd.Status = osd.Status
	err = o.UpdateServerIPsFromAddrs(ctx, osd.Addresses, &sd)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "unable to update server IPs", "sd", sd, "err", err)
		return &sd, fmt.Errorf("unable to update server IPs -- %v", err)
	}
	return &sd, nil
}

func (o *OpenstackPlatform) CreateAppVM(ctx context.Context, vmAppParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (o *OpenstackPlatform) CreateAppVMWithRootLB(ctx context.Context, vmAppParams, vmLbParams *infracommon.VMParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatCreateAppVMWithRootLB(ctx, vmAppParams, vmLbParams, updateCallback)
}

func (o *OpenstackPlatform) CreateRootLBVM(ctx context.Context, serverName, stackName, imgName string, vmspec *vmspec.VMCreationSpec, cloudletKey *edgeproto.CloudletKey, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("not implemented")
}

func (o *OpenstackPlatform) CreateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatCreateCluster(ctx, clusterInst, privacyPolicy, rootLBName, imgName, dedicatedRootLB, updateCallback)
}

func (o *OpenstackPlatform) UpdateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error {
	return o.HeatUpdateCluster(ctx, clusterInst, privacyPolicy, rootLBName, imgName, dedicatedRootLB, updateCallback)
}

func (o *OpenstackPlatform) DeleteClusterResources(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst) error {
	return fmt.Errorf("not implemented")
}

func (o *OpenstackPlatform) Resync(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// UpdateServerIPsFromAddrs gets the ServerIPs forthe given network from the addresses provided
func (o *OpenstackPlatform) UpdateServerIPsFromAddrs(ctx context.Context, addresses string, serverDetail *infracommon.ServerDetail) error {

	log.SpanLog(ctx, log.DebugLevelMexos, "UpdateServerIPsFromAddrs", "addresses", addresses, "serverDetail", serverDetail)
	its := strings.Split(addresses, ";")

	for _, it := range its {
		var serverIP infracommon.ServerIP
		sits := strings.Split(it, "=")
		if len(sits) != 2 {
			return fmt.Errorf("GetServerIPFromAddrs: Unable to parse '%s'", it)
		}
		network := sits[0]
		serverIP.Network = network
		addr := sits[1]
		// the comma indicates a floating IP is present.
		if strings.Contains(addr, ",") {
			addrs := strings.Split(addr, ",")
			if len(addrs) == 2 {
				serverIP.InternalAddr = strings.TrimSpace(addrs[0])
				serverIP.ExternalAddr = strings.TrimSpace(addrs[1])
				serverIP.ExternalAddrIsFloating = true
			} else {
				return fmt.Errorf("GetServerExternalIPFromAddr: Unable to parse '%s'", addr)
			}
		} else {
			// no floating IP, internal and external are the same
			addr = strings.TrimSpace(addr)
			serverIP.InternalAddr = addr
			serverIP.ExternalAddr = addr
		}
		serverDetail.Addresses = append(serverDetail.Addresses, serverIP)
	}
	log.SpanLog(ctx, log.DebugLevelMexos, "Updated ServerIPS", "serverDetail", serverDetail)
	return nil
}
