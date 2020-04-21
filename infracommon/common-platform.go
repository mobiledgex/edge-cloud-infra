package infracommon

// This file stores a global cloudlet infra properties object. The long term solution is for the controller to send this via the
// notification channel when the cloudlet is provisioned.   The controller will do the vault access and pass this data down; this
// is a stepping stone to start using edgepro data strucures to hold info abou the cloudlet rather than custom types and so the vault
// is still directly accessed here as are env variable to populate some variables

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/log"

	"github.com/mobiledgex/edge-cloud/vault"
)

// this is not exhaustive, currently only ResourceTypeSecurityGroup is needed
type ResourceType string

var ResourceTypeVM ResourceType = "VM"
var ResourceTypeSubnet ResourceType = "Subnet"
var ResourceTypeSecurityGroup ResourceType = "SecGrp"

type CommonPlatform struct {
	Properties        map[string]*PropertyInfo
	PlatformConfig    *platform.PlatformConfig
	VaultConfig       *vault.Config
	MappedExternalIPs map[string]string
}

// Package level test mode variable
var testMode = false

func (c *CommonPlatform) InitInfraCommon(ctx context.Context, platformConfig *pf.PlatformConfig, platformSpecificProps []map[string]*PropertyInfo, vaultConfig *vault.Config) error {
	if vaultConfig.Addr == "" {
		return fmt.Errorf("vaultAddr is not specified")
	}
	// set default properties
	c.Properties = infraCommonProps
	c.PlatformConfig = platformConfig
	c.VaultConfig = vaultConfig

	// append platform specific properties
	for _, propmap := range platformSpecificProps {
		for k, v := range propmap {
			c.Properties[k] = v
		}
	}

	// fetch properties from vault
	mexEnvPath := GetVaultCloudletCommonPath("mexenv.json")
	log.SpanLog(ctx, log.DebugLevelInfra, "interning vault", "addr", vaultConfig.Addr, "path", mexEnvPath)
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
			log.SpanLog(ctx, log.DebugLevelInfra, "Env variable MEX_CF_KEY not set")
		} else {
			return fmt.Errorf("Env variable MEX_CF_KEY not set")
		}
	}
	if c.GetCloudletCFUser() == "" {
		if testMode {
			log.SpanLog(ctx, log.DebugLevelInfra, "Env variable MEX_CF_USER not set")
		} else {
			return fmt.Errorf("Env variable MEX_CF_USER not set")
		}
	}
	err = c.initMappedIPs()
	if err != nil {
		return fmt.Errorf("unable to init Mapped IPs: %v", err)
	}
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
