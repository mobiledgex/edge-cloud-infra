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
var OpenstackProps edgeproto.OpenStackProperties

var MEXInfraVersion = "v2.0.3" //Stratus
var defaultOSImageName = "mobiledgex-" + MEXInfraVersion
var VaultAddr string

// Package level test mode variable
var testMode = false

func getVaultCloudletPath(filePath, vaultAddr string) string {
	return fmt.Sprintf(
		"%s/v1/secret/data/cloudlet/openstack/%s",
		vaultAddr, filePath,
	)
}

func InitInfraCommon(ctx context.Context, vaultAddr string) error {
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
	return nil
}

func InitOpenstackProps(ctx context.Context, operatorName, physicalName, vaultAddr string) error {
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

	OpenstackProps.OpenRcVars = make(map[string]string)

	OpenstackProps.OsExternalNetworkName = os.Getenv("MEX_EXT_NETWORK")
	if OpenstackProps.OsExternalNetworkName == "" {
		OpenstackProps.OsExternalNetworkName = "external-network-shared"
	}

	OpenstackProps.OsImageName = os.Getenv("MEX_OS_IMAGE")
	if OpenstackProps.OsImageName == "" {
		OpenstackProps.OsImageName = defaultOSImageName
	}

	// defaulting some value
	OpenstackProps.OsExternalRouterName = os.Getenv("MEX_ROUTER")
	if OpenstackProps.OsExternalRouterName == "" {
		OpenstackProps.OsExternalRouterName = "mex-k8s-router-1"
	} else if OpenstackProps.OsExternalRouterName == "NONE" {
		OpenstackProps.OsExternalRouterName = ""
	}
	OpenstackProps.OsMexNetwork = "mex-k8s-net-1"
	return nil
}

//GetCloudletExternalRouter returns default MEX external router name
func GetCloudletExternalRouter() string {
	//TODO validate existence and status
	return OpenstackProps.OsExternalRouterName
}

func GetCloudletExternalNetwork() string {
	// this will be unset if platform is not openstack
	// because InitOpenstackProps() will not have been called.
	return OpenstackProps.OsExternalNetworkName
}

// Utility functions that used to be within manifest.
//GetCloudletNetwork returns default MEX network, internal and prepped
func GetCloudletMexNetwork() string {
	//TODO validate existence and status
	return OpenstackProps.OsMexNetwork
}

func GetCloudletDNSZone() string {
	return CloudletInfraCommon.DnsZone
}

func GetCloudletNetworkScheme() string {
	return CloudletInfraCommon.NetworkScheme
}

func GetCloudletOSImage() string {
	return OpenstackProps.OsImageName
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
	sg := os.Getenv("MEX_SECURITY_GROUP")
	if sg == "" {
		return "default"
	}
	return sg
}
func GetCloudletMexosAgentPort() string {
	return "18889"
}

func GetCloudletFlavorMatchPattern() string {
	pattern := os.Getenv("FLAVOR_MATCH_PATTERN")
	if pattern == "" {
		return ".*"
	}
	return pattern
}

func GetCloudletCRMGatewayIPAndPort() (string, int) {
	gw := os.Getenv("MEX_CRM_GATEWAY_ADDR")
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
