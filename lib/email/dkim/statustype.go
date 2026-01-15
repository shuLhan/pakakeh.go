// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

// StatusType define type of status.
type StatusType byte

const (
	// StatusUnverify means that the signature has not been verified.
	StatusUnverify StatusType = iota

	// StatusNoSignature no dkim signature in message.
	StatusNoSignature

	// StatusOK the signature is valid.
	StatusOK

	// StatusTempFail the signature could not be verified at this time but
	// may be tried again later.
	StatusTempFail

	// StatusPermFail the signature failed and should not be reconsidered.
	StatusPermFail
)
