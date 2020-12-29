package vcd

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	// vu "github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd/vcdutils"
	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestCats(t *testing.T) {
	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestCats-I-tv init done\n")
		pcat := tv.Objs.PrimaryCat
		govcd.ShowCatalog(*pcat.Catalog)
	} else { //
		return
	}
}

// -tmpl from PrimaryCat
func TestRMTmpl(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")

	if live {
		fmt.Printf("Test Remove template %s from cat", *tmplName)
		err := tv.DeleteTemplate(ctx, *tmplName)
		if err != nil {
			fmt.Printf("TestRMTmpl delete %s returned %s\n", *tmplName, err.Error())
		}
		return
	}
}

// uses -ova
// If you run this live, use --timeout 0 to disable the default panic after 10 mins
// since this test runs ~ upload complete in 17m20.262586965s
//
func TestUploadOva(t *testing.T) {
	live, ctx, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("Live OVA  upload test\n")
		err = testOvaUpload(t, ctx)
		require.Nil(t, err, "testOVAUpload")
	} else {
		return
	}
}

// -live -href
func TestGetTmplByHref(t *testing.T) {

	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		//		fmt.Printf("Get by href org: %s href %s\n",
	}
}

// Or we can access tmpls via our vcd items looking for
// "application/vnd.vmware.vcloud.vAppTemplate+xml"
// In general, many resource Items can be discovered this way
//
func TestGetTemplates(t *testing.T) {

	//	var tmpls []string

	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		vdc := tv.Objs.Vdc
		fmt.Printf("TestGetTemplates\n")
		for _, res := range vdc.Vdc.ResourceEntities {
			for N, item := range res.ResourceEntity {
				if item.Type == "application/vnd.vmware.vcloud.vAppTemplate+xml" {
					fmt.Printf("%3d %-40s %s\n", N, item.Name, item.Type)
					//tmpls = append(tmpls, item.Name)

				}
				if item.Type == "application/vnd.vmware.vcloud.vm+xml" {
					fmt.Printf("VM found: %3d %-40s %s\n", N, item.Name, item.Type)
				}
			}
		}
		fmt.Println("")
	}
	return
}

// upload a local .vmdk to our catalog, actually, they prefer an entire .ova file
// uses -ova for what (local) file to upload
func testOvaUpload(t *testing.T, ctx context.Context) error {

	path := os.Getenv("HOME")
	fmt.Printf("Path: %s\n", path)

	url := path + "/vmware-lab/" + *ovaName

	//tname := *ovaName
	tname := "mobiledgex-v4.1.3-vcd"
	fmt.Printf("testMediaUpload-I-attempt uploading: %s naming it %s \n", url, tname)

	cat := tv.Objs.PrimaryCat
	elapse_start := time.Now()
	task, err := cat.UploadOvf(url, tname, "test-import-ova-vcd", (1024 * 100))

	if err != nil {
		fmt.Printf("\nError from UploadOvf: %s\n", err.Error())
	}
	fmt.Printf("Task: %+v\n", task)
	err = task.WaitTaskCompletion()
	fmt.Printf("upload complete in %s\n", time.Since(elapse_start).String())
	return err
}

// Ok, we have a case where qa2-vdc has no vapptemplate as vdc.resources.
// It has a catalog item though that we obtain using

// -live -tmpl
func TestImportVMTmpl(t *testing.T) {

	//	var tmpls []string

	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestImport tmpl %s\n", *tmplName)
		// we want to take a item (vcloud.vm+xml) and instanciate it to be a vdc.resource full vcloud.vapptemplate+xml type
		vdc := tv.Objs.Vdc
		templateVmQueryRecs, err := tv.Client.Client.QueryVmList(types.VmQueryFilterOnlyTemplates)
		qr := &types.QueryResultVMRecordType{}
		for _, qr = range templateVmQueryRecs {

			if qr.Name == *tmplName {

				fmt.Printf("Discover found template:\n\tName%s\n\tType: %s\n\tHref:%s\n", qr.Name, qr.Type, qr.HREF)

				tmpl, err := tv.Objs.PrimaryCat.GetVappTemplateByHref(qr.HREF)

				fmt.Printf("template:\n\tName: %s\n\tType: %s\n\tID: %s\n\tHREF: %s\n\t OperKey: %s\n\tStatus: %d\n\tOvfDescriptorUpLoaded: %s\n",
					tmpl.VAppTemplate.Name,
					tmpl.VAppTemplate.Type,
					tmpl.VAppTemplate.ID,
					tmpl.VAppTemplate.HREF,
					tmpl.VAppTemplate.OperationKey,
					tmpl.VAppTemplate.Status,
					tmpl.VAppTemplate.OvfDescriptorUploaded)

				if err != nil {
					fmt.Printf("\n\nDISCOVER: Error GetVappTemplateByHref: tmpl: %s err: %s \n", qr.Name, err.Error())
					return
				}
				fmt.Printf("Have tmpl as: %+v\n", tmpl)
				err = tmpl.Refresh()
				if err != nil {
					fmt.Printf("error refreshing tmpl: %s\n", err.Error())
				}
				//vu.DumpVAppTemplate(tmpl, 1)
				break
			}
		}
		vappTmplRef := &types.Reference{
			HREF: qr.HREF,
			ID:   qr.ID,
			Type: qr.Type,
			Name: qr.Name,
		}
		tmplParams := &types.InstantiateVAppTemplateParams{
			Name:             qr.Name,
			PowerOn:          false,
			Source:           vappTmplRef,
			AllEULAsAccepted: true,
		}

		err = vdc.InstantiateVAppTemplate(tmplParams)
		if err != nil {
			fmt.Printf("Instanciate error: %s\n", err.Error())
		}

		// now check our vdc.Resources, is it there now?
		for _, res := range vdc.Vdc.ResourceEntities {
			for N, item := range res.ResourceEntity {
				if item.Type == "application/vnd.vmware.vcloud.vAppTemplate+xml" {
					fmt.Printf("%3d %-40s %s\n", N, item.Name, item.Type)
					//tmpls = append(tmpls, item.Name)

				}
				if item.Type == "application/vnd.vmware.vcloud.vm+xml" {
					fmt.Printf("VM found: %3d %-40s %s\n", N, item.Name, item.Type)
				}
			}
		}
	}
}
