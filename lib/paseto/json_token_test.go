// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import (
	"fmt"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestJSONToken_Validate(t *testing.T) {
	now := time.Now().Round(time.Second)
	peer := Key{}

	issued1sAgo := now.Add(-1 * time.Second)
	issued6sAgo := now.Add(-6 * time.Second)

	cases := []struct {
		desc   string
		jtoken *JSONToken
		expErr string
	}{{
		desc: "With IssuedAt less than current time",
		jtoken: &JSONToken{
			IssuedAt: &issued1sAgo,
		},
	}, {
		desc: "With IssuedAt greater than drift",
		jtoken: &JSONToken{
			IssuedAt: &issued6sAgo,
		},
		expErr: fmt.Sprintf("token issued at %s before current time %s",
			issued6sAgo, now),
	}}

	for _, c := range cases {
		var gotErr string

		err := c.jtoken.Validate("", peer)
		if err != nil {
			gotErr = err.Error()
		}

		test.Assert(t, c.desc, c.expErr, gotErr)
	}
}
