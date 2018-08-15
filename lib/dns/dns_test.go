// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestQueryType(t *testing.T) {
	test.Assert(t, "QueryTypeA", QueryTypeA, QueryType(1), true)
	test.Assert(t, "QueryTypeTXT", QueryTypeTXT, QueryType(16), true)
	test.Assert(t, "QueryTypeAXFR", QueryTypeAXFR, QueryType(252), true)
	test.Assert(t, "QueryTypeALL", QueryTypeALL, QueryType(255), true)
}

func TestMain(m *testing.M) {
	debugLevel = 2

	os.Exit(m.Run())
}
