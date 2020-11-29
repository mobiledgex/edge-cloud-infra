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

// Should return an object of MexCluster type, not a cidr string
func (v *VcdPlatform) CreateCluster(ctx context.Context, cloud *MexCloudlet, tmpl *govcd.VAppTemplate, vmgp *vmlayer.VMGroupOrchestrationParams, updateCallback edgeproto.CacheUpdateCallback) (string, error) {

	if tmpl == nil {
		fmt.Printf("\nCreateCluster-E-can't create Cluster without a template = %+v\n", tmpl)
		return "", fmt.Errorf("template nil")
	}
	clusterName := vmgp.VMs[0].Name

	log.SpanLog(ctx, log.DebugLevelInfra, "CreateCluster", "create", clusterName, "cloudlet", cloud.CloudletName)

	nextCidr, err := v.GetNextInternalNet(ctx)
	if err != nil {
		fmt.Printf("GetNextInternalNet failed: %s\n", err.Error())
		return "", err
	}

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
	// Ok, set the network in the AddNewVM or wait and add it later?
	//numvms := len(vmgp.VMs)
	netConIdx := 0
	for n, vmparams := range vmgp.VMs {
		//		powered_on := false
		ncs := &types.NetworkConnectionSection{}
		tmpl.VAppTemplate.Name = vmparams.Name
		fmt.Printf("\n\tCreateCluster2-I-adding new vm %s #%d \n", vmparams.Name, n)

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

		if vmparams.Role == vmlayer.RoleAgent {
			lbvm = vm
			// We'll eventually add an external net to vm, make internal net iface idx 1
			//netConIdx = 1
			vmIp = baseAddr //v.IncrIP(baseAddr, 1) // gateway for new cidr, also needs a new ext net addr XXX
			fmt.Printf("\n\tCreateCluster2-I-adding vm role %s with IP  %s\n", vmparams.Role, vmIp)

		} else {
			// Single Internal Net, and 101 should be 100 + workerNode index XXX
			//netConIdx = 0
			vmIp = v.IncrIP(baseAddr, 100+n)
			fmt.Printf("\n\tCreateCluster2-I-creating docker node addr %s \n", vmIp)
			ncs.PrimaryNetworkConnectionIndex = 0
		}
		fmt.Printf("\n\tCreateCluster2-I-adding vm #%d Name: %s  role %s with IP  %s netname: %s idx: %d\n", n,
			vm.VM.Name, vmparams.Role, vmIp, internalNetName, netConIdx)

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
	fmt.Printf("\n\tCreateCluster2-I-update vm %s\n", lbvm.VM.Name)
	// Hopefully, adding this second will make it's default route first
	extAddr, err = v.GetNextExtAddrForVdcNet(ctx, cloud.ParentVdc)
	if err != nil {
		fmt.Printf("\tError obtaining next EXT net addr %s\n", err.Error())
		return clusterName, fmt.Errorf("Agent node failed to obtain external net IP")
	}
	v.Objs.Cloudlet.ExtVMMap[extAddr] = lbvm
	// make the external the primray index 0
	fmt.Printf("\nCreateCluster-I-set ext net addr as %s\n\n", extAddr)

	netName := cloud.ExtNet.OrgVDCNetwork.Name // just PrimaryNet
	err = v.AddExtNetToVm(ctx, lbvm, netName, extAddr)
	if err != nil {
		fmt.Printf("CreateCluster-E-AddExtNetToVm failed: %s\n", err.Error())
		return clusterName, err
	}

	//netConIdx = 1 // for our subsequent internal net
	fmt.Printf("CreateCluster-I-Add external net IP %s to vm %s OK\n", extAddr, lbvm.VM.Name)
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
