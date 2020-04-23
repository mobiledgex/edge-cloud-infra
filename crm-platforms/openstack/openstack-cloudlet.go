package openstack

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	ssh "github.com/mobiledgex/golang-ssh"
)

// TODO: there is still some code here that can eventually go to vmlayer

func (o *OpenstackPlatform) VerifyApiEndpoint(ctx context.Context, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error {
	// Verify if Openstack API Endpoint is reachable
	updateCallback(edgeproto.UpdateTask, "Verifying if Openstack API Endpoint is reachable")
	osAuthUrl := o.openRCVars["OS_AUTH_URL"]
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
		gatewayAddr, err := o.GetExternalGateway(ctx, o.vmPlatform.GetCloudletExternalNetwork())
		if err != nil {
			return fmt.Errorf("unable to fetch gateway IP for external network: %s, %v",
				o.vmPlatform.GetCloudletExternalNetwork(), err)
		}
		// Add route to reach API endpoint
		if out, err := client.Output(
			fmt.Sprintf(
				"sudo route add -host %s gw %s", urlObj.Hostname(), gatewayAddr,
			),
		); err != nil {
			return fmt.Errorf("unable to add route to reach API endpoint: %v, %s\n", err, out)
		}
		interfacesFile := vmlayer.GetCloudletNetworkIfaceFile()
		routeAddLine := fmt.Sprintf("up route add -host %s gw %s", urlObj.Hostname(), gatewayAddr)
		cmd := fmt.Sprintf("grep -l '%s' %s", routeAddLine, interfacesFile)
		_, err = client.Output(cmd)
		if err != nil {
			// grep failed so not there already
			log.SpanLog(ctx, log.DebugLevelInfra, "adding route to interfaces file", "route", routeAddLine, "file", interfacesFile)
			cmd = fmt.Sprintf("echo '%s'|sudo tee -a %s", routeAddLine, interfacesFile)
			out, err := client.Output(cmd)
			if err != nil {
				return fmt.Errorf("can't add route '%s' to interfaces file: %v, %s", routeAddLine, err, out)
			}
		} else {
			log.SpanLog(ctx, log.DebugLevelInfra, "route already present in interfaces file")
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
	return nil
}

func (o *OpenstackPlatform) DeleteCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Deleting cloudlet", "cloudletName", cloudlet.Key.Name)

	updateCallback(edgeproto.UpdateTask, "Deleting cloudlet")

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}

	// Source OpenRC file to access openstack API endpoint
	err = o.InitOpenstackProps(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig, cloudlet.EnvVar)
	if err != nil {
		// ignore this error, as no creation would've happened on infra, so nothing to delete
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to source platform variables", "cloudletName", cloudlet.Key.Name, "err", err)
		return nil
	}

	platformVMName := o.vmPlatform.GetPlatformVMName(&cloudlet.Key)
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting PlatformVM %s", platformVMName))
	err = o.HeatDeleteStack(ctx, platformVMName)
	if err != nil {
		return fmt.Errorf("DeleteCloudlet error: %v", err)
	}

	rootLBName := o.vmPlatform.GetRootLBName(&cloudlet.Key)
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Deleting RootLB %s", rootLBName))
	err = o.HeatDeleteStack(ctx, rootLBName)
	if err != nil {
		return fmt.Errorf("DeleteCloudlet error: %v", err)
	}

	// Not sure if it's safe to remove vars from Vault due to testing/virtual cloudlets,
	// so leaving them in Vault for the time being. We can always delete them manually

	return nil
}

