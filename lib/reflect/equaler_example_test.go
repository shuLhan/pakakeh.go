// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package reflect

import (
	"fmt"
	"log"
)

type ADT struct {
	vint int
}

func (rnp *ADT) Equal(v any) (err error) {
	var (
		logp = `Equal`
		got  *ADT
		ok   bool
	)
	got, ok = v.(*ADT)
	if !ok {
		return fmt.Errorf(`%s: v type is %T, want %T`, logp, got, v)
	}
	if rnp.vint != got.vint {
		return fmt.Errorf(`%s: vint: %d, want %d`,
			logp, got.vint, rnp.vint)
	}
	return nil
}

func ExampleEqualer() {
	var (
		rp1 = ADT{
			vint: 1,
		}
		rp2 = ADT{
			vint: 2,
		}
	)
	var err = DoEqual(&rp1, &rp2)
	if err == nil {
		log.Fatal(`expecting error, got nil`)
	}

	var exp = `Equal: vint: want 1, got 2`
	var got = err.Error()
	if exp != got {
		log.Fatalf(`want %q, got %q`, exp, got)
	}
}
