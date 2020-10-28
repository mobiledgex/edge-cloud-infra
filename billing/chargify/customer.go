package chargify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/billing"
)

var customerEndpoint = "/customers"

func (bs *BillingService) CreateCustomer(customer *billing.CustomerDetails, account *billing.AccountInfo) error {
	newCustomer := Customer{
		FirstName:    customer.FirstName,
		LastName:     customer.LastName,
		Organization: customer.OrgName,
		Email:        customer.Email,
		CcEmails:     customer.CcEmails,
		Address:      customer.Address1,
		Address2:     customer.Address2,
		City:         customer.City,
		State:        customer.State,
		Zip:          customer.Zip,
		Country:      customer.Country,
		Phone:        customer.Phone,
	}
	if customer.Type == billing.CUSTOMER_TYPE_CHILD {
		parentId, err := strconv.Atoi(customer.ParentId)
		if err != nil {
			return fmt.Errorf("Unable to parse parentId: %v", err)
		}
		newCustomer.ParentId = parentId
	}

	resp, err := newChargifyReq("POST", "/customers.json", CustomerWrapper{Customer: &newCustomer})
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}
	custResp := CustomerWrapper{}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&custResp)
	if err != nil {
		errorResp := ErrorResp{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return fmt.Errorf("Error parsing response: %v\n", err)
		}
		combineErrors(&errorResp)
		return fmt.Errorf("Errors: %s", strings.Join(errorResp.Errors, ","))
	}

	// if its a self or child org, create subscription for it to the public_edge product
	if customer.Type != billing.CUSTOMER_TYPE_PARENT {
		newSub := Subscription{
			CustomerId:    strconv.Itoa(custResp.Customer.Id),
			ProductHandle: publicEdgeProductHandle,
		}

		// set the billing cycle to the first of the month
		y, m, _ := time.Now().UTC().Date()
		// yyyy-mm-dd format
		newSub.NextBillingAt = time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

		resp, err := newChargifyReq("POST", "/subscriptions.json", SubscriptionWrapper{Subscription: &newSub})
		if err != nil {
			return fmt.Errorf("Error sending request: %v\n", err)
		}
		subResp := SubscriptionWrapper{}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&subResp)
		if err != nil {
			errorResp := ErrorResp{}
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			if err != nil {
				return fmt.Errorf("Error parsing response: %v\n", err)
			}
			combineErrors(&errorResp)
			return fmt.Errorf("Errors: %s", strings.Join(errorResp.Errors, ","))
		}
		account.SubscriptionId = strconv.Itoa(subResp.Subscription.Id)
	}
	if customer.Type == billing.CUSTOMER_TYPE_CHILD {
		account.ParentId = strconv.Itoa(custResp.Customer.ParentId)
	}
	account.AccountId = strconv.Itoa(custResp.Customer.Id)
	account.Type = customer.Type
	return nil
}

// this doesnt actually delete the customer, what it does is cancels the subscription associated with the customer
// if we delete the customer, we also have to delete the subscription first which would result in losing the transaction history of that sub
func (bs *BillingService) DeleteCustomer(customer *billing.AccountInfo) error {
	switch customer.Type {
	case billing.CUSTOMER_TYPE_SELF:
		url := siteName + "/subscriptions/" + customer.SubscriptionId + "/delayed_cancel.json"
		resp, err := newChargifyReq("POST", url, nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return nil
		}
		defer resp.Body.Close()
		errorResp := ErrorResp{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return fmt.Errorf("Error parsing response: %v\n", err)
		}
		combineErrors(&errorResp)
		return fmt.Errorf("Errors: %s", strings.Join(errorResp.Errors, ","))

	case billing.CUSTOMER_TYPE_PARENT:
		url := siteName + "/subscription_groups/" + customer.SubscriptionId + "/cancel.json"
		resp, err := newChargifyReq("POST", url, SubscriptionGroupCancel{ChargeUnbilledUsage: true})
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return nil
		}
		defer resp.Body.Close()
		errorResp := ErrorResp{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return fmt.Errorf("Error parsing response: %v\n", err)
		}
		combineErrors(&errorResp)
		return fmt.Errorf("Errors: %s", strings.Join(errorResp.Errors, ","))

	case billing.CUSTOMER_TYPE_CHILD:
		// for some reason individual subscriptions in groups can only be put on hold, so just do that
		url := siteName + "/subscriptions/" + customer.SubscriptionId + "/hold.json"
		resp, err := newChargifyReq("POST", url, nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return nil
		}
		defer resp.Body.Close()
		errorResp := ErrorResp{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return fmt.Errorf("Error parsing response: %v\n", err)
		}
		combineErrors(&errorResp)
		return fmt.Errorf("Errors: %s", strings.Join(errorResp.Errors, ","))
	}
	return nil
}

func (bs *BillingService) UpdateCustomer(account *billing.AccountInfo, customerDetails *billing.CustomerDetails) error {
	update := Customer{ // any fields that actually contain a value will be the ones that are updated
		FirstName: customerDetails.FirstName,
		LastName:  customerDetails.LastName,
		Email:     customerDetails.Email,
		CcEmails:  customerDetails.CcEmails,
		Address:   customerDetails.Address1,
		Address2:  customerDetails.Address2,
		City:      customerDetails.City,
		State:     customerDetails.State,
		Zip:       customerDetails.Zip,
		Country:   customerDetails.Country,
		Phone:     customerDetails.Phone,
	}
	endpoint := "/customers/" + account.AccountId + ".json"
	resp, err := newChargifyReq("POST", endpoint, CustomerWrapper{Customer: &update})
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}
	custResp := CustomerWrapper{}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&custResp)
	if err != nil {
		errorResp := ErrorResp{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return fmt.Errorf("Error parsing response: %v\n", err)
		}
		combineErrors(&errorResp)
		return fmt.Errorf("Errors: %s", strings.Join(errorResp.Errors, ","))
	}

	return nil
}

func (bs *BillingService) AddChild(parentAccount, childAccount *billing.AccountInfo, childDetails *billing.CustomerDetails) error {
	// dont modify the existing struct
	childCopy := *childDetails
	childCopy.ParentId = parentAccount.AccountId
	err := bs.CreateCustomer(&childCopy, childAccount)
	if err == nil {
		childAccount.Type = billing.CUSTOMER_TYPE_CHILD
	}
	if err != nil {
		return err
	}
	// if this is the first child, get the subscription group uid
	if parentAccount.SubscriptionId == "" {
		baseURL, err := url.Parse(siteName + "/subscription_groups/lookup.json")
		if err != nil {
			return err
		}
		params := url.Values{}
		params.Add("subscription_id", "36734274")
		baseURL.RawQuery = params.Encode()
		resp, err := newChargifyReq("GET", baseURL.String(), nil)
		if err != nil {
			return err
		}
		group := SubscriptionGroup{}
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&group)
		if err != nil {
			errorResp := ErrorResp{}
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			if err != nil {
				return fmt.Errorf("Error parsing response: %v\n", err)
			}
			combineErrors(&errorResp)
			return fmt.Errorf("Errors: %s", strings.Join(errorResp.Errors, ","))
		}
		// ensure this is the right subscription group
		if strconv.Itoa(group.CustomerId) != parentAccount.AccountId {
			return fmt.Errorf("Error setting up subscription group for children")
		}
		parentAccount.SubscriptionId = group.Uid
	}
	return nil
}

func (bs *BillingService) RemoveChild(parent *billing.AccountInfo, child *billing.AccountInfo) error {
	return bs.DeleteCustomer(child)
}
