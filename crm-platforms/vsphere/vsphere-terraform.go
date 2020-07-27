package vsphere

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/mobiledgex/edge-cloud-infra/infracommon"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer/terraform"
	"github.com/mobiledgex/edge-cloud/log"
)

var NumTerraformRetries = 2

const DoesNotExistError string = "does not exist"

const TagFieldDomain = "domain"
const TagFieldIp = "ip"
const TagFieldSubnetName = "subnetname"
const TagFieldCidr = "cidr"
const TagFieldVmName = "vmname"
const TagFieldNetName = "netname"

// for use when a port has to be detached but we don't want to reorder nics
const UnusedPortgroup = "UNUSED_PORTGROUP"

var vmOrchestrateLock sync.Mutex

type VSphereGeneralParams struct {
	VsphereUser              string
	VspherePassword          string
	VsphereServer            string
	TerraformProviderVersion string
	DataCenterName           string
	ResourcePool             string
	ComputeCluster           string
	ExternalNetwork          string
	ExternalNetworkId        string
	InternalDVS              string
	DataStore                string
	VmIpTagCategory          string
	VmDomainTagCategory      string
	SubnetTagCategory        string
	SessionPath              string
}

type VSphereVMGroupParams struct {
	*VSphereGeneralParams
	*vmlayer.VMGroupOrchestrationParams
}

const COMMENT_BEGIN = "BEGIN"
const COMMENT_END = "END"
const COMMENT_TAGS = "TAGS"

func getCommentLabel(beginOrEnd, objectType, object string) string {
	return fmt.Sprintf("## %s ADDITIONAL %s FOR %s", beginOrEnd, objectType, object)
}

func (v *VSpherePlatform) GetVmIpTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-vmip"
}

func (v *VSpherePlatform) GetSubnetTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-subnet"
}

func (v *VSpherePlatform) GetVMDomainTagCategory(ctx context.Context) string {
	return v.GetDatacenterName(ctx) + "-vmdomain"
}

func getTagFieldMap(tag string) (map[string]string, error) {
	fieldMap := make(map[string]string)
	ts := strings.Split(tag, ",")
	for _, field := range ts {
		fs := strings.Split(field, "=")
		if len(fs) != 2 {
			return nil, fmt.Errorf("incorrectly formatted tag: %s", tag)
		}
		fieldMap[fs[0]] = fs[1]
	}
	return fieldMap, nil
}

// GetDomainFromTag get the domain from the tag which is always the last field
func (v *VSpherePlatform) GetDomainFromTag(ctx context.Context, tag string) (string, error) {
	fm, err := getTagFieldMap(tag)
	if err != nil {
		return "", err
	}
	domain, ok := fm[TagFieldDomain]
	if !ok {
		return "", fmt.Errorf("No domain found for tag")
	}
	return domain, nil

}

func (v *VSpherePlatform) GetVmIpTag(ctx context.Context, vmName, network, ipaddr string) string {
	return TagFieldVmName + "=" + vmName + "," + TagFieldNetName + "=" + network + "," + TagFieldIp + "=" + ipaddr + "," + TagFieldDomain + "=" + string(v.vmProperties.Domain)
}

// ParseVMIpTag returns vmname, network, ipaddr, domain
func (v *VSpherePlatform) ParseVMIpTag(ctx context.Context, tag string) (string, string, string, string, error) {
	fm, err := getTagFieldMap(tag)
	if err != nil {
		return "", "", "", "", err
	}
	vmname, ok := fm[TagFieldVmName]
	if !ok {
		return "", "", "", "", fmt.Errorf("No vmname in vmip tag")
	}
	network, ok := fm[TagFieldNetName]
	if !ok {
		return "", "", "", "", fmt.Errorf("No netname in vmip tag")
	}
	ip, ok := fm[TagFieldIp]
	if !ok {
		return "", "", "", "", fmt.Errorf("No ip in vmip tag")
	}
	domain, ok := fm[TagFieldDomain]
	if !ok {
		return "", "", "", "", fmt.Errorf("No domain in vmip tag")
	}
	return vmname, network, ip, domain, nil
}

func (v *VSpherePlatform) GetSubnetTag(ctx context.Context, subnetName, cidr string) string {
	return TagFieldSubnetName + "=" + subnetName + "," + TagFieldCidr + "=" + cidr + "," + TagFieldDomain + "=" + string(v.vmProperties.Domain)
}

// ParseSubnetTag returns subnetName, cidr, domain
func (v *VSpherePlatform) ParseSubnetTag(ctx context.Context, tag string) (string, string, string, error) {
	fm, err := getTagFieldMap(tag)
	if err != nil {
		return "", "", "", err
	}
	subnetName, ok := fm[TagFieldSubnetName]
	if !ok {
		return "", "", "", fmt.Errorf("No subnetname in subnet tag")
	}
	cidr, ok := fm[TagFieldCidr]
	if !ok {
		return "", "", "", fmt.Errorf("No cidr in subnet tag")
	}
	domain, ok := fm[TagFieldDomain]
	if !ok {
		return "", "", "", fmt.Errorf("No domain in subnet tag")
	}
	return subnetName, cidr, domain, nil
}

