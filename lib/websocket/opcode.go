// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

type opcode byte

//
// List of valid operation code in frame.
//
const (
	opcodeCont  opcode = 0x0
	opcodeText  opcode = 0x1
	opcodeBin   opcode = 0x2
	opcodeClose opcode = 0x8
	opcodePing  opcode = 0x9
	opcodePong  opcode = 0xA
)
