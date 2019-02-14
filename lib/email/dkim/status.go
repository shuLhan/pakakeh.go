// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// Status represent the result of DKIM verify.
//
type Status struct {
	Type  StatusType // Type of status.
	SDID  []byte     // SDID in signature ("d=" tag).
	Error error      // Error contains the cause of failed verification.
}
