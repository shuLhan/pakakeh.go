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
	Exts   map[string][]string
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
	srvInfo.Info = res.Message

	srvInfo.Exts = make(map[string][]string, len(res.Body))
	for _, body := range res.Body {
		extParams := strings.Split(body, " ")
		extName := strings.ToLower(extParams[0])
		srvInfo.Exts[extName] = extParams[1:]
	}

	return srvInfo
}
