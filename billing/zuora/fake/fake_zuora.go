package fake

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/mobiledgex/edge-cloud-infra/billing/zuora"
)

var BillingPath = "fake"
var ClientID = "fakeID"
var ClientSecret = "fakeSecret"
var URL = "127.0.0.1:65122"

// for accounts and subs ids and nums will be the same
var accountMux sync.Mutex
var subMux sync.Mutex
var familyMux sync.Mutex
var accountIdCounter int
var subIdCounter int
var families map[string][]string
var deletedAccounts []string
var subOwners map[string]string // K: sub, V: acc

func RunFakeZuora() {
	// initialize variables
	accountIdCounter = 0
	subIdCounter = 0
	families = make(map[string][]string)
	deletedAccounts = make([]string, 0)
	subOwners = make(map[string]string)

	go runServer()
}

func runServer() {
	http.HandleFunc(zuora.OAuthEndpoint+"/", FakeAuth)
	http.HandleFunc(zuora.OrdersEndpoint+"/", FakeOrder)
	http.HandleFunc(zuora.AcccountsEndpoint+"/", FakeAccounts)
	http.HandleFunc(zuora.ObjectAccountsEndpoint+"/", FakeDeleteAccounts)
	http.HandleFunc(zuora.GetSubscriptionEndpoint+"/", FakeGetSubs)
	http.HandleFunc("/v1/subscriptions/", FakeCheckSubOwner)

	http.ListenAndServe(":65121", nil)
}
