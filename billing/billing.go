package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mobiledgex/edge-cloud/log"
	opentracing "github.com/opentracing/opentracing-go"
)

var Token OAuthToken

// sample curl to get oauth token: curl -X POST -H "Content-Type: application/x-www-form-urlencoded" -d "client_id=d0858528-8ed7-4790-bd0c-e1f689f54897" --data-urlencode "client_secret=G8uAaL/bEP3xBZsAhx1VlZwV3EA9efI1=am/7rs" -d "grant_type=client_credentials" "https://rest.apisandbox.zuora.com/oauth/token"
func getOauth(token *OAuthToken) error {
	data := url.Values{}
	data.Set("client_id", ClientId)
	data.Add("client_secret", ClientSecret)
	data.Add("grant_type", "client_credentials")
	req, err := http.NewRequest("POST", ZuoraUrl+OAuthEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("Error creating request: %v\n", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v\n", err)
	}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return fmt.Errorf("Error parsing response: %v\n", err)
	}
	return nil
}

func RunOAuth(ctx context.Context) {
	//timeout is an hour, run every 45
	oAuthSpan := log.StartSpan(log.DebugLevelInfo, "OAuth thread", opentracing.ChildOf(log.SpanFromContext(ctx).Context()))
	defer oAuthSpan.Finish()
	err := getOauth(&Token)
	if err != nil {
		log.FatalLog(fmt.Errorf("Error getting OAuth credentials: %v", err).Error())
	}
	for {
		// check if there are any new apps we need to start/stop scraping for
		select {
		case <-time.After(time.Minute * 45):
			err = getOauth(&Token)
			if err != nil {
				log.SpanLog(ctx, log.DebugLevelInfo, "Error getting OAuth credentials", "err", err)
			}
		}
	}
}
