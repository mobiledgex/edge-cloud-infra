package vcd

import (
	"context"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
)

// catalog releated functionality

// generic upload in cats_test
func (v *VcdPlatform) UploadOvaFile(ctx context.Context, tmplName string) error {

	baseurl := "" // ovaLocation
	tname := tmplName
	url := baseurl + "/tmplName" + "ova"

	log.SpanLog(ctx, log.DebugLevelInfra, "upload ova from", "URI", url, "tmpl", tname)
	cat := v.Objs.PrimaryCat
	elapse_start := time.Now()
	// MB
	task, err := cat.UploadOvf(url, tname, "mex ova base template", 8*1024)
	if err != nil {
		return err
	}
	err = task.WaitTaskCompletion()
	elapsed := time.Since(elapse_start).String()
	log.SpanLog(ctx, log.DebugLevelInfra, "tmpl uploaded ", "template", tmplName, "elapsed time", elapsed)

	return err
}

func (v *VcdPlatform) DeleteTemplate(ctx context.Context, name string) error {
	cat := v.Objs.PrimaryCat
	cItem, err := cat.GetCatalogItemByName(name, false)
	if err != nil {
		return err
	}
	return cItem.Delete()
}
