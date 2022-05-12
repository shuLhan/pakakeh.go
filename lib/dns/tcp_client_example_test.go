// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"log"
)

func ExampleTCPClient_Lookup() {
	var (
		q = MessageQuestion{
			Name: "kilabit.info",
		}

		cl  *TCPClient
		msg *Message
		rr  ResourceRecord
		err error
		x   int
	)

	cl, err = NewTCPClient("127.0.0.1:53")
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
