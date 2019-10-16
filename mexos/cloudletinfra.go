package mexos

// This file stores a global cloudlet infra properties object. The long term solution is for the controller to send this via the
// notification channel when the cloudlet is provisioned.   The controller will do the vault access and pass this data down; this
// is a stepping stone to start using edgepro data strucures to hold info abou the cloudlet rather than custom types and so the vault
// is still directly accessed here as are env variable to populate some variables

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var CloudletInfraCommon edgeproto.CloudletInfraCommon

var MEXInfraVersion = "v2.0.3" //Stratus
var defaultOSImageName = "mobiledgex-" + MEXInfraVersion
var VaultAddr string

// NoConfigExternalRouter is used for the case in which we don't manage the external
// router and don't add ports to it ourself, as happens with Contrail
var NoConfigExternalRouter = "NOCONFIG"

// Package level test mode variable
var testMode = false

// mapping of FQDNs the CRM knows about to externally mapped IPs. This
// is used mainly in lab environments that have NATed IPs which can be used to
// access the cloudlet externally but are not visible in any way to OpenStack
var mappedExternalIPs map[string]string

// Common Cloudlet Infra Properties
var infraProps = map[string]string{
	// Property: Default-Value
	"MEX_NETWORK_SCHEME": "",
}

// Openstack Infra Properties
var openstackProps = map[string]string{
	// Property: Default-Value
	"MEX_EXT_NETWORK":      "external-network-shared",
	"MEX_NETWORK":          "mex-k8s-net-1",
	"MEX_ROUTER":           "mex-k8s-router-1",
	"MEX_OS_IMAGE":         defaultOSImageName,
	"MEX_SECURITY_GROUP":   "default",
	"FLAVOR_MATCH_PATTERN": ".*",
	"MEX_CRM_GATEWAY_ADDR": "",
	"MEX_EXTERNAL_IP_MAP":  "",
}

func setPropsFromVars(ctx context.Context, props, vars map[string]string) {
	// Infra Props value is fetched in following order:
	// 1. Fetch props from vars passed, if nothing set then
	// 2. Fetch from env, if nothing set then
	// 3. Use default value
	for k, _ := range props {
		if val, ok := vars[k]; ok {
			log.SpanLog(ctx, log.DebugLevelMexos, "set infra property from vars", "key", k, "val", val)
			props[k] = val
		} else {
			val := os.Getenv(k)
			if val != "" {
				log.SpanLog(ctx, log.DebugLevelMexos, "set infra property from env", "key", k, "val", val)
				props[k] = val
			}
		}
	}
}

func getVaultCloudletPath(filePath, vaultAddr string) string {
	return fmt.Sprintf(
		"%s/v1/secret/data/cloudlet/openstack/%s",
		vaultAddr, filePath,
	)
}

func InitInfraCommon(ctx context.Context, vaultAddr string, vars map[string]string) error {
	if vaultAddr == "" {
		return fmt.Errorf("vaultAddr is not specified")
	}
	VaultAddr = vaultAddr
	mexEnvURL := getVaultCloudletPath("mexenv.json", vaultAddr)
	err := InternVaultEnv(ctx, mexEnvURL)
	if err != nil {
		if testMode {
			log.SpanLog(ctx, log.DebugLevelMexos, "failed to InternVaultEnv", "url", mexEnvURL, "err", err)
		} else {
			return fmt.Errorf("failed to InternVaultEnv %s: %v", mexEnvURL, err)
		}
	}
	CloudletInfraCommon.CfKey = os.Getenv("MEX_CF_KEY")
	if CloudletInfraCommon.CfKey == "" {
		if testMode {
			log.SpanLog(ctx, log.DebugLevelMexos, "Env variable MEX_CF_KEY not set")
		} else {
			return fmt.Errorf("Env variable MEX_CF_KEY not set")
		}
	}
	CloudletInfraCommon.CfUser = os.Getenv("MEX_CF_USER")
	if CloudletInfraCommon.CfKey == "" {
		if testMode {
			log.SpanLog(ctx, log.DebugLevelMexos, "Env variable MEX_CF_USER not set")
		} else {
			return fmt.Errorf("Env variable MEX_CF_USER not set")
		}
	}
	CloudletInfraCommon.DnsZone = "mobiledgex.net"
	CloudletInfraCommon.RegistryFileServer = "registry.mobiledgex.net"

	setPropsFromVars(ctx, infraProps, vars)

	err = initMappedIPs()
	if err != nil {
		return fmt.Errorf("unable to init Mapped IPs: %v", err)
	}
	return nil
}

