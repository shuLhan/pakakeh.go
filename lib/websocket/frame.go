// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"encoding/binary"
	"math"
	"math/rand"
)

//
// Frame represent a WebSocket data protocol.
//
type Frame struct {
	// fin Indicates that this is the final fragment in a message.
	// The first fragment MAY also be the final fragment.
	fin byte

	// rsv1, rsv2, and rsv3 is reserved bits in frame.
	rsv1 byte
	rsv2 byte
	rsv3 byte

	// opcode (4 bits) defines the interpretation of the "Payload data".
	// If an unknown opcode is received, the receiving endpoint MUST _Fail
	// the WebSocket Connection_.  The following values are defined.
	opcode Opcode

	//
	// masked (1 bit) defines whether the "Payload data" is masked.
	// If set to 1, a masking key is present in masking-key, and this is
	// used to unmask the "Payload data" as per Section 5.3.  All frames
	// sent from client to server have this bit set to 1.
	//
	masked byte

	// closeCode represent the status of control frame close request.
	closeCode CloseCode

	//
	// len represent Payload length:  7 bits, 7+16 bits, or 7+64 bits
	//
	// The length of the "Payload data", in bytes: if 0-125, that is the
	// payload length.  If 126, the following 2 bytes interpreted as a
	// 16-bit unsigned integer are the payload length.  If 127, the
	// following 8 bytes interpreted as a 64-bit unsigned integer (the
	// most significant bit MUST be 0) are the payload length.  Multibyte
	// length quantities are expressed in network byte order.  Note that
	// in all cases, the minimal number of bytes MUST be used to encode
	// the length, for example, the length of a 124-byte-long string
	// can't be encoded as the sequence 126, 0, 124.  The payload length
	// is the length of the "Extension data" + the length of the
	// "Application data".  The length of the "Extension data" may be
	// zero, in which case the payload length is the length of the
	// "Application data".
	//
	len uint64

	//
	// maskKey:  0 or 4 bytes
	//
	// All frames sent from the client to the server are masked by a
	// 32-bit value that is contained within the frame.  This field is
	// present if the mask bit is set to 1 and is absent if the mask bit
	// is set to 0.  See Section 5.3 for further information on client-
	// to-server masking.
	//
	maskKey []byte

	//
	// Payload data:  (x+y) bytes
	//
	// The "Payload data" is defined as "Extension data" concatenated
	// with "Application data".
	//
	// Extension data:  x bytes
	//
	// The "Extension data" is 0 bytes unless an extension has been
	// negotiated.  Any extension MUST specify the length of the
	// "Extension data", or how that length may be calculated, and how
	// the extension use MUST be negotiated during the opening handshake.
	// If present, the "Extension data" is included in the total payload
	// length.
	//
	// Application data:  y bytes
	//
	// Arbitrary "Application data", taking up the remainder of the frame
	// after any "Extension data".  The length of the "Application data"
	// is equal to the payload length minus the length of the "Extension
	// data".
	//
	payload []byte

	//
	// chopped contains the unfinished frame, excluding mask keys and
	// payload.
	//
	chopped []byte

	// isComplete will be true if all frame's field completely filled.
	isComplete bool
}

//
// NewFrameBin create a single binary data frame with optional payload.
// Client frame must be masked.
//
func NewFrameBin(isMasked bool, payload []byte) []byte {
	return NewFrame(OpcodeBin, isMasked, payload)
}

//
// NewFrameClose create control CLOSE frame.
// The optional code represent the reason why the endpoint send the CLOSE
// frame, for closure.
// The optional payload represent the human readable reason, usually for
// debugging.
//
func NewFrameClose(isMasked bool, code CloseCode, payload []byte) []byte {
	if code == 0 {
		code = StatusNormal
	}

	// If there is a body, the first two bytes of the body MUST be a
	// 2-byte unsigned integer (in network byte order) representing a
	// status code.
	packet := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(packet[:2], uint16(code))
	copy(packet[2:], payload)

	return newControlFrame(OpcodeClose, isMasked, packet)
}

//
// NewFramePing create a masked PING control frame.
//
func NewFramePing(isMasked bool, payload []byte) (packet []byte) {
	return newControlFrame(OpcodePing, isMasked, payload)
}

//
// NewFramePong create a masked PONG control frame to be used by client.
//
func NewFramePong(isMasked bool, payload []byte) (packet []byte) {
	return newControlFrame(OpcodePong, isMasked, payload)
}

