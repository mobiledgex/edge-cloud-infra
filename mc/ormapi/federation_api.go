package ormapi

import (
	fedcommon "github.com/mobiledgex/edge-cloud-infra/mc/federation/common"
)

// GORM objects
// ============
type Federator struct {
	// Globally unique string to identify an operator platform
	OwnerOperatorId string `gorm:"primary_key"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	OwnerCountryCode string `gorm:"primary_key"`
	// Globally unique string to identify an operator platform
	OperatorId string `gorm:"primary_key"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	CountryCode string `gorm:"primary_key"`
	// Type of federator: self or partner
	Type string
	// Globally unique string used to authenticate operations over federation interface
	FederationKey string
	// Federation access point address
	FederationAddr string
	// Mobile country code of operator sending the request
	MCC string
	// Comma separated list of mobile network codes of operator sending the request
	MNCs string
	// IP and Port of discovery service URL of OP
	LocatorEndPoint string
}

type Federation struct {
	// Self federator operator ID
	SelfOperatorId string `gorm:"primary_key"`
	// Self federator country code
	SelfCountryCode string `gorm:"primary_key"`
	// Partner federator operator ID
	PartnerOperatorId string `gorm:"primary_key"`
	// Partner federator country code
	PartnerCountryCode string `gorm:"primary_key"`
	// Role of the partner federator in federation
	PartnerRole fedcommon.FederatorRole `gorm:"primary_key"`
}

// Zone owned by a Federator. MC defines a zone as a group of cloudlets,
// but currently it is restricted to one cloudlet
type FederatorZone struct {
	// Globally unique string to identify an operator platform
	OperatorId string `gorm:"primary_key"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	CountryCode string `gorm:"primary_key"`
	// Globally unique string used to authenticate operations over federation interface
	ZoneId string `gorm:"primary_key"`
	// Mobile country code of operator sending the request
	MCC string `json:"MCC"`
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
	Cloudlets string `json:"cloudlet"`
}

// Information of the Federator with whom the zone is shared
type FederatorSharedZone struct {
	// Globally unique identifier of the federator zone
	ZoneId string `gorm:"primary_key"`
	// Federator operator ID who owns the zone
	OwnerOperatorId string `gorm:"primary_key"`
	// Federator country code who owns the zone
	OwnerCountryCode string `gorm:"primary_key"`
	// Federator operator ID with whom the zone is shared
	SharedWithOperatorId string `gorm:"primary_key"`
	// Federator country code with whom the zone is shared
	SharedWithCountryCode string `gorm:"primary_key"`
}

// Information of the Federator who has registered the zone
type FederatorRegisteredZone struct {
	// Globally unique identifier of the federator zone
	ZoneId string `gorm:"primary_key"`
	// Federator operator ID who owns the zone
	OwnerOperatorId string `gorm:"primary_key"`
	// Federator country code who owns the zone
	OwnerCountryCode string `gorm:"primary_key"`
	// Federator operator ID who has registered the zone
	RegisteredByOperatorId string `gorm:"primary_key"`
	// Federator country code who has registered the zone
	RegisteredByCountryCode string `gorm:"primary_key"`
}

// API Objects
// ===========
type FederatorRequest struct {
	// Self federator operator ID
	// required: true
	SelfOperatorId string
	// Self federator country code
	// required: true
	SelfCountryCode string
	// Globally unique string used to authenticate operations over federation interface
	FederationKey string
	// Federation access point address
	FederationAddr string
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string
	// Type of the federation
	Type string
	// Mobile country code of operator sending the request
	MCC string
	// List of mobile network codes of operator sending the request
	MNCs []string
	// IP and Port of discovery service URL of OP
	LocatorEndPoint string
}

type FederationRequest struct {
	// Operator ID of the self federator
	// required: true
	SelfOperatorId string
	// Country code of the self federator
	// required: true
	SelfCountryCode string
	// Operator ID of the partner federator
	// required: true
	PartnerOperatorId string
	// Country code of the partner federator
	// required: true
	PartnerCountryCode string
}

type FederatorZoneDetails struct {
	// Globally unique string used to authenticate operations over federation interface
	// required: true
	ZoneId string
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string
	// Operator region required for zone to get the list of cloudlets
	Region string
	// GPS co-ordinates associated with the zone (in decimal format)
	// Latitude and Longitude is separated by comma. For example: 44.4308975,-89.6884637
	GeoLocation string
	// Comma seperated list of cities under this zone
	City string
	// Comma seperated list of states under this zone
	State string
	// Type of locality eg rural, urban etc.
	Locality string
	// List of cloudlets belonging to the federation zone
	Cloudlets []string
	// List of partner federators by whom this zone is registered
	RegisteredByFederators []string
	// List of partner federators with whom this zone is shared
	SharedWithFederators []string
}

type FederatorZoneRequest struct {
	// Operator ID of the self federator
	// required: true
	SelfOperatorId string
	// Country code of the self federator
	// required: true
	SelfCountryCode string
	// Operator ID of the partner federator
	// required: true
	PartnerOperatorId string
	// Country code of the partner federator
	// required: true
	PartnerCountryCode string
	// Zone ID
	// required: true
	ZoneId string
}
