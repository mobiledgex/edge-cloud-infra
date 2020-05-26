package vmlayer

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	pf "github.com/mobiledgex/edge-cloud/cloud-resource-manager/platform"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

const (
	// Platform services
	ServiceTypeCRM             = "crmserver"
	ServiceTypeShepherd        = "shepherd"
	PlatformMaxWait            = 10 * time.Second
	PlatformVMReachableMaxWait = 2 * time.Minute

	// Chef Roles
	ChefRoleDocker         = "role[setup_crm_docker]"
	ChefRoleK8s            = "role[setup_crm_k8s]"
	ChefK8sMasterNodeCount = 1
	ChefK8sWorkerNodeCount = 2
)

var PlatformServices = []string{
	ServiceTypeCRM,
	ServiceTypeShepherd,
}

func (v *VMPlatform) GetPlatformVMName(key *edgeproto.CloudletKey) string {
	// Form platform VM name based on cloudletKey
	return v.VMProvider.NameSanitize(key.Name + "-" + key.Organization + "-pf")
}

func (v *VMPlatform) GetChefClientName(key *edgeproto.CloudletKey, region string) string {
	return v.VMProvider.NameSanitize(key.Name + "-" + key.Organization + "-" + region + "-pf")
}

func (v *VMPlatform) GetPlatformSubnetName(key *edgeproto.CloudletKey) string {
	return "mex-k8s-subnet-" + v.GetPlatformVMName(key)
}

func getCRMContainerVersion(ctx context.Context, client ssh.Client) (string, error) {
	var err error
	var out string

	log.SpanLog(ctx, log.DebugLevelInfra, "fetch crmserver container version")
	if out, err = client.Output(
		fmt.Sprintf("sudo docker ps --filter name=%s --format '{{.Image}}'", ServiceTypeCRM),
	); err != nil {
		return "", fmt.Errorf("unable to fetch crm version for %s, %v, %v",
			ServiceTypeCRM, err, out)
	}
	if out == "" {
		return "", fmt.Errorf("no container with name %s exists", ServiceTypeCRM)
	}
	imgParts := strings.Split(out, ":")
	return imgParts[len(imgParts)-1], nil
}

func getCRMPkgVersion(ctx context.Context, client ssh.Client) (string, error) {
	var err error
	var out string

	log.SpanLog(ctx, log.DebugLevelInfra, "fetch Cloudlet base image package version")
	if out, err = client.Output("sudo dpkg-query --showformat='${Version}' --show mobiledgex"); err != nil {
		return "", fmt.Errorf("failed to get mobiledgex debian package version, %v, %v", out, err)
	}
	return out, nil
}

