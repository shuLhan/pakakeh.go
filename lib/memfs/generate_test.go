// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import "testing"

func TestGenerate(t *testing.T) {
	mfs, err := New("testdata", nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	err = mfs.GoGenerate("test", "./generate_test/gen_test.go", "")
	if err != nil {
		t.Fatal(err)
	}
}
