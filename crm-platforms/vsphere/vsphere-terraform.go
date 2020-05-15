package vsphere

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer/terraform"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var NumRetries = 2

var terraformCreate string = "CREATE"
var terraformUpdate string = "UPDATE"
var terraformSync string = "SYNC"
var terraformTest string = "TEST"

const TagDelimiter = "__"

const DoesNotExistError string = "does not exist"

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
	VmIpTagCategory   string
	SubnetTagCategory string
	CustomizeVMs      bool
}

type VSphereVMGroupParams struct {
	*VSphereGeneralParams
	*vmlayer.VMGroupOrchestrationParams
}

func (v *VSpherePlatform) GetVmIpTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-vmip"
}

func (v *VSpherePlatform) GetSubnetTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-subnet"
}

func (v *VSpherePlatform) TerraformRefresh(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "TerraformRefresh")
	_, err := terraform.TimedTerraformCommand(ctx, terraform.TerraformDir, "terraform", "refresh")
	return err
}

/*
func (v *VSpherePlatform) ImportTagCategories(ctx context.Context) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ImportTagCategories")

	subnetCat := v.GetSubnetTagCategory(ctx)
	out, err := terraform.TimedTerraformCommand(ctx, terraform.TerraformDir, "terraform", "import", "vsphere_tag_category."+subnetCat, subnetCat)
	if strings.Contains(out, "not found") {
		return fmt.Errorf(DoesNotExistError)
	}
	if err != nil {
		return err
	}

	vmipCat := v.GetVmIpTagCategory(ctx)
	out, err = terraform.TimedTerraformCommand(ctx, terraform.TerraformDir, "terraform", "import", "vsphere_tag_category."+vmipCat, vmipCat)
	if strings.Contains(out, "not found") {
		return fmt.Errorf(DoesNotExistError)
	}
	if err != nil {
		return err
	}
	return nil
}*/

func (v *VSpherePlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName, portName string) error {
	fileName := terraform.TerraformDir + "/" + serverName + ".tf"
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer", "serverName", "serverName", "subnetName", subnetName, "portName", portName, "fileName", fileName)

	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	beginIf := "## BEGIN ADDITIONAL INTERFACE FOR " + subnetName + "__" + serverName
	endIf := "## END ADDITIONAL INTERFACE FOR " + subnetName + "__" + serverName

	beginTag := "## BEGIN ADDITIONAL TAGS FOR " + subnetName + "__" + serverName
	endTag := "## END ADDITIONAL TAGS FOR " + subnetName + "__" + serverName

	// remove the lines between the delimters above
	lines := strings.Split(string(input), "\n")
	var newlines []string
	skipLine := false
	for i, line := range lines {
		if strings.Contains(line, beginIf) || strings.Contains(line, beginTag) {
			log.SpanLog(ctx, log.DebugLevelInfra, "skipping lines starting from", "linenum", i, "fileName", fileName)
			skipLine = true
		}
		if !skipLine {
			newlines = append(newlines, line)
		}
		if strings.Contains(line, endIf) || strings.Contains(line, endTag) {
			skipLine = false
			log.SpanLog(ctx, log.DebugLevelInfra, "resuming lines starting from", "linenum", i, "fileName", fileName)

		}
	}
	output := strings.Join(newlines, "\n")
	err = ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer doing apply after removing interfaces and tags", "serverName", serverName, "portName", portName)

	out, err := terraform.TimedTerraformCommand(ctx, terraform.TerraformDir, "terraform", "apply", "--auto-approve")
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Terraform apply failed for detach port", "out", out, "fileName", fileName)
	}
	return err
}

func (v *VSpherePlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	fileName := terraform.TerraformDir + "/" + serverName + ".tf"
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer", "serverName", serverName, "fileName", fileName, "ipaddr", ipaddr, "action", action)
	tagName := serverName + TagDelimiter + subnetName + TagDelimiter + ipaddr
	tagId := v.IdSanitize(tagName)

	interfaceContents := fmt.Sprintf(`
		## BEGIN ADDITIONAL INTERFACE FOR `+subnetName+"__"+serverName+`
		network_interface {
			network_id = vsphere_distributed_port_group.%s.id
		}
		## END ADDITIONAL INTERFACE FOR `+subnetName+"__"+serverName+`
		`, subnetName)

	tagContents := fmt.Sprintf(`
	## BEGIN ADDITIONAL TAGS FOR `+subnetName+"__"+serverName+`
	## import vsphere_tag.%s {"category_name":"%s","tag_name":"%s"}
	resource "vsphere_tag" "%s" {
		name = "%s"
		category_id = "${vsphere_tag_category.%s.id}"
	}
	## END ADDITIONAL TAGS FOR `+subnetName+"__"+serverName+`
		`, tagId, v.GetVmIpTagCategory(ctx), tagName, tagId, tagName, v.GetVmIpTagCategory(ctx))

	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	var newlines []string
	for _, line := range lines {
		if strings.Contains(line, "## END NETWORK INTERFACES for "+serverName) {
			newlines = append(newlines, interfaceContents)
		}
		newlines = append(newlines, line)
	}
	newlines = append(newlines, tagContents)
	output := strings.Join(newlines, "\n")
	err = ioutil.WriteFile(fileName, []byte(output), 0644)

	if err != nil {
		return err
	}
	if action == vmlayer.ActionCreate {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer doing apply after adding interfaces and tags", "serverName", serverName, "portName", portName)
		out, err := terraform.TimedTerraformCommand(ctx, terraform.TerraformDir, "terraform", "apply", "--auto-approve")
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Terraform apply failed for attach port", "out", out, "fileName", fileName)
		}
	} else if action == vmlayer.ActionSync {
		return v.ImportTerraformPlan(ctx, serverName, edgeproto.DummyUpdateCallback)
	}
	return err
}

