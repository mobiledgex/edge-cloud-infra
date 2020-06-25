package zuora

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/mobiledgex/edge-cloud/vault"
)

var oAuthToken *OAuthToken
var oAuthMux sync.Mutex

// sample curl to get oauth token: curl -X POST -H "Content-Type: application/x-www-form-urlencoded" -d "client_id=d0858528-8ed7-4790-bd0c-e1f689f54897" --data-urlencode "client_secret=G8uAaL/bEP3xBZsAhx1VlZwV3EA9efI1=am/7rs" -d "grant_type=client_credentials" "https://rest.apisandbox.zuora.com/oauth/token"
func getOauth(token *OAuthToken) error {
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Add("client_secret", clientSecret)
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
	token.ExpireTime = time.Now().Add(time.Second * time.Duration(token.ExpiresIn))
	return nil
}

func getToken() (string, string, error) {
	oAuthMux.Lock()
	defer oAuthMux.Unlock()
	// give a 5 minute buffer to the expire time
	if oAuthToken == nil || time.Now().Add(time.Minute*5).After(oAuthToken.ExpireTime) {
		oAuthToken = &OAuthToken{}
		err := getOauth(oAuthToken)
		if err != nil {
			return "", "", nil
		}
	}
	return oAuthToken.AccessToken, oAuthToken.TokenType, nil
}

type accountCreds struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Url          string `json:"url"`
}

func InitZuora(vaultConfig *vault.Config, path string) error {
	if path == fakeBillingPath {
		runFakeZuora()
		clientId = fakeClientID
		clientSecret = fakeClientSecret
		ZuoraUrl = fakeURL
	}
	// pull it from vault and if you cant throw a fatal error
	creds := accountCreds{}
	err := vault.GetData(vaultConfig, vaultPath+path, 0, &creds)
	if err != nil {
		return err
	}

	clientId = creds.ClientId
	clientSecret = creds.ClientSecret
	ZuoraUrl = creds.Url
	return nil
}
