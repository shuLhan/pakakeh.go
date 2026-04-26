// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package time

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func ExampleDate_MarshalText() {
	type Sample struct {
		Birthday Date `json:"birthday"`
	}

	sample := Sample{
		Birthday: CreateDate(2026, 4, 24, time.UTC),
	}

	jsonb, err := json.Marshal(sample)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(`%s`, jsonb)
	// Output:
	// {"birthday":"2026-04-24"}
}

func ExampleDate_UnmarshalText() {
	type Sample struct {
		Birthday Date `json:"birthday"`
	}

	jsonb := []byte(`{"birthday":"2026-04-26"}`)
	sample := Sample{}
	err := json.Unmarshal(jsonb, &sample)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(`Birthday: %s`, sample.Birthday.String())
	// Output:
	// Birthday: 2026-04-26
}
