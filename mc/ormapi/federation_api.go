package ormapi

import (
	"fmt"

	"github.com/lib/pq"
)

type Federator struct {
	// Globally unique string used to indentify a federation with partner federation
	FederationId string `gorm:"primary_key" json:"federationid"`
	// Globally unique string to identify an operator platform
	OperatorId string `gorm:"type:citext" json:"operatorid"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	CountryCode string `json:"countrycode"`
	// Federation access point address
	FederationAddr string `json:"federationaddr"`
	// Mobile country code of operator sending the request
	MCC string `json:"mcc"`
	// List of mobile network codes of operator sending the request
	MNC pq.StringArray `gorm:"type:text[]" json:"mnc"`
	// IP and Port of discovery service URL of operator platform
	LocatorEndPoint string `json:"locatorendpoint"`
	// Revision ID to track object changes. We use timestamps but
	// this can differ with what partner federator uses
	// read_only: true
	Revision string `json:"revision"`
}

type Federation struct {
	// Internal ID to reference a federation
	// read_only: true
	Id int `gorm:"auto_increment:true; unique; not null"`
	// Self federation ID
	SelfFederationId string `gorm:"primary_key; unique" json:"selffederationid"`
	// Self operator ID
	SelfOperatorId string `json:"selfoperatorid"`
	// Partner Federator
	Federator `json:",inline"`
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
	// Globally unique string used to authenticate operations over federation interface
	ZoneId string `gorm:"primary_key" json:"zoneid"`
	// Globally unique string to identify an operator platform
	OperatorId string `gorm:"type:citext" json:"operatorid"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	CountryCode string `json:"countrycode"`
	// GPS co-ordinates associated with the zone (in decimal format)
	GeoLocation string `json:"geolocation"`
	// Comma seperated list of cities under this zone
	City string `json:"city"`
	// Comma seperated list of states under this zone
	State string `json:"state"`
	// Type of locality eg rural, urban etc.
	Locality string `json:"locality"`
	// Region in which cloudlets reside
	Region string `json:"region"`
	// List of cloudlets part of this zone
	Cloudlets pq.StringArray `gorm:"type:text[]" json:"cloudlets"`
	// Revision ID to track object changes. We use timestamps but
	// this can differ with what partner federator uses
	// read_only: true
	Revision string `json:"revision"`
}

// Information of the partner federator with whom the self federator zone is shared
type FederatedSelfZone struct {
	// Globally unique identifier of the federator zone
	ZoneId string `gorm:"primary_key;type:text REFERENCES federator_zones(zone_id)" json:"zoneid"`
	// Self operator ID
	SelfOperatorId string `json:"selfoperatorid"`
	// Self federation ID
	SelfFederationId string `gorm:"primary_key" json:"selffederationid"`
	// Partner federation ID
	PartnerFederationId string `gorm:"primary_key" json:"partnerfederationid"`
	// Zone registered by partner federator
	// read_only: true
	Registered bool
	// Revision ID to track object changes. We use timestamps but
	// this can differ with what partner federator uses
	// read_only: true
	Revision string `json:"revision"`
}

// Zones shared as part of federation with partner federator
type FederatedPartnerZone struct {
	// Self operator ID
	SelfOperatorId string `json:"selfoperatorid"`
	// Self federation ID
	SelfFederationId string `gorm:"primary_key" json:"selffederationid"`
	// Partner federation ID
	PartnerFederationId string `gorm:"primary_key" json:"partnerfederationid"`
	// Partner federator zone
	FederatorZone `json:",inline"`
	// Zone registered by self federator
	// read_only: true
	Registered bool
}

func federatorStr(operatorId, countryCode, federationId string) string {
	if federationId != "" {
		return fmt.Sprintf("OperatorId:%q/CountryCode:%q/FederationId:%q", operatorId, countryCode, federationId)
	}
	return fmt.Sprintf("OperatorId:%q/CountryCode:%q", operatorId, countryCode)
}

func federationIdStr(federationId string) string {
	return fmt.Sprintf("FederationId:%q", federationId)
}

func (s *Federator) IdString() string {
	return federatorStr(s.OperatorId, s.CountryCode, s.FederationId)
}

func (s *Federation) SelfIdString() string {
	return federationIdStr(s.SelfFederationId)
}

func (s *Federation) PartnerIdString() string {
	return s.Federator.IdString()
}

func (s *FederatorZone) IdString() string {
	return federatorStr(s.OperatorId, s.CountryCode, "")
}

func (s *FederatedSelfZone) SelfIdString() string {
	return federationIdStr(s.SelfFederationId)
}

func (s *FederatedSelfZone) PartnerIdString() string {
	return federationIdStr(s.PartnerFederationId)
}

func (s *FederatedPartnerZone) SelfIdString() string {
	return federationIdStr(s.SelfFederationId)
}

func (s *FederatedPartnerZone) PartnerIdString() string {
	return federationIdStr(s.PartnerFederationId)
}