func (v *VSpherePlatform) GetVmDomainTag(ctx context.Context, vmName string) string {
	return TagFieldVmName + "=" + vmName + "," + TagFieldDomain + "=" + string(v.vmProperties.Domain)
}

// ParseVMDomainTag returns vmname, domain
func (v *VSpherePlatform) ParseVMDomainTag(ctx context.Context, tag string) (string, string, error) {
	fm, err := getTagFieldMap(tag)
	if err != nil {
		return "", "", err
	}
	vmName, ok := fm[TagFieldVmName]
	if !ok {
		return "", "", fmt.Errorf("No subnetname in vmdomain tag")
	}
	domain, ok := fm[TagFieldDomain]
	if !ok {
		return "", "", fmt.Errorf("No domain in vmdomain tag")
	}
	return vmName, domain, nil
}

func (v *VSpherePlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName, portName string) error {
	fileName := v.getTerraformDir(ctx) + "/" + serverName + ".tf"
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer", "serverName", "serverName", "subnetName", subnetName, "portName", portName, "fileName", fileName)

	// backup file
	backupFile := fileName + ".bak"
	if err := infracommon.CopyFile(fileName, backupFile); err != nil {
		return fmt.Errorf("can't backup file %s, %v", fileName, err)
	}
	defer os.Remove(backupFile)
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	beginTag := getCommentLabel(COMMENT_BEGIN, COMMENT_TAGS, subnetName+"__"+serverName)
	endTag := getCommentLabel(COMMENT_END, COMMENT_TAGS, subnetName+"__"+serverName)

	// remove the lines between the delimters above
	lines := strings.Split(string(input), "\n")
	var newlines []string
	skipSection := false
	currentNetId := fmt.Sprintf("network_id = vsphere_distributed_port_group.%s.id", subnetName)
	unusedNetId := fmt.Sprintf("network_id = vsphere_distributed_port_group.%s.id", UnusedPortgroup)

	for i, line := range lines {
		line = strings.ReplaceAll(line, currentNetId, unusedNetId)
		if strings.Contains(line, beginTag) {
			log.SpanLog(ctx, log.DebugLevelInfra, "skipping lines starting from", "linenum", i, "fileName", fileName)
			skipSection = true
		}

		if !skipSection {
			newlines = append(newlines, line)
		}
		if strings.Contains(line, endTag) {
			skipSection = false
			log.SpanLog(ctx, log.DebugLevelInfra, "resuming lines starting from", "linenum", i, "fileName", fileName)
		}
	}
	output := strings.Join(newlines, "\n")
	err = ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer doing apply after removing interfaces and tags", "serverName", serverName, "portName", portName)

	out, err := terraform.TimedTerraformCommand(ctx, v.getTerraformDir(ctx), "terraform", "apply", "--auto-approve")
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Terraform apply failed for detach port", "out", out, "fileName", fileName)
		// revert backup file
		revertErr := infracommon.CopyFile(backupFile, fileName)
		if revertErr != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "error in reverting backup file", "err", err, "backupFile", backupFile)
		}
	}
	return err
}

func (v *VSpherePlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	fileName := v.getTerraformDir(ctx) + "/" + serverName + ".tf"
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer", "serverName", serverName, "subnetName", subnetName, "fileName", fileName, "ipaddr", ipaddr, "action", action)
	// backup file
	backupFile := fileName + ".bak"
	if err := infracommon.CopyFile(fileName, backupFile); err != nil {
		return fmt.Errorf("can't backup file %s, %v", fileName, err)
	}
	defer os.Remove(backupFile)

	newNetId := fmt.Sprintf("network_id = vsphere_distributed_port_group.%s.id", subnetName)
	unusedNetId := fmt.Sprintf("network_id = vsphere_distributed_port_group.%s.id", UnusedPortgroup)

	tagContents := ""
	if ipaddr != "" {
		tagName := v.GetVmIpTag(ctx, serverName, subnetName, ipaddr)
		tagId := v.IdSanitize(tagName)

		tagContents = fmt.Sprintf("	"+getCommentLabel(COMMENT_BEGIN, COMMENT_TAGS, subnetName+"__"+serverName)+`
	## import vsphere_tag.%s {"category_name":"%s","tag_name":"%s"}
	resource "vsphere_tag" "%s" {
		name = "%s"
		category_id = "${vsphere_tag_category.%s.id}"
	}
	`+getCommentLabel(COMMENT_END, COMMENT_TAGS, subnetName+"__"+serverName)+`
		`, tagId, v.GetVmIpTagCategory(ctx), tagName, tagId, tagName, v.GetVmIpTagCategory(ctx))
	}

	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	var newlines []string
	replacedUnusedIf := false
	newInterface := fmt.Sprintf("		"+`network_interface {
			%s
		}`, newNetId)

	for _, line := range lines {
		// find an unused entry to fill, unless we are doing a sync action
		// in which case we can be adding unused entries.
		if !replacedUnusedIf {
			if strings.Contains(line, unusedNetId) {
				line = strings.ReplaceAll(line, unusedNetId, newNetId)
				replacedUnusedIf = true
			}
		}
		if strings.Contains(line, "## END NETWORK INTERFACES for "+serverName) {
			if !replacedUnusedIf {
				newlines = append(newlines, newInterface)
			}
		}
		newlines = append(newlines, line)
	}

	if tagContents != "" {
		newlines = append(newlines, tagContents)
	}
	output := strings.Join(newlines, "\n")

	err = ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		return err
	}
	if action == vmlayer.ActionCreate {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer doing apply after adding interfaces and tags", "serverName", serverName, "portName", portName)
		out, err := terraform.TimedTerraformCommand(ctx, v.getTerraformDir(ctx), "terraform", "apply", "--auto-approve")
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "Terraform apply failed for attach port", "out", out, "fileName", fileName)
			// revert backup file
			revertErr := infracommon.CopyFile(backupFile, fileName)
			if revertErr != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "error in reverting backup file", "err", err, "backupFile", backupFile)
			}
			return err
		}
	}
	return err
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
	vgp.ResourcePool = v.IdSanitize(getResourcePool(planName, string(v.vmProperties.Domain)))
	vgp.SubnetTagCategory = v.GetSubnetTagCategory(ctx)
	vgp.VmIpTagCategory = v.GetVmIpTagCategory(ctx)
	vgp.VmDomainTagCategory = v.GetVMDomainTagCategory(ctx)
	vgp.SessionPath = "sessions"
	return nil
}

