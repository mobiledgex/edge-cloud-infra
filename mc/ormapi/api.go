package ormapi

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

// Data saved to persistent sql db, also used for API calls

type TOTPAuthType string

var (
	TOTPAuthenticator TOTPAuthType = "authenticator"
	TOTPEmail         TOTPAuthType = "email"
)

type User struct {
	// User name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen
	// required: true
	Name string `gorm:"primary_key;type:citext"`
	// User email
	Email string `gorm:"unique;not null"`
	// read only: true
	EmailVerified bool
	// read only: true
	Passhash string `gorm:"not null"`
	// read only: true
	Salt string `gorm:"not null"`
	// read only: true
	Iter int `gorm:"not null"`
	// Family Name
	FamilyName string
	// Given Name
	GivenName string
	// read only: true
	Picture string
	// Nick Name
	Nickname string
	// read only: true
	CreatedAt time.Time `json:",omitempty"`
	// read only: true
	UpdatedAt time.Time `json:",omitempty"`
	// read only: true
	Locked bool
	// read only: true
	PassEntropy float64
	// read only: true
	PassCrackTimeSec float64
	// read only: true
	TOTPType TOTPAuthType
	// read only: true
	DisableTOTP bool
	// read only: true
	TOTPSharedKey string
	// read only: true
	TOTPQRImage []byte
}

type UserResponse struct {
	Message       string
	TOTPSharedKey string
	TOTPQRImage   []byte
}

type Organization struct {
	// Organization name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen
	// required: true
	Name string `gorm:"primary_key;type:citext"`
	// Organization type: "developer" or "operator"
	Type string `gorm:"not null"`
	// Organization address
	Address string `json:",omitempty"`
	// Organization phone number
	Phone string `json:",omitempty"`
	// read only: true
	CreatedAt time.Time `json:",omitempty"`
	// read only: true
	UpdatedAt time.Time `json:",omitempty"`
	// read only: true
	PublicImages bool `json:",omitempty"`
	// read only: true
	DeleteInProgress bool `json:",omitempty"`
	// read only: true
	Parent string `json:",omitempty"`
}

type BillingOrganization struct {
	// BillingOrganization name. Can only contain letters, digits, underscore, period, hyphen. It cannot have leading or trailing spaces or period. It cannot start with hyphen
	// required: true
	Name string `gorm:"primary_key;type:citext"`
	// Organization type: "parent" or "self"
	Type string `gorm:"not null"`
	// Billing Info First Name
	FirstName string `json:",omitempty"`
	// Billing info last name
	LastName string `json:",omitempty"`
	// Organization email
	Email string `json:",omitempty"`
	// Organization address
	Address string `json:",omitempty"`
	// Organization city
	City string `json:",omitempty"`
	// Organization Country
	Country string `json:",omitempty"`
	// Organization State
	State string `json:",omitempty"`
	// Organization Postal code
	PostalCode string `json:",omitempty"`
	// Organization phone number
	Phone string `json:",omitempty"`
	// Currency
	Currency string `json:",omitempty"` // currently only allow "USD"
	// Children belonging to this BillingOrganization
	Children string `json:",omitempty"`
	// read only: true
	CreatedAt time.Time `json:",omitempty"`
	// read only: true
	UpdatedAt time.Time `json:",omitempty"`
	// read only: true
	DeleteInProgress bool `json:",omitempty"`
}

type Controller struct {
	Region    string    `gorm:"primary_key"`
	Address   string    `gorm:"unique;not null"`
	InfluxDB  string    `gorm:"type:text"`
	CreatedAt time.Time `json:",omitempty"`
	UpdatedAt time.Time `json:",omitempty"`
}

type Config struct {
	// read only: true
	ID int `gorm:"primary_key;auto_increment:false"`
	// Lock new accounts (must be unlocked by admin)
	LockNewAccounts bool
	// Email to notify when locked account is created
	NotifyEmailAddress string
	// Skip email verification for new accounts (testing only)
	SkipVerifyEmail bool
	// User accounts min password crack time seconds (a measure of strength)
	PasswordMinCrackTimeSec float64
	// Admin accounts min password crack time seconds (a measure of strength)
	AdminPasswordMinCrackTimeSec float64
}

