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
	// Status of the zone: Registered/DeRegistered
	Status int `json:"status"`
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

type OPZoneRegister struct {
	// Request id as sent in federation request
	RequestId string `json:"requestId"`
	// Globally unique string to identify an operator gMEC
	Operator string `json:"operator"`
	// ISO 3166-1 Alpha-2 code for the country where operator gMEC is located
	Country string `json:"country"`
	// Unique identifier for zone in the country of partner operator
	Zones []string `json:"zones"`
	// Origin OP federation ID
	OrigFederationId string `json:"origFederationId"`
	// Destination OP federation ID
	DestFederationId string `json:"destFederationId"`
}

// Resource details on a zone
type OPZoneResourceInfo struct {
	// Total maximum cpu physical cores that can be allocated for lead operator user apps
	CPU int64 `json:"cpu"`
	// Total maximum memory (GBs) that can be allocated for lead operator user apps
	RAM int64 `json:"ram"`
	// Total maximum disk (Gbs) that can be allocated for lead operator user apps
	Disk int64 `json:"disk"`
	// Total maximum gpu that can be allocated for lead operator user apps
	GPU int64 `json:"gpu"`
}

type OPZoneRegisterDetails struct {
	// Globally Unique identifier of the zone
	ZoneId string `json:"zoneId"`
	// Federation key of operator responding the request
	RegistrationToken string `json:"registrationToken"`
	// Guaranteed resource details on a zone
	GuaranteedResources OPZoneResourceInfo `json:"guaranteedResources"`
	// Upper limit quota of resources in a zone
	UpperLimitQuota OPZoneResourceInfo `json:"upperLimitQuota"`
}

type OPZoneRegisterResponse struct {
	// Request id as sent in federation request
	RequestId string `json:"requestId"`
	// Globally unique string to identify an operator gMEC
	LeadOperatorId string `json:"leadOperatorId"`
	// Globally unique string to identify an operator gMEC
	PartnerOperatorId string `json:"partnerOperatorId"`
	// Federation ID
	FederationId string `json:"federationId"`
	// Partner gMEC zone details
	Zone OPZoneRegisterDetails `json:"zone"`
}

type OPZoneDeRegister struct {
	// Request id as sent in federation request
	RequestId string `json:"requestId"`
	// Globally unique string to identify an operator gMEC
	Operator string `json:"operator"`
	// ISO 3166-1 Alpha-2 code for the country where operator gMEC is located
	Country string `json:"country"`
	// Zone identifier of partner operator. This zone will be de-registered
	Zone string `json:"zone"`
	// Origin OP federation ID
	OrigFederationId string `json:"origFederationId"`
	// Destination OP federation ID
	DestFederationId string `json:"destFederationId"`
}
