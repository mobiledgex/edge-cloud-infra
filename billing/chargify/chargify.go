package chargify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mobiledgex/edge-cloud/vault"
)

type BillingService struct{}

type accountCreds struct {
	ApiKey string `json:"apikey"`
	Url    string `json:"url"`
}

func (bs *BillingService) Init(vaultConfig *vault.Config, path string) error {
	creds := accountCreds{}
	err := vault.GetData(vaultConfig, vaultPath+path, 0, &creds)
	if err != nil {
		return err
	}

	apiKey = creds.ApiKey
	siteName = creds.Url
	fmt.Printf("apiKey: %s, siteName: %s\n", apiKey, siteName)
	return nil
}

func (bs *BillingService) GetType() string {
	return "chargify"
}

func newChargifyReq(method, endpoint string, payload interface{}) (*http.Response, error) {
	url := siteName + endpoint
	var body io.Reader
	if payload != nil {
		marshalled, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("Could not marshal %+v, err: %v", payload, err)
		}
		body = bytes.NewReader(marshalled)
	} else {
		body = strings.NewReader("{}")
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v\n", err)
	}
	req.SetBasicAuth(apiKey, apiPassword)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	return client.Do(req)
}

func combineErrors(e *ErrorResp) {
	e.Errors = append(e.Errors, e.Error)
}
