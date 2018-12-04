// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Environment define an interface for SMTP server environment.
//
type Environment interface {
	// Hostname return the primary domain name.
	Hostname() string

	// Domains return list of domains to be handled as final destination.
	Domains() []string
}
