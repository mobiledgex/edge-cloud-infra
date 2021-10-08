package ormapi

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Federator struct {
	// Globally unique string to identify an operator platform
	OperatorId string `gorm:"primary_key"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	CountryCode string `gorm:"primary_key"`
	// Globally unique string used to authenticate operations over federation interface
	FederationKey string
	// Federation access point address
	FederationAddr string
	// Mobile country code of operator sending the request
	MCC string
	// Comma separated list of mobile network codes of operator sending the request
	MNC pq.StringArray `gorm:"type:text[]"`
	// IP and Port of discovery service URL of operator platform
	LocatorEndPoint string
}

type Federation struct {
	// Self federator operator ID
	SelfOperatorId string `gorm:"primary_key"`
	// Self federator country code
	SelfCountryCode string `gorm:"primary_key"`
	// Partner Federator
	Federator
	// Partner shares its zones with self federator as part of federation
	// read_only: true
	PartnerRoleShareZonesWithSelf bool
	// Partner is allowed access to self federator zones as part of federation
	// read_only: true
	PartnerRoleAccessToSelfZones bool
}

// Details of zone owned by a federator. MC defines a zone as a group of cloudlets,
// but currently it is restricted to one cloudlet
type FederatorZone struct {
	// Globally unique string to identify an operator platform
	OperatorId string `gorm:"primary_key"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	CountryCode string `gorm:"primary_key"`
	// Globally unique string used to authenticate operations over federation interface
	ZoneId string `gorm:"primary_key"`
	// GPS co-ordinates associated with the zone (in decimal format)
	GeoLocation string `json:"geoLocation"`
	// Comma seperated list of cities under this zone
	City string `json:"city"`
	// Comma seperated list of states under this zone
	State string `json:"state"`
	// Type of locality eg rural, urban etc.
	Locality string `json:"locality"`
	// Region in which cloudlets reside
	Region string `json:"region"`
	// Delimited list of cloudlets part of this zone
	Cloudlets pq.StringArray `gorm:"type:text[]"`
}

// Information of the partner federator with whom the self federator zone is shared
type FederatedSelfZone struct {
	// Globally unique identifier of the federator zone
	ZoneId string `gorm:"primary_key"`
	// Self federator operator ID
	SelfOperatorId string `gorm:"primary_key"`
	// Self federator country code
	SelfCountryCode string `gorm:"primary_key"`
	// Partner federator operator ID
	PartnerOperatorId string `gorm:"primary_key"`
	// Partner federator country code
	PartnerCountryCode string `gorm:"primary_key"`
	// Zone registered by partner federator
	// read_only: true
	Registered bool
}

// Zones shared as part of federation with partner federator
type FederatedPartnerZone struct {
	// Self federator operator ID
	SelfOperatorId string `gorm:"primary_key"`
	// Self federator country code
	SelfCountryCode string `gorm:"primary_key"`
	// Partner federator zone
	FederatorZone
	// Zone registered by self federator
	// read_only: true
	Registered bool
}

func setForeignKeyConstraint(loggedDb *gorm.DB, fKeyTableName, fKeyFields, refTableName, refFields string) error {
	cmd := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT self_fk_constraint FOREIGN KEY (%s) "+
		"REFERENCES %s(%s)", fKeyTableName, fKeyFields, refTableName, refFields)
	err := loggedDb.Exec(cmd).Error
	if err != nil {
		return err
	}
	return nil
}

func InitFederationAPIConstraints(loggedDb *gorm.DB) error {
	// setup foreign key constraints, this is done separately here as gorm doesn't
	// support referencing composite primary key inline to the model

	// Federation's SelfOperatorId/SelfCountryCode references Federator's OperatorId/CountryCode
	scope := loggedDb.Unscoped().NewScope(&Federation{})
	fKeyTableName := scope.TableName()
	fKeyFields := fmt.Sprintf("%s, %s", scope.Quote("self_operator_id"), scope.Quote("self_country_code"))

	scope = loggedDb.Unscoped().NewScope(&Federator{})
	refTableName := scope.TableName()
	refFields := fmt.Sprintf("%s, %s", scope.Quote("operator_id"), scope.Quote("country_code"))
	err := setForeignKeyConstraint(loggedDb, fKeyTableName, fKeyFields, refTableName, refFields)
	if err != nil {
		return err
	}

	// FederatorZone's OperatorId/CountryCode references Federator's OperatorId/CountryCode
	scope = loggedDb.Unscoped().NewScope(&FederatorZone{})
	fKeyTableName = scope.TableName()
	fKeyFields = fmt.Sprintf("%s, %s", scope.Quote("operator_id"), scope.Quote("country_code"))

	scope = loggedDb.Unscoped().NewScope(&Federator{})
	refTableName = scope.TableName()
	refFields = fmt.Sprintf("%s, %s", scope.Quote("operator_id"), scope.Quote("country_code"))
	err = setForeignKeyConstraint(loggedDb, fKeyTableName, fKeyFields, refTableName, refFields)
	if err != nil {
		return err
	}

	// FederatedSelfZone's SelfOperatorId/SelfCountryCode/PartnerOperatorId/PartnerCountryCode references
	//   Federation's SelfOperatorId/SelfCountryCode/OperatorId/CountryCode
	scope = loggedDb.Unscoped().NewScope(&FederatedSelfZone{})
	fKeyTableName = scope.TableName()
	fKeyFields = fmt.Sprintf("%s, %s, %s, %s",
		scope.Quote("self_operator_id"), scope.Quote("self_country_code"),
		scope.Quote("partner_operator_id"), scope.Quote("partner_country_code"))

	scope = loggedDb.Unscoped().NewScope(&Federation{})
	refTableName = scope.TableName()
	refFields = fmt.Sprintf("%s, %s, %s, %s",
		scope.Quote("self_operator_id"), scope.Quote("self_country_code"),
		scope.Quote("operator_id"), scope.Quote("country_code"))
	err = setForeignKeyConstraint(loggedDb, fKeyTableName, fKeyFields, refTableName, refFields)
	if err != nil {
		return err
	}

	// FederatedPartnerZone's SelfOperatorId/SelfCountryCode/OperatorId/CountryCode references
	//   Federation's SelfOperatorId/SelfCountryCode/OperatorId/CountryCode
	scope = loggedDb.Unscoped().NewScope(&FederatedPartnerZone{})
	fKeyTableName = scope.TableName()
	fKeyFields = fmt.Sprintf("%s, %s, %s, %s",
		scope.Quote("self_operator_id"), scope.Quote("self_country_code"),
		scope.Quote("operator_id"), scope.Quote("country_code"))

	scope = loggedDb.Unscoped().NewScope(&Federation{})
	refTableName = scope.TableName()
	refFields = fmt.Sprintf("%s, %s, %s, %s",
		scope.Quote("self_operator_id"), scope.Quote("self_country_code"),
		scope.Quote("operator_id"), scope.Quote("country_code"))
	err = setForeignKeyConstraint(loggedDb, fKeyTableName, fKeyFields, refTableName, refFields)
	if err != nil {
		return err
	}

	return nil
}
