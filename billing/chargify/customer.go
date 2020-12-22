package chargify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mobiledgex/edge-cloud-infra/billing"
	"github.com/mobiledgex/edge-cloud/log"
)

var customerEndpoint = "/customers"

func (bs *BillingService) CreateCustomer(ctx context.Context, customer *billing.CustomerDetails, account *billing.AccountInfo, payment *billing.PaymentMethod) error {
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
	newPaymentSpecified := false
	if payment != nil && payment.PaymentType != "" {
		newPaymentSpecified = true
	}

	if customer.Type == billing.CUSTOMER_TYPE_CHILD {
		parentId, err := strconv.Atoi(customer.ParentId)
		fmt.Printf("creating child under: %d", parentId)
		if err != nil {
			return fmt.Errorf("Unable to parse parentId: %v", err)
		}
		newCustomer.ParentId = parentId
	} else if customer.Type == billing.CUSTOMER_TYPE_PARENT && !newPaymentSpecified {
		return fmt.Errorf("Parent type customers must have a payment profile specified")
	}

	resp, err := newChargifyReq("POST", "/customers.json", CustomerWrapper{Customer: &newCustomer})
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return getReqErr(resp.Body)
	}
	custResp := CustomerWrapper{}
	err = json.NewDecoder(resp.Body).Decode(&custResp)
	if err != nil {
		return fmt.Errorf("Error parsing response: %v\n", err)
	}

	paymentProfileId := 0
	if newPaymentSpecified {
		paymentProfileId, err = addPayment(custResp.Customer.Id, payment)
		if err != nil {
			endpoint := "/customers" + strconv.Itoa(custResp.Customer.Id) + ".json"
			resp, undoErr := newChargifyReq("DELETE", endpoint, nil)
			if undoErr != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Error undoing account creation", "err", err, "undoErr", undoErr)
				return fmt.Errorf("Error creating payment profile: %v", err)
			}
			if resp.StatusCode == http.StatusNoContent {
				return fmt.Errorf("Error creating payment profile: %v", err)
			}
			undoErr = getReqErr(resp.Body)
			log.SpanLog(ctx, log.DebugLevelInfo, "Error undoing account creation", "err", err, "undoErr", undoErr)
			return fmt.Errorf("Error creating payment profile: %v", err)
		}
	} else if payment.PaymentProfile != 0 {
		paymentProfileId = payment.PaymentProfile
	}

	// if its a self or child org, create subscription for it to the public_edge product
	if customer.Type != billing.CUSTOMER_TYPE_PARENT {
		newSub := Subscription{
			CustomerId:    strconv.Itoa(custResp.Customer.Id),
			ProductHandle: publicEdgeProductHandle,
		}

		// TODO: remove this and the function when we no longer offer free trials
		addFreeTrial(&newSub)

		if paymentProfileId == 0 {
			newSub.PaymentCollectionMethod = "invoice"
		} else {
			newSub.PaymentProfileId = strconv.Itoa(paymentProfileId)
		}

		// set the billing cycle to the first of the month
		y, m, _ := time.Now().UTC().Date()
		// yyyy-mm-dd format
		newSub.NextBillingAt = time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

		resp, err := newChargifyReq("POST", "/subscriptions.json", SubscriptionWrapper{Subscription: &newSub})
		if err != nil {
			return fmt.Errorf("Error sending request: %v\n", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			return getReqErr(resp.Body)
		}
		subResp := SubscriptionWrapper{}
		err = json.NewDecoder(resp.Body).Decode(&subResp)
		if err != nil {
			return fmt.Errorf("Error parsing response: %v\n", err)
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

// This function is temporary, adds a promotion to the sub that is 100% off
func addFreeTrial(sub *Subscription) {
	sub.CouponCode = "FREETRIALS"
}

// this doesnt actually delete the customer, what it does is cancels the subscription associated with the customer
// if we delete the customer, we also have to delete the subscription first which would result in losing the transaction history of that sub
func (bs *BillingService) DeleteCustomer(ctx context.Context, customer *billing.AccountInfo) error {
	switch customer.Type {
	case billing.CUSTOMER_TYPE_SELF:
		endpoint := "/subscriptions/" + customer.SubscriptionId + "/delayed_cancel.json"
		fmt.Printf("endpoint: %s\n", endpoint)
		resp, err := newChargifyReq("POST", endpoint, nil)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		defer resp.Body.Close()
		return getReqErr(resp.Body)

	case billing.CUSTOMER_TYPE_PARENT:
		endpoint := "/subscription_groups/" + customer.SubscriptionId + "/cancel.json"
		resp, err := newChargifyReq("POST", endpoint, SubscriptionGroupCancel{ChargeUnbilledUsage: true})
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		defer resp.Body.Close()
		return getReqErr(resp.Body)

	case billing.CUSTOMER_TYPE_CHILD:
		// for some reason individual subscriptions in groups can only be put on hold, so just do that
		endpoint := "/subscriptions/" + customer.SubscriptionId + "/hold.json"
		resp, err := newChargifyReq("POST", endpoint, nil)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		defer resp.Body.Close()
		return getReqErr(resp.Body)
	}
	return nil
}

func (bs *BillingService) UpdateCustomer(ctx context.Context, account *billing.AccountInfo, customerDetails *billing.CustomerDetails) error {
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
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return getReqErr(resp.Body)
	}

	return nil
}

func (bs *BillingService) AddChild(ctx context.Context, parentAccount, childAccount *billing.AccountInfo, childDetails *billing.CustomerDetails) error {
	// dont modify the existing struct
	childCopy := *childDetails
	childCopy.ParentId = parentAccount.AccountId
	childCopy.Type = billing.CUSTOMER_TYPE_CHILD
	err := bs.CreateCustomer(ctx, &childCopy, childAccount, &billing.PaymentMethod{PaymentProfile: parentAccount.DefaultPaymentProfile})
	if err != nil {
		return err
	}
	childAccount.Type = billing.CUSTOMER_TYPE_CHILD
	// if this is the first child, get the subscription group uid
	if parentAccount.SubscriptionId == "" {
		endpoint, err := url.Parse("/subscription_groups/lookup.json")
		if err != nil {
			return err
		}
		params := url.Values{}
		params.Add("subscription_id", childAccount.SubscriptionId)
		endpoint.RawQuery = params.Encode()
		resp, err := newChargifyReq("GET", endpoint.String(), nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return getReqErr(resp.Body)
		}
		group := SubscriptionGroup{}
		err = json.NewDecoder(resp.Body).Decode(&group)
		if err != nil {
			return fmt.Errorf("Error parsing response: %v\n", err)
		}
		// ensure this is the right subscription group
		if strconv.Itoa(group.CustomerId) != parentAccount.AccountId {
			return fmt.Errorf("Error setting up subscription group for children")
		}
		parentAccount.SubscriptionId = group.Uid
	}
	return nil
}

func (bs *BillingService) RemoveChild(ctx context.Context, parent, child *billing.AccountInfo) error {
	return bs.DeleteCustomer(ctx, child)
}
