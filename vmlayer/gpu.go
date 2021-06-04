package vmlayer

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/gcs"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

type GPUDrivers map[edgeproto.GPUDriverKey][]edgeproto.GPUDriverBuild

func (v *VMPlatform) getGCSStorageClient(ctx context.Context) (*gcs.GCSClient, error) {
	deploymentTag := v.VMProperties.GetDeploymentTag()
	bucketName := cloudcommon.GetGPUDriverBucketName(deploymentTag)
	accessApi := v.VMProperties.CommonPf.PlatformConfig.AccessApi
	credsObj, err := accessApi.GetGCSCreds(ctx)
	if err != nil {
		return nil, err
	}
	storageClient, err := gcs.NewClient(ctx, credsObj, bucketName, gcs.LongTimeout)
	if err != nil {
		return nil, fmt.Errorf("Unable to setup GCS client: %v", err)
	}
	return storageClient, nil
}

// Fetches driver package:
//        * From local cache
//        * In not in local cache, then fetch from cloud
func (v *VMPlatform) getGPUDriverPackagePath(ctx context.Context, storageClient *gcs.GCSClient, build *edgeproto.GPUDriverBuild) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "getGPUDriverPackagePath", "build", build)
	// Look in local cache first
	if _, err := os.Stat(v.CacheDir); os.IsNotExist(err) {
		return "", fmt.Errorf("Missing cache dir")
	}

	fileName := cloudcommon.GetGPUDriverBuildPathFromURL(build.DriverPath, v.VMProperties.GetDeploymentTag())
	localFilePath := v.CacheDir + "/" + strings.ReplaceAll(fileName, "/", "_")
	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		log.SpanLog(ctx, log.DebugLevelInfra, "GPU driver pkg not found in local cache, fetch it from GCS", "build.DriverPath", build.DriverPath)
		// In not in local cache, then fetch from cloud
		outBytes, err := storageClient.DownloadObject(ctx, fileName)
		if err != nil {
			return "", fmt.Errorf("Failed to download GPU driver package %s from GCS %v", fileName, err)
		}
		err = ioutil.WriteFile(localFilePath, outBytes, 0644)
		if err != nil {
			return "", fmt.Errorf("Failed to create local cache file %s, %v", localFilePath, err)
		}
	}
	return localFilePath, nil
}

// Fetches driver license config:
//        * From local cache
//        * In not in local cache, then fetch from cloud
func (v *VMPlatform) getGPUDriverLicenseConfigPath(ctx context.Context, storageClient *gcs.GCSClient, driverKey *edgeproto.GPUDriverKey) (string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "getGPUDriverLicenseConfigPath", "driverKey", driverKey)
	// Look in local cache first
	if _, err := os.Stat(v.CacheDir); os.IsNotExist(err) {
		return "", fmt.Errorf("Missing cache dir")
	}
	fileName := cloudcommon.GetGPUDriverLicenseStoragePath(driverKey)
	localFilePath := v.CacheDir + "/" + strings.ReplaceAll(fileName, "/", "_")
	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		log.SpanLog(ctx, log.DebugLevelInfra, "GPU driver license not found in local cache, fetch it from GCS", "fileName", fileName)
		outBytes, err := storageClient.DownloadObject(ctx, fileName)
		if err != nil {
			if err.Error() == gcs.NotFoundError {
				// license config doesn't exist
				return "", nil
			}
			return "", fmt.Errorf("Failed to download GPU driver license config %s from GCS %v", fileName, err)
		}
		err = ioutil.WriteFile(localFilePath, outBytes, 0644)
		if err != nil {
			return "", fmt.Errorf("Failed to create local cache file %s, %v", localFilePath, err)
		}
	}
	return localFilePath, nil
}

