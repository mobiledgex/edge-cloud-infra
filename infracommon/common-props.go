package infracommon

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

const MINIMUM_DISK_SIZE uint64 = 20

type PropertyInfo struct {
	Value  string
	Secret bool
}

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
	"MEX_EXT_NETWORK": &PropertyInfo{
		Value: "external-network-shared",
	},
	"MEX_NETWORK": &PropertyInfo{
		Value: "mex-k8s-net-1",
	},
	// note OS_IMAGE refers to Operating System
	"MEX_OS_IMAGE": &PropertyInfo{
		Value: DefaultOSImageName,
	},
	"MEX_SECURITY_GROUP": &PropertyInfo{
		Value: "default",
	},
	"FLAVOR_MATCH_PATTERN": &PropertyInfo{
		Value: ".*",
	},
	"MEX_CRM_GATEWAY_ADDR": &PropertyInfo{},
	"MEX_SHARED_ROOTLB_RAM": &PropertyInfo{
		Value: "4096",
	},
	"MEX_SHARED_ROOTLB_VCPUS": &PropertyInfo{
		Value: "2",
	},
	"MEX_SHARED_ROOTLB_DISK": &PropertyInfo{
		Value: "40",
	},
	"MEX_NETWORK_SCHEME": &PropertyInfo{
		Value: "name=mex-k8s-net-1,cidr=10.101.X.0/24",
	},
	"MEX_COMPUTE_AVAILABILITY_ZONE": &PropertyInfo{},
	"MEX_VOLUME_AVAILABILITY_ZONE":  &PropertyInfo{},
	"MEX_IMAGE_DISK_FORMAT": &PropertyInfo{
		Value: ImageFormatQcow2,
	},
	"CLEANUP_ON_FAILURE": &PropertyInfo{
		Value: "true",
	},
}

func GetVaultCloudletCommonPath(filePath string) string {
	return fmt.Sprintf("/secret/data/cloudlet/openstack/%s", filePath)
}

func GetCloudletVMImageName(imgVersion string) string {
	if imgVersion == "" {
		imgVersion = MEXInfraVersion
	}
	return ImageNamePrefix + imgVersion
}

func GetCloudletVMImagePath(imgPath, imgVersion string) string {
	vmRegistryPath := DefaultCloudletVMImagePath
	if imgPath != "" {
		vmRegistryPath = imgPath
	}
	if !strings.HasSuffix(vmRegistryPath, "/") {
		vmRegistryPath = vmRegistryPath + "/"
	}
	return vmRegistryPath + GetCloudletVMImageName(imgVersion) + ".qcow2"
}

// GetCloudletSharedRootLBFlavor gets the flavor from defaults
// or environment variables
func (c *CommonPlatform) GetCloudletSharedRootLBFlavor(flavor *edgeproto.Flavor) error {
	ram := c.Properties["MEX_SHARED_ROOTLB_RAM"].Value
	var err error
	if ram != "" {
		flavor.Ram, err = strconv.ParseUint(ram, 10, 64)
		if err != nil {
			return err
		}
	} else {
		flavor.Ram = 4096
	}
	vcpus := c.Properties["MEX_SHARED_ROOTLB_VCPUS"].Value
	if vcpus != "" {
		flavor.Vcpus, err = strconv.ParseUint(vcpus, 10, 64)
		if err != nil {
			return err
		}
	} else {
		flavor.Vcpus = 2
	}
	disk := c.Properties["MEX_SHARED_ROOTLB_DISK"].Value
	if disk != "" {
		flavor.Disk, err = strconv.ParseUint(disk, 10, 64)
		if err != nil {
			return err
		}
	} else {
		flavor.Disk = 40
	}
	return nil
}

// getCloudletSecurityGroupName returns the cloudlet-wide security group name.  This function cannot ever be called externally because
// this group name can be duplicated which can cause errors in some environments.   GetCloudletSecurityGroupID should be used instead.  Note
func (c *CommonPlatform) GetCloudletSecurityGroupName() string {
	return c.Properties["MEX_SECURITY_GROUP"].Value
}

//GetCloudletExternalRouter returns default MEX external router name
func (c *CommonPlatform) GetCloudletExternalRouter() string {
	return c.Properties["MEX_ROUTER"].Value
}

func (c *CommonPlatform) GetCloudletExternalNetwork() string {
	return c.Properties["MEX_EXT_NETWORK"].Value
}

// GetCloudletNetwork returns default MEX network, internal and prepped
func (c *CommonPlatform) GetCloudletMexNetwork() string {
	return c.Properties["MEX_NETWORK"].Value
}

func (c *CommonPlatform) GetCloudletNetworkScheme() string {
	return c.Properties["MEX_NETWORK_SCHEME"].Value
}

func (c *CommonPlatform) GetCloudletVolumeAvailabilityZone() string {
	return c.Properties["MEX_VOLUME_AVAILABILITY_ZONE"].Value
}

func (c *CommonPlatform) GetCloudletComputeAvailabilityZone() string {
	return c.Properties["MEX_COMPUTE_AVAILABILITY_ZONE"].Value
}

func (c *CommonPlatform) GetCloudletImageDiskFormat() string {
	return c.Properties["MEX_IMAGE_DISK_FORMAT"].Value
}

// GetServerSecurityGroupName gets the secgrp name based on the server name
func (c *CommonPlatform) GetServerSecurityGroupName(serverName string) string {
	return serverName + "-sg"
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
