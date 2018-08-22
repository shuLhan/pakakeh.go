// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

//
// A Handler responds to DNS request.
//
type Handler interface {
	ServeDNS(*Request) *Response
}
