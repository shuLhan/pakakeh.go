// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"log"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	testDefaultSection = newConfigSection()
	testParser         *configParser
)

func TestMain(m *testing.M) {
	var err error

	testParser, err = newConfigParser()
	if err != nil {
		log.Fatal(err)
	}

	testDefaultSection.postConfig(testParser.homeDir)

	os.Exit(m.Run())
}

func TestPatternToRegex(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in:  "*",
		exp: ".*",
	}, {
		in:  "?",
		exp: ".?",
	}, {
		in:  "192.*",
		exp: `192\..*`,
	}}

	for _, c := range cases {
		got := patternToRegex(c.in)
		test.Assert(t, "patternToRegex", c.exp, got, true)
	}
}
