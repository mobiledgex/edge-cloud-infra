package openstack

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/mobiledgex/edge-cloud/vmspec"
	ssh "github.com/mobiledgex/golang-ssh"
)

const (
	// Platform services
	ServiceTypeCRM      = "crmserver"
	ServiceTypeShepherd = "shepherd"
	PlatformMaxWait     = 10 * time.Second
)

var PlatformServices = []string{
	ServiceTypeCRM,
	ServiceTypeShepherd,
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
		pfConfig.RegistryPath + ":" + pfConfig.PlatformTag,
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

func setupPlatformService(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error {
	// Gather registry credentails from Vault
	updateCallback(edgeproto.UpdateTask, "Fetching registry auth credentials")
	regAuth, err := cloudcommon.GetRegistryAuth(ctx, pfConfig.RegistryPath, vaultConfig)
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
	if out, err := client.Output(
		fmt.Sprintf(
			"nc %s %s -w 5", addrPort[0], addrPort[1],
		),
	); err != nil {
		return fmt.Errorf("controller's notify port is unreachable: %v, %s\n", err, out)
	}

	// Verify if Openstack API Endpoint is reachable
	updateCallback(edgeproto.UpdateTask, "Verifying if Openstack API Endpoint is reachable")
	osAuthUrl := os.Getenv("OS_AUTH_URL")
	if osAuthUrl == "" {
		return fmt.Errorf("unable to find OS_AUTH_URL")
	}
	urlObj, err := url.Parse(osAuthUrl)
	if err != nil {
		return fmt.Errorf("unable to parse OS_AUTH_URL: %s, %v\n", osAuthUrl, err)
	}
	if _, err := client.Output(
		fmt.Sprintf(
			"nc %s %s -w 5", urlObj.Hostname(), urlObj.Port(),
		),
	); err != nil {
		updateCallback(edgeproto.UpdateTask, "Adding route for API endpoint as it is unreachable")
		// Fetch gateway IP of external network
		gatewayAddr, err := mexos.GetExternalGateway(ctx, mexos.GetCloudletExternalNetwork())
		if err != nil {
			return fmt.Errorf("unable to fetch gateway IP for external network: %s, %v",
				mexos.GetCloudletExternalNetwork(), err)
		}
		// Add route to reach API endpoint
		if out, err := client.Output(
			fmt.Sprintf(
				"sudo route add -host %s gw %s", urlObj.Hostname(), gatewayAddr,
			),
		); err != nil {
			return fmt.Errorf("unable to add route to reach API endpoint: %v, %s\n", err, out)
		}
		interfacesFile := mexos.GetCloudletNetworkIfaceFile()
		routeAddLine := fmt.Sprintf("up route add -host %s gw %s", urlObj.Hostname(), gatewayAddr)
		cmd := fmt.Sprintf("grep -l '%s' %s", routeAddLine, interfacesFile)
		_, err = client.Output(cmd)
		if err != nil {
			// grep failed so not there already
			log.SpanLog(ctx, log.DebugLevelMexos, "adding route to interfaces file", "route", routeAddLine, "file", interfacesFile)
			cmd = fmt.Sprintf("echo '%s'|sudo tee -a %s", routeAddLine, interfacesFile)
			out, err := client.Output(cmd)
			if err != nil {
				return fmt.Errorf("can't add route '%s' to interfaces file: %v, %s", routeAddLine, err, out)
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelMexos, "route already present in interfaces file")
		}
		// Retry
		updateCallback(edgeproto.UpdateTask, "Retrying verification of reachability of Openstack API endpoint")
		if out, err := client.Output(
			fmt.Sprintf(
				"nc %s %s -w 5", urlObj.Hostname(), urlObj.Port(),
			),
		); err != nil {
			return fmt.Errorf("Openstack API Endpoint is unreachable: %v, %s\n", err, out)
		}
	}

	// edge-cloud image already contains the certs
	_, crtFile := filepath.Split(pfConfig.TlsCertFile)
	ext := filepath.Ext(crtFile)
	if ext == "" {
		return fmt.Errorf("invalid tls cert file name: %s", crtFile)
	}
	pfConfig.TlsCertFile = "/root/tls/" + crtFile

	// Login to docker registry
	updateCallback(edgeproto.UpdateTask, "Setting up docker registry")
	if out, err := client.Output(
		fmt.Sprintf(
			`echo "%s" | sudo docker login -u %s --password-stdin %s`,
			regAuth.Password,
			regAuth.Username,
			pfConfig.RegistryPath,
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
//   * Brings up Platform VM (using HEAT stack)
//   * Sets up Security Group for access to Cloudlet
// Returns ssh client
func setupPlatformVM(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) (ssh.Client, error) {
	curStackNames, err := getCurrentStackName(ctx, &cloudlet.Key, mexos.PlatformVMDeployment)
	if err != nil {
		return nil, err
	}
	if len(curStackNames) > 0 {
		return nil, fmt.Errorf("Platform VM for this cloudlet already exists: %v", curStackNames)
	}

	// Get Flavor Info
	finfo, _, _, err := mexos.GetFlavorInfo(ctx)
	if err != nil {
		return nil, err
	}
	// Get Closest Platform Flavor
	vmspec, err := vmspec.GetVMSpec(finfo, *pfFlavor)
	if err != nil {
		return nil, fmt.Errorf("unable to find matching vm spec for platform: %v", err)
	}

	pfImageName, err := mexos.AddImageIfNotPresent(ctx, pfConfig.ImagePath, cloudlet.ImageVersion, updateCallback)
	if err != nil {
		return nil, err
	}

	// Form platform VM name based on cloudletKey
	platform_vm_name := mexos.GetPlatformVMName(&cloudlet.Key)
	stack_name := mexos.GetStackName(&cloudlet.Key, cloudlet.ImageVersion, mexos.PlatformVMDeployment)
	secGrp := mexos.GetSecurityGroupName(ctx, platform_vm_name)

	vmp, err := mexos.GetVMParams(ctx,
		mexos.PlatformVMDeployment,
		platform_vm_name,
		vmspec.FlavorName,
		vmspec.ExternalVolumeSize,
		pfImageName,
		secGrp,
		&cloudlet.Key,
		mexos.WithAccessPorts("tcp:22"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get vm params: %v", err)
	}

	// Deploy Platform VM
	updateCallback(edgeproto.UpdateTask, "Deploying Platform VM")
	log.SpanLog(ctx, log.DebugLevelMexos, "Deploying VM", "stackName", stack_name, "vmspec", vmspec)
	err = mexos.CreateHeatStackFromTemplate(ctx, vmp, stack_name, mexos.VmTemplate, updateCallback)
	if err != nil {
		return nil, fmt.Errorf("CreatePlatformVM error: %v", err)
	}
	updateCallback(edgeproto.UpdateTask, "Successfully Deployed Platform VM")

	external_ip, err := mexos.GetServerIPAddr(ctx, mexos.GetCloudletExternalNetwork(), platform_vm_name, mexos.ExternalIPType)
	if err != nil {
		return nil, err
	}
	updateCallback(edgeproto.UpdateTask, "Platform VM external IP: "+external_ip)

	client, err := mexos.GetSSHClient(ctx, platform_vm_name, mexos.GetCloudletExternalNetwork(), mexos.SSHUser)
	if err != nil {
		return nil, err
	}

	// setup SSH access to cloudlet for CRM
	updateCallback(edgeproto.UpdateTask, "Setting up security group for SSH access")

	if err := mexos.AddSecurityRuleCIDR(ctx, external_ip, "tcp", secGrp, "22"); err != nil {
		return nil, fmt.Errorf("unable to add security rule for ssh access, err: %v", err)
	}

	return client, nil
}

func (s *Platform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error

	log.SpanLog(ctx, log.DebugLevelMexos, "Creating cloudlet", "cloudletName", cloudlet.Key.Name)

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	// Source OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, "Sourcing access variables")
	err = mexos.InitOpenstackProps(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig)
	if err != nil {
		return err
	}

	// For real setups, ansible will always specify the correct
	// cloudlet container and vm image paths to the controller.
	// But for local testing convenience, we default to the hard-coded
	// ones if not specified.
	if pfConfig.RegistryPath == "" {
		pfConfig.RegistryPath = mexos.DefaultCloudletRegistryPath
	}

	client, err := setupPlatformVM(ctx, cloudlet, pfConfig, pfFlavor, updateCallback)
	if err != nil {
		return err
	}

	return setupPlatformService(ctx, cloudlet, pfConfig, vaultConfig, client, updateCallback)
}

func (s *Platform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Saving cloudlet access vars to vault", "cloudletName", cloudlet.Key.Name)
	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	openrcData, ok := accessVarsIn["OPENRC_DATA"]
	if !ok {
		return fmt.Errorf("Invalid accessvars, missing OPENRC_DATA")
	}
	out := strings.Split(openrcData, "\n")
	if len(out) <= 1 {
		return fmt.Errorf("Invalid accessvars, as OPENRC_DATA is invalid: %v", out)
	}
	accessVars := make(map[string]string)
	for _, v := range out {
		out1 := strings.Split(v, "=")
		if len(out1) != 2 {
			return fmt.Errorf("Invalid separator for key-value pair: %v", out1)
		}
		key := strings.TrimSpace(out1[0])
		value := strings.TrimSpace(out1[1])
		if !strings.HasPrefix(key, "OS_") {
			return fmt.Errorf("Invalid accessvars: %s, must start with 'OS_' prefix", key)
		}
		accessVars[key] = value
	}
	authURL, ok := accessVars["OS_AUTH_URL"]
	if !ok {
		return fmt.Errorf("Invalid accessvars, missing OS_AUTH_URL")
	}
	if strings.HasPrefix(authURL, "https") {
		certData, ok := accessVarsIn["CACERT_DATA"]
		if !ok {
			return fmt.Errorf("Invalid accessvars, missing CACERT_DATA")
		}
		certFile := mexos.GetCertFilePath(&cloudlet.Key)
		err := ioutil.WriteFile(certFile, []byte(certData), 0644)
		if err != nil {
			return err
		}
		accessVars["OS_CACERT"] = certFile
		accessVars["OS_CACERT_DATA"] = certData
	}
	updateCallback(edgeproto.UpdateTask, "Saving access vars to secure secrets storage (Vault)")
	var varList mexos.VaultEnvData
	for key, value := range accessVars {
		if key == "OS_CACERT" {
			continue
		}
		varList.Env = append(varList.Env, mexos.EnvData{
			Name:  key,
			Value: value,
		})
	}
	data := map[string]interface{}{
		"data": varList,
	}

	path := mexos.GetVaultCloudletPath(&cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, mexos.OSAccessVars)
	err = mexos.PutDataToVault(vaultConfig, path, data)
	if err != nil {
		updateCallback(edgeproto.UpdateTask, "Failed to save access vars to vault")
		log.SpanLog(ctx, log.DebugLevelMexos, err.Error(), "cloudletName", cloudlet.Key.Name)
		return fmt.Errorf("Failed to save access vars to vault: %v", err)
	}
	return nil
}

func (s *Platform) DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Deleting access vars from vault", "cloudletName", cloudlet.Key.Name)

	updateCallback(edgeproto.UpdateTask, "Deleting access vars from secure secrets storage")

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	path := mexos.GetVaultCloudletPath(&cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, mexos.OSAccessVars)
	err = mexos.DeleteDataFromVault(vaultConfig, path)
	if err != nil {
		return fmt.Errorf("Failed to delete access vars from vault: %v", err)
	}
	return nil
}

func (s *Platform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Deleting cloudlet", "cloudletName", cloudlet.Key.Name)

	updateCallback(edgeproto.UpdateTask, "Deleting cloudlet")

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	// Source OpenRC file to access openstack API endpoint
	err = mexos.InitOpenstackProps(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig)
	if err != nil {
		// ignore this error, as no creation would've happened on infra, so nothing to delete
		log.SpanLog(ctx, log.DebugLevelMexos, "failed to source platform variables", "cloudletName", cloudlet.Key.Name, "err", err)
		return nil
	}

	for _, vmType := range []mexos.DeploymentType{mexos.PlatformVMDeployment, mexos.RootLBVMDeployment} {
		stackNames, err := getCurrentStackName(ctx, &cloudlet.Key, vmType)
		if err != nil {
			return err
		}
		for _, stackName := range stackNames {
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting %s Stack %s", vmType, stackName))
			err = mexos.HeatDeleteStack(ctx, stackName)
			if err != nil {
				return fmt.Errorf("DeleteCloudlet error: %v", err)
			}
		}
	}

	// Not sure if it's safe to remove vars from Vault due to testing/virtual cloudlets,
	// so leaving them in Vault for the time being. We can always delete them manually

	return nil
}

func handleUpgradeError(ctx context.Context, client ssh.Client) error {
	for _, pfService := range PlatformServices {
		log.SpanLog(ctx, log.DebugLevelMexos, "restoring container names")
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

func getCurrentStackName(ctx context.Context, key *edgeproto.CloudletKey, vmType mexos.DeploymentType) ([]string, error) {
	stackNames := []string{}
	stacks, err := mexos.ListHeatStacks(ctx)
	if err != nil {
		return stackNames, fmt.Errorf("Failed to list heat stacks: %v", err)
	}
	// Check if PlatformVM image upgrade is required
	stackSuffix := mexos.GetStackSuffix(key, vmType)
	for _, heatStack := range stacks {
		if strings.HasSuffix(heatStack.StackName, stackSuffix) {
			prefix := strings.TrimSuffix(heatStack.StackName, stackSuffix)
			if prefix != "" && strings.Count(prefix, "_") > 1 {
				continue
			}
			stackNames = append(stackNames, heatStack.StackName)
		}
	}
	return stackNames, nil
}

func IsImageUpgradeRequired(ctx context.Context, key *edgeproto.CloudletKey, imgVersion string, vmType mexos.DeploymentType) (bool, error) {
	stackNames, err := getCurrentStackName(ctx, key, vmType)
	if err != nil {
		return false, err
	}
	if len(stackNames) == 0 {
		return true, nil
	}
	if len(stackNames) > 1 {
		return false, fmt.Errorf("More than one stack exists: %v, Please delete unused stacks", stackNames)
	}
	curStack := stackNames[0]
	serverDetails, err := mexos.GetServerDetails(ctx, curStack)
	if err != nil {
		return false, fmt.Errorf("Failed to server detail for %s: %v", curStack, err)
	}
	// VM exists, check if it is already using the same baseimage
	// Fetch CRM baseimage name
	pfImageName := mexos.GetCloudletVMImageName(imgVersion)
	if serverDetails.Image != pfImageName {
		return true, nil
	}
	return false, nil
}

func (s *Platform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Updating cloudlet", "cloudletName", cloudlet.Key.Name)

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	// Source OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Sourcing platform variables for %s cloudlet", cloudlet.PhysicalName))
	err = mexos.InitOpenstackProps(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig)
	if err != nil {
		return err
	}

	updateCallback(edgeproto.UpdateTask, "Verifying if Platform VM image version is updated")
	upgradeRequired, err := IsImageUpgradeRequired(ctx, &cloudlet.Key, cloudlet.ImageVersion, mexos.PlatformVMDeployment)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "Image upgrade check failed", "err", err)
	}

	var pfClient ssh.Client
	if upgradeRequired && cloudlet.ImageUpgrade {
		updateCallback(edgeproto.UpdateTask, "Upgrading Platform VM")
		pfClient, err = setupPlatformVM(ctx, cloudlet, pfConfig, pfFlavor, updateCallback)
		if err != nil {
			return err
		}
	} else {
		pfClient, err = mexos.GetSSHClient(ctx, mexos.GetPlatformVMName(&cloudlet.Key), mexos.GetCloudletExternalNetwork(), mexos.SSHUser)
		if err != nil {
			return err
		}

		// Rename existing containers
		for _, pfService := range PlatformServices {
			from := pfService
			to := pfService + "_old"
			log.SpanLog(ctx, log.DebugLevelMexos, "renaming existing services to bringup new ones", "from", from, "to", to)
			if out, err := pfClient.Output(
				fmt.Sprintf("sudo docker rename %s %s", from, to),
			); err != nil {
				errStr := fmt.Sprintf("unable to rename %s to %s: %v, %s\n",
					from, to, err, out)
				err = handleUpgradeError(ctx, pfClient)
				if err == nil {
					return errors.New(errStr)
				} else {
					return fmt.Errorf("%s. Cleanup failed as well: %v\n", errStr, err)
				}
			}
		}
	}

	err = setupPlatformService(ctx, cloudlet, pfConfig, vaultConfig, pfClient, updateCallback)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "failed to setup platform services", "err", err)
		if cloudlet.ImageUpgrade {
			// Delete currently created stacks
			pfStackName := mexos.GetStackName(&cloudlet.Key, cloudlet.ImageVersion, mexos.PlatformVMDeployment)
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting Heat Stack %s", pfStackName))
			err1 := mexos.HeatDeleteStack(ctx, pfStackName)
			if err1 != nil {
				return fmt.Errorf("Upgrade failed: %v and cloudlet deletion failed: %v", err, err1)
			}
			lbStackName := mexos.GetStackName(&cloudlet.Key, cloudlet.ImageVersion, mexos.RootLBVMDeployment)
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting Heat Stack %s", lbStackName))
			err1 = mexos.HeatDeleteStack(ctx, lbStackName)
			if err1 != nil {
				return fmt.Errorf("Upgrade failed: %v and cloudlet deletion failed: %v", err, err1)
			}
		} else {
			// Cleanup failed containers
			updateCallback(edgeproto.UpdateTask, "Upgrade failed, cleaning up")
			if out, err1 := pfClient.Output(
				fmt.Sprintf("sudo docker rm -f %s", strings.Join(PlatformServices, " ")),
			); err1 != nil {
				if strings.Contains(out, "No such container") {
					log.SpanLog(ctx, log.DebugLevelMexos, "no containers to cleanup")
				} else {
					return fmt.Errorf("upgrade failed: %v and cleanup failed: %v, %s\n", err, err1, out)
				}
			}
			// Cleanup container names
			for _, pfService := range PlatformServices {
				from := pfService + "_old"
				to := pfService
				log.SpanLog(ctx, log.DebugLevelMexos, "restoring old container name", "from", from, "to", to)
				if out, err1 := pfClient.Output(
					fmt.Sprintf("sudo docker rename %s %s", from, to),
				); err1 != nil {
					return fmt.Errorf("upgrade failed: %v and unable to rename old-container: %v, %s\n", err, err1, out)
				}
			}
		}
		return fmt.Errorf("Upgrade failed: %v", err)
	}

	if upgradeRequired && !cloudlet.ImageUpgrade {
		rootLBName := cloudcommon.GetRootLBFQDN(&cloudlet.Key)
		rlbClient, err := mexos.GetSSHClient(ctx, rootLBName, mexos.GetCloudletExternalNetwork(), mexos.SSHUser)
		if err != nil {
			return err
		}
		// Upgrade Cloudlet Packages
		upgradeMap := map[mexos.DeploymentType]ssh.Client{
			mexos.PlatformVMDeployment: pfClient,
			mexos.RootLBVMDeployment:   rlbClient,
		}
		for vmType, client := range upgradeMap {
			err = upgradeCloudletPkgs(ctx, vmType, cloudlet, pfConfig, vaultConfig, client, updateCallback)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelMexos, "Failed to upgrade cloudlet packages", "VM type", vmType, "err", err)
				return nil
			}
		}
	}
	return nil
}

func upgradeCloudletPkgs(ctx context.Context, vmType mexos.DeploymentType, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Upgrading cloudlet packages", "cloudletName", cloudlet.Key.Name, "to version", cloudlet.ImageVersion)

	pkgUrl := mexos.GetCloudletVMImagePkgPath(pfConfig.ImagePath, cloudlet.ImageVersion)
	if pkgUrl == "" {
		// No upgrade required
		log.SpanLog(ctx, log.DebugLevelMexos, "No upgrade of cloudlet packages required, as pkg path is missing")
		return nil
	}

	auth, err := cloudcommon.GetRegistryAuth(ctx, pkgUrl, vaultConfig)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelMexos, "warning, cannot get registry credentials from vault - assume public registry", "err", err)
		return nil
	}
	if auth == nil || auth.ApiKey == "" {
		return fmt.Errorf("Unable to find auth details for %s", pkgUrl)
	}
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Fetching cloudlet image package of version: %s", cloudlet.ImageVersion))
	if out, err := client.Output(
		fmt.Sprintf("curl -H 'X-JFrog-Art-Api:%s' -O %s", auth.ApiKey, pkgUrl),
	); err != nil {
		return fmt.Errorf("Failed to fetch mobiledgex pkg from %s, %v, %s", pkgUrl, err, out)
	}
	pkgName := mexos.GetCloudletVMImagePkgName(cloudlet.ImageVersion)
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Upgrading cloudlet image package to version: %s", cloudlet.ImageVersion))
	if out, err := client.Output(
		fmt.Sprintf("MEXVM_TYPE=%s dpkg -i %s", vmType, pkgName),
	); err != nil {
		return fmt.Errorf("Failed to upgrade mobiledgex pkg %s, %v, %s", pkgName, err, out)
	}
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Upgraded cloudlet image package to version %s successfully", cloudlet.ImageVersion))
	return nil
}

