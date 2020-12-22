package main

import (
	tls "crypto/tls"
	//"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	//"os"
	//"strconv"
	//"strings"
)

// We'd like to receive async notifications for certian events in vCD like
// when running VApp is about to suffer a lease expiration, so we can renew the least before that
// happens
// This file attempts to create a mqtt client and connect to the vCD endpoint and subscribe to whatever
// Topic will provide this event.
func main() {
	fmt.Printf("Create Client\n")

	// Our client will need to connect to the /messaging/mqtt path on our vCD endpoint
	// and provide a vaild JWT token in the header. Ensure our go client allows that.

	// Actual events are in Json, with nested json payload
	// Now, according to the vmWare doc we'll need to
	// 1) log into the vCD using the OpenApi endpoint
	// 2) Set Sec-WebScoket-Protocol propert to mqtt, set client to connect o
	//    the /messaging/mqtt paht, adding an outh header, and follow std mqtt connect procedure
	// So then subscribe to topcs
	// Org admins can subscribe using wildcards like
	// publish/$org_uuid/*
	// Ok, so I need the ID of our ORG which is
	//
	/*
		cliOpts := mqtt.ClientOptions{
			Username : os.Getenv("VCD_USER")
			Password : os.Getenv("VCD_PASSWD")
			broker
	*/
	cliOpts := mqtt.NewClientOptions()
	cliOpts.SetUsername(os.Getenv("VCD_USER"))
	cliOpts.SetPassword(os.Getenv("VCD_PASSWD"))
	cliOpts.SetClientID("vcd-notif-client")
	ip := os.Getenv("VCD_IP")
	broker := "wss://" + ip + "/messaging/mqtt"

	fmt.Printf("\nConnect to broker: %s\n\n", broker)
	cliOpts.AddBroker(broker)
	topic := fmt.Sprintf("%s/%s/%s", "publish", "urn:vcloud:org:dde40006-89a4-420d-b4db-3daf2d6c185e", "*")
	fmt.Printf("Topic: %s\n", topic)

	//cliOpts.AddBroker(fmt.Sprintf("ws://%s/messaging/mqtt", os.Getenv("VCD_IP")))

	// ws:// unsecure wss:// secure websockets

	tlsConfig := cliOpts.TLSConfig
	if cliOpts.TLSConfig == nil {
		fmt.Printf("TLSConfig nil in cliOpts creating...\n")
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		fmt.Printf("Adding InsecureSkipVerify to existing TLSConfig\n")
		// just add i t in
		tlsConfig.InsecureSkipVerify = true
	}
	// without this tlsConfig:  x509: certificate signed by unknown authority

	cliOpts.TLSConfig = tlsConfig
	// with this (and nothing else) we get:  bad status
	// which is what we get using ws:/
	// hmm...
	// could store to file, we'll  use default memory

	// we're not publishing (yet anyway) so
	reccnt := 0
	choke := make(chan [2]string)
	cliOpts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := mqtt.NewClient(cliOpts)

	fmt.Printf("have new mqtt client: %+v\n", client)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Printf("Topics to sub to %s\n", topic)
	qos := 0 // quality of service { 0,1,2 }
	num := 1
	if token := client.Subscribe(topic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	for reccnt < num {
		fmt.Printf("WAITING FOR MESSAGE...\n")
		incoming := <-choke
		fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming[0], incoming[1])
		reccnt++
	}

	client.Disconnect(250)
	fmt.Println("Sample Subscriber Disconnected")
}
