package vcd

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"
)

func TestNetAddrs(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	fmt.Printf("InitVcdTestEvn complete\n")

	testaddr := "10.101.5.10"
	N, err := tv.ThirdOctet(ctx, testaddr)
	require.Nil(t, err, "ThrirdOctet err")
	require.Equal(t, 5, N, "ThirdOctet")

	testaddr = "10.101.6.10/24"
	N, err = tv.ThirdOctet(ctx, testaddr)
	require.Nil(t, err, "ThrirdOctet err")
	require.Equal(t, 6, N, "ThirdOctet")

	if live {
		// how many vdcs do we have?
		// we should now have two

		fmt.Printf("TestNetAddrs-I-we have %d vdcs\n", len(tv.Objs.Vdcs))
		for name, _ := range tv.Objs.Vdcs {
			fmt.Printf("\t %s\n", name)
		}

		fmt.Printf("looking for vdc %s\n", *vdcName)
		// First, test case 1 no current entries
		vdc, err := tv.FindVdc(ctx, *vdcName)
		if err != nil {
			fmt.Printf("Vdc %s not found\n", *vdcName)
			return
		}

		fmt.Printf("Have vdc: %s\n", vdc.Vdc.Name)
		numClouds := len(tv.Objs.Cloudlets)

		// Expect our test begins with zero Cloudlets
		if numClouds == 0 {

			fmt.Printf("non-nominal zero cloudlet test in vdc %s\n", *vdcName)
			nextCidr, err := tv.GetNextInternalNet(ctx, nil)
			if err != nil {
				fmt.Printf("GetNextInternalNet return err: %s\n", err.Error())
				return
			}
			// expect 10.101.1.0/24
			fmt.Printf("GetNextInternalNet expect 10.101.1.0/24 cidr: %s\n", nextCidr)

		}
		// We still have no cloudlets, create one.

		cluster1 := make(CidrMap)
		cluster2 := make(CidrMap)

		vmIpMap1 := make(VMIPsMap)
		vmIpMap2 := make(VMIPsMap)

		vmMap1 := VmNet{
			vmName: "testVM1",
			vmRole: "roleAgent",
			//vmMeta: "",
		}
		vmMap2 := VmNet{
			vmName: "testVM2",
			vmRole: "roleNode",
			//vmMeta: "",
		}

		vmIpMap1["10.101.1.1"] = vmMap1
		vmIpMap2["10.101.2.1"] = vmMap2

		cluster1["10.101.1.0/24"] = vmIpMap1
		cluster2["10.101.2.0/24"] = vmIpMap2

		fmt.Printf("Nominal single Cloudlet test\n")
		vapp := &govcd.VApp{}
		// Create a test cloudlet obj

		tv.Objs.Cloudlets["testCloudlet1"] = &MexCloudlet{
			ParentVdc: vdc,
			CloudVapp: vapp,
			Clusters:  cluster1, // cluster right?
		}

		for _, cloud := range tv.Objs.Cloudlets {
			fmt.Printf("looking for vdc %s\n", *vdcName)
			if cloud.ParentVdc.Vdc.Name == *vdcName {
				fmt.Printf("vdc %s mapped to cloudlet %s\n", vdc.Vdc.Name, cloud.CloudletName)
				nextCidr, err := tv.GetNextInternalNet(ctx, cloud)
				if err != nil {
					fmt.Printf("GetNextInternalNet err: %s\n", err.Error())
					return
				}
				// first created should be 10.101.1.0/24
				fmt.Printf("next Cidr: %s\n", nextCidr)
			}

		}
		cli := tv.Client.Client
		vdc2 := govcd.NewVdc(&cli)

		vdc2.Vdc.Name = "testvdc2"

		tv.Objs.Cloudlets["testCloudlet2"] = &MexCloudlet{
			ParentVdc:    vdc,
			CloudVapp:    vapp,
			CloudletName: "testCloudlet2",
			Clusters:     cluster2,
		}
		cloud := tv.Objs.Cloudlets["testCloudlet2"]
		fmt.Printf("Text Case 3 seccond vdc/cloudlet\n")
		nextCidr, err := tv.GetNextInternalNet(ctx, cloud)
		if err != nil {
			fmt.Printf("GetNextInternalNet failed: %s\n", err.Error())
			return
		}
		fmt.Printf("next cider : %s\n", nextCidr)

		vdc3 := govcd.NewVdc(&cli)

		vdc3.Vdc.Name = "testvdc2"

		tv.Objs.Cloudlets["testCloudlet3"] = &MexCloudlet{
			ParentVdc:    vdc,
			CloudVapp:    vapp,
			CloudletName: "testCloudlet2",
			Clusters:     cluster2,
		}
		cloud = tv.Objs.Cloudlets["testCloudlet3"]
		fmt.Printf("Text Case 3 seccond vdc/cloudlet\n")
		nextCidr, err := tv.GetNextInternalNet(ctx, cloud)
		if err != nil {
			fmt.Printf("GetNextInternalNet failed: %s\n", err.Error())
			return
		}
		fmt.Printf("next cider : %s\n", nextCidr)

		// next, create a hole in our cidrs and ensure we fill it with the next new entry
		// TBI
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