func (v *VMPlatform) setupGPUDrivers(ctx context.Context, rootLBClient ssh.Client, clusterInst *edgeproto.ClusterInst, updateCallback edgeproto.CacheUpdateCallback, action ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "setupGPUDrivers", "clusterInst", clusterInst.Key)
	updateCallback(edgeproto.UpdateTask, "Gathering supported list of GPU drivers")
	gpuDrivers, err := v.GetCloudletGPUDriverBuilds(ctx, clusterInst.Key.CloudletKey.Organization)
	if err != nil {
		return err
	}
	if len(gpuDrivers) == 0 {
		return fmt.Errorf("No GPU drivers available")
	}

	targetNodes := []string{}
	switch clusterInst.Deployment {
	case cloudcommon.DeploymentTypeDocker:
		targetNodes = append(targetNodes, GetClusterMasterName(ctx, clusterInst))
	case cloudcommon.DeploymentTypeKubernetes:
		fallthrough
	case cloudcommon.DeploymentTypeHelm:
		targetNodes = append(targetNodes, GetClusterMasterName(ctx, clusterInst))
		for nn := uint32(1); nn <= clusterInst.NumNodes; nn++ {
			targetNodes = append(targetNodes, GetClusterNodeName(ctx, clusterInst, nn))
		}
	default:
		return fmt.Errorf("GPU driver installation not supported for deployment type %s", clusterInst.Deployment)
	}
	storageClient, err := v.getGCSStorageClient(ctx)
	if err != nil {
		return err
	}
	defer storageClient.Close()
	for _, node := range targetNodes {
		vmIP, err := v.GetIPFromServerName(ctx, v.VMProperties.GetCloudletMexNetwork(), GetClusterSubnetName(ctx, clusterInst), node)
		if err != nil {
			return err
		}

		client, err := rootLBClient.AddHop(vmIP.ExternalAddr, 22)
		if err != nil {
			return err
		}
		err = v.installGPUDriverBuild(ctx, storageClient, node, client, gpuDrivers, updateCallback)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *VMPlatform) GetCloudletGPUDriverBuilds(ctx context.Context, cloudletOrg string) (GPUDrivers, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCloudletGPUDriverBuilds", "cloudletOrg", cloudletOrg)
	gpuDrivers := make(GPUDrivers)
	if v.GPUConfig.GpuType == edgeproto.GPUType_GPU_TYPE_NONE {
		return gpuDrivers, nil
	}
	if v.Caches == nil {
		return nil, fmt.Errorf("caches is nil")
	}
	// Get all public GPU drivers
	var gpuDriver edgeproto.GPUDriver
	driverKey := edgeproto.GPUDriverKey{
		Name: v.GPUConfig.DriverName,
		Type: v.GPUConfig.GpuType,
	}
	gpuDrivers[driverKey] = []edgeproto.GPUDriverBuild{}
	if v.Caches.GPUDriverCache.Get(&driverKey, &gpuDriver) {
		gpuDrivers[driverKey] = append(gpuDrivers[driverKey], gpuDriver.Builds...)
	}
	// Get all operator owned GPU drivers
	driverKey.Organization = cloudletOrg
	gpuDrivers[driverKey] = []edgeproto.GPUDriverBuild{}
	if v.Caches.GPUDriverCache.Get(&driverKey, &gpuDriver) {
		gpuDrivers[driverKey] = append(gpuDrivers[driverKey], gpuDriver.Builds...)
	}
	return gpuDrivers, nil
}

func (v *VMPlatform) installGPUDriverBuild(ctx context.Context, storageClient *gcs.GCSClient, nodeName string, client ssh.Client, drivers GPUDrivers, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "installGPUDriverBuild", "nodeName", nodeName, "num drivers", len(drivers))
	// fetch linux kernel version
	out, err := client.Output("uname -sr")
	if err != nil {
		return err
	}
	if out == "" {
		return fmt.Errorf("failed to get kernel version for %s", nodeName)
	}
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return fmt.Errorf("invalid kernel version for %s: %s", nodeName, out)
	}
	os := parts[0]
	kernVers := parts[1]
	if os != "Linux" {
		return fmt.Errorf("unsupported os for %s: %s, only Linux is supported for now", nodeName, os)
	}
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching GPU driver supported for Linux kernel version %s", kernVers))
	var reqdDriverKey edgeproto.GPUDriverKey
	var reqdBuild edgeproto.GPUDriverBuild
	for driverKey, builds := range drivers {
		for _, build := range builds {
			if build.OperatingSystem == edgeproto.OSType_LINUX &&
				build.KernelVersion == kernVers {
				reqdBuild = build
				reqdDriverKey = driverKey
				break
			}
		}
	}
	if reqdDriverKey.Name == "" {
		return fmt.Errorf("Unable to find Linux GPU driver build for kernel version %s, node %s", kernVers, nodeName)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "found matching GPU driver", "nodename", nodeName, "driverkey", reqdDriverKey, "build", reqdBuild.Name)
	// Get path to GPU driver package file
	pkgPath, err := v.getGPUDriverPackagePath(ctx, storageClient, &reqdBuild)
	if err != nil {
		return err
	}
	// Get path to GPU driver license config file
	licenseConfigPath, err := v.getGPUDriverLicenseConfigPath(ctx, storageClient, &reqdDriverKey)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Copying GPU driver %s (%s) to node %s", reqdDriverKey.Name, reqdBuild.Name, nodeName))
	// Upload driver and license config to target node
	err = infracommon.SCPFilePath(client, pkgPath, "/tmp/")
	if err != nil {
		return fmt.Errorf("")
	}
	if licenseConfigPath != "" {
		err = infracommon.SCPFilePath(client, licenseConfigPath, "/tmp/")
		if err != nil {
			return fmt.Errorf("")
		}
	}
	// Install GPU driver, setup license and verify it
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Installing GPU driver %s (%s) on node %s", reqdDriverKey.Name, reqdBuild.Name, nodeName))
	cmd := fmt.Sprintf(
		"sudo bash /etc/mobiledgex/install-gpu-driver.sh -n %s -d %s -t %s",
		reqdDriverKey.Name,
		pkgPath,
		edgeproto.GPUType_CamelName[int32(reqdDriverKey.Type)],
	)
	if licenseConfigPath != "" {
		cmd += " -l " + licenseConfigPath
	}
	out, err = client.Output(cmd)
	if err != nil {
		return fmt.Errorf("Failed to setup GPU driver: %s, %v", out, err)
	}
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Successfully installed GPU driver on node %s", nodeName))
	return nil
}
