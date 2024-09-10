// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

// ContextKey define a type for context.
type ContextKey byte

// List of valid context key.
const (
	CtxKeyExternalJWT ContextKey = 1 << iota
	CtxKeyInternalJWT
	CtxKeyUID
)
