// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

import (
	"testing"
)

func TestMetadataIsEqual(t *testing.T) {
	type testCase struct {
		in     Metadata
		out    Metadata
		result bool
	}
	var listCase = []testCase{{
		Metadata{
			Name:      `A`,
			Separator: `,`,
		},
		Metadata{
			Name:      `A`,
			Separator: `,`,
		},
		true,
	}, {
		Metadata{
			Name:      `A`,
			Separator: `,`,
		},
		Metadata{
			Name:      `A`,
			Separator: `;`,
		},
		false,
	}}

	var c testCase
	for _, c = range listCase {
		var got = c.in.IsEqual(&c.out)
		if got != c.result {
			t.Error(`Test failed on `, c.in, c.out)
		}
	}
}
