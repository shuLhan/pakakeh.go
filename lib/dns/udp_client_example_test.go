// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"log"
)

// The following example show how to use send and Recv to query domain name
// address.
func ExampleUDPClient() {
	var (
		req = &Message{
			Header: MessageHeader{},
			Question: MessageQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeA,
				Class: RecordClassIN,
			},
		}

		cl  *UDPClient
		err error
		res *Message
		rr  ResourceRecord
		x   int
	)

	cl, err = NewUDPClient("127.0.0.1:53")
	if err != nil {
		log.Println(err)
		return
	}

	_, err = req.Pack()
	if err != nil {
		log.Fatal(err)
		return
	}

	res, err = cl.Query(req)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("Receiving DNS message: %s\n", res)
	for x, rr = range res.Answer {
		fmt.Printf("Answer %d: %s\n", x, rr.Value)
	}
	for x, rr = range res.Authority {
		fmt.Printf("Authority %d: %s\n", x, rr.Value)
	}
	for x, rr = range res.Additional {
		fmt.Printf("Additional %d: %s\n", x, rr.Value)
	}
}

func ExampleUDPClient_Lookup() {
	var (
		q = MessageQuestion{
			Name: "kilabit.info",
		}

		cl  *UDPClient
		msg *Message
		rr  ResourceRecord
		err error
		x   int
	)

	cl, err = NewUDPClient("127.0.0.1:53")
	if err != nil {
		log.Println(err)
		return
	}

	msg, err = cl.Lookup(q, false)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Receiving DNS message: %s\n", msg)
	for x, rr = range msg.Answer {
		fmt.Printf("Answer %d: %s\n", x, rr.Value)
	}
	for x, rr = range msg.Authority {
		fmt.Printf("Authority %d: %s\n", x, rr.Value)
	}
	for x, rr = range msg.Additional {
		fmt.Printf("Additional %d: %s\n", x, rr.Value)
	}
}