func (v *VSpherePlatform) populateGeneralParams(ctx context.Context, planName string, vgp *VSphereGeneralParams, action string) error {
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
	vgp.SubnetTagCategory = v.GetSubnetTagCategory(ctx)
	vgp.VmIpTagCategory = v.GetVmIpTagCategory(ctx)
	if action != terraformSync {
		vgp.CustomizeVMs = true
	}

	return nil
}

func (v *VSpherePlatform) populateVMOrchParams(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, action string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "populateVMOrchParams")

	masterIP := ""
	flavors, err := v.GetFlavorList(ctx)
	if err != nil {
		return nil
	}

	usedCidrs, err := v.GetUsedSubnetCIDRs(ctx)
	if err != nil {
		return nil
	}
	currentSubnetName := ""
	if action != terraformCreate {
		currentSubnetName = "mex-k8s-subnet-" + vmgp.GroupName
	}

	//find an available subnet or the current subnet for update and delete
	for i, s := range vmgp.Subnets {
		if s.CIDR != vmlayer.NextAvailableResource {
			// no need to compute the CIDR
			continue
		}
		found := false
		for octet := 0; octet <= 255; octet++ {
			subnet := fmt.Sprintf("%s.%s.%d.%d/%s", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 0, vmgp.Netspec.NetmaskBits)
			// either look for an unused one (create) or the current one (update)
			newSubnet := action == terraformCreate || action == terraformTest
			if (newSubnet && usedCidrs[subnet] == "") || (!newSubnet && usedCidrs[subnet] == currentSubnetName) {
				found = true
				vmgp.Subnets[i].CIDR = subnet
				vmgp.Subnets[i].GatewayIP = fmt.Sprintf("%s.%s.%d.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 1)
				vmgp.Subnets[i].NodeIPPrefix = fmt.Sprintf("%s.%s.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet)
				masterIP = fmt.Sprintf("%s.%s.%d.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 10)

				tagname := s.Name + TagDelimiter + subnet
				tagid := v.IdSanitize(tagname)
				vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetSubnetTagCategory(ctx), Id: tagid, Name: tagname})
				break
			}

		}
		if !found {
			return fmt.Errorf("cannot find subnet cidr")
		}
	}

	// populate vm fields
	for vmidx, vm := range vmgp.VMs {
		//var vmtags []string
		vmgp.VMs[vmidx].MetaData = vmlayer.GetVMMetaData(vm.Role, masterIP, vmsphereMetaDataFormatter)
		vmgp.VMs[vmidx].UserData = vmlayer.GetVMUserData(vm.SharedVolume, vm.DNSServers, vm.DeploymentManifest, vm.Command, vmsphereUserDataFormatter)
		vmgp.VMs[vmidx].DNSServers = "\"1.1.1.1\", \"1.0.0.1\""
		for _, f := range flavors {
			if f.Name == vm.FlavorName {
				vmgp.VMs[vmidx].Vcpus = f.Vcpus
				vmgp.VMs[vmidx].Disk = f.Disk
				vmgp.VMs[vmidx].Ram = f.Ram
			}
		}

		// populate external ips
		for _, portref := range vm.Ports {
			log.SpanLog(ctx, log.DebugLevelInfra, "updating VM port", "portref", portref)
			if portref.NetworkId == v.IdSanitize(v.vmProperties.GetCloudletExternalNetwork()) {
				var eip string
				if action == terraformUpdate || action == terraformSync {
					log.SpanLog(ctx, log.DebugLevelInfra, "using current ip for action", "action", action, "server", vm.Name)
					eip, err = v.GetExternalIPForServer(ctx, vm.Name)
				} else {
					eip, err = v.GetFreeExternalIP(ctx)
				}
				if err != nil {
					return err
				}

				fip := vmlayer.FixedIPOrchestrationParams{
					Subnet:  vmlayer.NewResourceReference(portref.Name, portref.Id, false),
					Mask:    v.GetExternalNetmask(),
					Address: eip,
				}
				vmgp.VMs[vmidx].FixedIPs = append(vmgp.VMs[vmidx].FixedIPs, fip)
				tagname := vm.Name + TagDelimiter + portref.NetworkId + TagDelimiter + eip
				tagid := v.IdSanitize(tagname)
				vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetVmIpTagCategory(ctx), Id: tagid, Name: tagname})
				vmgp.VMs[vmidx].ExternalGateway = v.GetExternalGateway()
			}
		}

		// update fixedips from subnet found
		for fipidx, fip := range vm.FixedIPs {
			if fip.Address == vmlayer.NextAvailableResource {
				found := false
				for _, s := range vmgp.Subnets {
					if s.Name == fip.Subnet.Name {
						found = true
						vmgp.VMs[vmidx].FixedIPs[fipidx].Address = fmt.Sprintf("%s.%d", s.NodeIPPrefix, fip.LastIPOctet)
						vmgp.VMs[vmidx].FixedIPs[fipidx].Mask = v.GetInternalNetmask()
						if vmgp.VMs[vmidx].ExternalGateway == "" {
							vmgp.VMs[vmidx].ExternalGateway = s.GatewayIP
						}
						tagname := vm.Name + TagDelimiter + s.Id + TagDelimiter + vmgp.VMs[vmidx].FixedIPs[fipidx].Address
						tagid := v.IdSanitize(tagname)
						vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetVmIpTagCategory(ctx), Id: tagid, Name: tagname})
						log.SpanLog(ctx, log.DebugLevelInfra, "updating address for VM", "vmname", vmgp.VMs[vmidx].Name, "address", vmgp.VMs[vmidx].FixedIPs[fipidx].Address)
						break
					}
				}
				if !found {
					return fmt.Errorf("found not find subnet for vm %s", vm.Name)
				}
			}
		}

	} //for vm

	return nil
}

