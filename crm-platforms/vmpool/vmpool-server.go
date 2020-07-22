package vmpool

import (
	"context"
	"fmt"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (o *VMPoolPlatform) GetServerDetail(ctx context.Context, serverName string) (*vmlayer.ServerDetail, error) {
	if o.caches == nil {
		return nil, fmt.Errorf("cache is nil")
	}

	cKey := o.GetCloudletKey()
	if cKey == nil || cKey.Name == "" {
		return nil, fmt.Errorf("missing cloudlet key")
	}

	sd := vmlayer.ServerDetail{}
	var vmPool edgeproto.VMPool
	if o.caches.VMPoolCache.Get(cKey, &vmPool) {
		log.SpanLog(ctx, log.DebugLevelInfra, "GetServerDetail", "server name", serverName, "vmPool", vmPool)
		for _, vm := range vmPool.Vms {
			if vm.InternalName != serverName {
				continue
			}
			sd.Name = vm.InternalName
			sd.Status = vmlayer.ServerActive
			sip := vmlayer.ServerIP{}
			if vm.NetInfo.InternalIp != "" {
				sip.InternalAddr = vm.NetInfo.InternalIp
			}
			if vm.NetInfo.ExternalIp == "" {
				sip.ExternalAddr = vm.NetInfo.InternalIp
			} else {
				sip.ExternalAddr = vm.NetInfo.ExternalIp
			}
			// Add two addresses with network name:
			// 1. External network
			// 2. Internal network
			// As there won't be more than one internal network interface
			// per VM
			sip.Network = o.VMProperties.GetCloudletExternalNetwork()
			sd.Addresses = append(sd.Addresses, sip)
			sip.Network = o.VMProperties.GetCloudletMexNetwork()
			sd.Addresses = append(sd.Addresses, sip)
			return &sd, nil
		}
	}
	return &sd, fmt.Errorf("No server with a name or ID: %s exists", serverName)
}

func (o *VMPoolPlatform) waitForAction(key *edgeproto.CloudletKey, action edgeproto.VMAction) (*edgeproto.VMPoolInfo, error) {
	info := edgeproto.VMPoolInfo{}
	var lastAction edgeproto.VMAction
	for i := 0; i < 10; i++ {
		if o.caches.VMPoolInfoCache.Get(key, &info) {
			if info.Action == action {
				return &info, nil
			}
			lastAction = info.Action
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("Unable to get desired Cloudlet VM Pool action, actual action %s, desired action %s", lastAction, action)
}

func (o *VMPoolPlatform) createVMsInternal(ctx context.Context, rootLBVMName string, info *edgeproto.VMPoolInfo, vmRoles map[string]vmlayer.VMRole, updateCallback edgeproto.CacheUpdateCallback) error {
	if o.caches == nil {
		return fmt.Errorf("cache is nil")
	}

	// Allocate VMs from the pool
	info.Action = edgeproto.VMAction_VM_ACTION_ALLOCATE
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs, requestion allocation of VMs", "info", info)

	updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Allocating VMs"))
	o.caches.VMPoolInfoCache.Update(ctx, info, 0)

	// wait for vmpoolinfo action to get changed to done
	infoFound, err := o.waitForAction(&info.Key, edgeproto.VMAction_VM_ACTION_DONE)
	if err != nil {
		return err
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs, received allocated VMs", "info", infoFound)
	if infoFound.Error != "" {
		return fmt.Errorf(infoFound.Error)
	}

	// Verify & get RootLB SSH Client
	rootLBVMIP := ""
	if rootLBVMName == o.VMProperties.SharedRootLBName {
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs, get shared rootlb IP", "rootLBVMName", rootLBVMName)
		sd, err := o.GetServerDetail(ctx, rootLBVMName)
		if err != nil {
			return fmt.Errorf("failed to get shared rootLB IP for %s, %v", rootLBVMName, err)
		}
		if sd == nil || len(sd.Addresses) == 0 || sd.Addresses[0].ExternalAddr == "" {
			return fmt.Errorf("missing shared rootLB IP for %s from info %v", rootLBVMName, sd)
		}
		rootLBVMIP = sd.Addresses[0].ExternalAddr
	} else {
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs, get dedicated rootlb IP", "rootLBVMName", rootLBVMName)
		for _, vm := range infoFound.Vms {
			if vm.InternalName == rootLBVMName {
				rootLBVMIP = vm.NetInfo.ExternalIp
				break
			}
		}
	}
	if rootLBVMIP == "" {
		return fmt.Errorf("failed to get rootLB IP for %s", rootLBVMName)
	}

	rootLBClient, err := o.VMProperties.GetSSHClientFromIPAddr(ctx, rootLBVMIP)
	if err != nil {
		return fmt.Errorf("can't get rootlb ssh client for %s %v", rootLBVMIP, err)
	}

	// Setup Cluster Nodes
	masterAddr := ""
	for _, vm := range infoFound.Vms {
		role, ok := vmRoles[vm.InternalName]
		if !ok {
			return fmt.Errorf("missing role for vm role %s", vm.InternalName)
		}

		client := rootLBClient
		if vm.InternalName != rootLBVMName {
			client, err = rootLBClient.AddHop(vm.NetInfo.InternalIp, 22)
			if err != nil {
				return err
			}

		}

		// Run cleanup script
		cmd := fmt.Sprintf("sudo bash /etc/mobiledgex/cleanup-vm.sh")
		out, err := client.Output(cmd)
		if err != nil {
			return fmt.Errorf("can't cleanup vm: %s, %v", out, err)
		}

		switch role {
		case vmlayer.RoleMaster:
			masterAddr = vm.NetInfo.InternalIp
		case vmlayer.RoleNode:
		default:
			// rootlb
			continue
		}

		// bringup k8s master nodes first, then k8s worker nodes
		if role == vmlayer.RoleMaster {
			updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Setting up kubernetes master node"))
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs, setup kubernetes master node")
			cmd := fmt.Sprintf("sudo sh -x /etc/mobiledgex/install-k8s-master.sh \"ens3\" \"%s\" \"%s\"", masterAddr, masterAddr)
			out, err := client.Output(cmd)
			if err != nil {
				return fmt.Errorf("can't setup k8s master on vm %s with masteraddr %s, %s, %v", vm.InternalName, masterAddr, out, err)
			}
		}
	}
	if masterAddr == "" {
		// See if existing k8s master node exists
		for vName, vRole := range vmRoles {
			if vRole == vmlayer.RoleMaster {
				sd, err := o.GetServerDetail(ctx, vName)
				if err != nil {
					return fmt.Errorf("failed to get k8s master node IP for %s, %v", vName, err)
				}
				if sd == nil || len(sd.Addresses) == 0 || sd.Addresses[0].ExternalAddr == "" {
					return fmt.Errorf("missing k8s master node IP for %s, %v", vName, sd)
				}
				masterAddr = sd.Addresses[0].InternalAddr
				break
			}
		}
	}
	if masterAddr != "" {
		// bring other nodes once master node is up (if deployment is k8s)
		updateCallback(edgeproto.UpdateTask, fmt.Sprintf("Setting up kubernetes worker nodes"))
		for _, vm := range infoFound.Vms {
			if vmRoles[vm.InternalName] != vmlayer.RoleNode {
				continue
			}
			client, err := rootLBClient.AddHop(vm.NetInfo.InternalIp, 22)
			if err != nil {
				return err
			}
			log.SpanLog(ctx, log.DebugLevelInfra, "CreateVMs, setup kubernetes worker node", "masterAddr", masterAddr, "nodename", vm.InternalName)
			cmd := fmt.Sprintf("sudo sh -x /etc/mobiledgex/install-k8s-node.sh \"ens3\" \"%s\" \"%s\"", masterAddr, masterAddr)
			out, err := client.Output(cmd)
			if err != nil {
				return fmt.Errorf("can't setup k8s node on vm %s with masteraddr %s, %s, %v", vm.InternalName, masterAddr, out, err)
			}
		}
	}

	return nil
}

func (o *VMPoolPlatform) CreateVMs(ctx context.Context, vmGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	cKey := o.GetCloudletKey()
	if cKey == nil || cKey.Name == "" {
		return fmt.Errorf("missing cloudlet key")
	}

	// Allocate VMs from the pool
	info := edgeproto.VMPoolInfo{}
	info.Key = *cKey
	info.GroupName = vmGroupOrchestrationParams.GroupName
	info.VmSpecs = []edgeproto.VMSpec{}
	info.Action = edgeproto.VMAction_VM_ACTION_ALLOCATE

	vmRoles := make(map[string]vmlayer.VMRole)
	rootLBVMName := o.VMProperties.SharedRootLBName
	for _, vm := range vmGroupOrchestrationParams.VMs {
		vmRoles[vm.Name] = vm.Role
		vmSpec := edgeproto.VMSpec{}
		vmSpec.InternalName = vm.Name
		for _, p := range vm.Ports {
			if p.NetworkType == vmlayer.NetTypeExternal {
				vmSpec.ExternalNetwork = true
				rootLBVMName = vm.Name
				break
			}
		}
		vmSpec.InternalNetwork = true
		info.VmSpecs = append(info.VmSpecs, vmSpec)
	}

	return o.createVMsInternal(ctx, rootLBVMName, &info, vmRoles, updateCallback)
}

func (o *VMPoolPlatform) UpdateVMs(ctx context.Context, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	if o.caches == nil {
		return fmt.Errorf("cache is nil")
	}

	cKey := o.GetCloudletKey()
	if cKey == nil || cKey.Name == "" {
		return fmt.Errorf("missing cloudlet key")
	}

	// Get already created VMs
	var vmPool edgeproto.VMPool
	existingVms := make(map[string]bool)
	if o.caches.VMPoolCache.Get(cKey, &vmPool) {
		log.SpanLog(ctx, log.DebugLevelInfra, "found vmpool", "vmPool", vmPool)
		for _, vm := range vmPool.Vms {
			if vm.GroupName != VMGroupOrchestrationParams.GroupName {
				continue
			}
			existingVms[vm.InternalName] = false
		}
	}

	info := edgeproto.VMPoolInfo{}
	info.Key = *cKey
	info.GroupName = VMGroupOrchestrationParams.GroupName
	info.VmSpecs = []edgeproto.VMSpec{}

	vmRoles := make(map[string]vmlayer.VMRole)
	rootLBVMName := o.VMProperties.SharedRootLBName
	for _, vm := range VMGroupOrchestrationParams.VMs {
		vmRoles[vm.Name] = vm.Role
		vmSpec := edgeproto.VMSpec{}
		vmSpec.InternalName = vm.Name
		for _, p := range vm.Ports {
			if p.NetworkType == vmlayer.NetTypeExternal {
				vmSpec.ExternalNetwork = true
				rootLBVMName = vm.Name
				break
			}
		}
		vmSpec.InternalNetwork = true
		if _, ok := existingVms[vm.Name]; ok {
			existingVms[vm.Name] = true
			continue
		}
		info.VmSpecs = append(info.VmSpecs, vmSpec)
	}

	updateAction := edgeproto.VMAction_VM_ACTION_ALLOCATE
	if len(info.VmSpecs) == 0 {
		// no new VMs to be added, see if something is to be removed
		for vName, vPresent := range existingVms {
			if !vPresent {
				vmSpec := edgeproto.VMSpec{}
				vmSpec.InternalName = vName
				info.VmSpecs = append(info.VmSpecs, vmSpec)
				updateAction = edgeproto.VMAction_VM_ACTION_RELEASE
			}
		}
	}

	if len(info.VmSpecs) == 0 {
		log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMs, nothing to update", "info", info)
		return nil
	}

	info.Action = updateAction
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateVMs", "info", info)

	switch updateAction {
	case edgeproto.VMAction_VM_ACTION_ALLOCATE:
		return o.createVMsInternal(ctx, rootLBVMName, &info, vmRoles, updateCallback)
	case edgeproto.VMAction_VM_ACTION_RELEASE:
		return o.deleteVMsInternal(ctx, &info)
	}

	return nil
}

func (o *VMPoolPlatform) SyncVMs(ctx context.Context, VMGroupOrchestrationParams *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "SyncVMs")
	// nothing to do right now
	return nil

}

func (o *VMPoolPlatform) deleteVMsInternal(ctx context.Context, info *edgeproto.VMPoolInfo) error {
	if o.caches == nil {
		return fmt.Errorf("cache is nil")
	}

	cKey := o.GetCloudletKey()
	if cKey == nil || cKey.Name == "" {
		return fmt.Errorf("missing cloudlet key")
	}

	deleteAll := false
	if len(info.VmSpecs) == 0 {
		// delete all Vms
		deleteAll = true
	}
	vmNames := map[string]struct{}{}
	for _, vmSpec := range info.VmSpecs {
		vmNames[vmSpec.InternalName] = struct{}{}
	}

	// Cleanup VMs if possible
	var vmPool edgeproto.VMPool
	if o.caches.VMPoolCache.Get(cKey, &vmPool) {
		log.SpanLog(ctx, log.DebugLevelInfra, "found vmpool", "vmPool", vmPool)
		rootLBVMIp := ""
		sharedRootLBVMIp := ""
		rootLBName := ""
		for _, vm := range vmPool.Vms {
			if vm.InternalName == o.VMProperties.SharedRootLBName {
				sharedRootLBVMIp = vm.NetInfo.ExternalIp
				continue
			}
			if vm.GroupName != info.GroupName {
				continue
			}
			if vm.NetInfo.ExternalIp == "" {
				continue
			}
			rootLBVMIp = vm.NetInfo.ExternalIp
			rootLBName = vm.InternalName
			break
		}
		if rootLBVMIp == "" {
			rootLBVMIp = sharedRootLBVMIp
		}
		rootLBClient, err := o.VMProperties.GetSSHClientFromIPAddr(ctx, rootLBVMIp)
		if err != nil {
			// skip, as cleanup happens as part of creation as well
			log.SpanLog(ctx, log.DebugLevelInfra, "DeleteVMs, can't get rootlb ssh client for %s %v", rootLBVMIp, err)
		} else {
			client := rootLBClient
			for _, vm := range vmPool.Vms {
				if vm.GroupName != info.GroupName {
					continue
				}
				if vm.InternalName != rootLBName {
					client, err = rootLBClient.AddHop(vm.NetInfo.InternalIp, 22)
					if err != nil {
						return err
					}
				}
				_, ok := vmNames[vm.InternalName]
				if ok || deleteAll {
					// Run cleanup script
					cmd := fmt.Sprintf("sudo bash /etc/mobiledgex/cleanup-vm.sh")
					out, err := client.Output(cmd)
					if err != nil {
						return fmt.Errorf("can't cleanup vm: %s, %v", out, err)
					}
				}
			}
		}
	}

	// Release VMs from the pool
	info.Action = edgeproto.VMAction_VM_ACTION_RELEASE
	o.caches.VMPoolInfoCache.Update(ctx, info, 0)

	// wait for vmpoolinfo action to get changed to done
	infoFound, err := o.waitForAction(&info.Key, edgeproto.VMAction_VM_ACTION_DONE)
	if err != nil {
		return err
	}
	if infoFound.Error != "" {
		return fmt.Errorf(infoFound.Error)
	}

	return nil
}

func (o *VMPoolPlatform) DeleteVMs(ctx context.Context, vmGroupName string) error {
	cKey := o.GetCloudletKey()
	if cKey == nil || cKey.Name == "" {
		return fmt.Errorf("missing cloudlet key")
	}

	// Release VMs from the pool
	info := edgeproto.VMPoolInfo{}
	info.Key = *cKey
	info.GroupName = vmGroupName
	info.Action = edgeproto.VMAction_VM_ACTION_RELEASE
	info.VmSpecs = []edgeproto.VMSpec{} // empty means delete all VMs

	return o.deleteVMsInternal(ctx, &info)
}

func (s *VMPoolPlatform) GetVMStats(ctx context.Context, key *edgeproto.AppInstKey) (*vmlayer.VMMetrics, error) {
	log.SpanLog(ctx, log.DebugLevelMetrics, "GetVMStats not supported")
	return &vmlayer.VMMetrics{}, nil
}

func (s *VMPoolPlatform) GetPlatformResourceInfo(ctx context.Context) (*vmlayer.PlatformResources, error) {
	log.SpanLog(ctx, log.DebugLevelMetrics, "GetPlatformResourceInfo not supported")
	return nil, nil
}