func upgradeCloudletPkgs(ctx context.Context, vmType VMType, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Updating apt package lists", "cloudletName", cloudlet.Key.Name, "vmType", vmType)
	if out, err := client.Output("sudo apt-get update"); err != nil {
		return fmt.Errorf("Failed to update apt package lists, %v, %v", out, err)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Upgrading mobiledgex base image package", "cloudletName", cloudlet.Key.Name, "vmType", vmType, "packageVersion", cloudlet.PackageVersion)
	if out, err := client.Output(
		fmt.Sprintf("MEXVM_TYPE=%s sudo apt-get install -y mobiledgex=%s", vmType, cloudlet.PackageVersion),
	); err != nil {
		return fmt.Errorf("Failed to upgrade mobiledgex pkg, %v, %v", out, err)
	}
	return nil
}

func startPlatformService(cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, client ssh.Client, serviceType string, updateCallback edgeproto.CacheUpdateCallback, cDone chan error) {
	var service_cmd string
	var envVars *map[string]string
	var err error

	switch serviceType {
	case ServiceTypeShepherd:
		service_cmd, envVars, err = intprocess.GetShepherdCmd(cloudlet, pfConfig)
		if err != nil {
			cDone <- fmt.Errorf("Unable to get shepherd service command: %v", err)
			return
		}

	case ServiceTypeCRM:
		service_cmd, envVars, err = cloudcommon.GetCRMCmd(cloudlet, pfConfig)
		if err != nil {
			cDone <- fmt.Errorf("Unable to get crm service command: %v", err)
			return
		}
	default:
		cDone <- fmt.Errorf("Unsupported service type: %s", serviceType)
		return
	}

	// Use service type as container name as there can only be one of them inside platform VM
	container_name := serviceType

	// Pull docker image and start service
	updateCallback(edgeproto.UpdateTask, "Starting "+serviceType)

	var envVarsAr []string
	for k, v := range *envVars {
		envVarsAr = append(envVarsAr, "-e")
		envVarsAr = append(envVarsAr, k+"="+v)
	}
	cmd := []string{
		"sudo docker run",
		"-d",
		"--network host",
		"-v /tmp:/tmp",
		"--restart=unless-stopped",
		"--name", container_name,
		strings.Join(envVarsAr, " "),
		pfConfig.ContainerRegistryPath + ":" + pfConfig.PlatformTag,
		service_cmd,
	}
	if out, err := client.Output(strings.Join(cmd, " ")); err != nil {
		cDone <- fmt.Errorf("Unable to start %s: %v, %s\n", serviceType, err, out)
		return
	}

	// - Wait for docker container to start running
	// - And also monitor the UP state for PlatformMaxTime to
	//   catch early Fatal Logs
	// - After which controller will monitor it using CloudletInfo
	start := time.Now()
	for {
		out, err := client.Output(`sudo docker ps -a -n 1 --filter name=` + container_name + ` --format '{{.Status}}'`)
		if err != nil {
			cDone <- fmt.Errorf("Unable to fetch %s container status: %v, %s\n", serviceType, err, out)
			return
		}
		if strings.Contains(out, "Up ") {
			break
		} else if !strings.Contains(out, "Created") {
			// container exited in failure state
			// Show Fatal Log, if not Fatal log found, then show last 10 lines of error
			out, err = client.Output(`sudo docker logs ` + container_name + ` 2>&1 | grep FATAL | awk '{for (i=1; i<=NF-3; i++) $i = $(i+3); NF-=3; print}'`)
			if err != nil || out == "" {
				out, err = client.Output(`sudo docker logs ` + container_name + ` 2>&1 | tail -n 10`)
				if err != nil {
					cDone <- fmt.Errorf("Failed to bringup %s: %s", serviceType, err.Error())
					return
				}
			}
			cDone <- fmt.Errorf("Failed to bringup %s: %s", serviceType, out)
			return
		}
		elapsed := time.Since(start)
		if elapsed >= (PlatformMaxWait) {
			// no issues in wait time
			break
		}
		time.Sleep(1 * time.Second)
	}
	cDone <- nil
	return
}

func handleUpgradeError(ctx context.Context, client ssh.Client) error {
	for _, pfService := range PlatformServices {
		log.SpanLog(ctx, log.DebugLevelInfra, "restoring container names")
		if out, err := client.Output(
			fmt.Sprintf("sudo docker rename %s_old %s", pfService, pfService),
		); err != nil {
			if strings.Contains(out, "No such container") {
				continue
			}
			return fmt.Errorf("unable to restore %s_old to %s: %v, %s\n",
				pfService, pfService, err, out)
		}
	}
	return nil
}

func (v *VMPlatform) SetupPlatformService(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error {
	// Gather registry credentails from Vault
	updateCallback(edgeproto.UpdateTask, "Fetching registry auth credentials")
	regAuth, err := cloudcommon.GetRegistryAuth(ctx, pfConfig.ContainerRegistryPath, vaultConfig)
	if err != nil {
		return fmt.Errorf("unable to fetch registry auth credentials")
	}
	if regAuth.AuthType != cloudcommon.BasicAuth {
		return fmt.Errorf("unsupported registry auth type %s", regAuth.AuthType)
	}
	// Verify if controller's notify port is reachable
	updateCallback(edgeproto.UpdateTask, "Verifying if controller notification channel is reachable")
	addrPort := strings.Split(pfConfig.NotifyCtrlAddrs, ":")

	if len(addrPort) != 2 {
		return fmt.Errorf("notifyctrladdrs format is incorrect")
	}

	start := time.Now()
	for {
		out, err := client.Output(fmt.Sprintf("nc %s %s -w 5", addrPort[0], addrPort[1]))
		if err == nil {
			break
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "error trying to connect to controller port via ssh", "addrPort", addrPort, "out", out, "error", err)
			if strings.Contains(err.Error(), "ssh client timeout") || strings.Contains(err.Error(), "ssh dial fail") {
				elapsed := time.Since(start)
				if elapsed > PlatformVMReachableMaxWait {
					return fmt.Errorf("timed out connecting to platform VM to test controller notification channel")
				}
				log.SpanLog(ctx, log.DebugLevelInfra, "sleeping 10 seconds before retry", "elapsed", elapsed)
				time.Sleep(10 * time.Second)
			} else {
				return fmt.Errorf("controller's notify port is unreachable: %v, %s\n", err, out)
			}
		}
	}

	err = v.VMProvider.VerifyApiEndpoint(ctx, client, updateCallback)

	// edge-cloud image already contains the certs
	if pfConfig.TlsCertFile != "" {
		_, crtFile := filepath.Split(pfConfig.TlsCertFile)
		ext := filepath.Ext(crtFile)
		if ext == "" {
			return fmt.Errorf("invalid tls cert file name: %s", crtFile)
		}
		pfConfig.TlsCertFile = "/root/tls/" + crtFile
	}

	// Login to docker registry
	updateCallback(edgeproto.UpdateTask, "Setting up docker registry")
	if out, err := client.Output(
		fmt.Sprintf(
			`echo "%s" | sudo docker login -u %s --password-stdin %s`,
			regAuth.Password,
			regAuth.Username,
			pfConfig.ContainerRegistryPath,
		),
	); err != nil {
		return fmt.Errorf("unable to login to docker registry: %v, %s\n", err, out)
	}

	// Get non-conflicting port for NotifySrvAddr if actual port is 0
	newAddr, err := cloudcommon.GetAvailablePort(cloudlet.NotifySrvAddr)
	if err != nil {
		return err
	}
	cloudlet.NotifySrvAddr = newAddr

	// Start platform service on PlatformVM
	crmChan := make(chan error, 1)
	shepherdChan := make(chan error, 1)
	go startPlatformService(cloudlet, pfConfig, client, ServiceTypeCRM, updateCallback, crmChan)
	go startPlatformService(cloudlet, pfConfig, client, ServiceTypeShepherd, updateCallback, shepherdChan)
	// Wait for platform services to come up
	crmErr := <-crmChan
	shepherdErr := <-shepherdChan
	if crmErr != nil {
		return crmErr
	}
	return shepherdErr

}

// setupPlatformVM:
//   * Downloads Cloudlet VM base image (if not-present)
//   * Brings up Platform VM (using vm provider stack)
//   * Sets up Security Group for access to Cloudlet
// Returns ssh client
func (v *VMPlatform) SetupPlatformVM(ctx context.Context, vaultConfig *vault.Config, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetupPlatformVM", "cloudlet", cloudlet)

	platformVmName := v.GetPlatformVMName(&cloudlet.Key)
	_, err := v.VMProvider.AddCloudletImageIfNotPresent(ctx, pfConfig.CloudletVmImagePath, cloudlet.VmImageVersion, updateCallback)
	if err != nil {
		return err
	}

	updateCallback(edgeproto.UpdateTask, "Deploying Platform VM")

	vms, err := v.GetCloudletVMsSpec(ctx, vaultConfig, cloudlet, pfConfig, pfFlavor)
	if err != nil {
		return err
	}

	if cloudlet.DeploymentType == edgeproto.DeploymentType_DEPLOYMENT_TYPE_DOCKER {
		_, err = v.CreateVMsFromVMSpec(
			ctx,
			platformVmName,
			vms,
			updateCallback,
			WithNewSecurityGroup(v.GetServerSecurityGroupName(platformVmName)),
			WithAccessPorts("tcp:22"),
			WithSkipDefaultSecGrp(true),
		)
	} else {
		subnetName := v.GetPlatformSubnetName(&cloudlet.Key)
		skipSubnetRangeCheck := false
		if cloudlet.InfraAccessType == edgeproto.InfraAccessType_ACCESS_TYPE_PRIVATE {
			// It'll be end-users responsibility to make sure subnet range
			// is not confliciting with existing subnets
			skipSubnetRangeCheck = true
		}
		_, err = v.CreateVMsFromVMSpec(
			ctx,
			platformVmName,
			vms,
			updateCallback,
			WithNewSecurityGroup(v.GetServerSecurityGroupName(platformVmName)),
			WithAccessPorts("tcp:22"),
			WithSkipDefaultSecGrp(true),
			WithNewSubnet(subnetName),
			WithSkipSubnetRangeCheck(skipSubnetRangeCheck),
		)
	}
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "error while creating platform VM", "vms request spec", vms)
		return err
	}

	updateCallback(edgeproto.UpdateTask, "Successfully Deployed Platform VM")

	return nil
}