//return nil
//}

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

	## import vsphere_tag_category.{{.VmIpTagCategory}} {{.VmIpTagCategory}}
	resource "vsphere_tag_category" "{{.VmIpTagCategory}}" {
		name        = "{{.VmIpTagCategory}}"
		cardinality = "SINGLE"
		description = "VM IP Addresses"
	  
		associable_types = [
		  "VirtualMachine",
		]
	}

	## import vsphere_tag_category.{{.SubnetTagCategory}} {{.SubnetTagCategory}}
	resource "vsphere_tag_category" "{{.SubnetTagCategory}}" {
		name        = "{{.SubnetTagCategory}}"
		cardinality = "SINGLE"
		description = "Subnets Allocated"
	  
		associable_types = [
		  "VirtualMachine",
		]
	}
	`

var vmGroupTemplate = `

	## import vsphere_resource_pool.{{.ResourcePool}} /{{.DataCenterName}}/host/{{.ComputeCluster}}/Resources/{{.ResourcePool}}
	resource "vsphere_resource_pool" "{{.ResourcePool}}" {  
		name          = "{{.ResourcePool}}"
		parent_resource_pool_id = "${data.vsphere_compute_cluster.{{.ComputeCluster}}.resource_pool_id}"
	}

	{{- range .Subnets}}
	## import vsphere_distributed_port_group.{{.Id}} /{{$.DataCenterName}}/network/{{.Name}}
	resource "vsphere_distributed_port_group" "{{.Id}}" {
		name                            = "{{.Name}}"
		distributed_virtual_switch_uuid = "${data.vsphere_distributed_virtual_switch.{{$.InternalDVS}}.id}"
		vlan_id                         = {{.Vlan}}
	}
	{{- end}}

	{{- range .Tags}}
	## import vsphere_tag.{{.Id}} {"category_name":"{{.Category}}","tag_name":"{{.Name}}"}
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

	## import vsphere_virtual_machine.{{.Id}} /{{$.DataCenterName}}/vm/{{.Name}}
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
			network_id = vsphere_distributed_port_group.{{.SubnetId}}.id
			{{- else}}
			network_id = data.vsphere_network.{{.NetworkId}}.id
			{{- end}}
		}
		{{- end}}
		## END NETWORK INTERFACES for {{.Name}}

  		disk {
			label = "disk0"
			size = {{.Disk}}
			thin_provisioned = true
			eagerly_scrub = false
  		}

		{{- if $.CustomizeVMs}}
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
		{{- end}}
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

