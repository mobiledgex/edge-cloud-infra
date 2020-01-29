package mexos

import (
	"fmt"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

var (
	MEXInfraVersion    = "3.0.3"
	ImageNamePrefix    = "mobiledgex-v"
	defaultOSImageName = ImageNamePrefix + MEXInfraVersion

	// Default CloudletVM/Registry paths should only be used for local testing.
	// Ansible should always specify the correct ones to the controller.
	// These are not used if running the CRM manually, because these are only
	// used by CreateCloudlet to set up the CRM VM and container.
	DefaultCloudletRegistryPath = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"
	DefaultVMRegistryPath       = "https://artifactory.mobiledgex.net/artifactory"
)

type DeploymentType string

const (
	RootLBVMDeployment   DeploymentType = "RootLB"
	UserVMDeployment     DeploymentType = "UserVM"
	PlatformVMDeployment DeploymentType = "PlatformVM"
	ClusterVMDeployment  DeploymentType = "ClusterVM"
	SharedCluster        DeploymentType = "sharedcluster"
)

func GetExt(vmType DeploymentType) string {
	ext := ""
	switch vmType {
	case RootLBVMDeployment:
		ext = "lb"
	case PlatformVMDeployment:
		ext = "pf"
	}
	return ext
}

func GetPlatformVMPrefix(key *edgeproto.CloudletKey) string {
	return key.Name + "." + key.OperatorKey.Name
}

func GetPlatformVMName(key *edgeproto.CloudletKey) string {
	return util.HeatSanitize(GetPlatformVMPrefix(key) + ".pf")
}

func GetStackName(key *edgeproto.CloudletKey, cloudletVersion string, vmType DeploymentType) string {
	// Form stack VM name based on cloudletKey & version
	version := cloudletVersion
	if cloudletVersion == "" {
		version = MEXInfraVersion
	}
	return util.HeatSanitize(GetPlatformVMPrefix(key) + "_" + version + "_" + GetExt(vmType))
}

func IsStackSame(stackName string, key *edgeproto.CloudletKey, vmType DeploymentType) bool {
	ext := GetExt(vmType)
	parts := strings.Split(stackName, "_")
	if parts[len(parts)-1] != ext {
		return false
	}
	if util.HeatSanitize(GetPlatformVMPrefix(key)) != strings.Join(parts[0:len(parts)-2], "_") {
		return false
	}
	return true
}

func GetCloudletVMImageName(imgVersion string) string {
	if imgVersion == "" {
		imgVersion = MEXInfraVersion
	}
	return ImageNamePrefix + imgVersion
}

func GetCloudletVMImagePath(imgPath, imgVersion string) string {
	vmRegistryPath := DefaultVMRegistryPath + "/baseimages/"
	if imgPath != "" {
		vmRegistryPath = imgPath
	}
	if !strings.HasSuffix(vmRegistryPath, "/") {
		vmRegistryPath = vmRegistryPath + "/"
	}
	return vmRegistryPath + GetCloudletVMImageName(imgVersion) + ".qcow2"
}

func GetCloudletVMImagePkgName(imgVersion string) string {
	if imgVersion == "" {
		imgVersion = MEXInfraVersion
	}
	return "mobiledgex_" + imgVersion + "_amd64.deb"
}

func GetCloudletVMImagePkgPath(imgPath, imgVersion string) string {
	if imgPath == "" {
		return imgPath
	}
	return DefaultVMRegistryPath + "/packages/pool/" + GetCloudletVMImagePkgName(imgVersion)
}

func GetVaultCloudletPath(key *edgeproto.CloudletKey, region, physicalName, filePath string) string {
	return fmt.Sprintf("/secret/data/%s/cloudlet/openstack/%s/%s/%s", region, key.OperatorKey.Name, physicalName, filePath)
}

func GetVaultCloudletCommonPath(filePath string) string {
	return fmt.Sprintf("/secret/data/cloudlet/openstack/%s", filePath)
}

func GetCertFilePath(key *edgeproto.CloudletKey) string {
	return fmt.Sprintf("/tmp/%s.%s.cert", key.Name, key.OperatorKey.Name)
}
