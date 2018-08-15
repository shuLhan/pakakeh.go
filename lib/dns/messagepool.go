// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync"
)

var msgPool = sync.Pool{
	New: func() interface{} {
		msg := &Message{
			Header: &SectionHeader{
				IsQuery: true,
				Op:      OpCodeQuery,
			},
			Question: &SectionQuestion{
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
		}

		return msg
	},
}
