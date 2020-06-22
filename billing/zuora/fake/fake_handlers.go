package fake

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mobiledgex/edge-cloud-infra/billing/zuora"
)

func FakeAuth(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Error parsing input", http.StatusInternalServerError)
			return
		}
		id := req.FormValue("client_id")
		secret := req.FormValue("client_secret")
		if id != ClientID || secret != ClientSecret {
			http.Error(w, fmt.Sprintf("invalid fake credentials, id: %s, secret: %s", id, secret), http.StatusInternalServerError)
			return
		}
		token := zuora.OAuthToken{
			AccessToken: "faketoken",
			TokenType:   "Bearer",
			ExpiresIn:   600000, //10,000 minutes, so this will only have to be called once per test run
		}
		resp, err := json.Marshal(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}

func FakeOrder(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, fmt.Sprintf("Unsupported method for order request %s", req.Method), http.StatusBadRequest)
		return
	}
	var order zuora.CreateOrder
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// right now all of our zuora order calls only have one order per call
	if len(order.Subscriptions) != 1 || len(order.Subscriptions[0].OrderActions) != 1 {
		http.Error(w, "too many order actions in a single request", http.StatusBadRequest)
		return
	}
	action := order.Subscriptions[0].OrderActions[0]
	resp = zuora.OrderResp{}
	switch action.Type {
	case "CreateSubscription":
		if action.CreateSubscription == nil || action.CreateSubscription.NewSubscriptionOwnerAccount == nil {
			http.Error(w, "create subscription fields empty", http.StatusBadRequest)
			return
		}
		acccount = action.CreateSubscriptionOwnerAccount
		accountMux.Lock()
		subMux.Lock()
		accountId := strconv.Itoa(accountIdCounter)
		subId := strconv.Itoa(subIdCounter)
		subOwners[accountId] = subId
		accountIdCounter = accountIdCounter + 1
		subIdCounter = subIdCounter + 1
		subMux.Unlock()
		accountMux.Unlock()
		if account.ParentId != "" {
			// add it to the family
			familyMux.Lock()
			families[account.ParentId] = append(families[account.ParentId], accountId)
			familyMux.Unlock()
		}
		resp.Success = true
		resp.Status = "Completeted"
		resp.AccountNumber = accountId
		resp.SubscriptionNumbers = []string{subId}
	case "CancelSubscription":
		if action.CancelSubscription == nil {
			http.Error(w, "cancel subscription fields empty", http.StatusBadRequest)
			return
		}
		resp.Success = true
	case "AddProduct":
		if action.AddProduct == nil {
			http.Error(w, "add product fields empty", http.StatusBadRequest)
			return
		}
		if resp.Success = true
	default:
		http.Error(w, "unknown order type "+action.Type, http.StatusBadRequest)
		return
	}
	payload, err := json.Marhsal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func FakeAccounts(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, fmt.Sprintf("Unsupported method for order request %s", req.Method), http.StatusBadRequest)
		return
	}
	switch req.Method {
	case "POST": // create a new customer(Parent)
		var order zuora.NewAccount
		err := json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := zoura.AccountResp{}
		resp.Success = true
		accountMux.Lock()
		resp.AccountID = strconv.Itoa(accountIdCounter)
		resp.AccountNumber = resp.AccountId
		accountIdCounter = accountIdCounter + 1
		accountMux.Unlock()
		// only parent orgs are created here so setup an entry for them in the family map
		familyMux.Lock()
		families[resp.AccountNumber] = make([]string,0)
		familyMux.Unlock()
		payload, err := json.Marhsal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(payload)
	case "PUT": //update, verify the account exists then do nothing
		pathSplit := strings.Split(req.URL.Path, "/")
		// path should be of the form: /v1/accounts/<accountID>
		if len(pathSplit) < 3 {
			http.Error(w, "Unable to find accountNumber", http.StatusBadRequest)
			return
		}
		accountStr := pathSplit[len(pathSplit)-1]
		accountNum, err := strconv.Atoi(accountStr)
		if err != nil {
			http.Error(w, "error parsing accountNumber: "+ accountStr, http.StatusBadRequest)
			return
		}
		if accountNum < accountIdCounter {
			for _, deleted := range deletedAccounts {
				if deleted == accountNum {
					http.Error(w, "cannot update a deleted account", http.StatusBadRequest)
					return
				}
			}
		} else {
			http.Error(w, fmt.Sprintf("account %d does not exist", accountNum), http.StatusBadRequest)
			return
		}

		resp := zoura.GenericResp{Success: true}
		payload, err := json.Marhsal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(payload)
	case "GET":
		pathSplit := strings.Split(req.URL.Path, "/")
		// path should be of the form: /v1/accounts/<accountID>
		if len(pathSplit) < 3 {
			http.Error(w, "Unable to find accountNumber", http.StatusBadRequest)
			return
		}
		accountStr := pathSplit[len(pathSplit)-1]
		accountNum, err := strconv.Atoi(accountStr)
		if err != nil {
			http.Error(w, "error parsing accountNumber: "+ accountStr, http.StatusBadRequest)
			return
		}
		resp := zuora.GetAccount{
			Success: true,
			BasicInfo: zuora.BasicInfo{
				Id: accountNum,
				AccountNumber: accountNum,
			}
		}
		payload, err := json.Marhsal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(payload)
	default:
		http.Error(w, "unsupported http method: "+ req.Method, http.StatusBadRequest)
		return
	}
}