func handleUpgradeError(ctx context.Context, client ssh.Client) error {
	for _, pfService := range vmlayer.PlatformServices {
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

func getCRMContainerVersion(ctx context.Context, client ssh.Client) (string, error) {
	var err error
	var out string

	log.SpanLog(ctx, log.DebugLevelInfra, "fetch crmserver container version")
	if out, err = client.Output(
		fmt.Sprintf("sudo docker ps --filter name=%s --format '{{.Image}}'", vmlayer.ServiceTypeCRM),
	); err != nil {
		return "", fmt.Errorf("unable to fetch crm version for %s, %v, %v",
			vmlayer.ServiceTypeCRM, err, out)
	}
	if out == "" {
		return "", fmt.Errorf("no container with name %s exists", vmlayer.ServiceTypeCRM)
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

func upgradeCloudletPkgs(ctx context.Context, vmType vmlayer.VMType, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, vaultConfig *vault.Config, client ssh.Client, updateCallback edgeproto.CacheUpdateCallback) error {
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

func (o *OpenstackPlatform) UpdateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) (edgeproto.CloudletAction, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "Updating cloudlet", "cloudletName", cloudlet.Key.Name)

	defCloudletAction := edgeproto.CloudletAction_ACTION_NONE

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return defCloudletAction, err
	}
	// Source OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Sourcing platform variables for %s cloudlet", cloudlet.PhysicalName))
	err = o.InitOpenstackProps(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig, cloudlet.EnvVar)
	if err != nil {
		return defCloudletAction, err
	}

	pfClient, err := o.vmPlatform.GetSSHClientForServer(ctx, o.vmPlatform.GetPlatformVMName(&cloudlet.Key), o.vmPlatform.GetCloudletExternalNetwork())
	if err != nil {
		return defCloudletAction, err
	}

	containerVersion, err := getCRMContainerVersion(ctx, pfClient)
	if err != nil {
		return defCloudletAction, err
	}

	rootLBName := cloudcommon.GetRootLBFQDN(&cloudlet.Key)
	rlbClient, err := o.vmPlatform.GetSSHClientForServer(ctx, rootLBName, o.vmPlatform.GetCloudletExternalNetwork())
	if err != nil {
		return defCloudletAction, err
	}
	upgradeMap := map[vmlayer.VMType]ssh.Client{
		vmlayer.VMTypePlatform: pfClient,
		vmlayer.VMTypeRootLB:   rlbClient,
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
	for _, pfService := range vmlayer.PlatformServices {
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

	err = o.vmPlatform.SetupPlatformService(ctx, cloudlet, pfConfig, vaultConfig, pfClient, updateCallback)

	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "failed to setup platform services", "err", err)
		// Cleanup failed containers
		updateCallback(edgeproto.UpdateTask, "Upgrade failed, cleaning up")
		if out, err1 := pfClient.Output(
			fmt.Sprintf("sudo docker rm -f %s", strings.Join(vmlayer.PlatformServices, " ")),
		); err1 != nil {
			if strings.Contains(out, "No such container") {
				log.SpanLog(ctx, log.DebugLevelInfra, "no containers to cleanup")
			} else {
				return defCloudletAction, fmt.Errorf("upgrade failed: %v and cleanup failed: %v, %s\n", err, err1, out)
			}
		}
		// Cleanup container names
		for _, pfService := range vmlayer.PlatformServices {
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

func (o *OpenstackPlatform) CleanupCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Cleaning up cloudlet", "cloudletName", cloudlet.Key.Name)

	client, err := o.vmPlatform.GetSSHClientForServer(ctx, o.vmPlatform.GetPlatformVMName(&cloudlet.Key), o.vmPlatform.GetCloudletExternalNetwork())
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "Removing old containers")
	for _, pfService := range vmlayer.PlatformServices {
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

func (o *OpenstackPlatform) CreateCloudlet(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error

	log.SpanLog(ctx, log.DebugLevelInfra, "Creating cloudlet", "cloudletName", cloudlet.Key.Name)

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	// Source OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, "Sourcing access variables")
	err = o.InitOpenstackProps(ctx, &cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName, vaultConfig, cloudlet.EnvVar)
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

	client, err := o.vmPlatform.SetupPlatformVM(ctx, cloudlet, pfConfig, pfFlavor, updateCallback)
	if err != nil {
		return err
	}

	return o.vmPlatform.SetupPlatformService(ctx, cloudlet, pfConfig, vaultConfig, client, updateCallback)
}

func (o *OpenstackPlatform) SaveCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, accessVarsIn map[string]string, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Saving cloudlet access vars to vault", "cloudletName", cloudlet.Key.Name)
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
		certFile := vmlayer.GetCertFilePath(&cloudlet.Key)
		err := ioutil.WriteFile(certFile, []byte(certData), 0644)
		if err != nil {
			return err
		}
		accessVars["OS_CACERT"] = certFile
		accessVars["OS_CACERT_DATA"] = certData
	}
	updateCallback(edgeproto.UpdateTask, "Saving access vars to secure secrets storage (Vault)")
	var varList infracommon.VaultEnvData
	for key, value := range accessVars {
		if key == "OS_CACERT" {
			continue
		}
		varList.Env = append(varList.Env, infracommon.EnvData{
			Name:  key,
			Value: value,
		})
	}
	data := map[string]interface{}{
		"data": varList,
	}

	path := o.vmPlatform.GetVaultCloudletAccessPath(&cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName)
	err = infracommon.PutDataToVault(vaultConfig, path, data)
	if err != nil {
		updateCallback(edgeproto.UpdateTask, "Failed to save access vars to vault")
		log.SpanLog(ctx, log.DebugLevelInfra, err.Error(), "cloudletName", cloudlet.Key.Name)
		return fmt.Errorf("Failed to save access vars to vault: %v", err)
	}
	return nil
}

func (o *OpenstackPlatform) DeleteCloudletAccessVars(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Deleting access vars from vault", "cloudletName", cloudlet.Key.Name)

	updateCallback(edgeproto.UpdateTask, "Deleting access vars from secure secrets storage")

	vaultConfig, err := vault.BestConfig(pfConfig.VaultAddr, vault.WithEnvMap(pfConfig.EnvVar))
	if err != nil {
		return err
	}
	path := o.vmPlatform.GetVaultCloudletAccessPath(&cloudlet.Key, pfConfig.Region, cloudlet.PhysicalName)
	err = infracommon.DeleteDataFromVault(vaultConfig, path)
	if err != nil {
		return fmt.Errorf("Failed to delete access vars from vault: %v", err)
	}
	return nil
}
