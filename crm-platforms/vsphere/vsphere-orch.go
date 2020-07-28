package vsphere

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

var orchVmLock sync.Mutex

var subnetToVlanLock sync.Mutex
var subnetToVlan map[string]uint32
var vlanToSubnet map[uint32]string

func init() {
	subnetToVlan = make(map[string]uint32)
	vlanToSubnet = make(map[uint32]string)
}

func getResourcePoolName(planName, domain string) string {
	return planName + "-pool" + "-" + domain
}

func (v *VSpherePlatform) GetVlanForSubnet(ctx context.Context, subnetName string) (uint32, error) {
	subnetToVlanLock.Lock()
	defer subnetToVlanLock.Unlock()
	vlan, ok := subnetToVlan[subnetName]
	if !ok {
		return 0, fmt.Errorf("No VLAN for subnet: %s", subnetName)
	}
	return vlan, nil
}

func (v *VSpherePlatform) GetSubnetForVlan(ctx context.Context, vlan uint32) (string, error) {
	subnetToVlanLock.Lock()
	defer subnetToVlanLock.Unlock()
	subnet, ok := vlanToSubnet[vlan]
	if !ok {
		return "", fmt.Errorf("No Subnet for vlan: %d", vlan)
	}
	return subnet, nil
}

func (v *VSpherePlatform) SetVlanForSubnet(ctx context.Context, subnetName string, vlan uint32) {
	log.SpanLog(ctx, log.DebugLevelInfra, "SetVlanForSubnet", "subnetName", subnetName, "vlan", vlan)
	subnetToVlanLock.Lock()
	defer subnetToVlanLock.Unlock()
	subnetToVlan[subnetName] = vlan
	vlanToSubnet[vlan] = subnetName
}

func (v *VSpherePlatform) DeleteResourcesForGroup(ctx context.Context, groupName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteResourcesForGroup", "groupName", groupName)
	// get all vm names
	vmtags, err := v.GetTagsMatchingField(ctx, TagFieldGroup, groupName, v.GetVMDomainTagCategory(ctx))
	if err != nil {
		return err
	}
	for _, vmtag := range vmtags {
		vmname, _, err := v.ParseVMDomainTag(ctx, vmtag.Name)
		if err != nil {
			return err
		}
		err = v.DeleteVM(ctx, vmname)
		if err != nil {
			return err
		}
		v.DeleteTag(ctx, vmtag.Name)
	}

	// delete subnet tag
	subTags, err := v.GetTagsMatchingField(ctx, TagFieldGroup, groupName, v.GetSubnetTagCategory(ctx))
	if err != nil {
		return err
	}
	for _, subTag := range subTags {
		v.DeleteTag(ctx, subTag.Name)
	}

	// delete vmip tags
	ipTags, err := v.GetTagsMatchingField(ctx, TagFieldGroup, groupName, v.GetVmIpTagCategory(ctx))
	if err != nil {
		return err
	}
	for _, ipTag := range ipTags {
		v.DeleteTag(ctx, ipTag.Name)
	}

	// delete resource pool
	poolName := getResourcePoolName(groupName, string(v.vmProperties.Domain))
	return v.DeletePool(ctx, poolName)
}

func getPortGroupNameForVlan(vlan uint32) string {
	return fmt.Sprintf("VLAN-%d", vlan)
}

// CreatePortGroup creates a portgroup on a DVS for a particular VLAN.  Since Govc does not currently support
// deleting port groups, we use a generic name "VLAN-x" and re-use the port groups when subnets are deleted/added
func (v *VSpherePlatform) CreatePortGroup(ctx context.Context, dvs string, pgName string, vlan uint32) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreatePortGroup", "dvs", dvs, "vlan", vlan)
	dcName := v.GetDatacenterName(ctx)
	out, err := v.TimedGovcCommand(ctx, "govc", "dvs.portgroup.add", "-dc", dcName, "-dvs", dvs, "-vlan", fmt.Sprintf("%d", vlan), "-nports", "100", pgName)
	if err != nil {
		if strings.Contains(string(out), "already exists") {
			log.SpanLog(ctx, log.DebugLevelInfra, "CreatePortGroup already exists", "pgName", pgName)
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in CreatePortGroup", "portGroupName", pgName, "out", out, "err", err)
		return fmt.Errorf("Failed to create port group: %s - %v", pgName, err)
	}
	return nil

}

