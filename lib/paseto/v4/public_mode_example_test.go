// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/paseto"
	pasetov4 "git.sr.ht/~shulhan/pakakeh.go/lib/paseto/v4"
)

func ExamplePublicMode_Sign() {
	secret := `b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`
	seed, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatal(err)
	}
	pmode := pasetov4.NewPublicMode(`testsign`, [32]byte(seed))

	payload := []byte(`{"data":"signed message!"}`)
	footer := []byte(`{"peer_id":"1000"}`)

	token, err := pmode.Sign(payload, footer, nil)
	if err != nil {
		log.Fatal(err)
	}

	msg := paseto.Message{}
	err = pasetov4.UnpackPublicToken(&msg, token, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Payload: %s\n", msg.RawPayload)
	fmt.Printf("Footer: %s\n", msg.RawFooter)

	// Output:
	// Payload: {"data":"signed message!"}
	// Footer: {"peer_id":"1000"}
}

func ExamplePublicMode_Verify() {
	secret := `b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`
	seed, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatal(err)
	}
	pmode := pasetov4.NewPublicMode(`testverify`, [32]byte(seed))

	token := `v4.public.eyJkYXRhIjoic2lnbmVkIG1lc3NhZ2UhIn2PyAoUJRvIViAgbEn0wVCLClzz26ZuJ7UY8z7rKRDXB0Vlo6i3QzYqmpg75TPJFQe2XX5_DmURV8hJnboNzH0H.eyJwZWVyX2lkIjoiMTAwMCJ9`

	gotPlain, gotFooter, err := pmode.Verify(token, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(gotPlain))
	fmt.Println(string(gotFooter))

	// Output:
	// {"data":"signed message!"}
	// {"peer_id":"1000"}
}

func ExamplePublicMode_Pack() {
	createPeerMode := func(id, secret string) (pmode *pasetov4.PublicMode) {
		seed, err := hex.DecodeString(secret)
		if err != nil {
			log.Fatal(err)
		}
		pmode = pasetov4.NewPublicMode(id, [32]byte(seed))
		return pmode
	}
	senderpm := createPeerMode(
		`john@example.com`,
		`b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`,
	)
	receivpm := createPeerMode(
		`jane@example.org`,
		`2222222222222222222222222222222222222222222222222222222222222222`,
	)

	sendMsg := paseto.NewMessage(senderpm.Peer, receivpm.Peer, `hello`)
	sendMsg.Payload.Data = map[string]string{
		`email`: `hello@example.com`,
	}
	sendMsg.Footer.Data = map[string]string{
		`quote`: `Live to eat`,
	}

	token, err := senderpm.Pack(*sendMsg, nil)
	if err != nil {
		log.Fatal(err)
	}

	msg := paseto.Message{}
	pasetov4.UnpackPublicToken(&msg, token, nil)

	pdata := paseto.Payload{}
	_ = json.Unmarshal(msg.RawPayload, &pdata)
	fmt.Printf("Payload.Data: %+v\n", pdata.Data)
	fmt.Printf("Footer: %s", msg.RawFooter)

	// Output:
	// Payload.Data: map[email:hello@example.com]
	// Footer: {"data":{"quote":"Live to eat"},"peer_id":"john@example.com"}
}

func ExamplePublicMode_Unpack() {
	createPeerMode := func(id, secret string) (
		pmode *pasetov4.PublicMode,
	) {
		seed, err := hex.DecodeString(secret)
		if err != nil {
			log.Fatal(err)
		}
		pmode = pasetov4.NewPublicMode(id, [32]byte(seed))
		return pmode
	}
	senderpm := createPeerMode(
		`john@example.com`,
		`b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`,
	)
	receivpm := createPeerMode(
		`jane@example.org`,
		`2222222222222222222222222222222222222222222222222222222222222222`,
	)

	// Add the sender to list of receiver's peer.
	receivpm.AddPeer(senderpm.Peer)

	// This is the token send by sender.
	token := `v4.public.eyJkYXRhIjp7ImVtYWlsIjoiaGVsbG9AZXhhbXBsZS5jb20ifSwiaXNzIjoiam9obkBleGFtcGxlLmNvbSIsInN1YiI6ImhlbGxvIiwiYXVkIjoiamFuZUBleGFtcGxlLm9yZyIsImV4cCI6MTc3NDg1OTg4OSwibmJmIjoxNzc0ODU5ODI5LCJpYXQiOjE3NzQ4NTk4Mjl9eGHMkOlYiVchKl1GqQwSMyTC8eWCpXhNJHGHq_MEJNqlKur-kPpHqY-hWHoh4I1RaBq68MN_TNveqwPIbk9iBQ.eyJkYXRhIjp7InF1b3RlIjoiTGl2ZSB0byBlYXQifSwicGVlcl9pZCI6ImpvaG5AZXhhbXBsZS5jb20ifQ`

	recvMsg := paseto.Message{}
	err := receivpm.Unpack(token, nil, &recvMsg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Payload.Data: %+v\n", recvMsg.Payload.Data)
	fmt.Printf("Footer: %+v\n", recvMsg.Footer)

	// Output:
	// Payload.Data: map[email:hello@example.com]
	// Footer: {Data:map[quote:Live to eat] PeerID:john@example.com}
}
