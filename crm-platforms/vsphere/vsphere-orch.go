package vsphere

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func getResourcePoolName(planName, domain string) string {
	return planName + "-pool" + "-" + domain
}

func (v *VSpherePlatform) DeleteResourcesForGroup(ctx context.Context, groupName string) error {
	return fmt.Errorf("DeleteResources TODO")
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
				masterIP = fmt.Sprintf("%s.%s.%d.%d", vmgp.Netspec.Octets[0], vmgp.Netspec.Octets[1], octet, 10)
				tagname := v.GetSubnetTag(ctx, s.Name, subnet)
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
		vmgp.VMs[vmidx].DNSServers = "\"1.1.1.1\", \"1.0.0.1\""
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
		for _, portref := range vm.Ports {
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
				tagname := v.GetVmIpTag(ctx, vm.Name, portref.NetworkId, eip)
				tagid := v.IdSanitize(tagname)
				vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetVmIpTagCategory(ctx), Id: tagid, Name: tagname})
				vmgp.VMs[vmidx].ExternalGateway, _ = v.GetExternalGateway(ctx, "")
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
						tagname := v.GetVmIpTag(ctx, vm.Name, s.Id, vmgp.VMs[vmidx].FixedIPs[fipidx].Address)
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
		tagname := v.GetVmDomainTag(ctx, vm.Name)
		tagid := v.IdSanitize(tagname)
		vmgp.Tags = append(vmgp.Tags, vmlayer.TagOrchestrationParams{Category: v.GetVMDomainTagCategory(ctx), Id: tagid, Name: tagname})

	} //for vm

	return nil
}

func (v *VSpherePlatform) CreateVM(ctx context.Context, vm *vmlayer.VMOrchestrationParams, poolName string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreatePool", "vmName", vm.Name, "poolName", poolName)

	dcName := v.GetDatacenterName(ctx)
	computeCluster := v.GetComputeCluster()
	pathPrefix := fmt.Sprintf("/%s/host/%s/Resources/", dcName, computeCluster)
	poolPath := pathPrefix + poolName

	if len(vm.Ports) == 0 {
		return fmt.Errorf("No networks assigned to VM")
	}
	primaryNet := vm.Ports[0].NetworkId
	out, err := v.TimedGovcCommand(ctx, "govc", "vm.clone",
		"-dc", dcName,
		"-net", primaryNet,
		"-on=false",
		"-vm", vm.ImageName,
		"-pool", poolPath,
		"-c", fmt.Sprintf("%d", vm.Vcpus),
		"-m", fmt.Sprintf("%d", vm.Ram),
		vm.Name)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in clone VM", "vmName", vm.Name, "out", string(out), "err", err)
		return fmt.Errorf("Failed to create VM: %s - %v", vm.Name, err)
	}
	// customize it
	netmask, err := maskLenToMask(v.GetExternalNetmask())
	if err != nil {
		return err
	}
	if len(vm.FixedIPs) == 0 {
		return fmt.Errorf("No IP for VM: %s", vm.Name)
	}
	eip := vm.FixedIPs[0].Address
	out, err = v.TimedGovcCommand(ctx, "govc", "vm.customize",
		"-name", vm.HostName,
		"-dc", dcName,
		"-ip", eip,
		"-gateway", vm.ExternalGateway,
		"-netmask", netmask,
		"-vm", vm.Name)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in customize VM", "vmName", vm.Name, "out", string(out), "err", err)
		return fmt.Errorf("Failed to create VM: %s - %v", vm.Name, err)
	}

	// update guestinfo
	masterIp := ""
	userdata, err := vmlayer.GetVMUserData(vm.SharedVolume, vm.DNSServers, vm.DeploymentManifest, vm.Command, vm.ChefParams, vmsphereUserDataFormatter)
	if err != nil {
		return err
	}
	out, err = v.TimedGovcCommand(ctx, "govc", "vm.change",
		"-dc", dcName,
		"-e", "guestinfo.metadata="+vmlayer.GetVMMetaData(vm.Role, masterIp, vmsphereMetaDataFormatter),
		"-e", "guestinfo.metadata.encoding=base64",
		"-e", "guestinfo.userdata="+userdata,
		"-e", "guestinfo.userdata.encoding=base64",
		"-vm", vm.Name)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "Error in change VM", "vmName", vm.Name, "out", string(out), "err", err)
		return fmt.Errorf("Failed to create VM: %s - %v", vm.Name, err)
	}

	v.SetPowerState(ctx, vm.Name, vmlayer.ActionStart)

	// if len(vm.Ports) > 1{
	//	....
	//}
	return nil
}

func (v *VSpherePlatform) CreateTag(ctx context.Context, tag *vmlayer.TagOrchestrationParams) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateTag", "tag", tag)

	out, err := v.TimedGovcCommand(ctx, "govc", "tags.create", "-c", tag.Category, tag.Name)
	if err != nil {
		if strings.Contains(string(out), "ALREADY_EXISTS") {
			log.SpanLog(ctx, log.DebugLevelInfra, "Tag already exists", "tag", tag)
			return nil
		}
		return fmt.Errorf("Error in creating tag: %s - %v", tag.Name, err)
	}
	return nil
}

func (v *VSpherePlatform) CreateVMs(ctx context.Context, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs")

	v.populateOrchestrationParams(ctx, vmgp, vmlayer.ActionCreate)

	for _, t := range vmgp.Tags {
		err := v.CreateTag(ctx, &t)
		if err != nil {
			return err
		}
	}

	poolName := getResourcePoolName(vmgp.GroupName, string(v.vmProperties.Domain))
	err := v.CreatePool(ctx, poolName)
	if err != nil {
		return err
	}
	errFound := false
	for vmidx, vm := range vmgp.VMs {
		log.SpanLog(ctx, log.DebugLevelInfra, "Creating VM", "vmName", vm.Name)
		err := v.CreateVM(ctx, &vmgp.VMs[vmidx], poolName)
		if err != nil {
			errFound = true
		}
	}
	if errFound {
		v.DeleteResourcesForGroup(ctx, vmgp.GroupName)
		return fmt.Errorf("CreateVMs failed")
	}
	return nil
}

func (v *VSpherePlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {
	return fmt.Errorf("TODO DELETEVMS")
}
func (o *VSpherePlatform) UpdateVMs(ctx context.Context, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	return fmt.Errorf("TODO UPDATEVMS")
}