func (v *VSpherePlatform) DeletePool(ctx context.Context, poolName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeletePool", "poolName", poolName)
	dcName := v.GetDatacenterName(ctx)
	computeCluster := v.GetComputeCluster()
	pathPrefix := fmt.Sprintf("/%s/host/%s/Resources/", dcName, computeCluster)
	poolPath := pathPrefix + poolName
	out, err := v.TimedGovcCommand(ctx, "govc", "pool.destroy", "-dc", dcName, poolPath)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in delete pool", "poolName", poolName, "out", out, "err", err)
		return fmt.Errorf("Failed to delete pool: %s - %v", poolName, err)
	}
	return nil
}

func (v *VSpherePlatform) CreatePool(ctx context.Context, poolName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreatePool", "poolName", poolName)

	dcName := v.GetDatacenterName(ctx)
	computeCluster := v.GetComputeCluster()
	pathPrefix := fmt.Sprintf("/%s/host/%s/Resources/", dcName, computeCluster)
	poolPath := pathPrefix + poolName
	out, err := v.TimedGovcCommand(ctx, "govc", "pool.create", "-dc", dcName, poolPath)
	if err != nil {
		if strings.Contains(string(out), "already exists") {
			log.SpanLog(ctx, log.DebugLevelInfra, "Pool already exists", "poolName", poolName)
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in create pool", "poolName", poolName, "out", out, "err", err)
		return fmt.Errorf("Failed to create pool: %s - %v", poolName, err)
	}
	return nil
}

func maskLenToMask(maskLen string) (string, error) {
	cidr := "255.255.255.255/" + maskLen
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	return ipnet.IP.String(), nil
}

func (v *VSpherePlatform) populateOrchestrationParams(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, action vmlayer.ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "populateOrchestrationParams")

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
	if action != vmlayer.ActionCreate {
		currentSubnetName = vmlayer.MexSubnetPrefix + vmgp.GroupName
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
			newSubnet := action == vmlayer.ActionCreate
			if (newSubnet && usedCidrs[subnet] == "") || (!newSubnet && usedCidrs[subnet] == currentSubnetName) {
				found = true
				vmgp.Subnets[i].CIDR = subnet
				vmgp.Subnets[i].GatewayIP = fmt.Sprintf("%s.%s.%d.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 1)
				vmgp.Subnets[i].NodeIPPrefix = fmt.Sprintf("%s.%s.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet)
				vmgp.Subnets[i].Vlan = 1000 + uint32(octet)
				v.SetVlanForSubnet(ctx, s.Name, vmgp.Subnets[i].Vlan)
				masterIP = fmt.Sprintf("%s.%s.%d.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 10)
				tagname := v.GetSubnetTag(ctx, vmgp.GroupName, s.Name, subnet)
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
		userdata, err := vmlayer.GetVMUserData(vm.SharedVolume, vm.DNSServers, vm.DeploymentManifest, vm.Command, vm.ChefParams, vmsphereUserDataFormatter)
		if err != nil {
			return err
		}
		vmgp.VMs[vmidx].UserData = userdata
		vmgp.VMs[vmidx].DNSServers = "1.1.1.1,1.0.0.1"
		flavormatch := false
		for _, f := range flavors {
			if f.Name == vm.FlavorName {
				vmgp.VMs[vmidx].Vcpus = f.Vcpus
				vmgp.VMs[vmidx].Disk = f.Disk
				vmgp.VMs[vmidx].Ram = f.Ram
				flavormatch = true
				break
			}
		}
		if vm.ImageName != "" {
			vmgp.VMs[vmidx].TemplateId = v.IdSanitize(vm.ImageName) + "-tmplt-" + vm.Id
		}
		if !flavormatch {
			return fmt.Errorf("No match in flavor cache for flavor name: %s", vm.FlavorName)
		}
		if vm.AttachExternalDisk {
			// AppVMs use a generic template with the disk attached separately
			var vol vmlayer.VolumeOrchestrationParams

			vol = vmlayer.VolumeOrchestrationParams{
				Name:               "disk0",
				ImageName:          vmgp.VMs[vmidx].ImageFolder + "/" + vmgp.VMs[vmidx].ImageName + ".vmdk",
				AttachExternalDisk: true,
			}

			vmgp.VMs[vmidx].Volumes = append(vmgp.VMs[vmidx].Volumes, vol)
			vmgp.VMs[vmidx].ImageName = ""
		} else {
			vol := vmlayer.VolumeOrchestrationParams{
				Name:               "disk0",
				Size:               vmgp.VMs[vmidx].Disk,
				AttachExternalDisk: false,
			}
			vmgp.VMs[vmidx].Volumes = append(vmgp.VMs[vmidx].Volumes, vol)

		}

		// populate external ips
		for pi, portref := range vm.Ports {
			log.SpanLog(ctx, log.DebugLevelInfra, "updating VM port", "portref", portref)
			if portref.NetworkId == v.IdSanitize(v.vmProperties.GetCloudletExternalNetwork()) {
				var eip string
				if action == vmlayer.ActionUpdate {
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
				tagname := v.GetVmIpTag(ctx, vmgp.GroupName, vm.Name, portref.NetworkId, eip)
				tagid := v.IdSanitize(tagname)
				vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetVmIpTagCategory(ctx), Id: tagid, Name: tagname})
				vmgp.VMs[vmidx].ExternalGateway, _ = v.GetExternalGateway(ctx, "")
			} else {
				vlan, ok := subnetToVlan[portref.SubnetId]
				if !ok {
					return fmt.Errorf("cannot find vlan for subnet: %s", portref.SubnetId)
				}
				vm.Ports[pi].PortGroup = getPortGroupNameForVlan(vlan)
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
						tagname := v.GetVmIpTag(ctx, vmgp.GroupName, vm.Name, s.Id, vmgp.VMs[vmidx].FixedIPs[fipidx].Address)
						tagid := v.IdSanitize(tagname)
						vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetVmIpTagCategory(ctx), Id: tagid, Name: tagname})
						log.SpanLog(ctx, log.DebugLevelInfra, "updating address for VM", "vmname", vmgp.VMs[vmidx].Name, "address", vmgp.VMs[vmidx].FixedIPs[fipidx].Address)
						break
					}
				}
				if !found {
					return fmt.Errorf("subnet for vm %s not found", vm.Name)
				}
			}
		}
		tagname := v.GetVmDomainTag(ctx, vmgp.GroupName, vm.Name)
		tagid := v.IdSanitize(tagname)
		vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetVMDomainTagCategory(ctx), Id: tagid, Name: tagname})

	} //for vm

	return nil
}

