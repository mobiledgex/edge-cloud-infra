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
func GetFederationClients(ctx context.Context, database *gorm.DB, region, operator string) ([]FederationClient, error) {
	if region == "" {
		return nil, fmt.Errorf("no region specified")
	}
	// NOTE: This client will abstract actions on partner federator's edge infra.
	//       Hence, consider region as CountryCode
	partnerFed := ormapi.Federator{
		OperatorId:  operator,
		CountryCode: region,
		Type:        fedcommon.TypePartner,
	}
	db := gormlog.LoggedDB(ctx, database)
	partnerFeds := []ormapi.Federator{}
	res := db.Where(&partnerFed).Find(&partnerFeds)
	if res.Error != nil {
		if res.RecordNotFound() {
			// return empty object if not found
			return []FederationClient{}, nil
		}
		return nil, res.Error
	}

	fedClients := []FederationClient{}
	for _, partnerFed := range partnerFeds {
		// Only access those partner federators whose zones can be accessed by self federators
		partnerFederation := ormapi.Federation{
			PartnerOperatorId:  partnerFed.OperatorId,
			PartnerCountryCode: partnerFed.CountryCode,
			PartnerRole:        fedcommon.RoleShareZonesWithSelf,
		}
		// There can only be one partner federator who can shares zones with MC,
		// this is enforced during addition of partner federation
		res := db.Where(&partnerFederation).First(&partnerFederation)
		if res.RecordNotFound() {
			continue
		}
		if res.Error != nil {
			return nil, ormutil.DbErr(res.Error)
		}
		fedClient := FederationClient{
			Database:  database,
			Federator: partnerFed,
		}
		fedClients = append(fedClients, fedClient)
	}
	return fedClients, nil
}

func (f *FederationClient) ShowCloudletStream(ctx context.Context, rc *ormutil.RegionContext, obj *edgeproto.Cloudlet, cb func(res *edgeproto.Cloudlet) error) error {
	return nil
}
