package ctrlapi

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/mobiledgex/edge-cloud-infra/mc/gormlog"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type FederationController struct {
	ormapi.OperatorFederation
	Database *gorm.DB
}

func (f *FederationController) loggedDB(ctx context.Context) *gorm.DB {
	return gormlog.LoggedDB(ctx, f.Database)
}

func (f *FederationController) GetOperatorFederationObj(ctx context.Context, region, operator string) (*ormapi.OperatorFederation, bool, error) {
	if operator == "" {
		return nil, false, fmt.Errorf("no operator specified")
	}
	allObjs, err := f.GetRegionFederationObjs(ctx, region, operator)
	if err != nil {
		return nil, false, err
	}
	if len(allObjs) == 0 {
		return nil, true, nil
	}
	// Since region+operator is a primary key, there can only be one output
	return &allObjs[0], true, nil
}

func (f *FederationController) GetRegionFederationObjs(ctx context.Context, region, operator string) ([]ormapi.OperatorFederation, error) {
	if region == "" {
		return nil, fmt.Errorf("no region specified")
	}
	fedCtrl := ormapi.OperatorFederation{
		CountryCode: region,
		OperatorId:  operator,
	}
	db := f.loggedDB(ctx)
	fedCtrls := []ormapi.OperatorFederation{}
	res := db.Where(&fedCtrl).Find(&fedCtrls)
	if res.Error != nil {
		if res.RecordNotFound() {
			// return empty object if not found
			return []ormapi.OperatorFederation{}, nil
		}
		return nil, res.Error
	}
	return fedCtrls, nil
}

func (f *FederationController) ShowCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Cloudlet) error) error {
	return nil
}
