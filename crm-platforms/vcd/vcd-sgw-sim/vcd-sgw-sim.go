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

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd"

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
	vcdPlatform  = vcd.VcdPlatform{} // used for creds
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
	if tokReq.ClientId != vcdPlatform.GetVcdOauthClientId() {
		log.Printf("wrong client id: %s", tokReq.ClientId)
		return http.StatusUnauthorized
	}
	if tokReq.ClientSecret != vcdPlatform.GetVcdOauthClientSecret() {
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

	vcdClient := govcd.NewVCDClient(*vcdUri, vcdPlatform.Creds.Insecure)
	authResp, err := vcdClient.GetAuthResponse(vcdPlatform.Creds.User, vcdPlatform.Creds.Password, vcdPlatform.Creds.Org)
	if err != nil {
		log.Printf("Unable to login to org %s at %s err: %s", vcdPlatform.Creds.Org, vcdPlatform.Creds.VcdApiUrl, err)
		// from the client's perspective this is a server error because these are not creds the client provides
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to login to org " + err.Error()))
		return
	}

	// the SGW simulator builds the oauth token from the original VCD client token plus the VCD
	// auth header separated by semicolon.  The AGW simulator will break these apart and send to VCD
	at := vcdClient.Client.VCDToken + ";" + authResp.Header.Get(vcd.VcdAuthHeader)

	log.Printf("Got authResp: %+v client: %+v ", authResp, vcdClient)
	tokenResponse := vcd.TokenResponse{
		AccessToken: at,
		TokenType:   "Bearer",
		ExpiresIn:   *expiresin,
		Scope:       "openid account.read customer.read customer.accounts.read",
		IdToken:     "aaaaaaaa.bbbbbbbb.cccccccc",
	}
	byt, _ := json.Marshal(tokenResponse)
	w.Write(byt)
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
	err := vcdPlatform.PopulateCredsForSimulator(context.TODO(), *region, *org, *physname, *vaultaddr)
	if err != nil {
		panic(err.Error())
	}

	vcdUri, err = url.ParseRequestURI(vcdPlatform.Creds.VcdApiUrl)
	if err != nil {
		panic(fmt.Errorf("Unable to parse vcd uri %s err: %s", vcdPlatform.Creds.VcdApiUrl, err))
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
