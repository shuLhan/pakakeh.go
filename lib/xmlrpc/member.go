// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xmlrpc

import (
	"fmt"
)

//
// Member represent member element, as sub-item of struct element.
//
type Member struct {
	Name  string
	Value Value
}

func (m Member) String() string {
	return fmt.Sprintf("<member><name>%s</name>%s</member>", m.Name,
		m.Value.String())
}