func (v *VMPlatform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error

	log.SpanLog(ctx, log.DebugLevelInfra, "Creating cloudlet", "cloudletName", cloudlet.Key.Name)

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}

	updateCallback(edgeproto.UpdateTask, "Fetching chef auth keys")
	chefAuth, err := GetChefAuthKeys(ctx, vaultConfig)
	if err != nil {
		return err
	}

	chefClient, err := GetChefClient(ctx, chefAuth.ApiKey, pfConfig.ChefServerPath)
	if err != nil {
		return err
	}

	// Source OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, "Sourcing access variables")
	log.SpanLog(ctx, log.DebugLevelInfra, "Sourcing access variables", "region", pfConfig.Region, "cloudletKey", cloudlet.Key, "PhysicalName", cloudlet.PhysicalName)
	err = v.VMProvider.InitApiAccessProperties(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig, cloudlet.EnvVar)
	if err != nil {
		return err
	}

	// TODO there's a lot of overlap between platform.PlatformConfig and edgeproto.PlatformConfig
	pc := pf.PlatformConfig{
		CloudletKey:         &cloudlet.Key,
		PhysicalName:        cloudlet.PhysicalName,
		VaultAddr:           pfConfig.VaultAddr,
		Region:              pfConfig.Region,
		TestMode:            pfConfig.TestMode,
		CloudletVMImagePath: pfConfig.CloudletVmImagePath,
		VMImageVersion:      cloudlet.VmImageVersion,
		PackageVersion:      cloudlet.PackageVersion,
		EnvVars:             pfConfig.EnvVar,
	}

	err = v.InitProps(ctx, &pc, vaultConfig)
	if err != nil {
		return err
	}

	// For real setups, ansible will always specify the correct
	// cloudlet container and vm image paths to the controller.
	// But for local testing convenience, we default to the hard-coded
	// ones if not specified.
	if pfConfig.ContainerRegistryPath == "" {
		pfConfig.ContainerRegistryPath = infracommon.DefaultContainerRegistryPath
	}

	chefAttributes, err := v.GetChefNodeAttributes(ctx, cloudlet, pfConfig)
	if err != nil {
		return err
	}

	chefRole := ChefRoleDocker
	if cloudlet.DeploymentType == edgeproto.DeploymentType_DEPLOYMENT_TYPE_K8S {
		chefRole = ChefRoleK8s
	}
	updateCallback(edgeproto.UpdateTask, "Creating chef node with cloudlet attributes")
	chefNodeName := v.GetChefClientName(&cloudlet.Key, pfConfig.Region)
	err = ChefNodeCreate(ctx, chefClient, chefNodeName, chefRole, chefAttributes)
	if err != nil {
		return err
	}
	if cloudlet.InfraAccessType == edgeproto.InfraAccessType_ACCESS_TYPE_PRIVATE {
		// Return as for PRIVATE access type, end-user will setup the platform VM
		return nil
	}

	startTime := time.Now()

	err = v.SetupPlatformVM(ctx, vaultConfig, cloudlet, pfConfig, pfFlavor, updateCallback)
	if err != nil {
		return err
	}

	updateCallback(edgeproto.UpdateTask, "Waiting for Platform VM to connect to Chef Server")
	// Wait for VM to connect to Chef Server
	clientName := v.GetChefClientName(&cloudlet.Key, pfConfig.Region)
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		found := false
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for platform VM to connect to Chef Server")
		case <-tick:
			found, err = ChefClientExists(ctx, chefClient, clientName)
			if err != nil {
				return err
			}
		}
		if found {
			updateCallback(edgeproto.UpdateTask, "Platform VM connected to Chef Server")
			break
		}
	}

	// Fetch chef run list status
	updateCallback(edgeproto.UpdateTask, "Waiting for run lists to be executed on Platform VM")
	timeout = time.After(10 * time.Minute)
	tick = time.Tick(5 * time.Second)
	for {
		var statusInfo []ChefStatusInfo
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for platform VM to connect to Chef Server")
		case <-tick:
			statusInfo, err = ChefNodeRunStatus(ctx, chefClient, clientName, startTime)
			if err != nil {
				return err
			}
		}
		if len(statusInfo) > 0 {
			updateCallback(edgeproto.UpdateTask, "Performed following actions:")
			for _, info := range statusInfo {
				if info.Failed {
					return fmt.Errorf(info.Message)
				}
				updateCallback(edgeproto.UpdateStep, info.Message)
			}
			break
		}
	}

	return nil
}

