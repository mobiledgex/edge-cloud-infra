package ctrlclient

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
	fedcommon "github.com/mobiledgex/edge-cloud-infra/mc/federation/common"
	"github.com/mobiledgex/edge-cloud-infra/mc/gormlog"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormutil"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type FederationClient struct {
	ormapi.OperatorFederation
	Database *gorm.DB
}

func (f *FederationClient) loggedDB(ctx context.Context) *gorm.DB {
	return gormlog.LoggedDB(ctx, f.Database)
}

func GetFederationClient(ctx context.Context, database *gorm.DB, region, operator string) (*FederationClient, bool, error) {
	if operator == "" {
		return nil, false, fmt.Errorf("no operator specified")
	}
	clients, err := GetFederationClients(ctx, database, region, operator)
	if err != nil {
		return nil, false, err
	}
	if len(clients) == 0 {
		return &FederationClient{}, false, nil
	}
	// Since region+operator is a primary key, there can only be one output
	return &clients[0], true, nil
}

func GetFederationClients(ctx context.Context, database *gorm.DB, region, operator string) ([]FederationClient, error) {
	if region == "" {
		return nil, fmt.Errorf("no region specified")
	}
	fedObj := ormapi.OperatorFederation{
		CountryCode: region,
		OperatorId:  operator,
	}
	db := gormlog.LoggedDB(ctx, database)
	fedObjs := []ormapi.OperatorFederation{}
	res := db.Where(&fedObj).Find(&fedObjs)
	if res.Error != nil {
		if res.RecordNotFound() {
			// return empty object if not found
			return []FederationClient{}, nil
		}
		return nil, res.Error
	}
	fedClients := []FederationClient{}
	for _, fedObj := range fedObjs {
		// Only access those partner OP's whose zones we can access
		if !fedcommon.FederationRoleExists(&fedObj, fedcommon.RoleAccessZones) {
			continue
		}
		fedClient := FederationClient{
			Database:           database,
			OperatorFederation: fedObj,
		}
		fedClients = append(fedClients, fedClient)
	}
	return fedClients, nil
}

func (f *FederationClient) ShowCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Cloudlet) error) error {
	return nil
}
