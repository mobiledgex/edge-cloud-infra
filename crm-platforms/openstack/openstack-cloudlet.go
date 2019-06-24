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
	CRMBaseImg     = "mobiledgex-crm-v1.0"
	CRMMinVcpus    = 2
	CRMMinRam      = 8
	CRMMinDisk     = 40
	CRMNotifyPort  = "37001"
	MaxWait        = 2 * time.Minute
	CRMImgRegistry = "registry.mobiledgex.net:5000/mobiledgex/edge-cloud"
	CRMBaseImgPath = "https://artifactory.mobiledgex.net/artifactory/crm-baseimages/openstack/" + CRMBaseImg + ".qcow2"
	CRMBaseImgSum  = "a605d8d5385e74b28f60381acd4a9433"
)

func getCRMName(cloudlet *edgeproto.Cloudlet) string {
	// Form CRM VM name based on cloudletKey
	return cloudlet.Key.Name + "-crm." + cloudlet.Key.OperatorKey.Name
}

func (s *Platform) CreateCloudlet(cloudlet *edgeproto.Cloudlet, updateCallback edgeproto.CacheUpdateCallback) error {
	var err error

	// Fetch CRM Image Tag
	crm_version, err = ioutil.ReadFile("/version.txt")
	if err != nil {
		return fmt.Errorf("unable to fetch crm image tag details: %v", err)
	}

	crm_registry_path := CRMImgRegistry + ":" + crm_version

	cloudlet.State = edgeproto.TrackedState_CREATING
	log.DebugLog(log.DebugLevelMexos, "Creating cloudlet", "cloudletName", cloudlet.Key.Name)

	// Soure OpenRC file to access openstack API endpoint
	updateCallback(edgeproto.UpdateTask, "Fetch and Source OpenRC file")
	err = mexos.InitOpenstackProps(cloudlet.Key.OperatorKey.Name, cloudlet.PhysicalName, cloudlet.VaultAddr)
	if err != nil {
		return err
	}

	// Use CRMBaseImage, if not present then fetch it from MobiledgeX VM registry
	imageDetail, err := mexos.GetImageDetail(CRMBaseImg)
	if err == nil && imageDetail.Status != "active" {
		return fmt.Errorf("image %s is not active", CRMBaseImg)
	}
	if err != nil {
		// Download CRM base image and Add to Openstack Glance
		updateCallback(edgeproto.UpdateTask, "Downloading CRM base image: ")
		err = mexos.CreateImageFromUrl(CRMBaseImg, CRMBaseImgPath, CRMBaseImgSum)
		if err != nil {
			return fmt.Errorf("Error downloading CRM base image: %v", err)
		}
	}

	// Fetch CRM Flavor
	crm_flavor := edgeproto.Flavor{
		Key: edgeproto.FlavorKey{
			Name: "crm-flavor",
		},
		Ram:   uint64(CRMMinRam),
		Vcpus: uint64(CRMMinVcpus),
		Disk:  uint64(CRMMinDisk),
	}
	finfo, err := mexos.GetFlavorInfo()
	if err != nil {
		return err
	}
	crm_flavor_name, err := flavor.GetClosestFlavor(finfo, crm_flavor)
	if err != nil {
		return fmt.Errorf("unable to find closest flavor for crm: %v", err)
	}

	// Form CRM VM name based on cloudletKey
	crm_name := getCRMName(cloudlet)

	// Generate SSH KeyPair
	keyPairPath := "/tmp/" + crm_name
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
 - echo MOBILEDGEX CRM CLOUD CONFIG START
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
		crm_name,
		crm_flavor_name,
		CRMBaseImg,
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
	regAuth := cloudcommon.GetRegistryAuth(crm_registry_path, cloudlet.VaultAddr)
	if regAuth == nil {
		return fmt.Errorf("unable to fetch registry auth credentials")
	}
	if regAuth.AuthType != cloudcommon.BasicAuth {
		return fmt.Errorf("unsupported registry auth type %s", regAuth.AuthType)
	}

	// Deploy CRM VM
	updateCallback(edgeproto.UpdateTask, "Deploying CRM VM")
	log.DebugLog(log.DebugLevelMexos, "Deploying VM", "stackName", crm_name, "flavor", crm_flavor_name)
	err = mexos.CreateHeatStackFromTemplate(vmp, crm_name, mexos.VmTemplate, updateCallback)
	if err != nil {
		return fmt.Errorf("CreateVMAppInst error: %v", err)
	}
	updateCallback(edgeproto.UpdateTask, "Successfully Deployed CRM VM")

	// Fetch external IP for further configuration
	external_ip, err := mexos.GetServerIPAddr(mexos.GetCloudletExternalNetwork(), crm_name)
	if err != nil {
		return err
	}
	updateCallback(edgeproto.UpdateTask, "External IP: "+external_ip)
	log.DebugLog(log.DebugLevelMexos, "external IP", "ip", external_ip)
	// TODO: Activate FDQN
	/*
		loc := util.DNSSanitize(cloudlet.Key.Name)
		oper := util.DNSSanitize(cloudletkey.OperatorKey.Name)
		fqdn := fmt.Sprintf("%s.%s.%s", loc, oper, cloudcommon.AppDNSRoot)
		if external_ip != "" {
			if err = mexos.ActivateFQDNA(fqdn, external_ip); err != nil {
				return err
			}
			log.DebugLog(log.DebugLevelMexos, "DNS A record activated",
				"fqdn", fqdn, "IP", external_ip)
		}
	*/

	// Setup SSH Client
	auth := ssh.Auth{Keys: []string{keyPairPath}}
	client, err := ssh.NewNativeClient("ubuntu", external_ip, "SSH-2.0-mobiledgex-ssh-client-1.0", 22, &auth, nil)
	if err != nil {
		return fmt.Errorf("cannot get ssh client for server %s with ip %s, %v", crm_name, external_ip, err)
	}

	// Verify if controller's notify port is reachable
	updateCallback(edgeproto.UpdateTask, "Verifying if controller is reachable")
	if out, err := client.Output(
		fmt.Sprintf(
			"nc %s %s -w 5", cloudlet.ControllerAddr, CRMNotifyPort,
		),
	); err != nil {
		return fmt.Errorf("controller's notify port is unreachable: %v, %s\n", err, out)
	}

	// Login to docker registry
	updateCallback(edgeproto.UpdateTask, "Setup docker registry")
	if out, err := client.Output(
		fmt.Sprintf(
			`echo "%s" | sudo docker login -u %s --password-stdin %s`,
			regAuth.Password,
			regAuth.Username,
			crm_registry_path,
		),
	); err != nil {
		return fmt.Errorf("unable to login to docker registry: %v, %s\n", err, out)
	}

	// Pull docker image and start crmserver
	updateCallback(edgeproto.UpdateTask, "Start CRMServer")
	if out, err := client.Output(
		`sudo docker run -d ` +
			`--name ` + crm_name +
			` -e VAULT_SECRET_ID="6e88cd75-7297-ba5e-6b27-9d612e3792b7"` + // TODO: How to fetch CRM specific Vault Role/Secret ID
			` -e VAULT_ROLE_ID="e017fc39-dff7-adc3-364f-bb8e04805454" ` +
			crm_registry_path +
			` crmserver ` +
			`"--notifyAddrs" ` + `"` + cloudlet.ControllerAddr + `:` + CRMNotifyPort + `" ` +
			`"--apiAddr" "0.0.0.0:55101" ` +
			`"--cloudletKey" "{\"operator_key\":{\"name\":\"` + cloudlet.Key.OperatorKey.Name + `\"},\"name\":\"` + cloudlet.Key.Name + `\"}" ` +
			`"--tls" "/root/tls/mex-server.crt" ` +
			`"-d" "api,notify,mexos" ` +
			`"--platform" "openstack" ` +
			`"-vaultAddr" "` + cloudlet.VaultAddr + `" ` +
			`"--physicalName" "` + cloudlet.PhysicalName + `"`,
	); err != nil {
		return fmt.Errorf("unable to start crmserver: %v, %s\n", err, out)
	}

	// Wait for CRM to come up: Since we cannot fetch information
	// regarding CRM state from docker ouput, we observe container
	// state for MaxWait time, if it is Up then we return
	// successfully and then controller will do the actual
	// check if CRM connected to it or not
	start := time.Now()
	for {
		out, err := client.Output(`sudo docker ps -a -n 1 --filter name=` + crm_name + ` --format '{{.Status}}'`)
		if err != nil {
			return fmt.Errorf("unable to fetch crm container status: %v, %s\n", err, out)
		}
		if !strings.Contains(out, "Up ") {
			// container exited in failure state
			out, err = client.Output(`sudo docker logs ` + crm_name + ` 2>&1 | grep FATAL | awk '{for (i=1; i<=NF-3; i++) $i = $(i+3); NF-=3; print}'`)
			if err != nil {
				return fmt.Errorf("failed to brinup crmserver")
			}
			return fmt.Errorf("failed to brinup crmserver: %s", out)
		}
		elapsed := time.Since(start)
		if elapsed >= (MaxWait) {
			// no issues in wait time
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (s *Platform) DeleteCloudlet(cloudlet *edgeproto.Cloudlet) error {
	log.DebugLog(log.DebugLevelMexos, "Deleting cloudlet", "cloudletName", cloudlet.Key.Name)
	crm_name := getCRMName(cloudlet)
	err := mexos.HeatDeleteStack(crm_name)
	if err != nil {
		return fmt.Errorf("DeleteCloudlet error: %v", err)
	}
	return nil
}
