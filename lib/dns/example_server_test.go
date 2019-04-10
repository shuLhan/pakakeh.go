package dns_test

import (
	"crypto/tls"
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

	msg, err := cl.Lookup(false, dns.QueryTypeA, dns.QueryClassIN, []byte("kilabit.info"))
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Receiving DNS message: %s\n", msg)
	for x, answer := range msg.Answer {
		fmt.Printf("Answer %d: %s\n", x, answer.RData())
	}
	for x, auth := range msg.Authority {
		fmt.Printf("Authority %d: %s\n", x, auth.RData())
	}
	for x, add := range msg.Additional {
		fmt.Printf("Additional %d: %s\n", x, add.RData())
	}
}

func ExampleServer() {
	serverAddress := "127.0.0.1:5300"

	cert, err := tls.LoadX509KeyPair("testdata/domain.crt", "testdata/domain.key")
	if err != nil {
		log.Fatal("dns: error loading certificate: " + err.Error())
	}

	serverOptions := &dns.ServerOptions{
		IPAddress:        "127.0.0.1",
		TCPPort:          5300,
		UDPPort:          5300,
		DoHPort:          8443,
		DoHCertificate:   &cert,
		DoHAllowInsecure: true,
	}

	server, err := dns.NewServer(serverOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Load records to be served from master file.
	server.LoadMasterFile("testdata/kilabit.info")

	server.Start()

	// Wait for all listeners running.
	time.Sleep(500 * time.Millisecond)

	clientLookup(serverAddress)

	server.Stop()
}
