// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package smtp

import (
	"strings"
)

// ServerInfo provide information about server from response of EHLO or HELO
// command.
type ServerInfo struct {
	Exts   map[string][]string
	Domain string
	Info   string
}

// NewServerInfo create and initialize ServerInfo from EHLO/HELO response.
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
