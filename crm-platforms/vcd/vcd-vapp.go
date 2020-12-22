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

	vapp, err := v.FindVApp(ctx, vappName)
	if err != nil {
		tmpVApp, err := vdc.GetVAppByName(vappName, true)
		if err != nil {
			return err
		}
		// Now, check their HREFS for equality
		if vapp.VApp.HREF != tmpVApp.VApp.HREF {
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

// Propteries make up the ProductSection and are retrievable in the VM
func makeProp(key, value string) *types.Property {
	prop := &types.Property{
		// We hard code UserConfigurable for now, as if false, it does not appear in the ovfenv fetched by vmtoolsd,
		// and what good is that to us?
		UserConfigurable: true,
		Type:             "string",
		Key:              key,
		Label:            key + "-label",
		Value: &types.Value{
			Value: value,
		},
	}
	return prop
}

func vcdUserDataFormatter(instring string) string {
	instring = strings.ReplaceAll(instring, "/dev/vd", "/dev/sd")
	return base64.StdEncoding.EncodeToString([]byte(instring))
}

func (v *VcdPlatform) populateProductSection(ctx context.Context, vm *govcd.VM, vmparams *vmlayer.VMOrchestrationParams) (*types.ProductSectionList, error) {

	command := ""
	manifest := ""
	// format vmparams.CloudConfigParams into yaml format, which we'll then base64 encode for the ovf datasource
	udata, err := vmlayer.GetVMUserData(vm.VM.Name, false, manifest, command, &vmparams.CloudConfigParams, vcdUserDataFormatter)
	if err != nil {
		return nil, err
	}
	guestCustomSec, err := vm.GetGuestCustomizationSection()
	if err != nil {
		return nil, err
	}
	if !*guestCustomSec.Enabled {
		guestCustomSec.Enabled = vu.TakeBoolPointer(true)
		// FixMe:  'AdminPassword' should either be reset or remain unchanged when auto"}
		// vault kv get -field=value secret/accounts/baseimage/password
		_, err := vm.SetGuestCustomizationSection(guestCustomSec)
		if err != nil {
			return nil, err

		}
	}

	psl, err := vm.GetProductSectionList()
	if err != nil {
		return nil, err
	}
	if psl.ProductSection == nil {
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
