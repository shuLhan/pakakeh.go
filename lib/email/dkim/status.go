// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

// Status represent the result of DKIM verification.
type Status struct {
	Error error      // Error contains the cause of failed verification.
	SDID  []byte     // SDID in signature ("d=" tag).
	Type  StatusType // Type of status.
}
