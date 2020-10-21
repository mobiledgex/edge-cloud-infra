package vcd

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"
)

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
