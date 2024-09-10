// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

// CloseCode represent the server close status.
type CloseCode uint16

// List of close code in network byte order.  The name of status is
// mimicking the "net/http" status code.
//
// Endpoints MAY use the following pre-defined status codes when sending
// a Close frame.
//
// Status code 1004-1006, and 1015 is reserved and MUST NOT be used on Close
// payload.
//
// See RFC6455 7.4.1-P45 for more information.
const (
	// StatusNormal indicates a normal closure, meaning that the purpose
	// for which the connection was established has been fulfilled.
	StatusNormal CloseCode = 1000

	// StatusGone indicates that an endpoint is "going away", such as a
	// server going down or a browser having navigated away from a page.
	StatusGone = 1001

	// StatusBadRequest indicates that an endpoint is terminating the
	// connection due to a protocol error.
	StatusBadRequest = 1002

	// StatusUnsupportedType indicates that an endpoint is terminating the
	// connection because it has received a type of data it cannot accept
	// (e.g., an endpoint that understands only text data MAY send this if
	// it receives a binary message).
	StatusUnsupportedType = 1003

	// StatusInvalidData indicates that an endpoint is terminating
	// the connection because it has received data within a message that
	// was not consistent with the type of the message (e.g., non-UTF-8
	// [RFC3629] data within a text message).
	StatusInvalidData = 1007

	// StatusForbidden indicates that an endpoint is terminating the
	// connection because it has received a message that violates its
	// policy.
	// This is a generic status code that can be returned when there is no
	// other more suitable status code (e.g., 1003 or 1009) or if there is
	// a need to hide specific details about the policy.
	StatusForbidden = 1008

	// StatusRequestEntityTooLarge indicates that an endpoint is
	// terminating the connection because it has received a message that
	// is too big for it to process.
	StatusRequestEntityTooLarge = 1009

	// StatusBadGateway indicates that an endpoint (client) is
	// terminating the connection because it has expected the server to
	// negotiate one or more extension, but the server didn't return them
	// in the response message of the WebSocket handshake.
	// The list of extensions that are needed SHOULD appear in the
	// "reason" part of the Close frame.
	// Note that this status code is not used by the server, because it
	// can fail the WebSocket handshake instead.
	StatusBadGateway = 1010

	// StatusInternalError indicates that a server is terminating the
	// connection because it encountered an unexpected condition that
	// prevented it from fulfilling the request.
	StatusInternalError = 1011
)
