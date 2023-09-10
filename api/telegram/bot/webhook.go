// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

import "crypto/tls"

// Webhook contains options to set Webhook to receive updates.
type Webhook struct {
	// Optional.  The certificate for Bot server when listening for
	// Webhook.
	ListenCertificate *tls.Certificate

	// HTTPS url to send updates to.
	// This option will be overridden by environment variable
	// TELEGRAM_WEBHOOK_URL.
	URL string

	// Optional.  The address and port where the Bot will listen for
	// Webhook in the following format "<address>:<port>".
	// Default to ":80" if ListenCertificate is nil or ":443" if not nil.
	ListenAddress string

	// Optional. Upload your public key certificate so that the root
	// certificate in use can be checked.
	Certificate []byte

	// Optional. A JSON-serialized list of the update types you want your
	// bot to receive.
	// For example, specify ["message", "edited_channel_post",
	// "callback_query"] to only receive updates of these types.
	//
	// Specify an empty list to receive all updates regardless of type
	// (default). If not specified, the previous setting will be used.
	AllowedUpdates []string

	// Optional.
	// Maximum allowed number of simultaneous HTTPS connections
	// to the webhook for update delivery, 1-100.
	// Defaults to 40.
	// Use lower values to limit the load on your bot‘s server, and higher
	// values to increase your bot’s throughput.
	MaxConnections int
}