type OrgCloudletPool struct {
	Org             string `gorm:"type:citext REFERENCES organizations(name)"`
	Region          string `gorm:"type:text REFERENCES controllers(region)"`
	CloudletPool    string `gorm:"not null"`
	CloudletPoolOrg string `gorm:"type:citext REFERENCES organizations(name)"`
}

// Structs used for API calls

type RolePerm struct {
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type Role struct {
	Org      string `form:"org" json:"org"`
	Username string `form:"username" json:"username"`
	Role     string `form:"role" json:"role"`
}

type OrgCloudlet struct {
	Region string `json:"region,omitempty"`
	Org    string `form:"org" json:"org"`
}

type UserLogin struct {
	// User's name or email address
	// required: true
	Username string `form:"username" json:"username"`
	// User's password
	// required: true
	Password string `form:"password" json:"password"`
	// read only: true
	TOTP string `form:"totp" json:"totp"`
	// Fallback to email to get TOTP
	// read only: true
	EmailTOTP bool `form:"emailtotp" json:"emailtotp"`
}

type NewPassword struct {
	Password string `form:"password" json:"password"`
}

type CreateUser struct {
	User   `json:",inline"`
	Verify EmailRequest `json:"verify"` // for verifying email
}

type AuditQuery struct {
	Username       string `json:"username"`
	Org            string `form:"org" json:"org"`
	Limit          int    `json:"limit"`
	util.TimeRange `json:",inline"`
	Operation      string            `json:"operation"`
	Tags           map[string]string `json:"tags"`
}

type AuditResponse struct {
	OperationName string               `json:"operationname"`
	Username      string               `json:"username"`
	Org           string               `json:"org"`
	ClientIP      string               `json:"clientip"`
	Status        int                  `json:"status"`
	StartTime     TimeMicroseconds     `json:"starttime"`
	Duration      DurationMicroseconds `json:"duration"`
	Request       string               `json:"request"`
	Response      string               `json:"response"`
	Error         string               `json:"error"`
	TraceID       string               `json:"traceid"`
	Tags          map[string]string    `json:"tags"`
}

// Email request is used for password reset and to resend welcome
// verification email. It contains the information need to send
// some kind of email to the user.
type EmailRequest struct {
	// read only: true
	Email string `form:"email" json:"email"`
	// read only: true
	OperatingSystem string `form:"operatingsystem" json:"operatingsystem"`
	// read only: true
	Browser string `form:"browser" json:"browser"`
	// Callback URL to verify user email
	CallbackURL string `form:"callbackurl" json:"callbackurl"`
	// read only: true
	ClientIP string `form:"clientip" json:"clientip"`
}

type PasswordReset struct {
	// Authentication token
	// required: true
	Token string `form:"token" json:"token"`
	// User's new password
	// required: true
	Password string `form:"password" json:"password"`
}

type Token struct {
	// Authentication token
	Token string `form:"token" json:"token"`
}

// Structs used in replies

type Result struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type Version struct {
	BuildMaster string `json:"buildmaster,omitempty"`
	BuildHead   string `json:"buildhead,omitempty"`
	BuildAuthor string `json:"buildauthor,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
}

// Data struct sent back for streaming (chunked) commands.
// Contains a data payload for incremental data, and a result
// payload for an error result. Only one of the two will be used
// in each chunk.

type StreamPayload struct {
	Data   interface{} `json:"data,omitempty"`
	Result *Result     `json:"result,omitempty"`
}

type WSStreamPayload struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// all data is for full create/delete

type AllData struct {
	Controllers      []Controller          `json:"controllers,omitempty"`
	BillingOrgs      []BillingOrganization `json:"billingorgs,omitempty"`
	AlertReceivers   []AlertReceiver       `json:"alertreceivers,omitempty"`
	Orgs             []Organization        `json:"orgs,omitempty"`
	Roles            []Role                `json:"roles,omitempty"`
	OrgCloudletPools []OrgCloudletPool     `json:"orgcloudletpools,omitempty"`
	RegionData       []RegionData          `json:"regiondata,omitempty"`
}

type RegionData struct {
	Region  string            `json:"region,omitempty"`
	AppData edgeproto.AllData `json:"appdata,omitempty"`
}

// Metrics data
type AllMetrics struct {
	Data []MetricData `json:"data"`
}

type MetricData struct {
	Series []MetricSeries `json:"Series"`
}

type MetricSeries struct {
	Columns []string        `json:"columns"`
	Name    string          `json:"name"`
	Values  [][]interface{} `json:"values"`
}

type RegionAppInstMetrics struct {
	Region    string
	AppInst   edgeproto.AppInstKey
	Selector  string
	StartTime time.Time `json:",omitempty"`
	EndTime   time.Time `json:",omitempty"`
	Last      int       `json:",omitempty"`
}

type RegionClusterInstMetrics struct {
	Region      string
	ClusterInst edgeproto.ClusterInstKey
	Selector    string
	StartTime   time.Time `json:",omitempty"`
	EndTime     time.Time `json:",omitempty"`
	Last        int       `json:",omitempty"`
}

type RegionCloudletMetrics struct {
	Region    string
	Cloudlet  edgeproto.CloudletKey
	Selector  string
	StartTime time.Time `json:",omitempty"`
	EndTime   time.Time `json:",omitempty"`
	Last      int       `json:",omitempty"`
}

type RegionClientMetrics struct {
	Region    string
	AppInst   edgeproto.AppInstKey
	Method    string `json:",omitempty"`
	CellId    int    `json:",omitempty"`
	Selector  string
	StartTime time.Time `json:",omitempty"`
	EndTime   time.Time `json:",omitempty"`
	Last      int       `json:",omitempty"`
}

type RegionAppInstEvents struct {
	Region    string
	AppInst   edgeproto.AppInstKey
	StartTime time.Time `json:",omitempty"`
	EndTime   time.Time `json:",omitempty"`
	Last      int       `json:",omitempty"`
}

type RegionClusterInstEvents struct {
	Region      string
	ClusterInst edgeproto.ClusterInstKey
	StartTime   time.Time `json:",omitempty"`
	EndTime     time.Time `json:",omitempty"`
	Last        int       `json:",omitempty"`
}

type RegionCloudletEvents struct {
	Region    string
	Cloudlet  edgeproto.CloudletKey
	StartTime time.Time `json:",omitempty"`
	EndTime   time.Time `json:",omitempty"`
	Last      int       `json:",omitempty"`
}

type RegionAppInstUsage struct {
	Region    string
	AppInst   edgeproto.AppInstKey
	StartTime time.Time `json:",omitempty"`
	EndTime   time.Time `json:",omitempty"`
	VmOnly    bool      `json:",omitempty"`
}

type RegionClusterInstUsage struct {
	Region      string
	ClusterInst edgeproto.ClusterInstKey
	StartTime   time.Time `json:",omitempty"`
	EndTime     time.Time `json:",omitempty"`
}

type RegionCloudletPoolUsage struct {
	Region       string
	CloudletPool edgeproto.CloudletPoolKey
	StartTime    time.Time `json:",omitempty"`
	EndTime      time.Time `json:",omitempty"`
}

type RegionCloudletPoolUsageRegister struct {
	Region          string
	CloudletPool    edgeproto.CloudletPoolKey
	UpdateFrequency time.Duration
	PushEndpoint    string
	StartTime       time.Time
}

// Configurable part of AlertManager Receiver
type AlertReceiver struct {
	// Receiver Name
	Name string
	// Receiver type. Eg. email, slack, pagerduty
	Type string
	// Alert severity filter
	Severity string
	// User that created this receiver
	User string `json:",omitempty"`
	// Custom receiving email
	Email string `json:",omitempty"`
	// Custom slack channel
	SlackChannel string `json:",omitempty"`
	// Custom slack webhook
	SlackWebhook string `json:",omitempty"`
	// Cloudlet spec for alerts
	Cloudlet edgeproto.CloudletKey `json:",omitempty"`
	// AppInst spec for alerts
	AppInst edgeproto.AppInstKey `json:",omitempty"`
}
