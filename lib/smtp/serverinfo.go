// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"strings"
)

//
// ServerInfo provide information about server from response of EHLO or HELO
// command.
//
type ServerInfo struct {
	Domain string
	Info   string
	Exts   []string
}

//
// NewServerInfo create and initialize ServerInfo from EHLO/HELO response.
//
func NewServerInfo(res *Response) (srvInfo *ServerInfo) {
	if res == nil {
		return nil
	}

	srvInfo = &ServerInfo{}

	domInfo := strings.Split(res.Message, " ")

	srvInfo.Domain = domInfo[0]
	if len(domInfo) == 2 {
		srvInfo.Info = domInfo[1]
	}

	srvInfo.Exts = make([]string, len(res.Body))
	for x, body := range res.Body {
		extParam := strings.Split(body, " ")
		srvInfo.Exts[x] = strings.ToLower(extParam[0])
	}

	return srvInfo
}
