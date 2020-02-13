// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestToMapStringFloat64(t *testing.T) {
	in := map[string]interface{}{
		"string": "1",
		"zero":   "0",
		"byte":   byte(3),
		"[]byte": []byte("4"),
	}

	exp := map[string]float64{
		"string": 1,
		"byte":   3,
		"[]byte": 4,
	}

	got, err := ToMapStringFloat64(in)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "ToMapStringFloat64", exp, got, true)
}
