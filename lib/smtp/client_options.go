// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

// ClientOptions contains all options to create new client.
type ClientOptions struct {
	// LocalName define the client domain address, used when issuing EHLO
	// command to server.
	// If its empty, it will set to current operating system's
	// hostname.
	// The LocalName only has effect when client is connecting from
	// server-to-server.
	LocalName string

	// ServerUrl use the following format,
	//
	//	ServerUrl = [ scheme "://" ](domain | IP-address)[":" port]
	//	scheme    = "smtp" / "smtps" / "smtp+starttls"
	//
	// If scheme is "smtp" and no port is given, client will connect to
	// remote address at port 25.
	// If scheme is "smtps" and no port is given, client will connect to
	// remote address at port 465 (implicit TLS).
	// If scheme is "smtp+starttls" and no port is given, client will
	// connect to remote address at port 587.
	ServerURL string

	// The user name to authenticate to remote server.
	//
	// AuthUser and AuthPass enable automatic authentication when creating
	// new Client, as long as one is not empty.
	AuthUser string

	// The user password to authenticate to remote server.
	AuthPass string

	// The SASL mechanism used for authentication.
	AuthMechanism SaslMechanism

	// Insecure if set to true it will disable verifying remote certificate when
	// connecting with TLS or STARTTLS.
	Insecure bool
}
