// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package sftp

import "encoding/binary"

// extensions contains mapping of extension-pair name and data, as defined in
// #section-4.2.
type extensions map[string]string

func unpackExtensions(payload []byte) (exts extensions) {
	exts = extensions{}
	for len(payload) > 0 {
		v := binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		name := string(payload[:v])
		payload = payload[v:]

		v = binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		exts[name] = string(payload[:v])
		payload = payload[v:]
	}
	return exts
}