func (v *VMPlatform) CleanupCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Cleaning up cloudlet", "cloudletName", cloudlet.Key.Name)

	client, err := v.GetSSHClientForServer(ctx, v.GetPlatformVMName(&cloudlet.Key), v.VMProperties.GetCloudletExternalNetwork())
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Removing old containers")
	for _, pfService := range PlatformServices {
		if out, err := client.Output(
			fmt.Sprintf("sudo docker rm -f %s_old", pfService),
		); err != nil {
			if strings.Contains(out, "No such container") {
				log.SpanLog(ctx, log.DebugLevelInfra, "no containers to cleanup")
				continue
			} else {
				return fmt.Errorf("cleanup failed: %v, %s\n", err, out)
			}
		}
	}
	return nil
}

func (v *VMPlatform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Deleting cloudlet", "cloudletName", cloudlet.Key.Name)

	updateCallback(edgeproto.UpdateTask, "Deleting cloudlet")

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}

	chefAuth, err := GetChefAuthKeys(ctx, vaultConfig)
	if err != nil {
		return err
	}

	chefClient, err := GetChefClient(ctx, chefAuth.ApiKey, pfConfig.ChefServerPath)
	if err != nil {
		return err
	}

	if cloudlet.InfraAccessType == edgeproto.InfraAccessType_ACCESS_TYPE_PUBLIC {
		// Source OpenRC file to access openstack API endpoint
		err = v.VMProvider.InitApiAccessProperties(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig, cloudlet.EnvVar)
		if err != nil {
			// ignore this error, as no creation would've happened on infra, so nothing to delete
			log.SpanLog(ctx, log.DebugLevelInfra, "failed to source platform variables", "cloudletName", cloudlet.Key.Name, "err", err)
			return nil
		}

		platformVMName := v.GetPlatformVMName(&cloudlet.Key)
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting PlatformVM %s", platformVMName))
		err = v.VMProvider.DeleteVMs(ctx, platformVMName)
		if err != nil {
			return fmt.Errorf("DeleteCloudlet error: %v", err)
		}

		rootLBName := v.GetRootLBName(&cloudlet.Key)
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting RootLB %s", rootLBName))
		err = v.VMProvider.DeleteVMs(ctx, rootLBName)
		if err != nil {
			return fmt.Errorf("DeleteCloudlet error: %v", err)
		}
	}

	clientName := v.GetChefClientName(&cloudlet.Key, pfConfig.Region)
	updateCallback(edgeproto.UpdateTask, "Deleting client from Chef Server")
	err = ChefClientDelete(ctx, chefClient, clientName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to delete client from Chef Server", "clientName", clientName, "err", err)
	}

	// Not sure if it's safe to remove vars from Vault due to testing/virtual cloudlets,
	// so leaving them in Vault for the time being. We can always delete them manually

	return nil
}

