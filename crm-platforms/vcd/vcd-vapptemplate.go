package vcd

import (
	"context"
	"fmt"
	vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// vapptemplate related operations

/*
 InstanciationParams represent each individual subsection of a standard OVF ffile.

type InstantiationParams struct {
	CustomizationSection         *CustomizationSection         `xml:"CustomizationSection,omitempty"`
	DefaultStorageProfileSection *DefaultStorageProfileSection `xml:"DefaultStorageProfileSection,omitempty"`
	GuestCustomizationSection    *GuestCustomizationSection    `xml:"GuestCustomizationSection,omitempty"`
	LeaseSettingsSection         *LeaseSettingsSection         `xml:"LeaseSettingsSection,omitempty"`
	NetworkConfigSection         *NetworkConfigSection         `xml:"NetworkConfigSection,omitempty"`
	NetworkConnectionSection     *NetworkConnectionSection     `xml:"NetworkConnectionSection,omitempty"`
	ProductSection               *ProductSection               `xml:"ProductSection,omitempty"`
	// TODO: Not Implemented
	// SnapshotSection              SnapshotSection              `xml:"SnapshotSection,omitempty"`
}
*/

func (v *VcdPlatform) FindTemplate(ctx context.Context, tmplName string) (*govcd.VAppTemplate, error) {

	for _, tmpl := range v.Objs.VAppTmpls {
		if tmpl.VAppTemplate.Name == tmplName {
			return tmpl, nil
		}
	}
	return nil, fmt.Errorf("template %s not found", tmplName)

}

// Return all templates found in our catalog(s)
func (v *VcdPlatform) GetAllVdcTemplates(ctx context.Context, cat *govcd.Catalog) ([]*govcd.VAppTemplate, error) {

	var tmpls []*govcd.VAppTemplate

	queryRes, err := v.Objs.Vdc.QueryVappTemplateList()
	if err != nil {
		fmt.Printf("QueryVappTemplateList error : %s\n", err.Error())
		return tmpls, err
	}
	for n, res := range queryRes {
		fmt.Printf("\t#%d Lookup res.Name: %sby HREF: %s\n", n, res.Name, res.HREF)

		tmpl, err := cat.GetVappTemplateByHref(res.HREF)
		// tmpl, err := cat.GetVappTemplateByBName(res.Name)
		if err != nil {
			// This can happen if we have objects using the same names?
			fmt.Printf("\tError from GetVappTemplateByHref for %s as %s Skipping\n", res.Name, res.HREF)
			continue
		} else {
			tmpls = append(tmpls, tmpl)
			fmt.Printf("\tAdded template %s to templs\n", res.Name)
		}
	}

	return tmpls, nil

}

// Return the list of VMs contained in this template
// This impl fails. But they have a new call, use that here
// 		if tmpl.Type == "application/vnd.vmware.vcloud.vm+xml" {
func (v *VcdPlatform) GetVmsFromTemplate(ctx context.Context, catName string, tmpl *govcd.VAppTemplate) ([]*govcd.VM, error) {

	var vms []*govcd.VM
	tmplName := tmpl.VAppTemplate.Name
	children := tmpl.VAppTemplate.Children.VM // VAppTemplateChildren

	for _, child := range children {
		fmt.Printf("GetVmsFromTemplate-I-request vm: %s\n", child.Name)
		if child.Type == "application/vnd.vmware.vcloud.vm+xml" {
			resultVmRecord, err := v.Objs.Vdc.QueryVappVmTemplate(catName, tmplName, child.Name)
			if err != nil {
				fmt.Printf("queryVappVmTemplate for vm %s  %s skipping \n", child.Name, err.Error())
				continue
			} else {
				// take the HREF from the query request
				vm, err := v.Client.Client.GetVMByHref(resultVmRecord.HREF)
				if err != nil {
					fmt.Printf("GetVmsFromTemplate client.GetVMByHref for  %s err: %s Skipping\n", resultVmRecord.Name, err.Error())
					continue
				} else {
					vu.DumpVM(vm.VM, 1)
					vms = append(vms, vm)
				}
			}
		}
	}
	return vms, nil
}
