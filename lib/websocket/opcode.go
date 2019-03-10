// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

type opcode byte

//
// List of valid operation code in frame.
//
const (
	opcodeCont        opcode = 0x0
	opcodeText        opcode = 0x1
	opcodeBin         opcode = 0x2
	opcodeDataRsv3    opcode = 0x3 // %x3-7 are reserved for further non-control frames
	opcodeDataRsv4    opcode = 0x4
	opcodeDataRsv5    opcode = 0x5
	opcodeDataRsv6    opcode = 0x6
	opcodeDataRsv7    opcode = 0x7
	opcodeClose       opcode = 0x8
	opcodePing        opcode = 0x9
	opcodePong        opcode = 0xA
	opcodeControlRsvB opcode = 0xB // %xB-F are reserved for further control frames
	opcodeControlRsvC opcode = 0xC
	opcodeControlRsvD opcode = 0xD
	opcodeControlRsvE opcode = 0xE
	opcodeControlRsvF opcode = 0xF
)
