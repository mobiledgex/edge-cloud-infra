package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// vapptemplate related operations

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
	queryRes, err := v.Objs.Vdc.QueryVappTemplateList()
	if err != nil {
		return nil, err
	}
	for _, res := range queryRes {
		for _, catContainer := range v.Objs.Cats {
			cat := catContainer.OrgCat

			tmpl, err := cat.GetVappTemplateByHref(res.HREF)
			if err != nil {
				// This can happen if we have a vm with no vapp, one gets created for it
				continue
			} else {
				v.Objs.VAppTmpls[tmpl.VAppTemplate.Name] = tmpl
				tmpls = append(tmpls, tmpl)
			}
		}
	}

	return tmpls, nil

}
