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
		fmt.Printf("\nCreateCluster-E-can't create Cluster without a template = %+v\n", tmpl)
		return "", fmt.Errorf("template nil")
	}

	clusterName := vmgp.VMs[0].Name
	cluster, err := v.FindCluster(ctx, clusterName)
	if err == nil {
		// we have a cluster by this name already
		log.SpanLog(ctx, log.DebugLevelInfra, "CreateCluster cluster exists", "cluster", clusterName)
		return clusterName, fmt.Errorf("Cluster already exists")
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCluster", "create", clusterName, "cloudlet", cloud.CloudletName)

	nextCidr, err := v.GetNextInternalNet(ctx)
	if err != nil {
		fmt.Printf("GetNextInternalNet failed: %s\n", err.Error())
		return "", err
	}
	cluster = &Cluster{}
	v.Objs.Cloudlet.Clusters[nextCidr] = cluster
	cluster.Name = clusterName

	fmt.Printf("\n\tCreateCluster2-I-new cluster's CIDR: %s on cloudlet: %s\n", nextCidr, cloud.CloudletName)
	vapp := cloud.CloudVapp
	//vdc := cloud.ParentVdc

	// Add the internal net to the vapp so we can create a networkConnectSection for the VM giving this internal network
	// Since I don't think we can add it only to the vm.
	// case 1

	internalNetName, err := v.CreateInternalNetworkForNewVm(ctx, vapp, vmgp, nextCidr)
	if err != nil {
		fmt.Printf("CreateCluster-E-error creating internal network for cluster vms: %s\n", err.Error())
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
		fmt.Printf("\n\tCreateCluster-I-adding new vm %s #%d \n", vmparams.Name, n)

		task, err := vapp.AddNewVM(vmparams.Name, *tmpl, ncs, true)
		if err != nil {
			fmt.Printf("CreateCluster-E-error AddNewVM: %s\n", err.Error())
			return "clusterName", err
		}
		err = task.WaitTaskCompletion()
		if err != nil { // fatal?
			fmt.Printf("CreateClusterE-error waiting for completion on AddNewVM: %s\n", err.Error())
		}

		vm, err := vapp.GetVMByName(vmparams.Name, true)
		if err != nil {
			return "clusterName", err
		}
		ncs, err = vm.GetNetworkConnectionSection()
		if err != nil {
			fmt.Printf("\n\tCreateCluster2-E-GetNetworkConnectionSection: %s\n", err.Error())
			return "", err
		}
		vmType := string(vmlayer.GetVmTypeForRole(string(vmparams.Role)))
		if vmparams.Role == vmlayer.RoleAgent {
			lbvm = vm
			// We'll eventually add an external net to vm, make internal net iface idx 1
			//netConIdx = 1
			vmIp = baseAddr //v.IncrIP(baseAddr, 1) // gateway for new cidr, also needs a new ext net addr XXX
			fmt.Printf("\n\tCreateCluster-I-adding vm role %s type %s with IP  %s\n", vmparams.Role, vmType, vmIp)

		} else {
			// Single Internal Net, and 101 should be 100 + workerNode index XXX
			//netConIdx = 0
			vmIp = v.IncrIP(baseAddr, 100+(n-1))
			fmt.Printf("\n\tCreateCluster2-I-creating docker node addr %s \n", vmIp)
			ncs.PrimaryNetworkConnectionIndex = 0
		}

		fmt.Printf("\n\nCreateCluster-I-adding vm #%d\n\tName: %s\n\t role %s\n\tType: %s\n\tInternalIP %s\n\t netname: %s\n\t idx: %d\n\n", n,
			vm.VM.Name, vmparams.Role, vmType, vmIp, internalNetName, netConIdx)

		cvm = ClusterVm{
			vmName:          vm.VM.Name,
			vmRole:          string(vmparams.Role),
			vmType:          vmType,
			vmFlavor:        vmparams.FlavorName,
			vmParentCluster: clusterName,
		}
		err = v.AddMetadataToVM(ctx, vm, vmparams, vmType, clusterName)
		if err != nil {
			fmt.Printf("\n\tCreateCluster-E-Error Adding metadata to vm: %s : %s\n\n", vm.VM.Name, err.Error())
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
			fmt.Printf("\n\tCreateCluster2-E-UpdateNetworkConnectionSection for %s failed : %s\n", vm.VM.Name, err.Error())
			return clusterName, err
		}
		subnet := ""
		err = v.updateVM(ctx, vm, vmparams, subnet)
	}
	// Ok, here, we have both cluster vm on the internal net. Now go back and
	// add the external to our LB
	extAddr := ""
	fmt.Printf("\n\tCreateCluster1-I-update vm %s\n", lbvm.VM.Name)
	// Hopefully, adding this second will make it's default route first
	extAddr, err = v.GetNextExtAddrForVdcNet(ctx, cloud.ParentVdc)
	if err != nil {
		fmt.Printf("\tError obtaining next EXT net addr %s\n", err.Error())
		return clusterName, fmt.Errorf("Agent node failed to obtain external net IP")
	}
	v.Objs.Cloudlet.ExtVMMap[extAddr] = lbvm
	cvm.vmIPs.ExternalIp = extAddr
	// make the external the primray index 0
	fmt.Printf("\nCreateCluster-I-set ext net addr as %s\n\n", extAddr)

	netName := cloud.ExtNet.OrgVDCNetwork.Name // just PrimaryNet
	err = v.AddExtNetToVm(ctx, lbvm, netName, extAddr)
	if err != nil {
		fmt.Printf("CreateCluster-E-AddExtNetToVm failed: %s\n", err.Error())
		return clusterName, err
	}

	fmt.Printf("CreateCluster-I-Add external net IP %s to vm %s OK\n", extAddr, lbvm.VM.Name)
	// Would be nice to have dedicated or shared here?
	task, err := lbvm.AddMetadata("ClusterVM", "true")
	if err != nil {
		fmt.Printf("\nError adding metadata ClusterVM true to lbvm: %s err: %s\n", lbvm.VM.Name, err.Error())
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
			fmt.Printf("CreateCluster-E-error powering on %s: %s\n", vm.VM.Name, err.Error())
			return clusterName, err
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			fmt.Printf("CreateVMs-E-error from wait : %s\n", err.Error())
			return clusterName, err
		}
		fmt.Printf("CreateCluster task complete %s power on...\n", vm.VM.Name)
	}

	return clusterName, nil
}

func (v *VcdPlatform) DeleteCluster(ctx context.Context, cloud MexCloudlet, vmMap *CidrMap) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "DeleteCluster")
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
				fmt.Printf("Found %s in cluster %s\n", vm.vmName, clust.Name)
				return clust.Name, &vm, nil
			}
		}
	}
	return "", nil, fmt.Errorf("VM not found")
}
