package ormapi

// GORM objects
// ============
type Federator struct {
	// Globally unique string used to authenticate operations over federation interface
	FederationId string `gorm:"primary_key"`
	// Federation access point address
	FederationAddr string `json:",omitempty"`
	// Type of the federation
	Type string `json:",omitempty"`
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string
	// Comma separated set of regions an operator belongs to
	Regions string
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string
	// Mobile country code of operator sending the request
	MCC string
	// Comma separated list of mobile network codes of operator sending the request
	MNCs string
	// IP and Port of discovery service URL of OP
	LocatorEndPoint string
}

type FederatorRole struct {
	// Self federation ID of the operator who is a federation partner with the partner federator
	SelfFederationId string `gorm:"primary_key"`
	// Partner federation ID to be associated with self federator
	PartnerFederationId string `gorm:"primary_key"`
	// Role of the partner federator
	Role string `json:",omitempty"`
}

// Zone owned by a Federator. MC defines a zone as a group of cloudlets,
// but currently it is restricted to one cloudlet
type FederatorZone struct {
	// Globally unique string used to authenticate operations over federation interface
	// required: true
	ZoneId string `gorm:"primary_key"`
	// Owner ID of the zone
	FederationId string `json:"federationId"`
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
	// List of cloudlets part of this zone (delineated by comma)
	Cloudlets string `json:"cloudlet"`
}

// Information of the Federator with whom the zone is shared
type FederatorSharedZone struct {
	// Globally unique string used to authenticate operations over federation interface
	// required: true
	ZoneId string `gorm:"primary_key"`
	// Federation ID of the partner with whom this zone is shared
	FederationId string `gorm:"primary_key"`
	// Operator ID of the partner with whom this zone is shared
	// required: true
	OperatorId string
	// Country code of the partner with whom this zone is shared
	// required: true
	CountryCode string
}

// Information of the Federator who has registered the zone
type FederatorRegisteredZone struct {
	// Globally unique string used to authenticate operations over federation interface
	// required: true
	ZoneId string `gorm:"primary_key"`
	// Federation ID of the partner who has registered the zone
	FederationId string `gorm:"primary_key"`
	// Operator ID of the partner who has registered the zone
	// required: true
	OperatorId string
	// Country code of the partner who has registered the zone
	// required: true
	CountryCode string
}

// API Objects
// ===========
type FederatorRequest struct {
	// Globally unique string used to authenticate operations over federation interface
	FederationId string
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

type FederatorPartnerRequest struct {
	// Self federator's operator ID
	// required: true
	SelfOperatorId string
	// Self federator's country code
	// required: true
	SelfCountryCode string
	// Partner federator's operator ID
	// required: true
	PartnerOperatorId string
	// Partner federator's country code
	// required: true
	PartnerCountryCode string
	// Partner federator's federation ID
	PartnerFederationId string
	// Partner federation access point address
	PartnerFederationAddr string
}

type FederatorRoleResponse struct {
	// Partner federator's operator ID
	PartnerOperatorId string
	// Partner federator's country code
	PartnerCountryCode string
	// Partner federator's role
	PartnerRole string
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
	RegisteredByOPs []string
	// List of partner federators with whom this zone is shared
	SharedWithOPs []string
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