func (v *VMPlatform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) (edgeproto.CloudletAction, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Updating cloudlet", "cloudletName", cloudlet.Key.Name)

	defCloudletAction := edgeproto.CloudletAction_ACTION_NONE

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return defCloudletAction, err
	}
	// Source OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Sourcing platform variables for %s cloudlet", cloudlet.PhysicalName))
	err = v.VMProvider.InitApiAccessProperties(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig, cloudlet.EnvVar)
	if err != nil {
		return defCloudletAction, err
	}

	pfClient, err := v.GetSSHClientForServer(ctx, v.GetPlatformVMName(&cloudlet.Key), v.VMProperties.GetCloudletExternalNetwork())
	if err != nil {
		return defCloudletAction, err
	}

	containerVersion, err := getCRMContainerVersion(ctx, pfClient)
	if err != nil {
		return defCloudletAction, err
	}

	rootLBName := cloudcommon.GetRootLBFQDN(&cloudlet.Key)
	rlbClient, err := v.GetSSHClientForServer(ctx, rootLBName, v.VMProperties.GetCloudletExternalNetwork())
	if err != nil {
		return defCloudletAction, err
	}
	upgradeMap := map[VMType]ssh.Client{
		VMTypePlatform: pfClient,
		VMTypeRootLB:   rlbClient,
	}
	for vmType, client := range upgradeMap {
		if cloudlet.PackageVersion == "" {
			// No package upgrade required
			break
		}
		pkgVersion, err := getCRMPkgVersion(ctx, client)
		if err != nil {
			return defCloudletAction, err
		}
		if cloudlet.PackageVersion == pkgVersion {
			continue
		}

		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Upgrading mobiledgex base image package for %s to version %s", vmType, cloudlet.PackageVersion))
		err = upgradeCloudletPkgs(ctx, vmType, cloudlet, pfConfig, vaultConfig, client, updateCallback)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Failed to upgrade cloudlet packages", "VM type", vmType, "Version", cloudlet.PackageVersion, "err", err)
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Failed to upgrade cloudlet packages of vm type %s to version %s, please upgrade them manually!", vmType, cloudlet.PackageVersion))
			return defCloudletAction, err
		}
	}

	if containerVersion == cloudlet.ContainerVersion {
		// No service upgrade required
		return edgeproto.CloudletAction_ACTION_DONE, nil
	}

	// Rename existing containers
	for _, pfService := range PlatformServices {
		from := pfService
		to := pfService + "_old"
		log.SpanLog(ctx, log.DebugLevelInfra, "renaming existing services to bringup new ones", "from", from, "to", to)
		if out, err := pfClient.Output(
			fmt.Sprintf("sudo docker rename %s %s", from, to),
		); err != nil {
			errStr := fmt.Sprintf("unable to rename %s to %s: %v, %s\n",
				from, to, err, out)
			err = handleUpgradeError(ctx, pfClient)
			if err == nil {
				return defCloudletAction, errors.New(errStr)
			} else {
				return defCloudletAction, fmt.Errorf("%s. Cleanup failed as well: %v\n", errStr, err)
			}
		}
	}

	err = v.SetupPlatformService(ctx, cloudlet, pfConfig, vaultConfig, pfClient, updateCallback)

	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to setup platform services", "err", err)
		// Cleanup failed containers
		updateCallback(edgeproto.UpdateTask, "Upgrade failed, cleaning up")
		if out, err1 := pfClient.Output(
			fmt.Sprintf("sudo docker rm -f %s", strings.Join(PlatformServices, " ")),
		); err1 != nil {
			if strings.Contains(out, "No such container") {
				log.SpanLog(ctx, log.DebugLevelInfra, "no containers to cleanup")
			} else {
				return defCloudletAction, fmt.Errorf("upgrade failed: %v and cleanup failed: %v, %s\n", err, err1, out)
			}
		}
		// Cleanup container names
		for _, pfService := range PlatformServices {
			from := pfService + "_old"
			to := pfService
			log.SpanLog(ctx, log.DebugLevelInfra, "restoring old container name", "from", from, "to", to)
			if out, err1 := pfClient.Output(
				fmt.Sprintf("sudo docker rename %s %s", from, to),
			); err1 != nil {
				return defCloudletAction, fmt.Errorf("upgrade failed: %v and unable to rename old-container: %v, %s\n", err, err1, out)
			}
		}
		return defCloudletAction, err
	}
	return edgeproto.CloudletAction_ACTION_IN_PROGRESS, nil
}

