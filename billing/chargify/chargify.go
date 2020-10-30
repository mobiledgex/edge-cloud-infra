package chargify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type BillingService struct{}

func (bs *BillingService) Init() error {
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
	fmt.Printf("payload: %s\n%s\n", url, body)
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
