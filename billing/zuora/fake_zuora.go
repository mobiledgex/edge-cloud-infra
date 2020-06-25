package zuora

import (
	"net/http"
	"sync"
)

var fakeBillingPath = "fake"
var fakeClientID = "fakeID"
var fakeClientSecret = "fakeSecret"
var fakeURL = "127.0.0.1:65122"

// for accounts and subs ids and nums will be the same
var accountMux sync.Mutex
var subMux sync.Mutex
var familyMux sync.Mutex
var accountIdCounter int
var subIdCounter int
var families map[string][]string
var deletedAccounts []string
var subOwners map[string]string // K: sub, V: acc

func runFakeZuora() {
	// initialize variables
	accountIdCounter = 0
	subIdCounter = 0
	families = make(map[string][]string)
	deletedAccounts = make([]string, 0)
	subOwners = make(map[string]string)

	go runServer()
}

func runServer() {
	http.HandleFunc(OAuthEndpoint+"/", fakeAuth)
	http.HandleFunc(OrdersEndpoint+"/", fakeOrder)
	http.HandleFunc(AccountsEndPoint+"/", fakeAccounts)
	http.HandleFunc(ObjectAccountsEndpoint+"/", fakeDeleteAccounts)
	http.HandleFunc(GetSubscriptionEndpoint+"/", fakeGetSubs)
	http.HandleFunc("/v1/subscriptions/", fakeCheckSubOwner)
	http.HandleFunc(UsageEndpoint, fakeUsageRecord)

	http.ListenAndServe(":65122", nil)
}
