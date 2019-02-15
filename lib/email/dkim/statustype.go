// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// StatusType define type of status.
//
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
