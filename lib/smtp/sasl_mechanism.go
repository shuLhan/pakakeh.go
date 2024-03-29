// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

// SaslMechanism represent Simple Authentication and Security Layer (SASL)
// mechanism (RFC 4422).
type SaslMechanism int

// List of available SASL mechanism.
const (
	SaslMechanismPlain SaslMechanism = 1
)