//
// NewFrameText create a single text data frame with optional payload.
// Client frame must be masked.
//
func NewFrameText(isMasked bool, payload []byte) []byte {
	return NewFrame(OpcodeText, isMasked, payload)
}

//
// newControlFrame create new control frame with specific operation code and
// optional payload.
//
func newControlFrame(opcode Opcode, isMasked bool, payload []byte) []byte {
	if len(payload) > frameSmallPayload {
		// All control frames MUST have a payload length of 125 bytes
		// or less and MUST NOT be fragmented.
		payload = payload[:frameSmallPayload]
	}
	return NewFrame(opcode, isMasked, payload)
}

//
// NewFrame create a single finished frame with specific operation code and
// optional payload.
//
func NewFrame(opcode Opcode, isMasked bool, payload []byte) []byte {
	f := &Frame{
		fin:     frameIsFinished,
		opcode:  opcode,
		payload: payload,
	}
	if isMasked {
		f.masked = frameIsMasked
	}
	return f.Pack()
}

//
// IsData return true if frame is either text or binary data frame.
//
func (f *Frame) IsData() bool {
	return f.opcode == OpcodeText || f.opcode == OpcodeBin
}

//
// isValid will return true if a frame is valid.
// If isMasked is true, the frame masked MUST be set, otherwise it will return
// false; and vice versa.
// Parameter allowRsv1, allowRsv2, and allowRsv3 are to allow one or more
// frame reserved bits to be set, in order.
// If reserved bit 1 is set but parameter allowRsv1 is false, it will return
// false; and so on.
// If its control frame the fin field should be set and payload must be less
// than 125.
//
func (f *Frame) isValid(isMasked, allowRsv1, allowRsv2, allowRsv3 bool) bool {
	if isMasked {
		if f.masked != frameIsMasked {
			return false
		}
	} else {
		if f.masked == frameIsMasked {
			return false
		}
	}
	if f.rsv1 > 0 && !allowRsv1 {
		return false
	}
	if f.rsv2 > 0 && !allowRsv2 {
		return false
	}
	if f.rsv3 > 0 && !allowRsv3 {
		return false
	}

	if f.opcode >= OpcodeClose {
		if f.fin == 0 {
			// Control frame must set the fin.
			return false
		}
		// Control frame payload must not larger than 125.
		if f.len > frameSmallPayload {
			return false
		}
	}

	return true
}

//
// Opcode return the frame operation code.
//
func (f *Frame) Opcode() Opcode {
	return f.opcode
}

//
// Pack WebSocket Frame into packet that can be written into socket.
//
// Frame payload len will be set based on length of payload.
//
// Frame maskKey will be set randomly only if its is empty.
//
// A server MUST NOT mask any frames that it sends to the client.
// (RFC 6455 5.1-P27).
//
func (f *Frame) Pack() (out []byte) {
	headerSize := uint64(2)
	payloadSize := uint64(len(f.payload))

	switch {
	case payloadSize > math.MaxUint16:
		f.len = frameLargePayload
		headerSize += 8
	case payloadSize > frameSmallPayload:
		f.len = frameMediumPayload
		headerSize += 2
	default:
		f.len = payloadSize
	}

	if f.masked == frameIsMasked {
		headerSize += 4
	}

	frameSize := headerSize + payloadSize
	out = make([]byte, frameSize)

	x := 0

	out[x] = f.fin | byte(f.opcode)
	x++

	out[x] = f.masked | uint8(f.len)
	x++

	switch f.len {
	case frameLargePayload:
		binary.BigEndian.PutUint64(out[x:x+8], payloadSize)
		x += 8
	case frameMediumPayload:
		binary.BigEndian.PutUint16(out[x:x+2], uint16(payloadSize))
		x += 2
	}

	if f.masked == frameIsMasked {
		if len(f.maskKey) != 4 {
			f.maskKey = make([]byte, 4)
			binary.LittleEndian.PutUint32(f.maskKey, rand.Uint32())
		}

		out[x] = f.maskKey[0]
		x++
		out[x] = f.maskKey[1]
		x++
		out[x] = f.maskKey[2]
		x++
		out[x] = f.maskKey[3]
		x++

		for y := uint64(0); y < payloadSize; y++ {
			out[x] = f.payload[y] ^ f.maskKey[y%4]
			x++
		}
	} else {
		copy(out[x:], f.payload)
	}

	return out
}

