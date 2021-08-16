package ormapi

// GORM objects
// ============
type OperatorFederation struct {
	// Globally unique string used to authenticate operations over federation interface
	FederationId string `gorm:"primary_key"`
	// Federation access point address
	FederationAddr string `json:",omitempty"`
	// Owner of this object, self or partner
	Type string `json:",omitempty"`
	// Globally unique string to identify an operator gMEC
	// required: true
	OperatorId string `gorm:"unique;not null"`
	// ISO 3166-1 Alpha-2 code for the country where operator gMEC is located
	// required: true
	CountryCode string `gorm:"unique;not null"`
	// Mobile country code of operator sending the request
	MCC string `json:"MCC"`
	// Comma separated list of mobile network codes of operator sending the request
	MNCs string `json:"MNCs"`
	// IP and Port of discovery service URL of gMEC
	LocatorEndPoint string `json:"locatorEndPoint"`
}

type OperatorZoneCloudlet struct {
	ZoneId string `gorm:"primary_key"`
	// Cloudlet name belonging to the federation zone
	CloudletName string `gorm:"unique;not null"`
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
	GeoLocation string
	// Comma seperated list of cities under this zone
	City string
	// Comma seperated list of states under this zone
	State string
	// Type of locality eg rural, urban etc.
	Locality string
	// List of cloudlets belonging to the federation zone
	Cloudlets []string
}

// Federation API Definitions
// ==========================
type OPRegistrationRequest struct {
	// Request ID for tracking federation request
	RequestId string `json:"requestId"`
	// Globally unique string to identify an operator gMEC
	// required: true
	OperatorId string `json:"operator"`
	// ISO 3166-1 Alpha-2 code for the country where operator gMEC is located
	// required: true
	CountryCode string `json:"country"`
	// Origin OP federation ID
	OrigFederationId string `json:"origFederationId"`
	// Destination OP federation ID
	DestFederationId string `json:"destFederationId"`
	// If partner gMEC shall endorse lead gMEC applications
	ApplicationAnchormentReq bool `json:"applicationAnchormentReq"`
}

type OPRegistrationResponse struct {
	// Request id as sent in federation request
	RequestId string `json:"requestId"`
	// Globally unique string to identify an operator gMEC
	OrigOperatorId string `json:"origOperatorId"`
	// Globally unique string to identify an operator gMEC
	PartnerOperatorId string `json:"partnerOperatorId"`
	// Origin OP federation ID
	OrigFederationId string `json:"origFederationId"`
	// Destination OP federation ID
	DestFederationId string `json:"destFederationId"`
	// Mobile country code of operator sending the request
	MCC string `json:"MCC"`
	// Mobile network codes of operator sending the request
	MNC []string `json:"MNC"`
	// IP and Port of discovery service URL of gMEC
	LocatorEndPoint string `json:"locatorEndPoint"`
	// List of zones partner gMEC is willing to share
	PartnerZone []OPZoneInfo `json:"partnerZone"`
}

type OPZoneInfo struct {
	// Globally Unique identifier of the zone
	ZoneId string `json:"zoneId"`
	// GPS co-ordinates associated with the zone (in decimal format)
	GeoLocation string `json:"geoLocation"`
	// Comma seperated list of cities under this zone
	City string `json:"city"`
	// Comma seperated list of states under this zone
	State string `json:"state"`
	// Type of locality eg rural, urban etc.
	Locality string `json:"locality"`
	// Number of edges in the zone
	EdgeCount int `json:"edgeCount"`
}
