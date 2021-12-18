package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 8080, "listen port")

	indexpath = "/"

	latencyPath    = "5g-latency/sessions/latency"
	throughputPath = "5g-latency/sessions/throughput"
)

func printUsage() {
	fmt.Println("\nUsage: \tsessions-server-sim [options]\n\noptions:")
	flag.PrintDefaults()
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf("doing showIndex for request: %v\n", r)
	rc := "QOS Priority Sessions server\n"
	rc += latencyPath + " -- Creates, Gets, or deletes a Latency priority session\n"
	rc += throughputPath + " -- Creates, Gets, or deletes a Throughput priority session\n"
	w.Write([]byte(rc))
}

func getToken(w http.ResponseWriter, r *http.Request) {
	log.Println("doing getToken")
}

func run() {
	http.HandleFunc(indexpath, showIndex)
	// http.HandleFunc(getTokenPath, getToken)
	// http.HandleFunc(getExpiredTokenPath, getToken)

	portstr := fmt.Sprintf(":%d", *port)

	log.Printf("Listening on http://127.0.0.1:%d", *port)
	if err := http.ListenAndServe(portstr, nil); err != nil {
		panic(err)
	}
}

func validateArgs() {
	flag.Parse()
	//nothing to check yet
}

func main() {
	validateArgs()
	run()
}
