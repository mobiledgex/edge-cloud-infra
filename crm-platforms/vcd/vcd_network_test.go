package vcd

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"
)

// -vdc
func TestNextExtAddr(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		vdc, err := tv.GetVdc(ctx, *vdcName)
		if err != nil {
			fmt.Printf("Error getting vdc %s : %s\n", *vdcName, err.Error())
			return
		}
		nextAddr, err := tv.GetNextExtAddrForVdcNet(ctx, vdc)
		if err != nil {
			fmt.Printf("Error getting next addr  : %s\n", err.Error())
			return
		}

		fmt.Printf("Next ext-net Address: %s\n", nextAddr)
	}
}

func TestNextIntAddr(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		cloud := &MexCloudlet{}
		if tv.Objs.Cloudlet == nil {
			fmt.Printf("Make Test Cloudlet\n")
			//cmap := make(CidrMap)
			tv.Objs.Cloudlet = cloud
			cloud.Clusters = make(CidrMap)
			cloud.Clusters["10.101.1.1"] = Cluster{}
		}
		nextAddr, err := tv.GetNextInternalNet(ctx)
		if err != nil {
			fmt.Printf("Error getting next addr  : %s\n", err.Error())
			return
		}

		fmt.Printf("Next ext-net Address: %s\n", nextAddr)
		cloud.Clusters[nextAddr] = Cluster{}
		nextAddr, err = tv.GetNextInternalNet(ctx)
		if err != nil {
			fmt.Printf("Error getting next addr  : %s\n", err.Error())
			return
		}
		fmt.Printf("Next ext-net Address: %s\n", nextAddr)
		delete(cloud.Clusters, "10.101.2.1")
		nextAddr, err = tv.GetNextInternalNet(ctx)
		if err != nil {
			fmt.Printf("Error getting next addr  : %s\n", err.Error())
			return
		}
		require.Equal(t, nextAddr, "10.101.2.1")

	}
}

// -vapp -net Return the current external net
func TestGetVappAddr(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		vapp, err := tv.FindVApp(ctx, *vappName)
		if err != nil {
			fmt.Printf("Test error finding vapp %s\n", *vappName)
			return
		}

		addr, err := tv.GetExtAddrOfVapp(ctx, vapp, *netName)
		if err != nil {
			fmt.Printf("Test error from GetExtAddrOfVapp  %s = %s \n", *vappName, err.Error())
			return
		}

		fmt.Printf("Next ext-net Address: %s\n", addr)
	}
}

func TestGetVMAddr(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		vm, err := tv.FindVM(ctx, *vmName)
		if err != nil {
			fmt.Printf("error finding vm %s\n", *vappName)
			return
		}

		addr, err := tv.GetExtAddrOfVM(ctx, vm, *netName)
		if err != nil {
			fmt.Printf("Test error from GetExtAddrOfVM  %s = %s \n", *vmName, err.Error())
			return
		}

		fmt.Printf("Next ext-net Address: %s\n", addr)
	}
}

