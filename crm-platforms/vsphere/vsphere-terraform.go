package vsphere

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer/terraform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var NumRetries = 2

var VmipTag = "vmip"

type VSphereGeneralParams struct {
	VsphereUser       string
	VspherePassword   string
	VsphereServer     string
	DataCenterName    string
	ResourcePool      string
	ComputeCluster    string
	ExternalNetwork   string
	ExternalNetworkId string
	InternalDVS       string
	DataStore         string
}

type VSphereVMGroupParams struct {
	*VSphereGeneralParams
	*vmlayer.VMGroupOrchestrationParams
}

func (v *VSpherePlatform) populateGeneralParams(ctx context.Context, planName string, vgp *VSphereGeneralParams) error {
	vcaddr, _, err := v.GetVCenterAddress()
	if err != nil {
		return err
	}
	vgp.VsphereUser = v.GetVCenterUser()
	vgp.VspherePassword = v.GetVCenterPassword()
	vgp.VsphereServer = vcaddr
	vgp.DataCenterName = v.GetDatacenterName(ctx)
	vgp.ExternalNetwork = v.vmProperties.GetCloudletExternalNetwork()
	vgp.ExternalNetworkId = v.IdSanitize(vgp.ExternalNetwork)
	vgp.ComputeCluster = v.GetComputeCluster()
	vgp.DataStore = v.GetDataStore()
	vgp.InternalDVS = v.GetInternalVSwitch()
	vgp.ResourcePool = v.IdSanitize(getResourcePool(planName))
	return nil
}

func (v *VSpherePlatform) populateVMOrchParams(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "populateVMOrchParams")

	masterIP := ""
	flavors, err := v.GetFlavorList(ctx)
	if err != nil {
		return nil
	}

	// populate vm fields
	for i, vm := range vmgp.VMs {
		var vmtags []string
		vmgp.VMs[i].MetaData = vmlayer.GetVMMetaData(vm.Role, masterIP, vmsphereMetaDataFormatter)
		vmgp.VMs[i].UserData = vmlayer.GetVMUserData(vm.SharedVolume, vm.DNSServers, vm.DeploymentManifest, vm.Command, vmsphereUserDataFormatter)
		vmgp.VMs[i].DNSServers = "\"1.1.1.1\", \"1.0.0.1\""
		for _, f := range flavors {
			if f.Name == vm.FlavorName {
				vmgp.VMs[i].Vcpus = f.Vcpus
				vmgp.VMs[i].Disk = f.Disk
				vmgp.VMs[i].Ram = f.Ram
			}
		}
		for _, port := range vm.Ports {
			log.SpanLog(ctx, log.DebugLevelInfra, "updating VM port", "port", port)
			if port.Id == v.IdSanitize(v.vmProperties.GetCloudletExternalNetwork()) {
				eip, err := v.GetFreeExternalIP(ctx)
				if err != nil {
					return err
				}
				fip := vmlayer.FixedIPOrchestrationParams{
					Subnet:  vmlayer.NewResourceReference(port.Name, port.Id, false),
					Mask:    v.GetExternalNetmask(),
					Address: eip,
				}
				vmgp.VMs[i].FixedIPs = append(vmgp.VMs[i].FixedIPs, fip)
				tagname := vm.Name + "__" + port.Id + "__" + eip
				tagid := v.IdSanitize(tagname)
				vmtags = append(vmtags, tagname)
				vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: VmipTag, Id: tagid, Name: tagname})
				vmgp.VMs[i].ExternalGateway = v.GetExternalGateway()
			}
		}
		vmgp.VMs[i].Tags = strings.Join(vmtags, ",")
	}

	return nil
}

func getResourcePool(planName string) string {
	return planName + "-pool"
}

