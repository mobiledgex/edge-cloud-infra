package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var (
	port      = flag.Int("port", 8443, "listen port")
	expiresin = flag.Int("expiresin", 28800, "expires in seconds")
	certdir   = flag.String("certdir", "", "certdir")
	certname  = flag.String("certname", "mex-server", "certname")
	vcdCreds  vcd.VcdConfigParams
	vcdUri    *url.URL

	getTokenPath = "/openid/connect/auth/oauth/v2/t3/org/token"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IdToken      string `json:"id_token"`
}

func printUsage() {
	fmt.Println("\nUsage: \nvcd-sgw-sim [options]\n\noptions:")
	flag.PrintDefaults()
}

func getToken(w http.ResponseWriter, r *http.Request) {
	log.Println("doing getToken")

	vcdClient := govcd.NewVCDClient(*vcdUri, vcdCreds.Insecure)
	_, err := vcdClient.GetAuthResponse(vcdCreds.User, vcdCreds.Password, vcdCreds.Org)
	if err != nil {
		log.Printf("Unable to login to org %s at %s err: %s", vcdCreds.Org, vcdCreds.Href, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to login to org"))
		return
	}
	tokenResponse := TokenResponse{
		AccessToken: vcdClient.Client.VCDToken, //get from vcd
		TokenType:   "Bearer",
		ExpiresIn:   *expiresin,
		Scope:       "openid account.read customer.read customer.accounts.read 3part.apiseg bla bla",
		IdToken:     "eyJ0eXAiOiJKV1QiLCJhbGciOiJQUzI1NiJ9.ewogInN1YiI6ICJtb25leS5wcm8ubW92aXN0YXIuZXMiLAogImF1ZCI6ICJkMWVkZjYyOC0zMzEzLTQyMTUtYmI5Ni1hMjg3ZjVhYjE2MTMiLAogImFjciI6ICJDRVJUIiwKICJhdXRoX3RpbWUiOiAxNTY3NTkwOTMwLAogImlzcyI6ICJodHRwczovL2FwaXNlZy50ZWxlZm9uaWNhLmVzIiwKICJleHAiOiAxNTY3NjE5NzMwLAogImlhdCI6IDE1Njc1OTA5MzAKfQ.BiPA5mSEJ9jsoMR3_SlGo_491rbxeRhB_EOL1ilgFHQcLvxFTdwGJYdsGR6cPe1ahKJBZWCCaJDxFxACWQAWzPD1bJtn6GkFA_TWnIKsurkKrYheXYD6Bp42N8FgnTv_TLJf05gyR_fV42MsQlxddzU3KWbbSaAdhDXd_js633nUc-f1mJLvkgA-wdPtuoqI_MKLCZgiVTIGGj8dIQCYjQ3wlepJiUBNMociH6oOUB69n0qEyA2Nm7osdlAcHBX4lKYZ7EzCgLBCHQJZNl-71btUd_QjBkwMD76wM4qGszdUHX_jXtzG9Yz_WRNref4BxVZghhgOq2pr2Tbx5zn_GA",
	}
	byt, _ := json.Marshal(tokenResponse)
	w.Write(byt)
}

func populateCreds() error {
	vcdCreds = vcd.VcdConfigParams{
		User:     os.Getenv("VCD_USER"),
		Password: os.Getenv("VCD_PASSWORD"),
		Org:      os.Getenv("VCD_ORG"),
		Href:     os.Getenv("VCD_IP") + "/api",
		VDC:      os.Getenv("VDC_NAME"),
		Insecure: true,
	}
	if vcdCreds.User == "" {
		return fmt.Errorf("VCD_USER not defined")
	}
	if vcdCreds.Password == "" {
		return fmt.Errorf("VCD_PASSWORD not defined")
	}
	if vcdCreds.Org == "" {
		return fmt.Errorf("VCD_ORG not defined")
	}
	if vcdCreds.Href == "/api" {
		return fmt.Errorf("VCD_IP")
	}
	if vcdCreds.VDC == "" {
		return fmt.Errorf("VDC_NAME not defined")
	}
	return nil
}

func run() {
	http.HandleFunc(getTokenPath, getToken)

	err := populateCreds()
	if err != nil {
		panic(err.Error())
	}

	if *certdir == "" {
		panic("--certdir is empty")
	}

	vcdUri, err = url.ParseRequestURI(vcdCreds.Href)
	if err != nil {
		panic(fmt.Errorf("Unable to parse vcd uri %s err: %s", vcdCreds.Href, err))
	}

	listenAddr := fmt.Sprintf("127.0.0.1:%d", *port)

	log.Printf("Listening on " + listenAddr)
	certfile := fmt.Sprintf("%s/%s.crt", *certdir, *certname)
	keyfile := fmt.Sprintf("%s/%s.key", *certdir, *certname)

	err = http.ListenAndServeTLS(listenAddr, certfile, keyfile, nil)
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