func FakeDeleteAccounts(w http.ResponseWriter, req *http.Request) {
	if req.Method != "DELETE" {
		http.Error(w, "unsupported http method: "+ req.Method, http.StatusBadRequest)
		return
	}
	pathSplit := strings.Split(req.URL.Path, "/")
	// path should be of the form: /v1/object/account/<accountID>
	if len(pathSplit) < 4 {
		http.Error(w, "Unable to find accountNumber", http.StatusBadRequest)
		return
	}
	accountStr := pathSplit[len(pathSplit)-1]
	accountNum, err := strconv.Atoi(accountStr)
	if err != nil {
		http.Error(w, "error parsing accountNumber: "+ accountStr, http.StatusBadRequest)
		return
	}
	deletedAccounts = append(deletedAccounts, accountNum)
}

func FakeGetSubs(w http.ResponseWriter, req *http.Request) {
	resp := zuora.CheckSubscriptions{
		Success: true,
		Subscriptions: []zuora.Subs{ zuora.Subs{
				RatePlans: []zuora.SubRatePlans {
					zuora.SubRatePlans{
						RatePlanCharges: []zuora.SubRatePlanCharges {
							// include the two rate plan numbers getRatePlanChargeId currently returns
							zuora.SubRatePlanCharges {
								ProductRatePlanChargeID: "2c92c0f8712986160171369e86d94ce9",
								Number: "Rate1",
							},
							zuora.SubRatePlanCharges {
								ProductRatePlanChargeID: "2c92c0f9712998b30171369c87bd3c44",
								Number: "Rate2",
							},
						},
					}
				},
			},
		},
	}
	payload, err := json.Marhsal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func FakeCheckSubOwner(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Unsupported http method: " + req.Method, http.StatusBadRequest)
		return
	}
	pathSplit := strings.Split(req.URL.Path, "/")
	// path should be of the form: /v1/subscriptions/<subID>
	if len(pathSplit) < 3 {
		http.Error(w, "Unable to find SubscriptionNumber", http.StatusBadRequest)
		return
	}
	subStr := pathSplit[len(pathSplit)-1]
	accountNum = subOwner[subStr]
	resp := zuora.GetSubscriptionByKey{
		Success: true,
		AccountId: accountNum.
		AccountNumber: accountNUm,
	}
	payload, err := json.Marhsal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}