var vcenterTemplate = `
	provider "vsphere" {
		user           = "{{.VsphereUser}}"
		password      = "{{.VspherePassword}}"
		vsphere_server = "{{.VsphereServer}}"
		# If you have a self-signed cert
		allow_unverified_ssl = true
	}
  
  	data "vsphere_datacenter" "dc" {
		name = "{{.DataCenterName}}"
	}

	data "vsphere_compute_cluster" "{{.ComputeCluster}}" {
		name          = "{{.ComputeCluster}}"
		datacenter_id = "${data.vsphere_datacenter.dc.id}"
	}

	data "vsphere_datastore" "datastore" {
		name          = "{{.DataStore}}"
		datacenter_id = data.vsphere_datacenter.dc.id
	}

	data "vsphere_network" "{{.ExternalNetworkId}}" {
		name          = "{{.ExternalNetwork}}"
		datacenter_id = data.vsphere_datacenter.dc.id
	}

	data "vsphere_distributed_virtual_switch" "{{.InternalDVS}}" {
		name          = "{{.InternalDVS}}"
		datacenter_id = "${data.vsphere_datacenter.dc.id}"
	}

	resource "vsphere_tag_category" "vmip" {
		name        = "vmip"
		cardinality = "SINGLE"
		description = "VM IP Addresses"
	  
		associable_types = [
		  "VirtualMachine",
		]
	}
	`

var vmGroupTemplate = `  
	resource "vsphere_resource_pool" "{{.ResourcePool}}" {
		name          = "{{.ResourcePool}}"
		parent_resource_pool_id = "${data.vsphere_compute_cluster.{{.ComputeCluster}}.resource_pool_id}"
	}

	{{- range .Subnets}}
	resource "vsphere_distributed_port_group" "{{.Id}}" {
		name                            = "{{.Name}}"
		distributed_virtual_switch_uuid = "${data.vsphere_distributed_virtual_switch.{{$.InternalDVS}}.id}"
		vlan_id                         = {{.Vlan}}
	}
	{{- end}}

	{{- range .Tags}}
	resource "vsphere_tag" "{{.Id}}" {
		name = "{{.Name}}"
		category_id = "${vsphere_tag_category.{{.Category}}.id}"
	}
	{{- end}}

	{{- range .VMs}}
	data "vsphere_virtual_machine" "{{.ImageName}}-tmplt-{{.Id}}" {
		name          = "{{.ImageName}}"
		datacenter_id = "${data.vsphere_datacenter.dc.id}"
	}

	resource "vsphere_virtual_machine" "{{.Id}}" {
		name             = "{{.Name}}"
		resource_pool_id = vsphere_resource_pool.{{$.ResourcePool}}.id
		datastore_id     = data.vsphere_datastore.datastore.id
		wait_for_guest_net_timeout = -1
		num_cpus = {{.Vcpus}}
		memory   = {{.Ram}}
		guest_id = "ubuntu64Guest"

  		{{- range .Ports}}
		network_interface {
			{{- if .SubnetId}}
			network_id = data.vsphere_distributed_port_group.{{.SubnetId}}.id
			{{- else}}
			network_id = data.vsphere_network.{{.NetworkId}}.id
			{{- end}}
  		}
		{{- end}}

  		disk {
			label = "disk0"
			size = {{.Disk}}
			thin_provisioned = "${data.vsphere_virtual_machine.{{.ImageName}}-tmplt-{{.Id}}.disks.0.thin_provisioned}"
			eagerly_scrub = "${data.vsphere_virtual_machine.{{.ImageName}}-tmplt-{{.Id}}.disks.0.eagerly_scrub}"
  		}

		extra_config = {
			"guestinfo.userdata" = "{{.UserData}}"
			"guestinfo.userdata.encoding" = "base64"
			"guestinfo.metadata" = "{{.MetaData}}"
			"guestinfo.metadata.encoding" = "base64"
		}

		clone {
			template_uuid = "${data.vsphere_virtual_machine.{{.ImageName}}-tmplt-{{.Id}}.id}"
			customize{
				linux_options {
					host_name = "{{.HostName}}"
					domain = "mobiledgex.net"
				}
				timeout = 2
				{{- range .FixedIPs}}
				network_interface {
					ipv4_address = "{{.Address}}"
					ipv4_netmask = "{{.Mask}}"
				}
				{{- end}}
	  			ipv4_gateway = "{{.ExternalGateway}}"
	  			dns_server_list = [{{.DNSServers}}]
			}
		}
		{{- if .Tags}}
		tags = ["{{.Tags}}"]
		{{- end}}
	}
	{{- end}}
`

