package dns_test

import (
	"fmt"
	"log"

	"github.com/shuLhan/share/lib/dns"
)

func ExampleTCPClient_Lookup() {
	cl, err := dns.NewTCPClient("127.0.0.1:53")
	if err != nil {
		log.Println(err)
		return
	}

	msg, err := cl.Lookup(false, dns.RecordTypeA, dns.QueryClassIN,
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