func TestNetAddrs(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	testaddr := "10.101.5.10"
	N, err := tv.Octet(ctx, testaddr, 2) // third octet)
	require.Nil(t, err, "ThrirdOctet err")
	require.Equal(t, 5, N, "ThirdOctet")

	testaddr = "10.101.6.10/24"
	N, err = tv.Octet(ctx, testaddr, 2)
	require.Nil(t, err, "ThrirdOctet err")
	require.Equal(t, 6, N, "ThirdOctet")

	if live {
		vdc := tv.Objs.Vdc

		// Expect our test begins with zero Cloudlets
		if tv.Objs.Cloudlet == nil {

			fmt.Printf("non-nominal zero cloudlet test in vdc")
			nextCidr, err := tv.GetNextInternalNet(ctx)
			if err != nil {
				fmt.Printf("GetNextInternalNet return err: %s\n", err.Error())
				return
			}
			// expect 10.101.1.0/24
			fmt.Printf("GetNextInternalNet expect 10.101.1.0/24 cidr: %s\n", nextCidr)

		}
		// We ask for the first external address in our PrimaryNet range
		CloudletAddr, err := tv.GetNextExtAddrForVdcNet(ctx, vdc)
		if err != nil {
			fmt.Printf("GetNextExternalAddrForVdcNet failed: %s\n", err.Error())
			return
		}

		vapp := &govcd.VApp{}

		fmt.Printf("Cloudlet Cider: %s\n", CloudletAddr)
		tv.Objs.Cloudlet = &MexCloudlet{
			ParentVdc:    vdc,
			CloudVapp:    vapp,
			CloudletName: "testCloudlet1",
			ExtIp:        CloudletAddr,
		}
		tv.Objs.Cloudlet.Clusters = make(CidrMap)
		// ExtVMMap represents all vms in the cloudlet assigned an external IP address
		tv.Objs.Cloudlet.ExtVMMap = make(CloudVMsMap)
		// Cloudlet's externa,l addr is our vapp
		// We still have no  clusters, get the next external Cidr for it

		cluster1Addr, err := tv.GetNextExtAddrForVdcNet(ctx, vdc)
		if err != nil {
			fmt.Printf("GetNextExternalAddrForVdcNet failed: %s\n", err.Error())
			return

		}
		cluster1 := Cluster{
			Name: "cluster1",
			VMs:  make(VMIPsMap),
		}

		tv.Objs.Cloudlet.ExtVMMap[cluster1Addr] = &govcd.VM{}
		tv.Objs.Cloudlet.Clusters[cluster1Addr] = cluster1
		fmt.Printf("Cluster1 received addr: %s\n", cluster1Addr)

		cluster2Addr, err := tv.GetNextExtAddrForVdcNet(ctx, vdc)
		if err != nil {
			fmt.Printf("GetNextExternalAddrForVdcNet failed: %s\n", err.Error())
			return
		}
		cluster2 := Cluster{
			Name: "cluster2",
			VMs:  make(VMIPsMap),
		}
		tv.Objs.Cloudlet.ExtVMMap[cluster2Addr] = &govcd.VM{}
		tv.Objs.Cloudlet.Clusters[cluster2Addr] = cluster2
		fmt.Printf("Cluster2 received addr: %s\n", cluster2Addr)

		cluster3Addr, err := tv.GetNextExtAddrForVdcNet(ctx, vdc)
		if err != nil {
			fmt.Printf("GetNextExternalAddrForVdcNet failed: %s\n", err.Error())
			return
		}
		cluster3 := Cluster{
			Name: "cluster3",
			VMs:  make(VMIPsMap),
		}
		tv.Objs.Cloudlet.ExtVMMap[cluster3Addr] = &govcd.VM{}
		tv.Objs.Cloudlet.Clusters[cluster3Addr] = cluster3
		fmt.Printf("Cluster3 received addr: %s\n", cluster3Addr)

		// Now delete 2, and create 4, should get what 2 had
		delete(tv.Objs.Cloudlet.ExtVMMap, cluster2Addr)
		delete(tv.Objs.Cloudlet.Clusters, cluster2Addr)

		cluster4Addr, err := tv.GetNextExtAddrForVdcNet(ctx, vdc)
		if err != nil {
			fmt.Printf("GetNextExternalAddrForVdcNet failed: %s\n", err.Error())
			return
		}
		if cluster4Addr != cluster2Addr {
			fmt.Printf("FAIL Cluster4addr %s vs Cluster2Addr: %s\n", cluster4Addr, cluster2Addr)
			return
		}
		// Next internal net tests for the vms

		//vmIpMap1 := make(VMIPsMap)
		//vmIpMap2 := make(VMIPsMap)
		/*
			cvm1 := ClusterVm{
				vmName: "testVM1",
				vmRole: "roleAgent",
				//vmMeta: "",
			}
			cvm2 := ClusterVm{
				vmName: "testVM2",
				vmRole: "roleNode",
				//vmMeta: "",
			}
			fmt.Printf("Nominal single Cloudlet test\n")
			// Create a test cloudlet obj

			cli := tv.Client.Client
			vdc2 := govcd.NewVdc(&cli)

			vdc2.Vdc.Name = "testvdc2"

			fmt.Printf("Text Case 3 seccond vdc/cloudlet\n")
			nextCidr, err := tv.GetNextInternalNet(ctx)
			if err != nil {
				fmt.Printf("GetNextInternalNet failed: %s\n", err.Error())
				return
			}
			fmt.Printf("next cider : %s\n", nextCidr)

			vdc3 := govcd.NewVdc(&cli)

			vdc3.Vdc.Name = "testvdc2"

			fmt.Printf("Text Case 3 seccond vdc/cloudlet\n")
			nextCidr, err = tv.GetNextInternalNet(ctx)
			if err != nil {
				fmt.Printf("GetNextInternalNet failed: %s\n", err.Error())
				return
			}
			fmt.Printf("next cider : %s\n", nextCidr)

			// next, create a hole in our cidrs and ensure we fill it with the next new entry
			// TBI
		*/
	}

}

