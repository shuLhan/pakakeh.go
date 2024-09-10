// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

// Opcode represent the websocket operation code.
type Opcode byte

// List of valid operation code in frame.
const (
	OpcodeCont        Opcode = 0x0
	OpcodeText        Opcode = 0x1
	OpcodeBin         Opcode = 0x2
	OpcodeDataRsv3    Opcode = 0x3 // %x3-7 are reserved for further non-control frames
	OpcodeDataRsv4    Opcode = 0x4
	OpcodeDataRsv5    Opcode = 0x5
	OpcodeDataRsv6    Opcode = 0x6
	OpcodeDataRsv7    Opcode = 0x7
	OpcodeClose       Opcode = 0x8
	OpcodePing        Opcode = 0x9
	OpcodePong        Opcode = 0xA
	OpcodeControlRsvB Opcode = 0xB // %xB-F are reserved for further control frames
	OpcodeControlRsvC Opcode = 0xC
	OpcodeControlRsvD Opcode = 0xD
	OpcodeControlRsvE Opcode = 0xE
	OpcodeControlRsvF Opcode = 0xF
)
