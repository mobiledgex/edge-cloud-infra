package vcd

import (
	"context"
	"fmt"
	//"os"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/vmlayer"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// catalog releated functionality

// Return catalog names found in our our org. Then we can get by Name.
func (v *VcdPlatform) GetCatalogNames(ctx context.Context) ([]string, error) {
	log.SpanLog(ctx, log.DebugLevelInfra, "GetCatalogNames from", "Org", v.Objs.Org.Org.Name)
	var catNames []string

	return catNames, nil
}

// Gather media records from our catalog(s)
func (v *VcdPlatform) GetMediaRecords(ctx context.Context) ([]*types.MediaRecordType, error) {
	c := CatContainer{}
	cname := ""
	for cname, c = range v.Objs.Cats {
		m, err := c.OrgCat.QueryMediaList()
		if err == nil {
			return nil, fmt.Errorf("Error from QueryMediaList cat: %s error %s", cname, err.Error())
		}
		c.MediaRecs = append(c.MediaRecs, m...)
	}
	return c.MediaRecs, nil
}

// generic upload in cats_test
func (v *VcdPlatform) UploadOvaFile(ctx context.Context, tmplName string) error {

	// The platform has some URL goodies to use
	// no longer exists	vconf := v.vmProperties.CommonPf.VaultConfig

	ovaLocation := vmlayer.DefaultCloudletVMImagePath + "vcd-" + vmlayer.MEXInfraVersion + ".ova"
	fmt.Printf("UploadOvaFile: ovaLocation: %s\n", ovaLocation)

	// need stdard URL I think platforms has a generic URL
	//path := os.Getenv("HOME")
	//fmt.Printf("Path: %s\n", path)

	url := ovaLocation // path + "/vmware-lab/" + *ovaName
	//
	tname := tmplName + "-tmpl"
	fmt.Printf("testMediaUpload-I-attempt uploading: %s naming it %s \n", url, tname)

	cat := v.Objs.PrimaryCat
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

	//afilePath, err := vmlayer.DownloadVMImage(ctx, v.vmProperties.CommonPf.VaultConfig, imageName, imageUrl, md5Sum)
	//if err != nil {
	//	return err
	//}

}
