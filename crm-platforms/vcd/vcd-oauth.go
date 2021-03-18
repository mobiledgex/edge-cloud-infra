package vcd

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IdToken      string `json:"id_token"`
	VcdAuthToken string `json:"vcd_auth_token"` // for simulator only
}

type TokenRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
}

const GrantTypeCert = "CERT"
const ScopeOpenId = "openid"
const ContentFormUrlEncoded = "application/x-www-form-urlencoded"
const VcdAuthHeader = "X-Vcloud-Authorization"
const VcdTokenHeader = "X-Vmware-Vcloud-Access-Token"
const VcdTokenTypeHeader = "X-Vmware-Vcloud-Token-Type"