func (v *VMPlatform) DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Deleting access vars from vault", "cloudletName", cloudlet.Key.Name)

	updateCallback(edgeproto.UpdateTask, "Deleting access vars from secure secrets storage")

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	path := GetVaultCloudletAccessPath(&cloudlet.Key, v.Type, pfConfig.Region, cloudlet.PhysicalName)
	err = infracommon.DeleteDataFromVault(vaultConfig, path)
	if err != nil {
		return fmt.Errorf("Failed to delete access vars from vault: %v", err)
	}
	return nil
}

func (v *VMPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	return v.VMProvider.SaveCloudletAccessVars(ctx, cloudlet, accessVarsIn, pfConfig, updateCallback)
}

func (v *VMPlatform) GatherCloudletInfo(ctx context.Context, info *edgeproto.CloudletInfo) error {
	return v.VMProvider.GatherCloudletInfo(ctx, info)
}

func (v *VMPlatform) GetChefNodeAttributes(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig) (map[string]interface{}, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetChefNodeAttributes", "region", pfConfig.Region, "cloudletKey", cloudlet.Key, "PhysicalName", cloudlet.PhysicalName)

	chefAttributes := make(map[string]interface{})

	if cloudlet.DeploymentType == edgeproto.DeploymentType_DEPLOYMENT_TYPE_K8S {
		chefAttributes["k8sNodeCount"] = ChefK8sMasterNodeCount + ChefK8sWorkerNodeCount
	}
	chefAttributes["edgeCloudImage"] = pfConfig.ContainerRegistryPath
	chefAttributes["edgeCloudVersion"] = cloudlet.ContainerVersion
	chefAttributes["notifyAddrs"] = pfConfig.NotifyCtrlAddrs

	for _, serviceType := range PlatformServices {
		serviceObj := make(map[string]interface{})
		var serviceCmdArgs []string
		var envVars *map[string]string
		var err error
		switch serviceType {
		case ServiceTypeShepherd:
			serviceCmdArgs, envVars, err = intprocess.GetShepherdCmdArgs(cloudlet, pfConfig)
			if err != nil {
				return nil, err
			}
		case ServiceTypeCRM:
			serviceCmdArgs, envVars, err = cloudcommon.GetCRMCmdArgs(cloudlet, pfConfig)
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
		chefArgs := GetChefArgs(ctx, serviceCmdArgs)
		serviceObj["args"] = chefArgs
		if envVars != nil {
			envVarArr := []string{}
			for k, v := range *envVars {
				envVar := fmt.Sprintf("%s=%s", k, v)
				envVarArr = append(envVarArr, envVar)
			}
			serviceObj["env"] = envVarArr
		}
		chefAttributes[serviceType] = serviceObj
	}

	apiAddr, err := v.VMProvider.GetApiEndpointAddr(ctx)
	if err != nil {
		return nil, err
	}
	urlObj, err := util.ImagePathParse(apiAddr)
	if err != nil {
		return nil, err
	}
	hostname := strings.Split(urlObj.Host, ":")
	if len(hostname) != 2 {
		return nil, fmt.Errorf("invalid api endpoint addr: %s", apiAddr)
	}
	// API Endpoint address might have hostname in it, hence resolve the addr
	endpointIp, err := infracommon.LookupDNS(hostname[0])
	if err != nil {
		return nil, err
	}
	chefAttributes["infraApiAddr"] = endpointIp
	chefAttributes["infraApiPort"] = hostname[1]
	if cloudlet.InfraAccessType == edgeproto.InfraAccessType_ACCESS_TYPE_PUBLIC {
		// Fetch gateway IP of external network
		gatewayAddr, err := v.VMProvider.GetExternalGateway(ctx, v.VMProperties.GetCloudletExternalNetwork())
		if err != nil {
			return nil, fmt.Errorf("unable to fetch gateway IP for external network: %s, %v",
				v.VMProperties.GetCloudletExternalNetwork(), err)
		}
		chefAttributes["infraApiGw"] = gatewayAddr
	}
	return chefAttributes, nil
}

func (v *VMPlatform) GetCloudletVMsSpec(ctx context.Context, vaultConfig *vault.Config, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor) ([]*VMRequestSpec, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Sourcing access variables", "region", pfConfig.Region, "cloudletKey", cloudlet.Key, "PhysicalName", cloudlet.PhysicalName)
	err := v.VMProvider.InitApiAccessProperties(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig, cloudlet.EnvVar)
	if err != nil {
		return nil, err
	}
	// edge-cloud image already contains the certs
	if pfConfig.TlsCertFile != "" {
		_, crtFile := filepath.Split(pfConfig.TlsCertFile)
		ext := filepath.Ext(crtFile)
		if ext == "" {
			return nil, fmt.Errorf("invalid tls cert file name: %s", crtFile)
		}
		pfConfig.TlsCertFile = "/root/tls/" + crtFile
	}
	// TODO there's a lot of overlap between platform.PlatformConfig and edgeproto.PlatformConfig
	pc := pf.PlatformConfig{
		CloudletKey:         &cloudlet.Key,
		PhysicalName:        cloudlet.PhysicalName,
		VaultAddr:           pfConfig.VaultAddr,
		Region:              pfConfig.Region,
		TestMode:            pfConfig.TestMode,
		CloudletVMImagePath: pfConfig.CloudletVmImagePath,
		VMImageVersion:      cloudlet.VmImageVersion,
		PackageVersion:      cloudlet.PackageVersion,
		EnvVars:             pfConfig.EnvVar,
	}

	err = v.InitProps(ctx, &pc, vaultConfig)
	if err != nil {
		return nil, err
	}

	if pfConfig.ContainerRegistryPath == "" {
		pfConfig.ContainerRegistryPath = infracommon.DefaultContainerRegistryPath
	}

	if cloudlet.InfraExternalNetwork != "" {
		v.VMProperties.SetCloudletExternalNetwork(cloudlet.InfraExternalNetwork)
	}

	flavorName := cloudlet.InfraFlavorName
	if cloudlet.InfraAccessType == edgeproto.InfraAccessType_ACCESS_TYPE_PUBLIC {
		// Validate infra external network provided by user
		if cloudlet.InfraExternalNetwork != "" {
			nets, err := v.VMProvider.GetNetworkList(ctx)
			if err != nil {
				return nil, err
			}

			found := false
			for _, n := range nets {
				if n == cloudlet.InfraExternalNetwork {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("cannot find infra external network %s", cloudlet.InfraExternalNetwork)
			}
		}
		flavorList, err := v.VMProvider.GetFlavorList(ctx)
		if err != nil {
			return nil, err
		}
		if cloudlet.InfraFlavorName == "" {
			vmspec, err := vmspec.GetVMSpec(flavorList, *pfFlavor)
			if err != nil {
				return nil, fmt.Errorf("unable to find VM spec for Shared RootLB: %v", err)
			}
			flavorName = vmspec.FlavorName
		} else {
			// Validate infra flavor name provided by user
			for _, finfo := range flavorList {
				if finfo.Name == cloudlet.InfraFlavorName {
					flavorName = cloudlet.InfraFlavorName
					break
				}
			}
			if flavorName == "" {
				return nil, fmt.Errorf("invalid infraflavorname, does not exist")
			}
		}

	}
	if flavorName == "" {
		return nil, fmt.Errorf("unable to fetch platform flavor")
	}

	platformVmName := v.GetPlatformVMName(&cloudlet.Key)
	imgPath := GetCloudletVMImagePath(pfConfig.CloudletVmImagePath, cloudlet.VmImageVersion)
	pfImageName, err := cloudcommon.GetFileName(imgPath)
	if err != nil {
		return nil, err
	}

	// Setup Chef parameters in cloud-config
	chefNodeName := v.GetChefClientName(&cloudlet.Key, pfConfig.Region)
	chefAuth, err := GetChefAuthKeys(ctx, vaultConfig)
	if err != nil {
		return nil, err
	}
	chefParams := ChefParams{
		NodeName:       chefNodeName,
		ServerPath:     pfConfig.ChefServerPath,
		ValidationKey:  chefAuth.ValidationKey,
		ClientInterval: pfConfig.ChefClientInterval,
	}

	buf, err := ExecTemplate(chefParams.NodeName, VmChefConfig, chefParams)
	if err != nil {
		return nil, err
	}
	chefManifest := buf.String()

	var vms []*VMRequestSpec

	subnetName := v.GetPlatformSubnetName(&cloudlet.Key)

	if cloudlet.DeploymentType == edgeproto.DeploymentType_DEPLOYMENT_TYPE_DOCKER {
		platvm, err := v.GetVMRequestSpec(
			ctx,
			VMTypePlatform,
			platformVmName,
			flavorName,
			pfImageName,
			true, //connect external
			WithDeploymentManifest(VmCloudConfig+chefManifest),
		)
		if err != nil {
			return nil, err
		}
		vms = append(vms, platvm)
	} else {
		rootlb, err := v.GetVMRequestSpec(
			ctx,
			VMTypeRootLB,
			platformVmName+"-lb",
			flavorName,
			v.VMProperties.GetCloudletOSImage(),
			true, //connect external
			WithSubnetConnection(subnetName),
			WithDeploymentManifest(VmCloudConfig+chefManifest),
		)
		if err != nil {
			return nil, err
		}
		vms = append(vms, rootlb)
		// create ports on the rootLB
		rootlbPort, err := v.GetVMSpecForRootLBPorts(ctx, platformVmName+"-lb", subnetName)
		if err != nil {
			return nil, err
		}
		vms = append(vms, rootlbPort)
		master, err := v.GetVMRequestSpec(
			ctx,
			VMTypeClusterMaster,
			platformVmName+"-master",
			flavorName,
			v.VMProperties.GetCloudletOSImage(),
			false, //connect external
			WithSubnetConnection(subnetName),
			WithDeploymentManifest(VmCloudConfig),
		)
		if err != nil {
			return nil, err
		}
		vms = append(vms, master)
		for nn := uint32(1); nn <= ChefK8sWorkerNodeCount; nn++ {
			node, err := v.GetVMRequestSpec(ctx,
				VMTypeClusterNode,
				fmt.Sprintf("%s-node-%d", platformVmName, nn),
				flavorName,
				v.VMProperties.GetCloudletOSImage(),
				false, //connect external
				WithSubnetConnection(subnetName),
				WithDeploymentManifest(VmCloudConfig),
			)
			if err != nil {
				return nil, err
			}
			vms = append(vms, node)
		}
	}

	return vms, nil
}

func (v *VMPlatform) GetCloudletManifest(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor) (*edgeproto.CloudletManifest, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Get cloudlet manifest", "cloudletName", cloudlet.Key.Name)

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return nil, err
	}

	platvms, err := v.GetCloudletVMsSpec(ctx, vaultConfig, cloudlet, pfConfig, pfFlavor)
	if err != nil {
		return nil, err
	}

	platformVmName := v.GetPlatformVMName(&cloudlet.Key)

	var gp *VMGroupOrchestrationParams
	if cloudlet.DeploymentType == edgeproto.DeploymentType_DEPLOYMENT_TYPE_DOCKER {
		gp, err = v.GetVMGroupOrchestrationParamsFromVMSpec(
			ctx,
			platformVmName,
			platvms,
			WithNewSecurityGroup(v.GetServerSecurityGroupName(platformVmName)),
			WithAccessPorts("tcp:22"),
			WithSkipDefaultSecGrp(true),
		)
	} else {
		subnetName := v.GetPlatformSubnetName(&cloudlet.Key)
		skipSubnetRangeCheck := false
		if cloudlet.InfraAccessType == edgeproto.InfraAccessType_ACCESS_TYPE_PRIVATE {
			// It'll be end-users responsibility to make sure subnet range
			// is not confliciting with existing subnets
			skipSubnetRangeCheck = true
		}
		gp, err = v.GetVMGroupOrchestrationParamsFromVMSpec(
			ctx,
			platformVmName,
			platvms,
			WithNewSecurityGroup(v.GetServerSecurityGroupName(platformVmName)),
			WithAccessPorts("tcp:22"),
			WithSkipDefaultSecGrp(true),
			WithNewSubnet(subnetName),
			//WithSkipSubnetGw(true),
			WithSkipSubnetRangeCheck(skipSubnetRangeCheck),
		)
	}
	if err != nil {
		return nil, err
	}
	manifest, err := v.VMProvider.GetCloudletManifest(ctx, platformVmName, gp)
	if err != nil {
		return nil, err
	}
	imgPath := GetCloudletVMImagePath(pfConfig.CloudletVmImagePath, cloudlet.VmImageVersion)

	return &edgeproto.CloudletManifest{
		Manifest:  manifest,
		ImagePath: imgPath,
	}, nil
}