func (s *Platform) CleanupCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelMexos, "Cleaning up cloudlet", "cloudletName", cloudlet.Key.Name)

	// Cleanup old Stacks
	oldStackNames, err := getCurrentStackName(ctx, &cloudlet.Key, mexos.PlatformVMDeployment)
	if err != nil {
		return err
	}
	curStackName := mexos.GetStackName(&cloudlet.Key, cloudlet.ImageVersion, mexos.PlatformVMDeployment)
	for _, oldStack := range oldStackNames {
		if oldStack == curStackName {
			continue
		}
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting old Heat Stack %s", oldStack))
		err = mexos.HeatDeleteStack(ctx, oldStack)
		if err != nil {
			return fmt.Errorf("CleanupCloudlet error: %v", err)
		}
	}

	client, err := mexos.GetSSHClient(ctx, mexos.GetPlatformVMName(&cloudlet.Key), mexos.GetCloudletExternalNetwork(), mexos.SSHUser)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Removing old containers")
	for _, pfService := range PlatformServices {
		if out, err := client.Output(
			fmt.Sprintf("sudo docker rm -f %s_old", pfService),
		); err != nil {
			if strings.Contains(out, "No such container") {
				log.SpanLog(ctx, log.DebugLevelMexos, "no containers to cleanup")
				continue
			} else {
				return fmt.Errorf("cleanup failed: %v, %s\n", err, out)
			}
		}
	}

	return nil
}
