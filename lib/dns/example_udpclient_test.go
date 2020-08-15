package dns_test

import (
	"fmt"
	"log"

	"github.com/shuLhan/share/lib/dns"
)

//
// The following example show how to use send and Recv to query domain name
// address.
//
func ExampleUDPClient() {
	cl, err := dns.NewUDPClient("127.0.0.1:53")
	if err != nil {
		log.Println(err)
		return
	}

	req := &dns.Message{
		Header: dns.SectionHeader{},
		Question: dns.SectionQuestion{
			Name:  "kilabit.info",
			Type:  dns.QueryTypeA,
			Class: dns.QueryClassIN,
		},
	}

	_, err = req.Pack()
	if err != nil {
		log.Fatal(err)
		return
	}

	res, err := cl.Query(req)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("Receiving DNS message: %s\n", res)
	for x, answer := range res.Answer {
		fmt.Printf("Answer %d: %s\n", x, answer.Value)
	}
	for x, auth := range res.Authority {
		fmt.Printf("Authority %d: %s\n", x, auth.Value)
	}
	for x, add := range res.Additional {
		fmt.Printf("Additional %d: %s\n", x, add.Value)
	}
}

func ExampleUDPClient_Lookup() {
	cl, err := dns.NewUDPClient("127.0.0.1:53")
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
