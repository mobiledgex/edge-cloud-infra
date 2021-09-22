package ormapi

// GORM objects
// ============
type OperatorFederation struct {
	// Globally unique string used to authenticate operations over federation interface
	FederationId string `gorm:"primary_key"`
	// Federation access point address
	FederationAddr string `json:",omitempty"`
	// Type of the federation
	Type string `json:",omitempty"`
	// Role of the federation
	Role string `json:",omitempty"`
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string
	// Mobile country code of operator sending the request
	MCC string `json:"MCC"`
	// Comma separated list of mobile network codes of operator sending the request
	MNCs string `json:"MNCs"`
	// IP and Port of discovery service URL of OP
	LocatorEndPoint string `json:"locatorEndPoint"`
}

type OperatorZoneCloudlet struct {
	ZoneId string `gorm:"primary_key"`
	// Cloudlet name belonging to the federation zone
	CloudletName string `gorm:"unique;not null"`
}

type OperatorRegisteredZone struct {
	// Globally unique string used to authenticate operations over federation interface
	// required: true
	ZoneId string `gorm:"primary_key"`
	// Owner ID of the zone
	FederationId string `gorm:"primary_key"`
	// Globally unique string to identify an operator platform
	// required: true
	OperatorId string
	// ISO 3166-1 Alpha-2 code for the country where operator platform is located
	// required: true
	CountryCode string
}

type OperatorZone struct {
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
}

// API Objects
// ===========
type OperatorZoneCloudletMap struct {
	// Owner ID of the zone
	FederationId string
	// Globally unique string used to authenticate operations over federation interface
	// required: true
	ZoneId string
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
	// Zone is registered by partner OP or not
	RegisteredOPs []string
}
