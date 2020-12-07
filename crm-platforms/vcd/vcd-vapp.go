package vcd

import (
	"context"
	"encoding/base64"
	"fmt"
	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"strings"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func (v *VcdPlatform) addVApp(ctx context.Context, vappName string, vdc *govcd.Vdc) error {
	// shouldn't we just pass it in here?

	fmt.Printf("addVapp-I-adding newly created vapp named: %s to our local cache\n", vappName)

	// check if it's already here?
	vapp, err := v.FindVApp(ctx, vappName)
	if err != nil {
		fmt.Printf("addVApp-W-%s exists is the same?\n", vappName)
		tmpVApp, err := vdc.GetVAppByName(vappName, true)
		if err != nil {
			fmt.Printf("addVApp-E-asking to add an unknown Vapp: %s\n", vappName)
			return err
		}
		// Now, check their HREFS for equality
		if vapp.VApp.HREF != tmpVApp.VApp.HREF {
			fmt.Printf("addVApp-E-two vapps same name, different apps %s vs %s\n",
				vapp.VApp.HREF, tmpVApp.VApp.HREF)
			return fmt.Errorf("Duplicate VApp Names found")
		}
	}
	// update
	VApp := VApp{
		VApp: vapp,
	}
	v.Objs.VApps[vapp.VApp.Name] = &VApp
	return nil
}

func (v *VcdPlatform) FindVApp(ctx context.Context, vappName string) (*govcd.VApp, error) {

	for name, vapp := range v.Objs.VApps {
		if vappName == name {
			return vapp.VApp, nil
		}
	}

	// Use GetAvailableQuery incase something new was created behind our backs.
	return nil, fmt.Errorf("Server does not exist")
}

func (v *VcdPlatform) populateINetConnectSection(ctx context.Context, vmparams *vmlayer.VMOrchestrationParams) (*types.NetworkConnectionSection, error) {
	// xlate orch -> netconnectsec
	//nc := &types.NetworkConnectionSection{}
	fmt.Printf("populateINetConnectSection-I-vm: %s role %s \n", vmparams.Name, vmparams.Role)
	//	vu.DumpOrchParamsVM(vmparams, 1)

	fmt.Printf("\n\npopulate Network connection section for VM %s role %s\n", vmparams.Name, vmparams.Role)
	//panic("popINetConnectionSection debug")
	//	return nc, nil
	return nil, nil
}

// Propteries make up the ProductSection and are retrievable in the VM
func makeProp(key, value string) *types.Property {
	prop := &types.Property{
		// We hard code UserConfigurable for now, as if false, it does not appear in the ovfenv fetched by vmtoolsd,
		// and what good is that to us?
		UserConfigurable: true,
		Type:             "string",
		Key:              key,
		Label:            key + "label",
		Value: &types.Value{
			Value: value,
		},
	}
	return prop
}

// pick off meta data to add to the VMs product section, where it will become available using
// vmware tools vmtoolsd --get for mobiledgex-init.sh
// Possible values used by our init script include:
//
// set_metadata_param HOSTNAME .name
//set_metadata_param UPDATE .meta.update
//set_metadata_param SKIPINIT .meta.skipinit
//set_metadata_param INTERFACE .meta.interface
//set_metadata_param ROLE .meta.role
//set_metadata_param SKIPK8S .meta.skipk8s
//set_metadata_param MASTERADDR .meta.k8smaster
//set_metadata_param UPDATEHOSTNAME .meta.updatehostname

//set_network_param IPADDR '.networks[0].ip_address'
//set_network_param NETMASK '.networks[0].netmask'
//set_network_param NETTYPE '.networks[0].type'

func vcdUserDataFormatter(instring string) string {
	// despite the use of paravirtualized drivers, vSphere gets get name sda, sdb
	fmt.Printf("\n\nvcdUserDataFormater received instring as: %s\n\n", instring)
	instring = strings.ReplaceAll(instring, "/dev/vd", "/dev/sd")
	return base64.StdEncoding.EncodeToString([]byte(instring))
}

func (v *VcdPlatform) populateProductSection(ctx context.Context, vm *govcd.VM, vmparams *vmlayer.VMOrchestrationParams) (*types.ProductSectionList, error) {

	//cldCfg := vmlayer.VMCloudConfigParams{}
	command := ""
	manifest := ""
	//	cloudParams := &vmlayer.VMCloudCnnfigParams{}
	// format vmparams.CloudConfigParams into yaml format, which we'll then base64 encode for the ovf datasource
	udata, err := vmlayer.GetVMUserData(vm.VM.Name, false, manifest, command, &vmparams.CloudConfigParams, vcdUserDataFormatter)
	if err != nil {
		fmt.Printf("GetVMUserData returns %s\n", err.Error())
		return nil, err
	}
	fmt.Printf("\npopProdSec-I-udata: %s \n", udata)
	// create userdata string?
	guestCustomSec, err := vm.GetGuestCustomizationSection()
	if err != nil {
		return nil, err
	}
	if !*guestCustomSec.Enabled {
		guestCustomSec.Enabled = vu.TakeBoolPointer(true)
		// guestCustomSec.AdminPassword = "2b|!2b-titq" // xxx
		// FixMe:  'AdminPassword' should either be reset or remain unchanged when auto"}
		//
		// vault kv get -field=value secret/accounts/baseimage/password
		gcs, err := vm.SetGuestCustomizationSection(guestCustomSec)
		if err != nil {
			fmt.Printf("popProdSec-E-SetGuestCustomizationSectionFailed: %s\n", err.Error())
			return nil, err

		}
		fmt.Printf("\nCustomSect enabled : %+v\n", gcs)
	}

	psl, err := vm.GetProductSectionList()
	if err != nil {
		return nil, err
	}
	if psl.ProductSection == nil {
		fmt.Printf("\n\t vm %s prod section nil, creating one\n\n", vm.VM.Name)
		psl = &types.ProductSectionList{
			ProductSection: &types.ProductSection{
				Info:     "Guest Properties",
				Property: []*types.Property{},
			},
		}
	}

	var props []*types.Property

	// manditory
	props = append(props, makeProp("instance-id", vm.VM.ID))
	// test user ubuntu getting my local ssh pub key

	if udata != "" {
		props = append(props, makeProp("user-data", udata))
	}

	log.SpanLog(ctx, log.DebugLevelInfra, "populateProdcutSection", "name", vmparams.Name, "role", vmparams.Role)
	role := vmparams.Role
	props = append(props, makeProp("ROLE", string(role)))
	skipk8s := vmlayer.SkipK8sYes
	if role == vmlayer.RoleMaster || role == vmlayer.RoleNode {
		skipk8s = vmlayer.SkipK8sNo
	}
	props = append(props, makeProp("SKIPK8S", string(skipk8s)))
	psl.ProductSection.Property = props
	// XXX what else?

	return psl, nil
}
