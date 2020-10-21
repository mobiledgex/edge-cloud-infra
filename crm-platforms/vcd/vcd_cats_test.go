package vcd

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"os"
	"testing"
	"time"
)

func TestCats(t *testing.T) {
	live, _, err := InitVcdTestEnv()
	require.Nil(t, err, "InitVcdTestEnv")
	if live {
		fmt.Printf("TestCats-I-tv init done\n")
		pcat := tv.Objs.PrimaryCat
		fmt.Printf("PrimaryCat perhaps not set? : %+v\n", pcat)
		govcd.ShowCatalog(*pcat.Catalog)
	} else { //
		return
	}
}

// Or we can access tmpls via our vcd items looking for
// "application/vnd.vmware.vcloud.vAppTemplate+xml"
// In general, many resource Items can be discovered this way
//
func testGetAllVdcTemplates(t *testing.T, ctx context.Context) []string {
	var tmpls []string

	fmt.Printf("\nvdc items\n")
	for _, res := range tv.Objs.Vdc.Vdc.ResourceEntities {
		for N, item := range res.ResourceEntity {
			if item.Type == "application/vnd.vmware.vcloud.vAppTemplate+xml" {
				fmt.Printf("%3d %-40s %s\n", N, item.Name, item.Type)
				tmpls = append(tmpls, item.Name)

			}
		}
	}
	fmt.Println("")
	return tmpls
}

// upload a local .vmdk to our catalog, actually, they prefer an entire .ova file
func testOvfUpload(t *testing.T, ctx context.Context) error {

	catalog := &govcd.Catalog{}
	path := os.Getenv("HOME")
	fmt.Printf("Path: %s\n", path)

	url := path + "/vmware-lab/mex-ova/mobiledgex-v3.1.6-v14-vapp.ova"
	fmt.Printf("testMediaUpload-I-attempt uploading: %s\n", url)
	cats := tv.Objs.Cats
	for name, cat := range cats {
		catalog = cat.OrgCat
		fmt.Printf("selecting cat %s\n", name)
		break
	}
	elapse_start := time.Now()
	// units for upload check size? MB? dunno... yet.

	task, err := catalog.UploadOvf(url, "mobiledgex-3.1.6.ova", "test-import-ova-vsphere", 10)

	if err != nil {
		fmt.Printf("\nError from UploadOvf: %s\n", err.Error())
	}
	fmt.Printf("Task: %+v\n", task)
	err = task.WaitTaskCompletion()
	fmt.Printf("upload complete in %s\n", time.Since(elapse_start).String())
	return err
}
