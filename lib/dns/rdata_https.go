// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"io"
)

// RDataHTTPS the resource record for type 65 [HTTPS RR].
//
// [HTTPS RR]: https://datatracker.ietf.org/doc/html/rfc9460
type RDataHTTPS struct {
	RDataSVCB
}

// WriteTo write the SVCB record as zone format to out.
func (https *RDataHTTPS) WriteTo(out io.Writer) (_ int64, err error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `HTTPS %d %s`, https.Priority, https.TargetName)

	var (
		keys = https.keys()

		keyid int
	)
	for _, keyid = range keys {
		buf.WriteByte(' ')

		if keyid == svcbKeyIDNoDefaultALPN {
			buf.WriteString(svcbKeyNameNoDefaultALPN)
			continue
		}

		https.writeParam(&buf, keyid)
	}
	buf.WriteByte('\n')

	var n int

	n, err = out.Write(buf.Bytes())

	return int64(n), err
}
