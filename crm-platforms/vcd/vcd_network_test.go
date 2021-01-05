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
		nextAddr, err := tv.GetNextExtAddrForVdcNet(ctx)
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
		nextAddr, err := tv.GetNextInternalNet(ctx)
		if err != nil {
			fmt.Printf("Error getting next addr  : %s\n", err.Error())
			return
		}
		/*
			fmt.Printf("Next ext-net Address: %s\n", nextAddr)
			cloud.Clusters[nextAddr] = &Cluster{}
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
		*/require.Equal(t, nextAddr, "10.101.2.1")

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
		vm, err := tv.FindVMByName(ctx, *vmName)
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
		_, err := tv.GetVdc(ctx)
		if err != nil {
			fmt.Printf("Error from GetVdc : %s\n", err.Error())
			return
		}

		// Expect our test begins with zero Cloudlets

		// We ask for the first external address in our PrimaryNet range
		_ /*CloudletAddr,*/, err = tv.GetNextExtAddrForVdcNet(ctx)
		if err != nil {
			fmt.Printf("GetNextExternalAddrForVdcNet failed: %s\n", err.Error())
			return
		}
	}
}

// -vdc -net  is some OrgVdcNetwork in vdc
func TestGetAllocatedIPs(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		org, err := tv.GetOrg(ctx)
		if err != nil {
			fmt.Printf("GetOrgs failed: %s\n", err.Error())
			return
		}
		vdc, err := org.GetVdcByName(*vdcName)

		if err != nil {
			fmt.Printf("vdc %s not found in org %s\n", *vdcName, org.Org.Name)
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

func TestDumpNet(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestNetworks\n")
		// monitor.go
		net, err := tv.GetExtNetwork(ctx)
		if err != nil {
			fmt.Printf("GetExtNetwork error; %s\n", err.Error())
			return
		}
		govcd.ShowNetwork(*net.OrgVDCNetwork)
	} else {
		return
	}
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

func TestGetExtNet(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {

		extNetMask := tv.GetExternalNetmask()
		fmt.Printf("extNetMask: %s\n", extNetMask)
		orgvdcNet, err := tv.GetExtNetwork(ctx)
		if err != nil {
			fmt.Printf("Error retrieving network object  err: %s\n", err.Error())
			return
		}
		fmt.Printf("Found network %s\n", orgvdcNet.OrgVDCNetwork.Name)
		govcd.ShowNetwork(*orgvdcNet.OrgVDCNetwork)
	}
}

// -net -live
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
