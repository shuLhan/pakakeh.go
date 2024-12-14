// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package reflect

import (
	"fmt"
	"testing"
)

type recvNotPointer struct {
	vint int
}

func (rnp recvNotPointer) Equal(v any) (err error) {
	var (
		logp = `Equal`
		got  *recvNotPointer
		ok   bool
	)
	got, ok = v.(*recvNotPointer)
	if !ok {
		return fmt.Errorf(`%s: v type is %T, want %T`, logp, v, got)
	}
	if rnp.vint != got.vint {
		return fmt.Errorf(`%s: vint: want %d, got %d`,
			logp, rnp.vint, got.vint)
	}
	return nil
}

func TestEqualerRecvNotPointer(t *testing.T) {
	var (
		rnp1 = recvNotPointer{
			vint: 1,
		}
		rnp2 = recvNotPointer{
			vint: 2,
		}
	)

	var err = DoEqual(&rnp1, &rnp2)
	if err == nil {
		t.Fatal(`expecting error, got nil`)
	}

	var exp = `Equal: vint: want 1, got 2`
	var got = err.Error()
	if exp != got {
		t.Fatalf(`want %q, got %q`, exp, got)
	}

	var rnp3 = recvNotPointer{
		vint: 1,
	}
	err = DoEqual(&rnp1, &rnp3)
	if err != nil {
		t.Fatalf(`expecting no error, got %s`, err)
	}
}
