package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/mobiledgex/edge-cloud/vault"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd"
	"github.com/mobiledgex/edge-cloud-infra/infracommon"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var (
	port      = flag.Int("port", 8443, "listen port")
	expiresin = flag.Int("expiresin", 28800, "expires in seconds")
	certdir   = flag.String("certdir", "", "cert directory")
	certname  = flag.String("certname", "mex-server", "cert name")
	caname    = flag.String("caname", "mex-ca", "CA cert name")
	vaultaddr = flag.String("vaultaddr", "", "vault addr for vcd creds, e.g. https://vault-dev.mobiledgex.net")
	region    = flag.String("region", "", "region (US or EU)")
	physname  = flag.String("physname", "", "cloudlet physical name")
	org       = flag.String("org", "", "cloudlet org")

	getTokenPath = flag.String("gettokenpath", "/openid/connect/auth/oauth/v2/t3/org/token", "path to gettoken")
	indexpath    = "/"
	vcdVars      map[string]string
	vcdCreds     vcd.VcdConfigParams
	vcdUri       *url.URL
)

func showIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf("doing showIndex for request: %v\n", r)
	rc := *getTokenPath + " -- token auth\n"
	w.Write([]byte(rc))
}

func printUsage() {
	fmt.Println("\nUsage: \nvcd-sgw-sim [options]\n\noptions:")
	flag.PrintDefaults()
}

func validateRequest(r *http.Request) int {
	contentType := r.Header.Get("Content-Type")
	if contentType != vcd.ContentFormUrlEncoded {
		log.Printf("Bad Content-Type: " + contentType)
		return http.StatusBadRequest
	}
	var tokReq vcd.TokenRequest
	err := json.NewDecoder(r.Body).Decode(&tokReq)
	if err != nil {
		log.Printf("Unable to decode request %v", err)
		return http.StatusBadRequest
	}
	log.Printf("Got Token Request: %v contentType: %s", tokReq, contentType)
	if tokReq.GrantType != vcd.GrantTypeCert {
		log.Printf("Bad grant type: %s", tokReq.GrantType)
		return http.StatusBadRequest
	}
	if tokReq.Scope != vcd.ScopeOpenId {
		log.Printf("Bad scope: %s", tokReq.Scope)
		return http.StatusBadRequest
	}
	if tokReq.ClientId != vcdCreds.OauthClientId {
		log.Printf("wrong client id: %s", tokReq.ClientId)
		return http.StatusUnauthorized
	}
	if tokReq.ClientSecret != vcdCreds.OauthClientSecret {
		log.Printf("wrong client secret: %s", tokReq.ClientSecret)
		return http.StatusUnauthorized
	}
	return http.StatusOK
}

func getToken(w http.ResponseWriter, r *http.Request) {
	log.Println("doing getToken")
	code := validateRequest(r)
	if code != http.StatusOK {
		w.WriteHeader(code)
		return
	}

	vcdClient := govcd.NewVCDClient(*vcdUri, vcdCreds.Insecure)
	authResp, err := vcdClient.GetAuthResponse(vcdCreds.User, vcdCreds.Password, vcdCreds.Org)
	if err != nil {
		log.Printf("Unable to login to org %s at %s err: %s", vcdCreds.Org, vcdCreds.Href, err)
		// from the client's perspective this is a server error because these are not creds the client provides
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to login to org " + err.Error()))
		return
	}

	log.Printf("Got authResp: %+v client: %+v ", authResp, vcdClient)
	tokenResponse := vcd.TokenResponse{
		AccessToken:  vcdClient.Client.VCDToken, //get from vcd
		TokenType:    "Bearer",
		ExpiresIn:    *expiresin,
		Scope:        "openid account.read customer.read customer.accounts.read",
		IdToken:      "aaaaaaaa.bbbbbbbb.cccccccc",
		VcdAuthToken: authResp.Header.Get(vcd.VcdAuthHeader),
	}
	byt, _ := json.Marshal(tokenResponse)
	w.Write(byt)
}

func getCredsFromVault() error {
	path := fmt.Sprintf("/secret/data/%s/cloudlet/vcd/%s/%s/vcd.json", "EU", "packet", "qa-lab")
	ctx := context.TODO()
	vaultConfig, err := vault.BestConfig(*vaultaddr)
	if err != nil {
		return fmt.Errorf("Unable to get vault config - %v", err)
	}
	vcdVars, err = infracommon.GetEnvVarsFromVault(ctx, vaultConfig, path)
	if err != nil {
		return fmt.Errorf("Unable to get vars from vault: %s -  %v", *vaultaddr, err)
	}
	return nil
}

func populateCreds() error {
	vcdCreds = vcd.VcdConfigParams{
		User:              vcdVars["VCD_USER"],
		Password:          vcdVars["VCD_PASSWORD"],
		Org:               vcdVars["VCD_ORG"],
		Href:              vcdVars["VCD_IP"] + "/api",
		VDC:               vcdVars["VDC_NAME"],
		OauthClientId:     vcdVars["OAUTH_CLIENT_ID"],
		OauthClientSecret: vcdVars["OAUTH_CLIENT_SECRET"],
		Insecure:          true,
	}
	if vcdCreds.User == "" {
		return fmt.Errorf("VCD_USER not found")
	}
	if vcdCreds.Password == "" {
		return fmt.Errorf("VCD_PASSWORD not found")
	}
	if vcdCreds.Org == "" {
		return fmt.Errorf("VCD_ORG not found")
	}
	if vcdCreds.Href == "/api" {
		return fmt.Errorf("VCD_IP not found")
	}
	if vcdCreds.VDC == "" {
		return fmt.Errorf("VDC_NAME not found")
	}
	if vcdCreds.OauthClientId == "" {
		return fmt.Errorf("OAUTH_CLIENT_ID not found")
	}
	if vcdCreds.OauthClientSecret == "" {
		return fmt.Errorf("OAUTH_CLIENT_SECRET not found")
	}
	return nil
}

func run() {
	http.HandleFunc(indexpath, showIndex)
	http.HandleFunc(*getTokenPath, getToken)

	if *certdir == "" {
		panic("--certdir is empty")
	}
	if *caname == "" {
		panic("--caname is empty")
	}
	if *physname == "" {
		panic("--physname is empty")
	}
	if *org == "" {
		panic("--org is empty")
	}
	if *region == "" {
		panic("--region is empty")
	}
	if *vaultaddr == "" {
		panic("--vaultaddr is empty")
	}

	err := getCredsFromVault()
	if err != nil {
		panic(err.Error())
	}
	err = populateCreds()
	if err != nil {
		panic(err.Error())
	}

	vcdUri, err = url.ParseRequestURI(vcdCreds.Href)
	if err != nil {
		panic(fmt.Errorf("Unable to parse vcd uri %s err: %s", vcdCreds.Href, err))
	}
	certfile := fmt.Sprintf("%s/%s.crt", *certdir, *certname)
	keyfile := fmt.Sprintf("%s/%s.key", *certdir, *certname)
	cafile := fmt.Sprintf("%s/%s.crt", *certdir, *caname)

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile(cafile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", *port),
		TLSConfig: tlsConfig,
	}

	log.Printf("listening on port %d\n", *port)
	err = server.ListenAndServeTLS(certfile, keyfile)
	if err != nil {
		panic(fmt.Sprintf("Error in ListenAndServeTLS: %v", err))
	}
}

func validateArgs() {
	flag.Parse()
}

func main() {
	validateArgs()
	run()
}
