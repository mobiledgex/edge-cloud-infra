package infracommon

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
)

type PropertyInfo struct {
	Value  string
	Secret bool
}

var ImageFormatQcow2 = "qcow2"
var ImageFormatVmdk = "vmdk"

var MEXInfraVersion = "3.1.0"
var ImageNamePrefix = "mobiledgex-v"
var DefaultOSImageName = ImageNamePrefix + MEXInfraVersion

// Default CloudletVM/Registry paths should only be used for local testing.
// Ansible should always specify the correct ones to the controller.
// These are not used if running the CRM manually, because these are only
// used by CreateCloudlet to set up the CRM VM and container.
var DefaultContainerRegistryPath = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"

// Cloudlet Infra Common Properties
var infraCommonProps = map[string]*PropertyInfo{
	// Property: Default-Value
	"MEX_CF_KEY": &PropertyInfo{
		Secret: true,
	},
	"MEX_CF_USER":         &PropertyInfo{},
	"MEX_EXTERNAL_IP_MAP": &PropertyInfo{},
	"MEX_REGISTRY_FILE_SERVER": &PropertyInfo{
		Value: "registry.mobiledgex.net",
	},
	"MEX_DNS_ZONE": &PropertyInfo{
		Value: "mobiledgex.net",
	},
	"FLAVOR_MATCH_PATTERN": &PropertyInfo{
		Value: ".*",
	},
	"MEX_CRM_GATEWAY_ADDR": &PropertyInfo{},
	"MEX_SUBNET_DNS":       &PropertyInfo{},
	"CLEANUP_ON_FAILURE": &PropertyInfo{
		Value: "true",
	},
}

func SetPropsFromVars(ctx context.Context, props map[string]*PropertyInfo, vars map[string]string) {
	// Infra Props value is fetched in following order:
	// 1. Fetch props from vars passed, if nothing set then
	// 2. Fetch from env, if nothing set then
	// 3. Use default value
	for k, v := range props {
		if val, ok := vars[k]; ok {
			if props[k].Secret {
				log.SpanLog(ctx, log.DebugLevelInfra, "set infra property (secret) from vars", "key", k)
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "set infra property from vars", "key", k, "val", val)
			}
			props[k].Value = val
		} else if val, ok := os.LookupEnv(k); ok {
			if props[k].Secret {
				log.SpanLog(ctx, log.DebugLevelInfra, "set infra property (secret) from env", "key", k)
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "set infra property from env", "key", k, "val", val)
			}
			props[k].Value = val
		} else {
			if props[k].Secret {
				log.SpanLog(ctx, log.DebugLevelInfra, "using default infra property (secret)", "key", k)
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "using default infra property", "key", k, "val", v.Value)
			}
		}
	}
}

func GetVaultCloudletCommonPath(filePath string) string {
	// TODO this path really should not be openstack
	return fmt.Sprintf("/secret/data/cloudlet/openstack/%s", filePath)
}

// GetCleanupOnFailure should be true unless we want to debug the failure,
// in which case this env var can be set to no.  We could consider making
// this configurable at the controller but really is only needed for debugging.
func (v *CommonPlatform) GetCleanupOnFailure(ctx context.Context) bool {
	cleanup := v.Properties["CLEANUP_ON_FAILURE"].Value
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCleanupOnFailure", "cleanup", cleanup)
	cleanup = strings.ToLower(cleanup)
	cleanup = strings.ReplaceAll(cleanup, "'", "")
	if cleanup == "no" || cleanup == "false" {
		return false
	}
	return true
}

func (c *CommonPlatform) GetCloudletCRMGatewayIPAndPort() (string, int) {
	gw := c.Properties["MEX_CRM_GATEWAY_ADDR"].Value
	if gw == "" {
		return "", 0
	}
	host, portstr, err := net.SplitHostPort(gw)
	if err != nil {
		log.FatalLog("Error in MEX_CRM_GATEWAY_ADDR format")
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		log.FatalLog("Error in MEX_CRM_GATEWAY_ADDR port format")
	}
	return host, port
}

func (c *CommonPlatform) GetCloudletOSImage() string {
	return c.Properties["MEX_OS_IMAGE"].Value
}

func (c *CommonPlatform) GetCloudletFlavorMatchPattern() string {
	return c.Properties["FLAVOR_MATCH_PATTERN"].Value
}

//GetCloudletExternalRouter returns default MEX external router name
func (c *CommonPlatform) GetCloudletExternalRouter() string {
	return c.Properties["MEX_ROUTER"].Value
}

func (c *CommonPlatform) GetSubnetDNS() string {
	return c.Properties["MEX_SUBNET_DNS"].Value
}
