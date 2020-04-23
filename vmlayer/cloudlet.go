package vmlayer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	intprocess "github.com/mobiledgex/edge-cloud-infra/e2e-tests/int-process"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
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
)

var PlatformServices = []string{
	ServiceTypeCRM,
	ServiceTypeShepherd,
}

func (v *VMPlatform) GetPlatformVMName(key *edgeproto.CloudletKey) string {
	// Form platform VM name based on cloudletKey
	return v.vmProvider.NameSanitize(key.Name + "." + key.Organization + ".pf")
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
			log.SpanLog(ctx, log.DebugLevelInfra, "error trying to connect to controller port via ssh", "out", out, "error", err)
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

	err = v.vmProvider.VerifyApiEndpoint(ctx, client, updateCallback)

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
func (v *VMPlatform) SetupPlatformVM(ctx context.Context, cloudlet *edgeproto.Cloudlet, pfConfig *edgeproto.PlatformConfig, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) (ssh.Client, error) {
	// Get Closest Platform Flavor
	platformVmName := v.GetPlatformVMName(&cloudlet.Key)
	vmspec, err := vmspec.GetVMSpec(v.FlavorList, *pfFlavor)
	if err != nil {
		return nil, fmt.Errorf("unable to find VM spec for Shared RootLB: %v", err)
	}
	az := vmspec.AvailabilityZone
	if az == "" {
		az = v.GetCloudletComputeAvailabilityZone()
	}
	pfImageName, err := v.vmProvider.AddCloudletImageIfNotPresent(ctx, pfConfig.CloudletVmImagePath, cloudlet.VmImageVersion, updateCallback)
	if err != nil {
		return nil, err
	}
	vmreqspec, err := v.GetVMRequestSpec(ctx, VMTypePlatform, pfImageName, platformVmName, vmspec.FlavorName, true, WithExternalVolume(vmspec.ExternalVolumeSize))
	var vms []*VMRequestSpec
	vms = append(vms, vmreqspec)

	updateCallback(edgeproto.UpdateTask, "Deploying Platform VM")
	_, err = v.CreateVMsFromVMSpec(ctx, platformVmName, vms, updateCallback)

	updateCallback(edgeproto.UpdateTask, "Successfully Deployed Platform VM")
	ip, err := v.vmProvider.GetIPFromServerName(ctx, v.GetCloudletExternalNetwork(), platformVmName)
	if err != nil {
		return nil, err
	}
	updateCallback(edgeproto.UpdateTask, "Platform VM external IP: "+ip.ExternalAddr)

	client, err := v.GetSSHClientForServer(ctx, platformVmName, v.GetCloudletExternalNetwork())
	if err != nil {
		return nil, err
	}
	return client, nil
}
