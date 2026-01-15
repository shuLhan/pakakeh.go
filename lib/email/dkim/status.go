// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

// Status represent the result of DKIM verification.
type Status struct {
	Error error      // Error contains the cause of failed verification.
	SDID  []byte     // SDID in signature ("d=" tag).
	Type  StatusType // Type of status.
}
