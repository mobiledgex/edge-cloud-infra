package vcd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"
)

// Relevant types.go:

/*
type ProductSection struct {
	Info     string      `xml:"Info,omitempty"`
	Property []*Property `xml:"http://schemas.dmtf.org/ovf/envelope/1 Property,omitempty"`
}

type Property struct {
	Key              string `xml:"http://schemas.dmtf.org/ovf/envelope/1 key,attr,omitempty"`
	Label            string `xml:"http://schemas.dmtf.org/ovf/envelope/1 Label,omitempty"`
	Description      string `xml:"http://schemas.dmtf.org/ovf/envelope/1 Description,omitempty"`
	DefaultValue     string `xml:"http://schemas.dmtf.org/ovf/envelope/1 value,attr"`
	Value            *Value `xml:"http://schemas.dmtf.org/ovf/envelope/1 Value,omitempty"`
	Type             string `xml:"http://schemas.dmtf.org/ovf/envelope/1 type,attr,omitempty"`
	UserConfigurable bool   `xml:"http://schemas.dmtf.org/ovf/envelope/1 userConfigurable,attr"`
}

type Value struct {
	Value string `xml:"http://schemas.dmtf.org/ovf/envelope/1 value,attr,omitempty"`
}
*/

func TestProdSec(t *testing.T) {

	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestProdSec:")

		// Test setting ProductSection as a means of guest customization
		// Create a test vm
		// Create some Properties, let's start with the VMRole of the node as an example
		// *Note: mobiledgex-init.sh wants values for
		// HOSTNAME, UPDATES, SKIPINIT, INTERFACE, ROLE, SKIPK8S, MASTERADDR, UPDATEHOSTNAME IPADDR, NETMASK, NETTYPE
		// And checks $VMWARE_CLOUDINIT, and if set, does vmtoolsd --cmd "info-get guestinfo.metadata";
		// Which vCD had that, but it does not. It's special and uses these properties.
		// So, lets see if we can set them all eventually via properties.
		// How the guest accesses them is another mater. TBD
		//
		// "guest custimization  use GuestCustomizationSection
		// pass runtime info to the vm, can't suppy at compose/recompose time, rather create the vapp
		// and then go to the vm's productSection to update the runtime info which  you want to pass int"
		// Uh huh... sure buddy, we'll see
		//
		vmRecs, err := tv.GetAvailableVMs(ctx) // or should we just create a vm named test?
		if err != nil {
			fmt.Printf("TestProcSec-E-GetAvalableVMs return error: %s\n", err.Error())
			return
		}

		fmt.Printf("Our current set of VM names:\n")
		for _, vm := range vmRecs {
			fmt.Printf("\t%s : %s \n", vm.Name, vm.HREF)
		}

		//cat := &govcd.Catalog{}
		targetVAppName := "clusterVapp1"

		targetVApp := &govcd.VApp{}

		targetVApp = tv.Objs.VApps[targetVAppName].VApp
		// The other object in VappsMap is a vmMap of all vms in this VApp, is it populated?
		vmMap := tv.Objs.VApps[targetVAppName].VMs
		fmt.Printf("len of vmMap for %s is %d\n", targetVAppName, len(vmMap))

		fmt.Printf("Test using vapp %s  = %+v\n", targetVAppName, targetVApp)
		targetVMName := "mex-plat"

		targetVM := &govcd.VM{}

		targetVM, err = targetVApp.GetVMByName(targetVMName, false)
		if err != nil {
			fmt.Printf("Error getting vm by name: %s using  vapp: %s\n", targetVMName, targetVApp.VApp.Name)
		}

		vmProperties := &types.ProductSectionList{
			ProductSection: &types.ProductSection{
				Info:     "Guest Properties",
				Property: []*types.Property{},
			},
		}

		// this works, just make UserConfigurable true, since if you set this to false,
		// you'll see the key in the ovfenv, but value will be "" for some reason.
		// In the vm, use vmtoolsd --cmd "get-info guestinfo.ovfenv"
		// and find the <PropertySection> </PropertySecion>
		// I've read indications that you can create your own "named sections" for catagories of properties,
		// maybe find time to play with that for init params vs env vars for instance...
		//
		prop := createProp("user-data", "encoded", true)
		vmProperties.ProductSection.Property = append(vmProperties.ProductSection.Property, prop)
		prop = createProp("ROLE", "platform", true)
		vmProperties.ProductSection.Property = append(vmProperties.ProductSection.Property, prop)
		// how about setting env vars in the host?
		prop = createProp("MASTERADDR", "10.1.101.10", true) // XXX testing
		vmProperties.ProductSection.Property = append(vmProperties.ProductSection.Property, prop)

		_, err = targetVM.SetProductSectionList(vmProperties)
		if err != nil {
			fmt.Printf("error Setting guest properties: %s", err)
			return
		}
		section, err := targetVM.SetProductSectionList(vmProperties)
		if err != nil {
			fmt.Printf("SetProductSectionList failed: %s\n", err.Error())
			return
		}
		fmt.Printf("section returned : %+v\n", section)

	} else {
		return
	}
}

func createProp(key, value string, config bool) *types.Property {
	prop := &types.Property{
		UserConfigurable: config,
		Type:             "string",
		Key:              key,
		Label:            key + "label",
		Value: &types.Value{
			Value: value,
		},
	}
	return prop
}
