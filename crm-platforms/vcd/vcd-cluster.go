package vcd

import (
	"context"
	"fmt"
	//vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strings"
)

// Cluster related operations
func (v *VcdPlatform) FindCluster(ctx context.Context, clusterName string) (*Cluster, error) {
	if v.Objs.Cloudlet == nil {
		return nil, fmt.Errorf("Cluster not found")
	}
	for _, cluster := range v.Objs.Cloudlet.Clusters {
		if cluster.Name == clusterName {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("Cluster not found")
}

// Should return an object of MexCluster type, not a cidr string
func (v *VcdPlatform) CreateCluster(ctx context.Context, cloud *MexCloudlet, tmpl *govcd.VAppTemplate, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) (string, error) {

	if tmpl == nil {
		return "", fmt.Errorf("template nil")
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "create", "groupName", vmgp.GroupName, "[vm0]", vmgp.VMs[0].Name)

	clusterName := vmgp.VMs[0].Name
	cluster, err := v.FindCluster(ctx, clusterName)
	if err == nil {
		// we have a cluster by this name already
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateCluster cluster exists", "cluster", clusterName)
		return clusterName, nil
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCluster", "create", clusterName, "cloudlet", cloud.CloudletName)

	nextCidr, err := v.GetNextInternalNet(ctx)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "create next internal net failed: ", "GroupName", vmgp.GroupName, "err", err)
		return "clusterName", err
	}
	cluster = &Cluster{}
	v.Objs.Cloudlet.Clusters[nextCidr] = cluster
	cluster.Name = clusterName

	log.SpanLog(ctx, log.DebugLevelInfra, "create", "cluster", clusterName, "cidr", nextCidr)
	vapp := cloud.CloudVapp

	// Add the internal net to the vapp so we can create a networkConnectSection for the VM giving this internal network
	// Since I don't think we can add it only to the vm.
	// case 1

	internalNetName, err := v.CreateInternalNetworkForNewVm(ctx, vapp, vmgp, nextCidr)
	if err != nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "create cluster internal net failed", "err", err)
		return clusterName, err
	}

	// verify we have a new internal network
	a := strings.Split(nextCidr, "/")
	baseAddr := string(a[0])
	vmIp := ""
	lbvm := &govcd.VM{}
	netConIdx := 0
	cvm := ClusterVm{}

	for n, vmparams := range vmgp.VMs {

		ncs := &types.NetworkConnectionSection{}
		tmpl.VAppTemplate.Name = vmparams.Name

		task, err := vapp.AddNewVM(vmparams.Name, *tmpl, ncs, true)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelInfra, "create add vm failed", "err", err)
			return "clusterName", err
		}
		err = task.WaitTaskCompletion()
		if err != nil { // fatal?
			log.SpanLog(ctx, log.DebugLevelInfra, "wait add vm failed", "err", err)
		}

		vm, err := vapp.GetVMByName(vmparams.Name, true)
		if err != nil {
			// internal error
			return "clusterName", err
		}
		ncs, err = vm.GetNetworkConnectionSection()
		if err != nil {
			return "", err
		}
		vmType := string(vmlayer.GetVmTypeForRole(string(vmparams.Role)))
		if vmparams.Role == vmlayer.RoleAgent {
			lbvm = vm
			vmIp = baseAddr
		} else {
			vmIp = v.IncrIP(baseAddr, 100+(n-1))
			ncs.PrimaryNetworkConnectionIndex = 0
		}

		if v.Verbose {
			log.SpanLog(ctx, log.DebugLevelInfra, "create add vm", "Name", vm.VM.Name, "Role", vmparams.Role, "Type", vmType, "IP", vmIp, "netname", "internalNetName", "idx", netConIdx)

		}
		// create our clusterVM object
		cvm = ClusterVm{
			vmName:          vm.VM.Name,
			vmRole:          string(vmparams.Role),
			vmType:          vmType,
			vmFlavor:        vmparams.FlavorName,
			vmParentCluster: clusterName,
			vm:              vm,
		}
		err = v.AddMetadataToVM(ctx, vm, vmparams, vmType, clusterName)
		if err != nil {
			return clusterName, err
		}
		cvm.vmIPs.InternalIp = vmIp
		// vu.DumpNetworkConnectionSection(ncs, 1)

		ncs.NetworkConnection = append(ncs.NetworkConnection,
			&types.NetworkConnection{
				Network:                 internalNetName,
				NetworkConnectionIndex:  netConIdx,
				IPAddress:               vmIp,
				IsConnected:             true,
				IPAddressAllocationMode: types.IPAllocationModeManual,
			})

		err = vm.UpdateNetworkConnectionSection(ncs)
		if err != nil {
			return clusterName, err
		}
		subnet := ""
		err = v.updateVM(ctx, vm, vmparams, subnet)
	}
	// Ok, here, we have both cluster vm on the internal net. Now go back and
	// add the external to our LB
	extAddr := ""
	// Hopefully, adding this second will make it's default route first
	extAddr, err = v.GetNextExtAddrForVdcNet(ctx, cloud.ParentVdc)
	if err != nil {
		return clusterName, fmt.Errorf("Agent node failed to obtain external net IP")
	}
	v.Objs.Cloudlet.ExtVMMap[extAddr] = lbvm
	cvm.vmIPs.ExternalIp = extAddr
	// make the external the primray index 0
	netName := cloud.ExtNet.OrgVDCNetwork.Name // just PrimaryNet
	err = v.AddExtNetToVm(ctx, lbvm, netName, extAddr)
	if err != nil {
		return clusterName, err
	}
	// Would be nice to have dedicated or shared here?
	task, err := lbvm.AddMetadata("ClusterVM", clusterName)
	if err != nil {
		return clusterName, err
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return clusterName, err
	}
	// try power_on with only our ext net and add the internal net after that, do we still
	// get blocked by the internal nets default route coming first?
	for _, vmparams := range vmgp.VMs {
		vm, err := vapp.GetVMByName(vmparams.Name, true)
		if err != nil {
			return "clusterName", err
		}
		task, err := vm.PowerOn()
		if err != nil {
			return clusterName, err
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return clusterName, err
		}

		log.SpanLog(ctx, log.DebugLevelInfra, "create cluster vm power on ", "vmName", vm.VM.Name)
	}
	log.SpanLog(ctx, log.DebugLevelInfra, "create cluster power on ", "cluster", clusterName)
	return clusterName, nil
}

