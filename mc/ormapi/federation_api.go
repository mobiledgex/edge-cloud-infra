package ormapi

// GORM objects
// ============
type SelfFederator struct {
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string `gorm:"primary_key"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string `gorm:"primary_key"`
	// Globally unique string used to authenticate operations over federation interface
	FederationKey string
	// Federation access point address
	FederationAddr string
	// Type of the federation
	// Type string
	// Comma separated set of regions an operator belongs to
	Regions string
	// Mobile country code of operator sending the request
	MCC string
	// Comma separated list of mobile network codes of operator sending the request
	MNCs string
	// IP and Port of discovery service URL of OP
	LocatorEndPoint string
}

type PartnerFederator struct {
	// Self federator operator ID
	// required: true
	SelfOperatorId string `gorm:"primary_key"`
	// Self federator country code
	// required: true
	SelfCountryCode string `gorm:"primary_key"`
	// Partner federator operator ID
	// required: true
	PartnerOperatorId string `gorm:"primary_key"`
	// Partner federator country code
	// required: true
	PartnerCountryCode string `gorm:"primary_key"`
	// Partner federator federation key used to authenticate operations over federation interface
	PartnerFederationKey string
	// Partner federation access point address
	PartnerFederationAddr string
	// Mobile country code of operator sending the request
	PartnerMCC string
	// Comma separated list of mobile network codes of operator sending the request
	PartnerMNCs string
	// IP and Port of discovery service URL of OP
	PartnerLocatorEndPoint string
	// Partner federator can share zones with self federator
	RoleShareZonesWithSelf bool
	// Partner federator can access zones of self federator
	RoleAccessToSelfZones bool
}

// Zone owned by a Federator. MC defines a zone as a group of cloudlets,
// but currently it is restricted to one cloudlet
type FederatorZone struct {
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string `gorm:"primary_key"`
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string `gorm:"primary_key"`
	// Globally unique string used to authenticate operations over federation interface
	// required: true
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
	// Globally unique string used to authenticate operations over federation interface
	FederationKey string
	// Federation access point address
	FederationAddr string
	// Type of the federation
	Type string
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string
	// Set of regions an operator belongs to
	Regions []string
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string
	// Mobile country code of operator sending the request
	MCC string
	// List of mobile network codes of operator sending the request
	MNCs []string
	// IP and Port of discovery service URL of OP
	LocatorEndPoint string
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

type FederatorZoneShare struct {
	// Operator ID of the federator whose zone is to be shared/unshared
	// required: true
	SelfOperatorId string
	// Country code of the federator whose zone is to be shared/unshared
	// required: true
	SelfCountryCode string
	// Operator ID of the federator with whom the zone is to be shared/unshared
	// required: true
	PartnerOperatorId string
	// Country code of the federator with whom the zone is to be shared/unshared
	// required: true
	PartnerCountryCode string
	// Zone ID to be shared/unshared with the partner federator
	// required: true
	ZoneId string
}

type FederatorZoneRegister struct {
	// Operator ID of the federator who wants to register/deregister a partner zone
	// required: true
	SelfOperatorId string
	// Country code of the federator who wants to register/deregister a partner zone
	// required: true
	SelfCountryCode string
	// Operator ID of the partner federator whose zone is to be registered/deregistered
	// required: true
	PartnerOperatorId string
	// Country code of the partner federator whose zone is to be registered/deregistered
	// required: true
	PartnerCountryCode string
	// Zone ID of the partner federator to be registered/deregistered
	// required: true
	ZoneId string
}
