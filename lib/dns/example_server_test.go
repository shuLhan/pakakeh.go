package dns_test

import (
	"fmt"
	"log"
	"time"

	"github.com/shuLhan/share/lib/dns"
)

func clientLookup(nameserver string) {
	cl, err := dns.NewUDPClient(nameserver)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err := cl.Lookup(false, dns.QueryTypeA, dns.QueryClassIN,
		"kilabit.info")
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Receiving DNS message: %s\n", msg)
	for x, answer := range msg.Answer {
		fmt.Printf("Answer %d: %s\n", x, answer.Value)
	}
	for x, auth := range msg.Authority {
		fmt.Printf("Authority %d: %s\n", x, auth.Value)
	}
	for x, add := range msg.Additional {
		fmt.Printf("Additional %d: %s\n", x, add.Value)
	}
}

func ExampleServer() {
	serverAddress := "127.0.0.1:5300"

	serverOptions := &dns.ServerOptions{
		ListenAddress:    "127.0.0.1:5300",
		HTTPPort:         8443,
		TLSCertFile:      "testdata/domain.crt",
		TLSPrivateKey:    "testdata/domain.key",
		TLSAllowInsecure: true,
	}

	server, err := dns.NewServer(serverOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Load records to be served from zone file.
	zoneFile, err := dns.ParseZoneFile("testdata/kilabit.info", "", 0)
	if err != nil {
		log.Fatal(err)
	}

	server.PopulateCaches(zoneFile.Messages(), zoneFile.Path)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for all listeners running.
	time.Sleep(500 * time.Millisecond)

	clientLookup(serverAddress)

	server.Stop()
}
