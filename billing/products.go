package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var usageFlavorRatePlanId = "2c92c0f9712998a401712de88cc44c9f"

// Creates a product rate plan for the flavors product
func CreateFlavor(FlavorName string) (string, error) {
	name := FlavorName
	description := "TODO: list flavor details here"
	productId := FlavorProductId
	return createProductRatePlan(name, description, productId)
}

func SetFlavorPrice(flavorId, pricingTier string, price float64) (string, error) {
	if price < 0 {
		return "", fmt.Errorf("Invalid Price: %f", price)
	}
	return createProductRatePlanCharge(flavorId, pricingTier, "%s pricing plan", PER_UNIT_PRICING, USAGE, price, MINUTE)
}

// Creates a product and returns the product id
// This should be rarely called, you probably want createProductRatePlan or createRatePlan
func createProduct(name, description string) (string, error) {
	productReq := CreateProductReq{
		Name:               name,
		Description:        description,
		EffectiveStartDate: StartDate,
		EffectiveEndDate:   EndDate,
	}
	payload, err := json.Marshal(productReq)
	if err != nil {
		return "", fmt.Errorf("Could not marshal %s, err: %v", productReq, err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", ZuoraUrl+ProductEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v\n", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", Token.TokenType+" "+Token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v\n", err)
	}
	productResp := GenericResp{}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&productResp)
	if err != nil {
		return "", fmt.Errorf("Error parsing response: %v\n", err)
	}
	if !productResp.Success {
		return "", fmt.Errorf("Error creating product")
	}
	return productResp.ID, nil
}

func createProductRatePlan(name, description, productId string) (string, error) {
	reqBody := CreateProductRatePlanReq{
		Name:        name,
		Description: description,
		ProductID:   productId,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("Could not marshal %+v, err: %v", reqBody, err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", ZuoraUrl+ProductRatePlanEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v\n", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", Token.TokenType+" "+Token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v\n", err)
	}
	productResp := GenericResp{}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&productResp)
	if err != nil {
		return "", fmt.Errorf("Error parsing response: %v\n", err)
	}
	if !productResp.Success {
		return "", fmt.Errorf("Error creating flavor")
	}
	return productResp.ID, nil
}

// only PER_UNIT_PRICING and FLAT_FEE_PRICING are supported right now
// we might not even need this, rateplancharges might end up being created manually
func createProductRatePlanCharge(ratePlanId, name, description string, chargeModel ChargeModel, chargeType ChargeType, listPrice float64, uom UOM) (string, error) {
	reqBody := CreateProductRatePlanChargeReq{
		ProductRatePlanID:      ratePlanId,
		Name:                   name,
		Description:            description,
		Model:                  chargeModel,
		Type:                   chargeType,
		ListPrice:              listPrice,
		Uom:                    uom,
		BillingPeriod:          "Month",
		TriggerEvent:           "ServiceActivation",
		EndDateCondition:       "SubscriptionEnd",
		BillingPeriodAlignment: "AlignToSubscriptionStart",
		BillCycleType:          "SubscriptionStartDay",
	}
	if chargeModel == PER_UNIT_PRICING {
		reqBody.RatingGroup = "ByUsageRecord"
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("Could not marshal %+v, err: %v", reqBody, err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", ZuoraUrl+ProductRatePlanChargeEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v\n", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", Token.TokenType+" "+Token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v\n", err)
	}
	productResp := GenericResp{}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&productResp)
	if err != nil {
		return "", fmt.Errorf("Error parsing response: %v\n", err)
	}
	if !productResp.Success {
		return "", fmt.Errorf("Error creating Product Rate Plan Charge for Product Rate Plan: %s", ratePlanId)
	}
	return productResp.ID, nil
}

func getProductRatePlanChargeId(key *edgeproto.ClusterInstKey, flavorName string) string {
	// return a set rateplan charge id for now until we actually figure out what the pricing model is going to be
	// product rate plan id: 2c92c0f9712998a401712de88cc44c9f
	return "2c92c0f9712998b30171369c87bd3c44" // usage prices rate 1
}