// user data is encoded as base64
func vmsphereUserDataFormatter(instring string) string {
	return base64.StdEncoding.EncodeToString([]byte(instring))
}

// meta data needs to have an extra layer "meta" for vsphere
func vmsphereMetaDataFormatter(instring string) string {
	indented := ""
	for _, v := range strings.Split(instring, "\n") {
		indented += strings.Repeat(" ", 4) + v + "\n"
	}
	withMeta := fmt.Sprintf("meta:\n%s", indented)
	return base64.StdEncoding.EncodeToString([]byte(withMeta))
}

func (v *VSpherePlatform) ImportVSphereData(ctx context.Context) error {
	return fmt.Errorf("importVSphereData not implemented")
}

func (v *VSpherePlatform) TerraformSetupVsphere(ctx context.Context, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "TerraformSetupVsphere")

	planName := v.NameSanitize(v.GetDatacenterName(ctx))
	err := os.Mkdir(terraform.TerraformDir, 0755)
	if err != nil {
		if strings.Contains(err.Error(), "file exists") {
			log.SpanLog(ctx, log.DebugLevelInfra, "terraform dir already exists")
		} else {
			return fmt.Errorf("unable to create terraformDir: %s - %v", terraform.TerraformDir, err)
		}
	}

	var vgp VSphereGeneralParams
	err = v.populateGeneralParams(ctx, planName, &vgp)
	if err != nil {
		return err
	}

	return terraform.CreateTerraformPlanFromTemplate(
		ctx,
		vgp,
		planName,
		vcenterTemplate, updateCallback,
		terraform.WithInit(true),
		terraform.WithCleanupOnFailure(v.vmProperties.CommonPf.GetCleanupOnFailure(ctx)),
		terraform.WithRetries(1),
	)
}

func (v *VSpherePlatform) CreateVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Terraform CreateVMs")

	planName := v.NameSanitize(vmGroupOrchestrationParams.GroupName)
	var vvgp VSphereVMGroupParams
	var vgp VSphereGeneralParams
	err := v.populateGeneralParams(ctx, planName, &vgp)
	if err != nil {
		return err
	}
	err = v.populateVMOrchParams(ctx, vmGroupOrchestrationParams)
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "Terraform orch params", "vmGroupOrchestrationParams", vmGroupOrchestrationParams)

	vvgp.VMGroupOrchestrationParams = vmGroupOrchestrationParams
	vvgp.VSphereGeneralParams = &vgp

	return terraform.CreateTerraformPlanFromTemplate(
		ctx,
		vvgp,
		planName,
		vmGroupTemplate,
		updateCallback,
		terraform.WithCleanupOnFailure(v.vmProperties.CommonPf.GetCleanupOnFailure(ctx)),
		terraform.WithRetries(1),
	)
}

func (v *VSpherePlatform) UpdateVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMs", "vmGroupOrchestrationParams", vmGroupOrchestrationParams)

	return terraform.CreateTerraformPlanFromTemplate(
		ctx,
		vmGroupOrchestrationParams,
		vmGroupOrchestrationParams.GroupName,
		vmGroupTemplate,
		updateCallback,
		terraform.WithCleanupOnFailure(v.vmProperties.CommonPf.GetCleanupOnFailure(ctx)),
		terraform.WithRetries(1),
	)
}

func (v *VSpherePlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {
	return terraform.DeleteTerraformPlan(ctx, vmGroupName)
}
