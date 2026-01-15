// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package dkim

// QueryMethod define a type and option to retrieve public key.
// An empty QueryMethod will use default type and option, which is "dns/txt".
type QueryMethod struct {
	Type   QueryType
	Option QueryOption
}
