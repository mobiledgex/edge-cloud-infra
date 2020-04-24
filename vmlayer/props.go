package vmlayer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var ImageFormatQcow2 = "qcow2"
var ImageFormatVmdk = "vmdk"

var MEXInfraVersion = "3.1.0"
var ImageNamePrefix = "mobiledgex-v"
var DefaultOSImageName = ImageNamePrefix + MEXInfraVersion

const MINIMUM_DISK_SIZE uint64 = 20

// NoSubnetDNS means that DNS servers are not specified when creating the subnet
var NoSubnetDNS = "NONE"

// NoConfigExternalRouter is used for the case in which we don't manage the external
// router and don't add ports to it ourself, as happens with Contrail.  The router does exist in
// this case and we use it to route from the LB to the pods
var NoConfigExternalRouter = "NOCONFIG"

// NoExternalRouter means there is no router at all and we connect the LB to the k8s pods on the same subnet
// this may eventually be the default and possibly only option
var NoExternalRouter = "NONE"

var DefaultCloudletVMImagePath = "https://artifactory.mobiledgex.net/artifactory/baseimages/"

// properties common to all VM providers
var VMProviderProps = map[string]*infracommon.PropertyInfo{
	// Property: Default-Value

	"MEX_EXT_NETWORK": &infracommon.PropertyInfo{
		Value: "external-network-shared",
	},
	"MEX_NETWORK": &infracommon.PropertyInfo{
		Value: "mex-k8s-net-1",
	},
	// note OS_IMAGE refers to Operating System
	"MEX_OS_IMAGE": &infracommon.PropertyInfo{
		Value: DefaultOSImageName,
	},
	"MEX_SECURITY_GROUP": &infracommon.PropertyInfo{
		Value: "default",
	},
	"MEX_SHARED_ROOTLB_RAM": &infracommon.PropertyInfo{
		Value: "4096",
	},
	"MEX_SHARED_ROOTLB_VCPUS": &infracommon.PropertyInfo{
		Value: "2",
	},
	"MEX_SHARED_ROOTLB_DISK": &infracommon.PropertyInfo{
		Value: "40",
	},
	"MEX_NETWORK_SCHEME": &infracommon.PropertyInfo{
		Value: "name=mex-k8s-net-1,cidr=10.101.X.0/24",
	},
	"MEX_COMPUTE_AVAILABILITY_ZONE": &infracommon.PropertyInfo{},
	"MEX_VOLUME_AVAILABILITY_ZONE":  &infracommon.PropertyInfo{},
	"MEX_IMAGE_DISK_FORMAT": &infracommon.PropertyInfo{
		Value: ImageFormatQcow2,
	},
	"MEX_ROUTER": &infracommon.PropertyInfo{
		Value: NoExternalRouter,
	},
}

func GetVaultCloudletCommonPath(filePath string) string {
	// TODO this path really should not be openstack
	return fmt.Sprintf("/secret/data/cloudlet/openstack/%s", filePath)
}

func GetCloudletVMImageName(imgVersion string) string {
	if imgVersion == "" {
		imgVersion = MEXInfraVersion
	}
	return ImageNamePrefix + imgVersion
}

func GetCertFilePath(key *edgeproto.CloudletKey) string {
	return fmt.Sprintf("/tmp/%s.%s.cert", key.Name, key.Organization)
}

func (v *VMPlatform) GetVaultCloudletAccessPath(key *edgeproto.CloudletKey, region, physicalName string) string {
	return fmt.Sprintf("/secret/data/%s/cloudlet/%s/%s/%s/%s", region, v.Type, key.Organization, physicalName, "openrc.json")
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
func (v *VMPlatform) GetCloudletSharedRootLBFlavor(flavor *edgeproto.Flavor) error {
	ram := v.CommonPf.Properties["MEX_SHARED_ROOTLB_RAM"].Value
	var err error
	if ram != "" {
		flavor.Ram, err = strconv.ParseUint(ram, 10, 64)
		if err != nil {
			return err
		}
	} else {
		flavor.Ram = 4096
	}
	vcpus := v.CommonPf.Properties["MEX_SHARED_ROOTLB_VCPUS"].Value
	if vcpus != "" {
		flavor.Vcpus, err = strconv.ParseUint(vcpus, 10, 64)
		if err != nil {
			return err
		}
	} else {
		flavor.Vcpus = 2
	}
	disk := v.CommonPf.Properties["MEX_SHARED_ROOTLB_DISK"].Value
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

func (v *VMPlatform) GetCloudletSecurityGroupName() string {
	return v.CommonPf.Properties["MEX_SECURITY_GROUP"].Value
}

func (v *VMPlatform) GetCloudletExternalNetwork() string {
	return v.CommonPf.Properties["MEX_EXT_NETWORK"].Value
}

// GetCloudletNetwork returns default MEX network, internal and prepped
func (v *VMPlatform) GetCloudletMexNetwork() string {
	return v.CommonPf.Properties["MEX_NETWORK"].Value
}

func (v *VMPlatform) GetCloudletNetworkScheme() string {
	return v.CommonPf.Properties["MEX_NETWORK_SCHEME"].Value
}

func (v *VMPlatform) GetCloudletVolumeAvailabilityZone() string {
	return v.CommonPf.Properties["MEX_VOLUME_AVAILABILITY_ZONE"].Value
}

func (v *VMPlatform) GetCloudletComputeAvailabilityZone() string {
	return v.CommonPf.Properties["MEX_COMPUTE_AVAILABILITY_ZONE"].Value
}

func (v *VMPlatform) GetCloudletImageDiskFormat() string {
	return v.CommonPf.Properties["MEX_IMAGE_DISK_FORMAT"].Value
}

func (v *VMPlatform) GetCloudletOSImage() string {
	return v.CommonPf.Properties["MEX_OS_IMAGE"].Value
}

func (v *VMPlatform) GetCloudletFlavorMatchPattern() string {
	return v.CommonPf.Properties["FLAVOR_MATCH_PATTERN"].Value
}

//GetCloudletExternalRouter returns default MEX external router name
func (v *VMPlatform) GetCloudletExternalRouter() string {
	return v.CommonPf.Properties["MEX_ROUTER"].Value
}

func (v *VMPlatform) GetSubnetDNS() string {
	return v.CommonPf.Properties["MEX_SUBNET_DNS"].Value
}