func (v *VSpherePlatform) DeleteVM(ctx context.Context, vmName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVM", "vmName", vmName)
	dcName := v.GetDatacenterName(ctx)
	out, err := v.TimedGovcCommand(ctx, "govc", "vm.destroy", "-dc", dcName, vmName)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in vm.destroy", "vmName", vmName, "out", out, "err", err)

		if strings.Contains(string(out), "not found") {
			log.SpanLog(ctx, log.DebugLevelInfra, "VM already gone", "vmName", vmName)
		} else {
			return fmt.Errorf("Error in deleting VM: %s", vmName)
		}
	}
	return nil
}

func (v *VSpherePlatform) CreateVM(ctx context.Context, vm *vmlayer.VMOrchestrationParams, poolName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVM", "vmName", vm.Name, "poolName", poolName)

	dcName := v.GetDatacenterName(ctx)
	computeCluster := v.GetComputeCluster()
	pathPrefix := fmt.Sprintf("/%s/host/%s/Resources/", dcName, computeCluster)
	poolPath := pathPrefix + poolName

	if len(vm.Ports) == 0 {
		return fmt.Errorf("No networks assigned to VM")
	}
	cloneArgs := []string{"vm.clone",
		"-dc", dcName,
		"-on=false",
		"-vm", vm.ImageName,
		"-pool", poolPath,
		"-c", fmt.Sprintf("%d", vm.Vcpus),
		"-m", fmt.Sprintf("%d", vm.Ram)}

	for _, port := range vm.Ports {
		netname := port.NetworkId
		if port.PortGroup != "" {
			netname = port.PortGroup
		}
		cloneArgs = append(cloneArgs, []string{"-net", netname}...)
	}
	cloneArgs = append(cloneArgs, vm.Name)
	out, err := v.TimedGovcCommand(ctx, "govc", cloneArgs...)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in clone VM", "vmName", vm.Name, "out", string(out), "err", err)
		return fmt.Errorf("Failed to create VM: %s - %v", vm.Name, err)
	}
	// customize it
	custArgs := []string{"vm.customize",
		"-vm", vm.Name,
		"-name", vm.HostName,
		"-gateway", vm.ExternalGateway,
		"-dc", dcName}

	if len(vm.FixedIPs) == 0 {
		return fmt.Errorf("No IP for VM: %s", vm.Name)
	}
	for _, ip := range vm.FixedIPs {
		netmask, err := maskLenToMask(ip.Mask)
		if err != nil {
			return err
		}
		custArgs = append(custArgs, []string{"-dns-server", vm.DNSServers}...)
		custArgs = append(custArgs, []string{"-ip", ip.Address}...)
		custArgs = append(custArgs, []string{"-netmask", netmask}...)
	}
	out, err = v.TimedGovcCommand(ctx, "govc", custArgs...)

	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in customize VM", "vmName", vm.Name, "out", string(out), "err", err)
		return fmt.Errorf("Failed to create VM: %s - %v", vm.Name, err)
	}

	// update guestinfo
	out, err = v.TimedGovcCommand(ctx, "govc", "vm.change",
		"-dc", dcName,
		"-e", "guestinfo.metadata="+vm.MetaData,
		"-e", "guestinfo.metadata.encoding=base64",
		"-e", "guestinfo.userdata="+vm.UserData,
		"-e", "guestinfo.userdata.encoding=base64",
		"-vm", vm.Name)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in change VM", "vmName", vm.Name, "out", string(out), "err", err)
		return fmt.Errorf("Failed to create VM: %s - %v", vm.Name, err)
	}
	return v.SetPowerState(ctx, vm.Name, vmlayer.ActionStart)
}

