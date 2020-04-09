package infracommon

// This file stores a global cloudlet infra properties object. The long term solution is for the controller to send this via the
// notification channel when the cloudlet is provisioned.   The controller will do the vault access and pass this data down; this
// is a stepping stone to start using edgepro data strucures to hold info abou the cloudlet rather than custom types and so the vault
// is still directly accessed here as are env variable to populate some variables

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

const ServerDoesNotExistError string = "Server does not exist"

type InfraProvider interface {
	NameSanitize(string) string
	AddCloudletImageIfNotPresent(ctx context.Context, imgPathPrefix, imgVersion string, updateCallback edgeproto.CacheUpdateCallback) (string, error)
	AddAppImageIfNotPresent(ctx context.Context, app *edgeproto.App, updateCallback edgeproto.CacheUpdateCallback) error
	GetServerDetail(ctx context.Context, serverName string) (*ServerDetail, error)
	GetIPFromServerName(ctx context.Context, networkName, serverName string) (*ServerIP, error)
	GetClusterMasterNameAndIP(ctx context.Context, clusterInst *edgeproto.ClusterInst) (string, string, error)
	AttachPortToServer(ctx context.Context, serverName, portName string) error
	DetachPortFromServer(ctx context.Context, serverName, portName string) error
	AddSecurityRuleCIDRWithRetry(ctx context.Context, cidr string, proto string, group string, port string, serverName string) error
	CreateAppVM(ctx context.Context, vmAppParams *VMParams, updateCallback edgeproto.CacheUpdateCallback) error
	CreateAppVMWithRootLB(ctx context.Context, vmAppParams, vmLbParams *VMParams, updateCallback edgeproto.CacheUpdateCallback) error
	CreateRootLBVM(ctx context.Context, serverName, stackName, imgName string, vmspec *vmspec.VMCreationSpec, cloudletKey *edgeproto.CloudletKey, updateCallback edgeproto.CacheUpdateCallback) error
	CreateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error
	UpdateClusterVMs(ctx context.Context, clusterInst *edgeproto.ClusterInst, privacyPolicy *edgeproto.PrivacyPolicy, rootLBName string, imgName string, dedicatedRootLB bool, updateCallback edgeproto.CacheUpdateCallback) error
	DeleteClusterResources(ctx context.Context, client ssh.Client, clusterInst *edgeproto.ClusterInst) error
	DeleteResources(ctx context.Context, resourceGroupName string) error
	NetworkSetupForRootLB(ctx context.Context, client ssh.Client, rootLBName string) error
	WhitelistSecurityRules(ctx context.Context, secGrpName string, serverName string, allowedCIDR string, ports []dme.AppPort) error
	RemoveWhitelistSecurityRules(ctx context.Context, secGrpName string, allowedCIDR string, ports []dme.AppPort) error
	GetVMParams(ctx context.Context, depType DeploymentType, serverName, flavorName string, externalVolumeSize uint64, imageName, secGrp string, cloudletKey *edgeproto.CloudletKey, opts ...VMParamsOp) (*VMParams, error)
	Resync(ctx context.Context) error
}

// CommonInfraPlatform embeds Platform and InfraProvider
type CommonPlatformProvider interface {
	platform.Platform
	InfraProvider
}

type CommonPlatform struct {
	Properties        map[string]*PropertyInfo
	MappedExternalIPs map[string]string
	SharedRootLBName  string
	SharedRootLB      *MEXRootLB
	FlavorList        []*edgeproto.FlavorInfo
	//Platform          platform.Platform
	PlatformConfig *platform.PlatformConfig
	VaultConfig    *vault.Config
	infraProvider  CommonPlatformProvider
}

var MEXInfraVersion = "3.0.3"
var ImageNamePrefix = "mobiledgex-v"
var DefaultOSImageName = ImageNamePrefix + MEXInfraVersion
var ImageFormatQcow2 = "qcow2"
var ImageFormatVmdk = "vmdk"

// Default CloudletVM/Registry paths should only be used for local testing.
// Ansible should always specify the correct ones to the controller.
// These are not used if running the CRM manually, because these are only
// used by CreateCloudlet to set up the CRM VM and container.
var DefaultContainerRegistryPath = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"
var DefaultCloudletVMImagePath = "https://artifactory.mobiledgex.net/artifactory/baseimages/"

// NoConfigExternalRouter is used for the case in which we don't manage the external
// router and don't add ports to it ourself, as happens with Contrail.  The router does exist in
// this case and we use it to route from the LB to the pods
var NoConfigExternalRouter = "NOCONFIG"

// NoExternalRouter means there is no router at all and we connect the LB to the k8s pods on the same subnet
// this may eventually be the default and possibly only option
var NoExternalRouter = "NONE"

// Package level test mode variable
var testMode = false

