package vcd

import (
	"context"
	"fmt"
	//vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
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

// Return all templates found in our catalog
func (v *VcdPlatform) GetAllVdcTemplates(ctx context.Context /*, cat *govcd.Catalog*/) ([]*govcd.VAppTemplate, error) {

	var tmpls []*govcd.VAppTemplate
	fmt.Printf("\n\n-----GetAddVdcTemplates----\n\n")
	queryRes, err := v.Objs.Vdc.QueryVappTemplateList()
	if err != nil {
		fmt.Printf("\nQueryVappTemplList err: %s\n", err.Error())
		return nil, err
	}
	fmt.Printf("\n\n-----GetAddVdcTemplates----have %d queryResults \n\n", len(queryRes))
	for n, res := range queryRes {
		for catName, catContainer := range v.Objs.Cats {
			cat := catContainer.OrgCat
			fmt.Printf("\n\n\t#%d Lookup in cat %s  res.Name: %sby HREF: %s\n\n", n, catName, res.Name, res.HREF)

			tmpl, err := cat.GetVappTemplateByHref(res.HREF)
			if err != nil {
				fmt.Printf("\tError fetching %s in cat %s err %s\n\n", res.HREF, catName, err.Error())
				// This can happen if we have a vm with no vapp, one gets created for it
				continue
			} else {
				v.Objs.VAppTmpls[tmpl.VAppTemplate.Name] = tmpl
				if tmpl.VAppTemplate.Name == v.GetVDCTemplateName() {
					fmt.Printf("\nFound our VDCTEMPLATE in cat %s\n", catName)
				}
				tmpls = append(tmpls, tmpl)
			}
		}
	}

	return tmpls, nil

}

// Return the list of VMs contained in this template
// This impl fails. But they have a new call, use that here
// 		if tmpl.Type == "application/vnd.vmware.vcloud.vm+xml" {

/*
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
*/
