package client

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
	ormapi.Federator
	Database *gorm.DB
}

func (f *FederationClient) loggedDB(ctx context.Context) *gorm.DB {
	return gormlog.LoggedDB(ctx, f.Database)
}

func GetFederationClient(ctx context.Context, database *gorm.DB, region, operator string) (*FederationClient, bool, error) {
	if operator == "" {
		// No operator specified
		return &FederationClient{}, false, nil
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

// Get Federation Clients using region as CountryCode and optionally operator as OperatorId
// NOTE: This client will abstract actions on partner federator's edge infra. Hence,
//       consider region as CountryCode
func GetFederationClients(ctx context.Context, database *gorm.DB, region, operator string) ([]FederationClient, error) {
	if region == "" {
		return nil, fmt.Errorf("no region specified")
	}
	fedObj := ormapi.Federator{
		CountryCode: region,
		OperatorId:  operator,
		Type:        fedcommon.TypePartner,
	}
	db := gormlog.LoggedDB(ctx, database)
	fedObjs := []ormapi.Federator{}
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
		// Only access those partner federators whose zones can be accessed by self federators
		roleLookup := ormapi.FederatorRole{
			PartnerFederationId: fedObj.FederationId,
		}
		partnerFederatorRole := ormapi.FederatorRole{}
		res := db.Where(&roleLookup).Find(&partnerFederatorRole)
		if !res.RecordNotFound() && res.Error != nil {
			return nil, ormutil.DbErr(res.Error)
		}
		if !fedcommon.ValueExistsInDelimitedList(partnerFederatorRole.Role, fedcommon.RoleAccessPartnerZones) {
			continue
		}
		fedClient := FederationClient{
			Database:  database,
			Federator: fedObj,
		}
		fedClients = append(fedClients, fedClient)
	}
	return fedClients, nil
}

func (f *FederationClient) ShowCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Cloudlet) error) error {
	return nil
}