func (v *VSpherePlatform) ImportTerraformPlan(ctx context.Context, planName string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ImportTerraformPlan", "planName", planName)

	fileName := terraform.TerraformDir + "/" + planName + ".tf"
	notfoundReg := regexp.MustCompile("Error: .* not found")
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## import ") {
			importCmd := strings.ReplaceAll(line, "## import", "import")
			log.SpanLog(ctx, log.DebugLevelInfra, "Found import", "importCmd", importCmd)
			args := strings.Split(importCmd, " ")
			out, err := terraform.TimedTerraformCommand(ctx, terraform.TerraformDir, "terraform", args...)
			if err != nil {
				if strings.Contains(out, "Resource already managed by Terraform") {
					log.SpanLog(ctx, log.DebugLevelInfra, "resource already in terraform state")
				} else if notfoundReg.MatchString(out) {
					log.SpanLog(ctx, log.DebugLevelInfra, "resource does not exist")
				} else {
					return fmt.Errorf("Terraform import fail: %v", err)
				}
			} else {
				log.SpanLog(ctx, log.DebugLevelInfra, "Import success", "importCmd", importCmd)
			}
		}
	}
	return nil
}

func (v *VSpherePlatform) TerraformSetupVsphere(ctx context.Context, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "TerraformSetupVsphere")

	planName := v.NameSanitize(v.GetDatacenterName(ctx))
	_, staterr := os.Stat(terraform.TerraformDir)
	if staterr == nil {
		timestamp := time.Now().Format("2006-01-02T150405")
		backdir := terraform.TerraformDir + "-" + timestamp
		log.SpanLog(ctx, log.DebugLevelInfra, "backing up terraformdir", "backdir", backdir)

		err := os.Rename(terraform.TerraformDir, backdir)
		if err != nil {
			return fmt.Errorf("unable to backup terraformDir: %s %s - %v", terraform.TerraformDir, timestamp, err)
		}
	}
	err := os.Mkdir(terraform.TerraformDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to create terraformDir: %s - %v", terraform.TerraformDir, err)
	}

	var vgp VSphereGeneralParams
	err = v.populateGeneralParams(ctx, planName, &vgp, terraformCreate)
	if err != nil {
		return err
	}
	terraformFile, err := terraform.CreateTerraformPlanFromTemplate(
		ctx,
		vgp,
		planName,
		vcenterTemplate, updateCallback,
		terraform.WithInit(true),
	)
	if err != nil {
		return err
	}
	err = v.ImportTerraformPlan(ctx, planName, updateCallback)
	if err != nil {
		return err
	}

	return terraform.ApplyTerraformPlan(ctx, terraformFile, updateCallback)
}

func (v *VSpherePlatform) orchestrateVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, action string, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "Terraform orchestrateVMs", "action", action)

	planName := v.NameSanitize(vmGroupOrchestrationParams.GroupName)
	var vvgp VSphereVMGroupParams
	var vgp VSphereGeneralParams
	err := v.populateGeneralParams(ctx, planName, &vgp, action)
	if err != nil {
		return err
	}
	err = v.populateVMOrchParams(ctx, vmGroupOrchestrationParams, action)
	if err != nil {
		return err
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "Terraform orch params", "vmGroupOrchestrationParams", vmGroupOrchestrationParams)

	vvgp.VMGroupOrchestrationParams = vmGroupOrchestrationParams
	vvgp.VSphereGeneralParams = &vgp

	terraformFile, err := terraform.CreateTerraformPlanFromTemplate(
		ctx,
		vvgp,
		planName,
		vmGroupTemplate,
		updateCallback,
	)
	if err != nil {
		return err
	}
	if action == terraformSync {
		return v.ImportTerraformPlan(ctx, planName, updateCallback)
	}
	return terraform.ApplyTerraformPlan(
		ctx,
		terraformFile,
		updateCallback,
		terraform.WithCleanupOnFailure(v.vmProperties.CommonPf.GetCleanupOnFailure(ctx)),
		terraform.WithRetries(1))
}

func (v *VSpherePlatform) CreateVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs")
	return v.orchestrateVMs(ctx, vmGroupOrchestrationParams, terraformCreate, updateCallback)
}

func (v *VSpherePlatform) UpdateVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMs", "vmGroupOrchestrationParams", vmGroupOrchestrationParams)
	return v.orchestrateVMs(ctx, vmGroupOrchestrationParams, terraformUpdate, updateCallback)
}

func (v *VSpherePlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {
	return terraform.DeleteTerraformPlan(ctx, vmGroupName)
}

func (v *VSpherePlatform) SyncVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncVMs", "vmGroupOrchestrationParams", vmGroupOrchestrationParams)
	return v.orchestrateVMs(ctx, vmGroupOrchestrationParams, terraformSync, updateCallback)
}