func getResourcePool(planName, domain string) string {
	return planName + "-pool" + "-" + domain
}

var vcenterTemplate = `
	provider "vsphere" {
		user           = "{{.VsphereUser}}"
		password      = "{{.VspherePassword}}"
		vsphere_server = "{{.VsphereServer}}"
		# If you have a self-signed cert
		allow_unverified_ssl = true
		version = "{{.TerraformProviderVersion}}"
		persist_session = true
		rest_session_path = "{{.SessionPath}}"
		vim_session_path = "{{.SessionPath}}"
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

	## import vsphere_tag_category.{{.VmDomainTagCategory}} {{.VmDomainTagCategory}}
	resource "vsphere_tag_category" "{{.VmDomainTagCategory}}" {
		name        = "{{.VmDomainTagCategory}}"
		cardinality = "SINGLE"
		description = "compute or platform domain"
	  
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
	resource "vsphere_distributed_port_group" "UNUSED_PORTGROUP" {
		name                            = "UNUSED_PORTGROUP"
		distributed_virtual_switch_uuid = "${data.vsphere_distributed_virtual_switch.{{$.InternalDVS}}.id}"
		vlan_id                         = 999
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
	{{- if .ImageName}}
	data "vsphere_virtual_machine" "{{.TemplateId}}" {
		name          = "{{.ImageName}}"
		datacenter_id = "${data.vsphere_datacenter.dc.id}"
	}
	{{- end}}

	## import vsphere_virtual_machine.{{.Id}} /{{$.DataCenterName}}/vm/{{.Name}}
	resource "vsphere_virtual_machine" "{{.Id}}" {
		name             = "{{.Name}}"
		resource_pool_id = vsphere_resource_pool.{{$.ResourcePool}}.id
		datastore_id     = data.vsphere_datastore.datastore.id
		wait_for_guest_net_timeout = -1
		num_cpus = {{.Vcpus}}
		memory   = {{.Ram}}
		memory_reservation = {{.Ram}}
		guest_id = "ubuntu64Guest"
		scsi_type = "pvscsi"

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

		{{- range .Volumes}}
		{{if .AttachExternalDisk}}
		disk {
			label = "{{.Name}}"
			path = "{{.ImageName}}"
			datastore_id = data.vsphere_datastore.datastore.id
			attach = true
		}
		{{- else}}
		disk {
			label = "{{.Name}}"
			size = {{.Size}}
			thin_provisioned = true
			eagerly_scrub = false
			unit_number = {{.UnitNumber}}
		}
		{{- end}}
		{{- end}}

		extra_config = {
			"guestinfo.userdata" = "{{.UserData}}"
			"guestinfo.userdata.encoding" = "base64"
			"guestinfo.metadata" = "{{.MetaData}}"
			"guestinfo.metadata.encoding" = "base64"
		}
		clone {
			template_uuid = "${data.vsphere_virtual_machine.{{.TemplateId}}.id}"
			customize{
				linux_options {
					host_name = "{{.HostName}}"
					domain = "{{.DNSDomain}}"
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
	}
	{{- end}}
`

// user data is encoded as base64
func vmsphereUserDataFormatter(instring string) string {
	// despite the use of paravirtualized drivers, vSphere gets get name sda, sdb
	instring = strings.ReplaceAll(instring, "/dev/vd", "/dev/sd")
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

func (v *VSpherePlatform) getTerraformDir(ctx context.Context) string {
	return "terraform-" + v.GetDatacenterName(ctx)
}

func (v *VSpherePlatform) IsTerraformInitialized(ctx context.Context) bool {
	terraformDir := v.getTerraformDir(ctx)
	_, err := os.Stat(terraformDir)
	return err == nil
}