func (v *VSpherePlatform) CreateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs")

	// lock until all the tags are created, meaning we have the IPs picked
	orchVmLock.Lock()
	v.populateOrchestrationParams(ctx, vmgp, vmlayer.ActionCreate)
	log.SpanLog(ctx, log.DebugLevelInfra, "Updated Group Orch Parms", "vmgp", vmgp)

	updateCallback(edgeproto.UpdateTask, "Creating vCenter Tags")

	for _, t := range vmgp.Tags {
		err := v.CreateTag(ctx, &t)
		if err != nil {
			orchVmLock.Unlock()
			return err
		}
	}
	orchVmLock.Unlock()
	updateCallback(edgeproto.UpdateTask, "Creating Distributed Port Groups")

	for _, s := range vmgp.Subnets {
		pgName := getPortGroupNameForVlan(s.Vlan)
		err := v.CreatePortGroup(ctx, v.GetInternalVSwitch(), pgName, s.Vlan)
		if err != nil {
			return err
		}
	}

	poolName := getResourcePoolName(vmgp.GroupName, string(v.vmProperties.Domain))
	err := v.CreatePool(ctx, poolName)
	if err != nil {
		return err
	}
	vmCreateResults := make(chan string, len(vmgp.VMs))
	updateCallback(edgeproto.UpdateTask, "Creating VMs")
	for vmidx := range vmgp.VMs {
		log.SpanLog(ctx, log.DebugLevelInfra, "Creating VM", "vmName", vmgp.VMs[vmidx].Name)
		go func(idx int) {
			err := v.CreateVM(ctx, &vmgp.VMs[idx], poolName)
			if err == nil {
				vmCreateResults <- ""
			} else {
				vmCreateResults <- err.Error()
			}
		}(vmidx)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "Waiting for VM create results")

	errFound := false
	for range vmgp.VMs {
		result := <-vmCreateResults
		if result != "" {
			errFound = true
		}
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "All VMs finished, checking results")
	if errFound {
		if !vmgp.SkipCleanupOnFailure {
			updateCallback(edgeproto.UpdateTask, "Cleaning up after failure")
			err := v.DeleteResourcesForGroup(ctx, vmgp.GroupName)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfra, "cleanup failed: %v", err)

			}
		}
		return fmt.Errorf("CreateVMs failed")
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs complete")
	return nil
}

