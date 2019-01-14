// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Mechanism represent Simple Authentication and Security Layer (SASL)
// mechanism (RFC 4422).
//
type Mechanism int

// List of available SASL mechanism.
const (
	MechanismPLAIN Mechanism = 1
)
