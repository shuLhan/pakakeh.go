// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/paseto"
)

func ExamplePublicMode() {
	secret := `b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`
	seed, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatal(err)
	}
	pmode := NewPublicMode([32]byte(seed))

	plain := []byte(`{"data":"signed message!"}`)
	footer := []byte(`{"kid":1000}`)
	token, err := pmode.Sign(plain, footer, nil)
	if err != nil {
		log.Fatal(err)
	}

	gotPlain, gotFooter, err := pmode.Verify(token, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(gotPlain))
	fmt.Println(string(gotFooter))

	// Output:
	// {"data":"signed message!"}
	// {"kid":1000}
}

type DummyData struct {
	Email string `json:"email"`
}
type DummyFooterData struct {
	Quote string `json:"quote"`
}

func ExamplePublicMode_Pack() {
	createPeer := func(id, secret string) (peer paseto.Peer, pmode *PublicMode) {
		seed, err := hex.DecodeString(secret)
		if err != nil {
			log.Fatal(err)
		}
		pmode = NewPublicMode([32]byte(seed))
		peer = paseto.Peer{
			ID:      id,
			Private: pmode.PrivateKey,
			Public:  pmode.PublicKey,
		}
		return peer, pmode
	}
	sender, senderpm := createPeer(
		`john@example.com`,
		`b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`,
	)
	receiver, _ := createPeer(
		`jane@example.org`,
		`2222222222222222222222222222222222222222222222222222222222222222`,
	)

	sendMsg := paseto.NewMessage(sender, receiver, `hello`)
	sendMsg.Payload.Data = DummyData{
		Email: `hello@example.com`,
	}
	sendMsg.Footer.Data = DummyFooterData{
		Quote: `Live to eat`,
	}

	token, err := senderpm.Pack(*sendMsg, nil)
	if err != nil {
		log.Fatal(err)
	}

	pieces := strings.Split(token, `.`)
	fmt.Printf("Header: %s.%s\n", pieces[0], pieces[1])

	paysig, _ := base64.RawURLEncoding.DecodeString(pieces[2])
	payload := paysig[:len(paysig)-64]
	pdata := paseto.Payload{
		Data: &DummyData{},
	}
	_ = json.Unmarshal(payload, &pdata)
	fmt.Printf("Payload.Data: %+v\n", pdata.Data)

	footer, _ := base64.RawURLEncoding.DecodeString(pieces[3])
	fmt.Printf("Footer: %s", footer)

	// Output:
	// Header: v4.public
	// Payload.Data: &{Email:hello@example.com}
	// Footer: {"data":{"quote":"Live to eat"},"peer_id":"john@example.com"}
}

func ExamplePublicMode_Unpack() {
	createPeer := func(id, secret string) (peer paseto.Peer, pmode *PublicMode) {
		seed, err := hex.DecodeString(secret)
		if err != nil {
			log.Fatal(err)
		}
		pmode = NewPublicMode([32]byte(seed))
		peer = paseto.Peer{
			ID:      id,
			Private: pmode.PrivateKey,
			Public:  pmode.PublicKey,
		}
		return peer, pmode
	}
	sender, _ := createPeer(
		`john@example.com`,
		`b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`,
	)
	_, recvpm := createPeer(
		`jane@example.org`,
		`2222222222222222222222222222222222222222222222222222222222222222`,
	)
	recvpm.AddPeer(sender)

	token := `v4.public.eyJpc3MiOiJqb2huQGV4YW1wbGUuY29tIiwic3ViIjoiaGVsbG8iLCJhdWQiOiJqYW5lQGV4YW1wbGUub3JnIiwiZXhwIjoxNzc0ODE5NzYzLCJuYmYiOjE3NzQ4MTk3MDMsImlhdCI6MTc3NDgxOTcwMywiZGF0YSI6eyJlbWFpbCI6ImhlbGxvQGV4YW1wbGUuY29tIn19Ui741xeDUhQE2w31smKyAJ_d-fhMH20SRLkIY3lDbIBfj_tq7es9-H2wQR3q6EZZRPrlAPxEug712IEgomTnBA.eyJkYXRhIjp7InF1b3RlIjoiTGlmZSBpcyBmb3IgZWF0aW5nIn0sInBlZXJfaWQiOiJqb2huQGV4YW1wbGUuY29tIn0`
	pdata := DummyData{}
	fdata := DummyFooterData{}
	recvMsg := paseto.Message{
		Payload: paseto.Payload{
			Data: &pdata,
		},
		Footer: paseto.Footer{
			Data: &fdata,
		},
	}
	err := recvpm.Unpack(token, nil, &recvMsg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", recvMsg.Payload.Data)
	fmt.Printf("%+v\n", recvMsg.Footer.Data)

	// Output:
	// &{Email:hello@example.com}
	// &{Quote:Life is for eating}
}
