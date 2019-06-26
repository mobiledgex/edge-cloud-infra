package openstack

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/flavor"
	"github.com/mobiledgex/edge-cloud/log"
	ssh "github.com/mobiledgex/golang-ssh"
)

const (
	PlatformMaxWait = 5 * time.Minute
	/*
		PlatformBaseImgPath  = "https://artifactory.mobiledgex.net/artifactory/crm-baseimages/openstack/" + PlatformBaseImg + ".qcow2"
		PlatformBaseImg      = "mobiledgex-crm-v1.0"
		PlatformBaseImgSum   = "a605d8d5385e74b28f60381acd4a9433"
	*/
)

func getPlatformVMName(cloudlet *edgeproto.Cloudlet) string {
	// Form platform VM name based on cloudletKey
	return cloudlet.Key.Name + cloudlet.Key.OperatorKey.Name
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, pf *edgeproto.Platform, pfFlavor *edgeproto.Flavor, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error

	cloudlet.State = edgeproto.TrackedState_CREATING
	log.DebugLog(log.DebugLevelMexos, "Creating cloudlet", "cloudletName", cloudlet.Key.Name)

	// Soure OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, "Fetch and Source OpenRC file")
	err = mexos.InitOpenstackProps(cloudlet.Key.OperatorKey.Name, cloudlet.PhysicalName, cloudlet.VaultAddr)
	if err != nil {
		return err
	}

	// Fetch platform base image name and md5sum
	pfImageName, err := cloudcommon.GetFileName(pf.ImagePath)
	if err != nil {
		return err
	}
	_, md5Sum, err := mexos.GetUrlInfo(pf.ImagePath)
	if err != nil {
		return err
	}

	// Use PlatformBaseImage, if not present then fetch it from MobiledgeX VM registry
	imageDetail, err := mexos.GetImageDetail(pfImageName)
	if err == nil && imageDetail.Status != "active" {
		return fmt.Errorf("image %s is not active", pfImageName)
	}
	if err != nil {
		// Download platform base image and Add to Openstack Glance
		updateCallback(edgeproto.UpdateTask, "Downloading platform base image: ")
		err = mexos.CreateImageFromUrl(pfImageName, pf.ImagePath, md5Sum)
		if err != nil {
			return fmt.Errorf("Error downloading platform base image: %v", err)
		}
	}

	// Get Closest Platform Flavor
	finfo, err := mexos.GetFlavorInfo()
	if err != nil {
		return err
	}
	platform_flavor_name, err := flavor.GetClosestFlavor(finfo, *pfFlavor)
	if err != nil {
		return fmt.Errorf("unable to find closest flavor for platform: %v", err)
	}

	// Form platform VM name based on cloudletKey
	platform_vm_name := getPlatformVMName(cloudlet)

	// Generate SSH KeyPair
	keyPairPath := "/tmp/" + platform_vm_name
	err = mexos.GenerateSSHKeyPair(keyPairPath)
	if err != nil {
		return err
	}
	pubKeyBytes, err := ioutil.ReadFile(keyPairPath + ".pub")
	if err != nil {
		return err
	}

	// Use cloud-config to configure SSH access
	cloud_config := `#cloud-config
bootcmd:
 - echo MOBILEDGEX PLATFORM CLOUD CONFIG START
 - echo 'APT::Periodic::Enable "0";' > /etc/apt/apt.conf.d/10cloudinit-disable
hostname: ` + cloudlet.Key.Name + `
chpasswd: { expire: False }
ssh_pwauth: False
timezone: UTC
ssh_authorized_keys:
 - ` + string(pubKeyBytes)

	// TODO Upload SSHKeyPair to Vault so that all the VMs in the cloudlet use this KeyPair

	vmp, err := mexos.GetVMParams(
		mexos.UserVMDeployment,
		platform_vm_name,
		platform_flavor_name,
		pfImageName,
		"",           // AuthPublicKey
		"tcp:22",     // AccessPorts
		cloud_config, // DeploymentManifest,
		"",           // Command,
		nil,          // NetSpecInfo
	)
	if err != nil {
		return fmt.Errorf("unable to get vm params: %v", err)
	}

	// Gather registry credentails from Vault
	updateCallback(edgeproto.UpdateTask, "Fetch registry auth credentials")
	regAuth := cloudcommon.GetRegistryAuth(pf.RegistryPath, cloudlet.VaultAddr)
	if regAuth == nil {
		return fmt.Errorf("unable to fetch registry auth credentials")
	}
	if regAuth.AuthType != cloudcommon.BasicAuth {
		return fmt.Errorf("unsupported registry auth type %s", regAuth.AuthType)
	}

	// Deploy Platform VM
	updateCallback(edgeproto.UpdateTask, "Deploying Platform VM")
	log.DebugLog(log.DebugLevelMexos, "Deploying VM", "stackName", platform_vm_name, "flavor", platform_flavor_name)
	err = mexos.CreateHeatStackFromTemplate(vmp, platform_vm_name, mexos.VmTemplate, updateCallback)
	if err != nil {
		return fmt.Errorf("CreateVMAppInst error: %v", err)
	}
	updateCallback(edgeproto.UpdateTask, "Successfully Deployed Platform VM")

	// Fetch external IP for further configuration
	external_ip, err := mexos.GetServerIPAddr(mexos.GetCloudletExternalNetwork(), platform_vm_name)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "External IP: "+external_ip)
	log.DebugLog(log.DebugLevelMexos, "external IP", "ip", external_ip)

	// Setup SSH Client
	auth := ssh.Auth{Keys: []string{keyPairPath}}
	client, err := ssh.NewNativeClient("ubuntu", external_ip, "SSH-2.0-mobiledgex-ssh-client-1.0", 22, &auth, nil)
	if err != nil {
		return fmt.Errorf("cannot get ssh client for server %s with ip %s, %v", platform_vm_name, external_ip, err)
	}

	// Verify if controller's notify port is reachable
	updateCallback(edgeproto.UpdateTask, "Verifying if controller notification channel is reachable")
	for _, ctrlAddrPort := range strings.Split(cloudlet.NotifyCtrlAddrs, ",") {
		addrPort := strings.Split(ctrlAddrPort, ":")
		if len(addrPort) != 2 {
			return fmt.Errorf("notifyctrladdrs format is incorrect")
		}
		if out, err := client.Output(
			fmt.Sprintf(
				"nc %s %s -w 5", addrPort[0], addrPort[2],
			),
		); err != nil {
			return fmt.Errorf("controller's notify port is unreachable: %v, %s\n", err, out)
		}
	}

	// Login to docker registry
	updateCallback(edgeproto.UpdateTask, "Setup docker registry")
	if out, err := client.Output(
		fmt.Sprintf(
			`echo "%s" | sudo docker login -u %s --password-stdin %s`,
			regAuth.Password,
			regAuth.Username,
			pf.RegistryPath,
		),
	); err != nil {
		return fmt.Errorf("unable to login to docker registry: %v, %s\n", err, out)
	}

	// Pull docker image and start crmserver
	updateCallback(edgeproto.UpdateTask, "Start CRMServer")
	if out, err := client.Output(
		`sudo docker run -d --restart=unless-stopped ` +
			`--name ` + platform_vm_name +
			` -e VAULT_SECRET_ID="` + cloudlet.CrmSecretId + `"` +
			` -e VAULT_ROLE_ID="` + cloudlet.CrmRoleId + `" ` +
			pf.RegistryPath +
			` crmserver ` +
			`"--notifyAddrs" ` + `"` + cloudlet.NotifyCtrlAddrs + `" ` +
			`"--cloudletKey" "{\"operator_key\":{\"name\":\"` + cloudlet.Key.OperatorKey.Name + `\"},\"name\":\"` + cloudlet.Key.Name + `\"}" ` +
			`"--tls" "` + cloudlet.TlsCertFile + `" ` +
			`"-d" "api,notify,mexos" ` +
			`"--platform" "openstack" ` +
			`"-vaultAddr" "` + cloudlet.VaultAddr + `" ` +
			`"--physicalName" "` + cloudlet.PhysicalName + `"`,
	); err != nil {
		return fmt.Errorf("unable to start crmserver: %v, %s\n", err, out)
	}

	// Wait for CRM to come up: Since we cannot fetch information
	// regarding CRM state from docker ouput, we observe container
	// state for PlatformMaxWait time, if it is Up then we return
	// successfully and then controller will do the actual
	// check if CRM connected to it or not
	start := time.Now()
	for {
		out, err := client.Output(`sudo docker ps -a -n 1 --filter name=` + platform_vm_name + ` --format '{{.Status}}'`)
		if err != nil {
			return fmt.Errorf("unable to fetch crm container status: %v, %s\n", err, out)
		}
		if !strings.Contains(out, "Up ") {
			// container exited in failure state
			out, err = client.Output(`sudo docker logs ` + platform_vm_name + ` 2>&1 | grep FATAL | awk '{for (i=1; i<=NF-3; i++) $i = $(i+3); NF-=3; print}'`)
			if err != nil {
				return fmt.Errorf("failed to brinup crmserver")
			}
			return fmt.Errorf("failed to brinup crmserver: %s", out)
		}
		elapsed := time.Since(start)
		if elapsed >= (PlatformMaxWait) {
			// no issues in wait time
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "Deleting cloudlet", "cloudletName", cloudlet.Key.Name)
	platform_vm_name := getPlatformVMName(cloudlet)
	err := mexos.HeatDeleteStack(platform_vm_name)
	if err != nil {
		return fmt.Errorf("DeleteCloudlet error: %v", err)
	}
	return nil
}
