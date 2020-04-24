package zuora

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var referenceTime = "2006-01-02T15:04:05"

func RecordClusterUsage(account *AccountInfo, clusterInstKey *edgeproto.ClusterInstKey, flavorName string, startTime, endTime time.Time, runTime float64) error {
	chargeId := getProductRatePlanChargeId(clusterInstKey, flavorName)
	chargeNumber, err := getSubChargeNumber(account.AccountNumber, chargeId)
	if err != nil {
		return fmt.Errorf("unable to get charge number: %v", err)
	}
	desc := fmt.Sprintf("Clusterinst: %s, Cloudlet: %s, Flavor: %s",
		clusterInstKey.ClusterKey.Name, clusterInstKey.CloudletKey.Name, flavorName)
	newUsage := CreateUsage{
		AccountNumber:      account.AccountNumber,
		SubscriptionNumber: account.SubscriptionNumber,
		ChargeNumber:       chargeNumber,
		Quantity:           runTime / 60, // assume we port right from influx and runTime is in seconds still
		StartDateTime:      startTime.Format(referenceTime),
		EndDateTime:        endTime.Format(referenceTime),
		UOM:                "Minute",
		Description:        desc,
	}

	payload, err := json.Marshal(newUsage)
	fmt.Printf("payload: %s\n", payload)
	if err != nil {
		return fmt.Errorf("Could not marshal %+v, err: %v", newUsage, err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", ZuoraUrl+UsageEndpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("Error creating request: %v\n", err)
	}
	token, tokentype, err := getToken()
	if err != nil {
		return fmt.Errorf("Unable to retrieve oAuth token")
	}
	req.Header.Add("Authorization", tokentype+" "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}
	usageResp := GenericResp{}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&usageResp)
	if err != nil {
		return fmt.Errorf("Error parsing response: %v\n", err)
	}
	if !usageResp.Success {
		return fmt.Errorf("Error creating usage record")
	}
	return nil
}

// for some reason zuora needs to have the charge id generated with the subscription, not the actual original charge id of the rateplancharge
func getSubChargeNumber(accountNum, ratePlanChargeId string) (string, error) {
	subInfo, err := getSubscription(accountNum)
	if err != nil {
		return "", err
	}
	for _, ratePlan := range subInfo.Subscriptions[0].RatePlans {
		for _, ratePlanCharges := range ratePlan.RatePlanCharges {
			if ratePlanChargeId == ratePlanCharges.ProductRatePlanChargeID {
				return ratePlanCharges.Number, nil
			}
		}
	}
	return "", fmt.Errorf("Rate plan charge does not exist in the subscription")
}
