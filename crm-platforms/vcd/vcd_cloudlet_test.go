package vcd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	//"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// -vdc

// XXX restarted crm rebuilds curstate.
func TestRebuildCloudlet(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		cname := ""
		// Get our vapp, that's the cloudlet
		vapps := tv.Objs.VApps
		if len(vapps) == 0 {
			fmt.Printf("No VApp/Cloudlet exists, nothing to rebuild\n")
			return
		}
		cloudlet := &MexCloudlet{}
		var clusters []*Cluster
		for name, vapp := range vapps {
			fmt.Printf("consider vapp %s\n", name)
			// If we have a vapp, we need to create a cloudlet from it

			cloudlet, cname, err = testBuildCloudlet(ctx, vapp.VApp)

			// Next need to recreate our clusters, run though our vm map and build the ClusterVMs
			// map and ExtVMMaps
			clusters, err = testBuildClusters(ctx, cname, *cloudlet)
		}
		fmt.Printf("Cloudlet %s has %d clusters:\n", cloudlet.CloudletName, len(clusters))

	}

}

func testBuildCloudlet(ctx context.Context, vapp *govcd.VApp) (*MexCloudlet, string, error) {

	cname := ""
	cloudlet := &MexCloudlet{}

	cloudlet.CloudletName = vapp.VApp.Name
	cloudlet.ParentVdc = tv.Objs.Vdc
	cloudlet.CloudVapp = vapp
	cloudlet.ExtNet = tv.Objs.PrimaryNet

	mdata, err := vapp.GetMetadata()
	if err != nil {
		fmt.Printf("\n\nError getting vapp metadata for vapp: %s\n", vapp.VApp.Name)
		return nil, "", err
	}
	for _, data := range mdata.MetadataEntry {
		if data.Key == "CloudletName" {
			fmt.Printf("Vapp %s has metadata for CloudletName: %s\n", vapp.VApp.Name, data.TypedValue.Value)
			cname = data.TypedValue.Value
		}
	}
	extAddr, err := tv.GetExtAddrOfVapp(ctx, vapp, cloudlet.ExtNet.OrgVDCNetwork.Name)
	if err != nil {
		fmt.Printf("testBuildCloudlet-E-getting ext addr of vapp: %s err: %s\n",
			vapp.VApp.Name, err.Error())
		return nil, "", err
	}
	cloudlet.ExtIp = extAddr

	return cloudlet, cname, nil
}

func testBuildClusters(ctx context.Context, cldName string, cloudlet MexCloudlet) ([]*Cluster, error) {

	// cldName needs to be from metadata or split o ut

	// Consider name formats:
	// cld3.tdg.mobiledgex.net              -- the cloudlet  (only external nic)
	// clust1.cld3.tdg.mobiledgex.net       -- a cluster-rootLB (both external and internal nic)
	// mex-docker-vm-cld3-clust1-mobiledgex -- a worker node (internal nic only)

	clusterMap := make(CidrMap)
	fmt.Printf("test rebuilding clusters for cloudlet %s\n", cldName)
	var waitingVMs []*govcd.VM

	// need a getClusterNameFromVMName() which should just be the first element of vmName
	for vmName, vm := range tv.Objs.VMs {
		// first, check if this vm is the cloudlet's vm

		targetCluster := ""
		parts := strings.Split(vmName, ".")

		if len(parts) == 1 { // some internal nic only worker node
			bits := strings.Split(vmName, "-")
			if len(bits) > 1 {
				// fuzzier, find our cldName, and cluster should be +1
				for n, component := range bits {
					if string(component) == cldName {
						targetCluster = bits[n+1]
						fmt.Printf("Vm %s belongs to cluster : %s\n", vmName, targetCluster)
						break
					}
				}
				if clusterMap[targetCluster] == nil {
					// haven't found the clusterNode yet save
					waitingVMs = append(waitingVMs, vm)
				}
			}

		} else {
			fmt.Printf("\tmulti-part split by . len: %d name: %s\n", len(parts), vmName)

			if string(parts[0]) == cldName {
				// this is the cloudlet itself, skip it
				fmt.Printf("Skipping VM %s it's the cloudlet %s\n", vmName, cldName)
				continue
			}
			// set this guys external addr as key in our cluster map
			targetCluster = parts[1]
			extAddr, err := tv.GetExtAddrOfVM(ctx, vm, tv.Objs.PrimaryNet.OrgVDCNetwork.Name)
			if err != nil {
				fmt.Printf("Error getting external addr of  vm %s : %s\n", vmName, err.Error())
			}
			fmt.Printf("LB vm %s is the targetCluster vm extAddr: %s\n", vmName, extAddr)
			clusterMap[extAddr] = &Cluster{
				Name: vmName,
			}

			c := clusterMap[extAddr]
			c.VMs = make(VMIPsMap)

			tv.Objs.Cloudlet.ExtVMMap[extAddr] = vm
			for _, waiter := range waitingVMs {
				// vms in the cluster are cvms key'ed by vmName eh?
				// ok, here's where we cached some metadata in this vm
				// for role, type, vmMeta, vmIPs, ParentCluster (wait what? yeah!)
				c.VMs[waiter.VM.Name] = ClusterVm{
					vmName: waiter.VM.Name,
					// pull metadata from this VM and fill this in. XXX
				}
			}
		}

		fmt.Printf("next vm name: %s belongs in cluster %s \n", vmName, targetCluster)
		// Create ClusterVM but we've lost the vmparams hmm...
	}
	// Run our list of tv.Objs.VMs, all must have

	return nil, nil
}
