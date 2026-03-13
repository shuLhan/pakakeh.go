// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package uuidv7_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/uuidv7"
)

func ExampleGenerate() {
	now := time.Date(2026, 3, 9, 14, 20, 0, 123456700, time.UTC)
	uuidv7.Now = func() time.Time {
		now = now.Add(time.Second)
		return now
	}
	uuidv7.Rand = func() uint64 {
		return 0xFFFF_FFFF_FFFF_FFFF
	}

	id := uuidv7.Generate()
	fmt.Printf("%s %s\n", id, id.Time().UTC())
	id = uuidv7.Generate()
	fmt.Printf("%s %s\n", id, id.Time().UTC())

	// Output:
	// 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF 2026-03-09 14:20:01.123 +0000 UTC
	// 019CD2F8-1ECB-774E-BFFF-FFFFFFFFFFFF 2026-03-09 14:20:02.123 +0000 UTC
}

func ExampleParse() {
	id := uuidv7.Parse(``)
	fmt.Printf("%t %s %s\n", id.IsZero(), id, id.Time().UTC())

	id = uuidv7.Parse(`019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF`)
	fmt.Printf("%t %s %s\n", id.IsZero(), id, id.Time().UTC())

	// Output:
	// true 00000000-0000-0000-0000-000000000000 1970-01-01 00:00:00 +0000 UTC
	// false 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF 2026-03-09 14:20:01.123 +0000 UTC
}

func ExampleUUIDv7_Equal() {
	id := uuidv7.Parse(`019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF`)
	id2 := uuidv7.Parse(`019CD2F8-2AE3-774E-BFFF-FFFFFFFFFFFF`)

	err := id.Equal(id)
	fmt.Printf("%v\n", err)

	err = id.Equal(&id)
	fmt.Printf("%v\n", err)

	err = id.Equal(&id2)
	fmt.Printf("%v\n", err)

	// Output:
	// <nil>
	// <nil>
	// uuidv7: not equal, want 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF, got 019CD2F8-2AE3-774E-BFFF-FFFFFFFFFFFF
}

func ExampleUUIDv7_MarshalBinary() {
	now := time.Date(2026, 3, 9, 14, 20, 0, 123456700, time.UTC)
	uuidv7.Now = func() time.Time {
		now = now.Add(time.Second)
		return now
	}
	uuidv7.Rand = func() uint64 {
		return 0xFFFF_FFFF_FFFF_FFFF
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	id := uuidv7.Generate()
	err := enc.Encode(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: %s\n", id)

	dec := gob.NewDecoder(&buf)
	var gotID uuidv7.UUIDv7
	err = dec.Decode(&gotID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Receive: %s\n", gotID)

	// Output:
	// Sent: 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF
	// Receive: 019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF
}

func ExampleUUIDv7_MarshalText() {
	now := time.Date(2026, 3, 9, 14, 20, 0, 123456700, time.UTC)
	uuidv7.Now = func() time.Time {
		now = now.Add(time.Second)
		return now
	}
	uuidv7.Rand = func() uint64 {
		return 0xFFFF_FFFF_FFFF_FFFF
	}

	type T struct {
		ID uuidv7.UUIDv7
	}

	t := T{
		ID: uuidv7.Generate(),
	}
	jsonb, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", jsonb)

	var got T
	err = json.Unmarshal(jsonb, &got)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", got)

	// Output:
	// {"ID":"019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF"}
	// {ID:019CD2F8-1AE3-774E-BFFF-FFFFFFFFFFFF}
}

func ExampleUUIDv7_Scan() {
	var id uuidv7.UUIDv7

	// Scan the value from the database:
	//
	// dbc.QueryRow(`SELECT id ...`, &id)

	err := id.Scan([]byte(`017F22E2-79B0-7CC3-98C4-DC0C0C07398F`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", id)
	// Output:
	// 017F22E2-79B0-7CC3-98C4-DC0C0C07398F
}
