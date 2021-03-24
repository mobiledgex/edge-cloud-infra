package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net/http"

	"context"
	"io/ioutil"
	"strings"

	"github.com/mobiledgex/edge-cloud-infra/crm-platforms/vcd"
)

var (
	port         = flag.Int("port", 8443, "listen port")
	caname       = flag.String("caname", "mex-ca", "CA cert name")
	certdir      = flag.String("certdir", "", "certdir")
	certname     = flag.String("certname", "mex-server", "certname")
	apiprefix    = flag.String("apiprefix", "/api/rest/TelefonicaApiVCloud/v1/", "api gw path prefix")
	vcdapivers   = flag.String("vcdapivers", "32.0", "vcd api version")
	vcdapiprefix = flag.String("vcdapiprefix", "", "api prefix for call to vcd")
	region       = flag.String("region", "", "region (US or EU)")
	physname     = flag.String("physname", "", "cloudlet physical name")
	org          = flag.String("org", "", "cloudlet org")
	vaultaddr    = flag.String("vaultaddr", "", "vault addr for vcd creds, e.g. https://vault-dev.mobiledgex.net")
	indexpath    = "/"
	vcdVars      map[string]string
	vcdPlatform  = vcd.VcdPlatform{} // used for creds
)

func showIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf("doing showIndex for request: %v\n", r)
	rc := *apiprefix + " -- prefix to prepend to vcd apis\n"
	w.Write([]byte(rc))
}

func doApi(w http.ResponseWriter, r *http.Request) {
	log.Println("doing doApi URL: " + r.URL.Path + " QueryParams: " + r.URL.RawQuery)

	token := r.Header.Get("Authorization")
	stoken := strings.Split(token, "Bearer")
	if len(stoken) != 2 {
		log.Printf("Bad auth token header: %s", token)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokval := strings.TrimSpace(stoken[1])
	// the SGW simulator builds the token as vcdtoken;vcdauth break these apart
	ts := strings.Split(tokval, ";")
	if len(ts) != 2 {
		log.Printf("Bad bearer token, should be 2 parts: %s", token)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vcdClientToken := ts[0]
	vcdAuthToken := ts[1]

	vcdapi := strings.TrimPrefix(r.URL.Path, *apiprefix)

	// now forward to vcd
	urlString := fmt.Sprintf("%s%s/%s", vcdPlatform.GetVcdUrl(), *vcdapiprefix, vcdapi)
	if r.URL.RawQuery != "" {
		urlString += "?" + r.URL.RawQuery
	}
	vcdreq, err := http.NewRequest(r.Method, urlString, r.Body)
	if err != nil {
		log.Printf("error creating vcd request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	vcdreq.Header.Add(vcd.VcdTokenTypeHeader, "Bearer")
	vcdreq.Header.Add(vcd.VcdTokenHeader, vcdClientToken)
	vcdreq.Header.Add(vcd.VcdAuthHeader, vcdAuthToken)
	vcdreq.Header.Add("Accept", fmt.Sprintf("application/*+xml;version=%s", *vcdapivers))
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	log.Printf("---> Sending to VCD -- Method: %s URL: %s HEADER: %+v\n\n", vcdreq.Method, vcdreq.URL, vcdreq.Header)

	resp, err := client.Do(vcdreq)
	if err != nil {
		log.Printf("error in vcd response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("<--- Received VCD status code: %d", resp.StatusCode)
	if resp.Body == nil {
		log.Printf("nil body in vcd response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading vcd response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	log.Printf("<--- VCD body bytes: %s\n", body)
	w.Write(body)
}

func run() {
	http.HandleFunc(indexpath, showIndex)
	http.HandleFunc(*apiprefix, doApi)
	if *certdir == "" {
		panic("--certdir is empty")
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
	listenAddr := fmt.Sprintf(":%d", *port)
	err := vcdPlatform.PopulateCredsForSimulator(context.TODO(), *region, *org, *physname, *vaultaddr)
	if err != nil {
		panic(err.Error())
	}

	log.Printf("Listening on " + listenAddr)
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
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", *port),
		TLSConfig: tlsConfig,
	}
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