// -vdc -net  is some OrgVdcNetwork in vdc
func TestGetAllocatedIPs(t *testing.T) {
	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		vdc, err := tv.Objs.Org.GetVdcByName(*vdcName)
		if err != nil {
			fmt.Printf("vdc %s not found in org %s\n", *vdcName, tv.Objs.Org.Org.Name)
			return
		}
		vdcnet, err := vdc.GetOrgVdcNetworkByName(*netName, false)
		if err != nil {
			fmt.Printf("net %s not found in vdc %s\n", *netName, *vdcName)
			return
		}
		// look in IPScope for AllocatedIPAddresses *IPAddresses
		// vdcnet.Configuration.IPScopes.
		addrs := &types.IPAddresses{}

		IPScopes := vdcnet.OrgVDCNetwork.Configuration.IPScopes
		for _, ipscope := range IPScopes.IPScope {

			addrs = ipscope.AllocatedIPAddresses
			fmt.Printf("net %s addrs: ===>>%s<<===\n", *netName, addrs)

			/*
				fmt.Printf("Net  %s has these allocated IP addresses\n", *netName)
				for _, address := range ipscope.AllocatedIPAddresses.IPAddress {
					fmt.Printf("\t%d %+v\n", n, address)
				}
			*/

		}

	}
}
func TestGetNetList(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		nets, err := tv.GetNetworkList(ctx)
		require.Nil(t, err, "GetNetworkList")
		for n, net := range nets {
			fmt.Printf("%d : %s\n", n, net)
		}
	} else {
		return
	}
}

func TestDumpNet(t *testing.T) {
	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestNetworks\n")
		// monitor.go
		govcd.ShowNetwork(*tv.Objs.PrimaryNet.OrgVDCNetwork)
	} else {
		return
	}
}

func getNetByName(t *testing.T, ctx context.Context, netName string) (*govcd.OrgVDCNetwork, error) {

	if netName == "" {
		return nil, fmt.Errorf("Nil netName encountered")
	}
	for name, net := range tv.Objs.Nets {
		if name == netName {
			return net, nil
		}
	}
	return nil, fmt.Errorf("net %s not found", netName)
}

// only operate on vapp networks, not OrgVDCNetwworks
func TestRMNet(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestRmNetwork\n")
		_, err = testDeleteVAppNetwork(t, ctx, *vappName, *netName)
		if err != nil {
			fmt.Printf("error from testDelete VAppNetwork: %s\n", err.Error())
		}
	} else {
		return
	}

}

func TestNets(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestNetworks\n")
		net, err := getNetByName(t, ctx, *netName)
		if err != nil {
			fmt.Printf("Error getting net %s\n", *netName)
			return
		}
		// ok, let's see if we can enable FirewallService on netName Should be a vapp network.
		netconfig := net.OrgVDCNetwork.Configuration
		gatewayFeatures := net.OrgVDCNetwork.ServiceConfig

		//govcd.ShowNetwork(*net.OrgVDCNetwork)

		fmt.Printf("netconfig %+v\n", netconfig)
		fmt.Printf("gatewayFeatures %+v\n", gatewayFeatures)

		// Ok, so grab our Vapp, and create a new VApp
		// VappNetworkSettings, and enable the firewall service in the config, and call
		// vapp.UpdateNetwork( newsettings, orgvdcnetwork)
		vappName := "clusterVapp1"
		vapp, err := tv.FindVApp(ctx, vappName)
		if err != nil {
			fmt.Printf("Error getting %s : %s\n", vappName, err.Error())
			return
		}
		netID := net.OrgVDCNetwork.ID
		vappnet, err := vapp.GetVappNetworkById(netID, false)
		if err != nil {
			fmt.Printf("Error from GetVappNetworkById : %s\n", err.Error())
			return
		}
		fmt.Printf("result vapp net: %+v\n", vappnet)

		protocols := types.FirewallRuleProtocols{
			Any: true,
		}
		var rules []*types.FirewallRule

		// allow our host
		rule1 := types.FirewallRule{
			SourceIP:             "73.252.170.111",
			Policy:               "allow",
			Protocols:            &protocols,
			DestinationPortRange: "22",
			IcmpSubType:          "any",
			IsEnabled:            true, // default
		}
		// drop everything else
		rules = append(rules, &rule1)
		rule2 := types.FirewallRule{
			SourceIP: "Any",
			Policy:   "drop",
		}
		rules = append(rules, &rule2)
		// this drop = defaultAction, we have explictly defined each.
		vappNetwork, err := vapp.UpdateNetworkFirewallRules(netID, rules, true, "drop", false)
		if err != nil {
			fmt.Printf("Error UpdateNetworkFirewalRules: %s\n", err.Error())
			return
		}
		fmt.Printf("Result vappnetwork: \n %+v\n", vappNetwork)
	} else {
		return
	}
}

// what can we enable or not?
func testEnableVDCNetFirewall(t *testing.T, ctx context.Context, netName string) error {

	return nil
}

func testEnableVAppFirewall(t *testing.T, ctx context.Context, vappName string) error {

	return nil
}

func testDeleteVAppNetwork(t *testing.T, ctx context.Context, vappName, networkName string) (*types.NetworkConfigSection, error) {

	vapp, err := tv.FindVApp(ctx, vappName)
	if err != nil {
		fmt.Printf("Error finding Vapp: %s : %s \n", vappName, err.Error())
		return nil, err
	}
	netConfig, err := vapp.RemoveNetwork(networkName)
	if err != nil {
		fmt.Printf("Error from RemoveNetwork %s\n", err.Error())
		return nil, err
	}
	return netConfig, nil
}
