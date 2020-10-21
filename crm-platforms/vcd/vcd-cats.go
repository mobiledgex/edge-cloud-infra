package vcd

import (
	"context"
	"fmt"
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