func SetPropsFromVars(ctx context.Context, props map[string]*PropertyInfo, vars map[string]string) {
	// Infra Props value is fetched in following order:
	// 1. Fetch props from vars passed, if nothing set then
	// 2. Fetch from env, if nothing set then
	// 3. Use default value
	for k, v := range props {
		if val, ok := vars[k]; ok {
			if props[k].Secret {
				log.SpanLog(ctx, log.DebugLevelMexos, "set infra property (secret) from vars", "key", k)
			} else {
				log.SpanLog(ctx, log.DebugLevelMexos, "set infra property from vars", "key", k, "val", val)
			}
			props[k].Value = val
		} else if val, ok := os.LookupEnv(k); ok {
			if props[k].Secret {
				log.SpanLog(ctx, log.DebugLevelMexos, "set infra property (secret) from env", "key", k)
			} else {
				log.SpanLog(ctx, log.DebugLevelMexos, "set infra property from env", "key", k, "val", val)
			}
			props[k].Value = val
		} else {
			if props[k].Secret {
				log.SpanLog(ctx, log.DebugLevelMexos, "using default infra property (secret)", "key", k)
			} else {
				log.SpanLog(ctx, log.DebugLevelMexos, "using default infra property", "key", k, "val", v.Value)
			}
		}
	}
}

func (c *CommonPlatform) InitInfraCommon(ctx context.Context, platformConfig *pf.PlatformConfig, platformSpecificProps map[string]*PropertyInfo, vaultConfig *vault.Config, provider CommonPlatformProvider) error {
	if vaultConfig.Addr == "" {
		return fmt.Errorf("vaultAddr is not specified")
	}
	// set default properties
	c.Properties = infraCommonProps
	c.PlatformConfig = platformConfig
	c.VaultConfig = vaultConfig
	c.infraProvider = provider

	// append platform specific properties
	for k, v := range platformSpecificProps {
		c.Properties[k] = v
	}

	// fetch properties from vault
	mexEnvPath := GetVaultCloudletCommonPath("mexenv.json")
	log.SpanLog(ctx, log.DebugLevelMexos, "interning vault", "addr", vaultConfig.Addr, "path", mexEnvPath)
	envData := &VaultEnvData{}
	err := vault.GetData(vaultConfig, mexEnvPath, 0, envData)
	if err != nil {
		if strings.Contains(err.Error(), "no secrets") {
			return fmt.Errorf("Failed to source access variables as mexenv.json " +
				"does not exist in secure secrets storage (Vault)")
		}
		return fmt.Errorf("Failed to source access variables from %s, %s: %v", vaultConfig.Addr, mexEnvPath, err)
	}
	for _, envData := range envData.Env {
		if _, ok := c.Properties[envData.Name]; ok {
			c.Properties[envData.Name].Value = envData.Value
		} else {
			c.Properties[envData.Name] = &PropertyInfo{
				Value: envData.Value,
			}
		}
	}
	// fetch properties from user input
	SetPropsFromVars(ctx, c.Properties, c.PlatformConfig.EnvVars)

	if c.GetCloudletCFKey() == "" {
		if testMode {
			log.SpanLog(ctx, log.DebugLevelMexos, "Env variable MEX_CF_KEY not set")
		} else {
			return fmt.Errorf("Env variable MEX_CF_KEY not set")
		}
	}
	if c.GetCloudletCFUser() == "" {
		if testMode {
			log.SpanLog(ctx, log.DebugLevelMexos, "Env variable MEX_CF_USER not set")
		} else {
			return fmt.Errorf("Env variable MEX_CF_USER not set")
		}
	}
	err = c.initMappedIPs()
	if err != nil {
		return fmt.Errorf("unable to init Mapped IPs: %v", err)
	}
	c.SharedRootLBName = c.GetRootLBName(c.PlatformConfig.CloudletKey)
	return nil
}

func (c *CommonPlatform) GetCloudletDNSZone() string {
	return c.Properties["MEX_DNS_ZONE"].Value
}

func (c *CommonPlatform) GetCloudletRegistryFileServer() string {
	return c.Properties["MEX_REGISTRY_FILE_SERVER"].Value
}

func (c *CommonPlatform) GetCloudletCFKey() string {
	return c.Properties["MEX_CF_KEY"].Value
}

func (c *CommonPlatform) GetCloudletCFUser() string {
	return c.Properties["MEX_CF_USER"].Value
}

func SetTestMode(tMode bool) {
	testMode = tMode
}

func GetCloudletNetworkIfaceFile() string {
	return "/etc/network/interfaces.d/50-cloud-init.cfg"
}

// initMappedIPs takes the env var MEX_EXTERNAL_IP_MAP contents like:
// fromip1=toip1,fromip2=toip2 and populates mappedExternalIPs
func (c *CommonPlatform) initMappedIPs() error {
	c.MappedExternalIPs = make(map[string]string)
	meip := c.Properties["MEX_EXTERNAL_IP_MAP"].Value
	if meip != "" {
		ippair := strings.Split(meip, ",")
		for _, i := range ippair {
			ia := strings.Split(i, "=")
			if len(ia) != 2 {
				return fmt.Errorf("invalid format for mapped ip, expect fromip=destip")
			}
			fromip := ia[0]
			toip := ia[1]
			c.MappedExternalIPs[fromip] = toip
		}

	}
	return nil
}

// GetMappedExternalIP returns the IP that the input IP should be mapped to. This
// is used for environments which used NATted external IPs
func (c *CommonPlatform) GetMappedExternalIP(ip string) string {
	mappedip, ok := c.MappedExternalIPs[ip]
	if ok {
		return mappedip
	}
	return ip
}
