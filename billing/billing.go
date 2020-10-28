package billing

import (
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

const CUSTOMER_TYPE_PARENT = "parent"
const CUSTOMER_TYPE_CHILD = "child"
const CUSTOMER_TYPE_SELF = "self"

type AccountInfo struct {
	OrgName           string `gorm:"primary_key;type:citext"`
	AccountId         string
	SubscriptionId    string
	ParentId          string
	Type              string
	LastManualRefresh time.Time
}

type CustomerDetails struct {
	OrgName   string
	FirstName string
	LastName  string
	Email     string
	CcEmails  string // comma separated list of additional emails
	Address1  string
	Address2  string
	City      string
	State     string
	Zip       string
	Country   string
	Phone     string
	Type      string // parent or child
	ParentId  string
}

type CreditCard struct {
	FirstName       string
	LastName        string
	CardNumber      string
	CardType        string
	ExpirationMonth int
	ExpirationYear  int
	BillingAddress  string
	BillingAddress2 string
	City            string
	State           string
	Zip             string
	Country         string
}

type UsageRecord struct {
	FlavorName  string
	NodeCount   int
	ClusterInst *edgeproto.ClusterInstKey
	AppInst     *edgeproto.AppInstKey
	StartTime   time.Time
	EndTime     time.Time
}

type BillingService interface {
	// Init is called once during startup
	Init() error
	// The Billing service's type ie. "chargify" or "zuora"
	GetType() string
	// Create Customer
	CreateCustomer(customer *CustomerDetails, account *AccountInfo) error
	// Delete Customer
	DeleteCustomer(account *AccountInfo) error
	// Update Customer
	UpdateCustomer(account *AccountInfo, customerDetails *CustomerDetails) error
	// Add a child to a parent
	AddChild(parentAccount, childAccount *AccountInfo, childDetails *CustomerDetails) error
	// Remove a child from a parent
	RemoveChild(child *AccountInfo) error
	// Records usage
	RecordUsage(account *AccountInfo, usageRecords []UsageRecord) error
}
