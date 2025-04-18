// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func FuzzParseZone(f *testing.F) {
	f.Skip()
	flagNoServer = true
	var tdata *test.Data
	var err error
	tdata, err = test.LoadData(`testdata/zone_fuzz_test.txt`)
	if err != nil {
		f.Fatal(`LoadData:`, err)
	}

	for origin, input := range tdata.Input {
		f.Add(input, origin)
	}
	f.Fuzz(func(t *testing.T, input []byte, origin string) {
		_, err = ParseZone(input, origin, 60)
		if err != nil {
			return
		}
	})
}