func InitOpenstackProps(ctx context.Context, operatorName, physicalName, vaultAddr string, vars map[string]string) error {
	openRcURL := getVaultCloudletPath(physicalName+"/openrc.json", vaultAddr)
	err := InternVaultEnv(ctx, openRcURL)
	if err != nil {
		return fmt.Errorf("failed to InternVaultEnv %s: %v", openRcURL, err)
	}
	VaultAddr = vaultAddr
	authURL := os.Getenv("OS_AUTH_URL")
	if strings.HasPrefix(authURL, "https") {
		caCertURL := getVaultCloudletPath(physicalName+"/os_cacert", vaultAddr)
		if caCertURL != "" {
			certFile := fmt.Sprintf("/tmp/%s.%s.cert", operatorName, physicalName)
			err := GetVaultDataToFile(caCertURL, certFile)
			if err != nil {
				return fmt.Errorf("failed to GetVaultDataToFile %s: %v", caCertURL, err)
			}
			os.Setenv("OS_CACERT", certFile)
		}
	}

	setPropsFromVars(ctx, openstackProps, vars)

	return nil
}

//GetCloudletExternalRouter returns default MEX external router name
func GetCloudletExternalRouter() string {
	//TODO validate existence and status
	return openstackProps["MEX_ROUTER"]
}

func GetCloudletExternalNetwork() string {
	// this will be unset if platform is not openstack
	// because InitopenstackProps() will not have been called.
	return openstackProps["MEX_EXT_NETWORK"]
}

// Utility functions that used to be within manifest.
//GetCloudletNetwork returns default MEX network, internal and prepped
func GetCloudletMexNetwork() string {
	//TODO validate existence and status
	return openstackProps["MEX_NETWORK"]
}

func GetCloudletDNSZone() string {
	return CloudletInfraCommon.DnsZone
}

func GetCloudletOSImage() string {
	return openstackProps["MEX_OS_IMAGE"]
}

func GetCloudletRegistryFileServer() string {
	return CloudletInfraCommon.RegistryFileServer
}

func GetCloudletCFKey() string {
	return CloudletInfraCommon.CfKey
}

func GetCloudletCFUser() string {
	return CloudletInfraCommon.CfUser
}

func GetCloudletNetworkScheme() string {
	return infraProps["MEX_NETWORK_SCHEME"]
}

func SetCloudletNetworkScheme(val string) {
	infraProps["MEX_NETWORK_SCHEME"] = val
}

func SetTestMode(tMode bool) {
	testMode = tMode
}

// GetCleanupOnFailure should be true unless we want to debug the failure,
// in which case this env var can be set to no.  We could consider making
// this configurable at the controller but really is only needed for debugging.
func GetCleanupOnFailure(ctx context.Context) bool {
	cleanup := os.Getenv("CLEANUP_ON_FAILURE")
	log.SpanLog(ctx, log.DebugLevelMexos, "GetCleanupOnFailure", "cleanup", cleanup)
	cleanup = strings.ToLower(cleanup)
	cleanup = strings.ReplaceAll(cleanup, "'", "")
	if cleanup == "no" || cleanup == "false" {
		return false
	}
	return true
}

// These not in the proto file yet because they may not change for a while
func GetCloudletTenant() string {
	return "null"
}
func GetCloudletSecurityGroup() string {
	return openstackProps["MEX_SECURITY_GROUP"]
}
func GetCloudletMexosAgentPort() string {
	return "18889"
}

func GetCloudletFlavorMatchPattern() string {
	pattern := openstackProps["FLAVOR_MATCH_PATTERN"]
	if pattern == "" {
		return ".*"
	}
	return pattern
}

func GetCloudletCRMGatewayIPAndPort() (string, int) {
	gw := openstackProps["MEX_CRM_GATEWAY_ADDR"]
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

// initMappedIPs takes the env var MEX_EXTERNAL_IP_MAP contents like:
// fromip1=toip1,fromip2=toip2 and populates mappedExternalIPs
func initMappedIPs() error {
	mappedExternalIPs = make(map[string]string)
	meip := openstackProps["MEX_EXTERNAL_IP_MAP"]
	if meip != "" {
		ippair := strings.Split(meip, ",")
		for _, i := range ippair {
			ia := strings.Split(i, "=")
			if len(ia) != 2 {
				return fmt.Errorf("invalid format for mapped ip, expect fromip=destip")
			}
			fromip := ia[0]
			toip := ia[1]
			mappedExternalIPs[fromip] = toip
		}

	}
	return nil
}

// GetMappedExternalIP returns the IP that the input IP should be mapped to. This
// is used for environments which used NATted external IPs
func GetMappedExternalIP(ip string) string {
	mappedip, ok := mappedExternalIPs[ip]
	if ok {
		return mappedip
	}
	return ip
}