//
// Payload return the frame payload.
//
func (f *Frame) Payload() []byte {
	return f.payload
}

//
// unpack the WebSocket data protocol from raw bytes into single frame.
//
// On success it will return the rest of unpacked frame.
//
// WebSocket data protocol,
//
//	   0                   1                   2                   3
//	   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	  +-+-+-+-+-------+-+-------------+-------------------------------+
//	  |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
//	  |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
//	  |N|V|V|V|       |S|             |   (if payload len==126/127)   |
//	  | |1|2|3|       |K|             |                               |
//	  +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
//	  |     Extended payload length continued, if payload len == 127  |
//	  + - - - - - - - - - - - - - - - +-------------------------------+
//	  |                               |Masking-key, if MASK set to 1  |
//	  +-------------------------------+-------------------------------+
//	  | Masking-key (continued)       |          Payload Data         |
//	  +-------------------------------- - - - - - - - - - - - - - - - +
//	  :                     Payload Data continued ...                :
//	  + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
//	  |                     Payload Data continued ...                |
//	  +---------------------------------------------------------------+
//
func (f *Frame) unpack(packet []byte) []byte {
	var isHaveLen bool
	for !isHaveLen {
		switch len(f.chopped) {
		case 0:
			f.fin = packet[0] & frameIsFinished
			f.rsv1 = packet[0] & 0x40
			f.rsv2 = packet[0] & 0x20
			f.rsv3 = packet[0] & 0x10
			f.opcode = Opcode(packet[0] & 0x0F)
			f.chopped = append(f.chopped, packet[0])
			packet = packet[1:]
			if len(packet) == 0 {
				return nil
			}
		case 1:
			f.masked = packet[0] & frameIsMasked
			f.len = uint64(packet[0] & 0x7F)
			f.chopped = append(f.chopped, packet[0])
			packet = packet[1:]
		default:
			// We got the masked and len, lets check and get the
			// extended length.
			switch f.len {
			case frameLargePayload:
				if len(f.chopped) < 10 {
					exp := 10 - len(f.chopped)
					if len(packet) < exp {
						f.chopped = append(f.chopped, packet...)
						return nil
					}
					// chopped: 81 FF 0 0 0 1 0 0 = 10 - 8) = 2
					// exp: 0 0
					f.chopped = append(f.chopped, packet[:exp]...)
					f.len = binary.BigEndian.Uint64(f.chopped[2:10])
					packet = packet[exp:]
				}
			case frameMediumPayload:
				if len(f.chopped) < 4 {
					exp := 4 - len(f.chopped)
					if len(packet) < exp {
						f.chopped = append(f.chopped, packet...)
						return nil
					}
					f.chopped = append(f.chopped, packet[:exp]...)
					f.len = uint64(binary.BigEndian.Uint16(f.chopped[2:4]))
					packet = packet[exp:]
				}
			}
			isHaveLen = true
		}
	}
	if f.masked == frameIsMasked && len(f.maskKey) != 4 {
		if len(packet) == 0 {
			return nil
		}

		exp := 4 - len(f.maskKey)
		if len(packet) < exp {
			f.maskKey = append(f.maskKey, packet...)
			return nil
		}

		f.maskKey = append(f.maskKey, packet[:exp]...)

		packet = packet[exp:]
	}
	if f.len == 0 {
		if f.opcode == OpcodeClose {
			f.closeCode = StatusNormal
		}
		f.isComplete = true
		f.chopped = nil
		return packet
	}
	if len(packet) == 0 {
		return nil
	}

	exp := f.len - uint64(len(f.payload))
	if uint64(len(packet)) < exp {
		exp = uint64(len(packet))
	}

	if f.masked == frameIsMasked {
		start := len(f.payload) % 4
		for x := uint64(0); x < exp; x++ {
			packet[x] ^= f.maskKey[start%4]
			start++
		}
	}

	f.payload = append(f.payload, packet[:exp]...)
	packet = packet[exp:]

	if uint64(len(f.payload)) == f.len {
		if f.opcode == OpcodeClose {
			switch len(f.payload) {
			case 0:
				f.closeCode = StatusNormal
			case 1:
				f.closeCode = StatusBadRequest
			default:
				f.closeCode = CloseCode(binary.BigEndian.Uint16(f.payload[:2]))
			}
		}
		f.isComplete = true
		f.chopped = nil
	}

	return packet
}
