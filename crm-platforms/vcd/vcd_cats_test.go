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

// Or we can access tmpls via our vcd items looking for
// "application/vnd.vmware.vcloud.vAppTemplate+xml"
// In general, many resource Items can be discovered this way
//
func testGetAllVdcTemplates(t *testing.T, ctx context.Context) []string {
	var tmpls []string
	vdc := tv.Objs.Vdc

	fmt.Printf("\nvdc items\n")
	for _, res := range vdc.Vdc.ResourceEntities {
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
// uses -ova for what (local) file to upload
func testOvaUpload(t *testing.T, ctx context.Context) error {

	path := os.Getenv("HOME")
	fmt.Printf("Path: %s\n", path)

	url := path + "/vmware-lab/" + *ovaName

	tname := *ovaName + "-tmpl"
	fmt.Printf("testMediaUpload-I-attempt uploading: %s naming it %s \n", url, tname)

	cat := tv.Objs.PrimaryCat
	elapse_start := time.Now()
	// units for upload check size? MB? dunno... yet.

	task, err := cat.UploadOvf(url, tname, "test-import-ova-vcd", 1024)

	if err != nil {
		fmt.Printf("\nError from UploadOvf: %s\n", err.Error())
	}
	fmt.Printf("Task: %+v\n", task)
	err = task.WaitTaskCompletion()
	fmt.Printf("upload complete in %s\n", time.Since(elapse_start).String())
	return err
}
