package mexos

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var (
	MEXInfraVersion    = "3.0.1"
	defaultOSImageName = "mobiledgex-v" + MEXInfraVersion

	// Default CloudletVM/Registry paths should only be used for local testing.
	// Ansible should always specify the correct ones to the controller.
	// These are not used if running the CRM manually, because these are only
	// used by CreateCloudlet to set up the CRM VM and container.
	DefaultCloudletRegistryPath = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"
	DefaultVMRegistryPath       = "https://artifactory.mobiledgex.net/artifactory"
)

type DeploymentType string

const (
	RootLBVMDeployment   DeploymentType = "mexrootlb"
	UserVMDeployment     DeploymentType = "mexuservm"
	PlatformVMDeployment DeploymentType = "mexplatformvm"
	ClusterVMDeployment  DeploymentType = "mexclustervm"
	SharedCluster        DeploymentType = "sharedcluster"
)

func GetPlatformVMPrefix(key *edgeproto.CloudletKey) string {
	return key.Name + "." + key.OperatorKey.Name
}

func GetPlatformVMName(key *edgeproto.CloudletKey) string {
	return GetPlatformVMPrefix(key) + ".pf"
}

func GetStackSuffix(key *edgeproto.CloudletKey, vmType DeploymentType) string {
	ext := ""
	switch vmType {
	case RootLBVMDeployment:
		ext = ".lb"
	case PlatformVMDeployment:
		ext = ".pf"
	}
	return GetPlatformVMPrefix(key) + ext
}

func GetStackName(key *edgeproto.CloudletKey, cloudletVersion string, vmType DeploymentType) string {
	// Form stack VM name based on cloudletKey & version
	version := cloudletVersion
	if cloudletVersion == "" {
		version = MEXInfraVersion
	}
	return version + "_" + GetStackSuffix(key, vmType)
}

func GetCloudletVMImageName(imgVersion string) string {
	if imgVersion == "" {
		imgVersion = MEXInfraVersion
	}
	return "mobiledgex-v" + imgVersion
}

func GetCloudletVMImagePath(imgPath, imgVersion string) string {
	if imgPath != "" {
		return imgPath
	}
	return DefaultVMRegistryPath + "/baseimages/" + GetCloudletVMImageName(imgVersion) + ".qcow2"
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