func (v *VcdPlatform) DeleteCluster(ctx context.Context, clusterName string /* cloud MexCloudlet, vmMap *CidrMap*/) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCluster")
	cld := v.Objs.Cloudlet
	if cld == nil {
		log.SpanLog(ctx, log.DebugLevelInfra, "No such cloudlet")
		return nil
	}
	clusters := v.Objs.Cloudlet.Clusters
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			for _, cvm := range cluster.VMs {
				err := v.DeleteVM(ctx, cvm.vm)
				if err != nil {
					log.SpanLog(ctx, log.DebugLevelInfra, "err deleting cvm:", "vmName", cvm.vm.VM.Name, "cluster", clusterName)
				}
			}
		}
	}

	return nil

}

func (v *VcdPlatform) UpdateCluster(ctx context.Context, cloud MexCloudlet, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) (*CidrMap, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "UpdateCluster")
	return nil, nil

}

func (v *VcdPlatform) RestartCluster(ctx context.Context, vmMap *CidrMap) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "RestartCluster")
	return nil

}

func (v *VcdPlatform) StartCluster(ctx context.Context, vmMap *CidrMap) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "StopCluster")
	return nil

}

func (v *VcdPlatform) StopCluster(ctx context.Context, vmMap *CidrMap) error {
	return nil
}

func (v *VcdPlatform) FindClusterVM(ctx context.Context, name string) (string, *ClusterVm, error) {
	if v.Objs.Cloudlet == nil {
		return "", nil, fmt.Errorf("No Cloudlet exists yet")
	}
	if v.Objs.Cloudlet.Clusters == nil {
		return "", nil, fmt.Errorf("No clusters exist in Cloudlet")
	}

	for _, clust := range v.Objs.Cloudlet.Clusters {
		for _, vm := range clust.VMs {
			if vm.vmName == name {
				return clust.Name, &vm, nil
			}
		}
	}
	return "", nil, fmt.Errorf("VM not found")
}