func (v *VSpherePlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVMs", "vmGroupName", vmGroupName)
	return v.DeleteResourcesForGroup(ctx, vmGroupName)
}
func (v *VSpherePlatform) UpdateVMs(ctx context.Context, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("TODO UPDATEVMS")
}

func (v *VSpherePlatform) AttachPortToServer(ctx context.Context, serverName, subnetName, portName, ipaddr string, action vmlayer.ActionType) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer", "serverName", serverName, "subnetName", subnetName)
	vlan, err := v.GetVlanForSubnet(ctx, subnetName)

	if err != nil {
		return err
	}
	portGrp := getPortGroupNameForVlan(vlan)
	attached, err := v.IsPortgrpAttached(ctx, serverName, portGrp)
	if err != nil {
		return err
	}
	if attached {
		log.SpanLog(ctx, log.DebugLevelInfra, "AttachPortToServer port already attached")
	} else {
		dcName := v.GetDatacenterName(ctx)
		out, err := v.TimedGovcCommand(ctx, "govc", "vm.network.add", "-dc", dcName, "-vm", serverName, "-net", portGrp)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "vm.network.add failed", "out", out, "err", err)
			return fmt.Errorf("AttachPortToServer failed")
		}
	}
	// now create the tag
	tagName := v.GetVmIpTag(ctx, serverName, serverName, subnetName, ipaddr)
	tagId := v.IdSanitize(tagName)
	tag := vmlayer.TagOrchestrationParams{
		Name:     tagName,
		Id:       tagId,
		Category: v.GetVmIpTagCategory(ctx),
	}
	return v.CreateTag(ctx, &tag)
}

// we never actually detach the port, just keep it attached to the vlan which will be reused
func (v *VSpherePlatform) DetachPortFromServer(ctx context.Context, serverName, subnetName string, portName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DetachPortFromServer", "subnetName", subnetName, "portName", portName)
	// get all the ip tags for this server
	tags, err := v.GetTagsMatchingField(ctx, TagFieldVmName, serverName, v.GetVmIpTagCategory(ctx))
	if err != nil {
		return err
	}
	// delete the tag matching this subnet
	for _, t := range tags {
		_, tagnet, _, _, err := v.ParseVMIpTag(ctx, t.Name)
		if err != nil {
			return err
		}
		if tagnet == subnetName {
			return v.DeleteTag(ctx, t.Name)
		}
	}
	return fmt.Errorf("DetachPortFromServer failed: no IP tag found")
}
