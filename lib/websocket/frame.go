// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"encoding/binary"
	"math"
	"math/rand"
	"time"
)

//
// Frame represent a websocket data protocol.
//
type Frame struct {
	//
	// Websocket data protocol,
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

	// fin Indicates that this is the final fragment in a message.
	// The first fragment MAY also be the final fragment.
	fin byte

	// opcode (4 bits) defines the interpretation of the "Payload data".
	// If an unknown opcode is received, the receiving endpoint MUST _Fail
	// the WebSocket Connection_.  The following values are defined.
	opcode opcode

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
	maskKey [4]byte

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
}

//
// NewFrameBin create a single binary data frame with optional payload.
// Client frame must be masked.
//
func NewFrameBin(isMasked bool, payload []byte) []byte {
	return newFrame(opcodeBin, isMasked, payload)
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
		return newControlFrame(opcodeClose, isMasked, nil)
	}

	// If there is a body, the first two bytes of the body MUST be a
	// 2-byte unsigned integer (in network byte order) representing a
	// status code.
	packet := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(packet[:2], uint16(code))
	copy(packet[2:], payload)

	return newControlFrame(opcodeClose, isMasked, packet)
}

//
// NewFramePing create a masked PING control frame.
//
func NewFramePing(isMasked bool, payload []byte) (packet []byte) {
	return newControlFrame(opcodePing, isMasked, payload)
}

//
// NewFramePong create a masked PONG control frame to be used by client.
//
func NewFramePong(isMasked bool, payload []byte) (packet []byte) {
	return newControlFrame(opcodePong, isMasked, payload)
}

//
// NewFrameText create a single text data frame with optional payload.
// Client frame must be masked.
//
func NewFrameText(isMasked bool, payload []byte) []byte {
	return newFrame(opcodeText, isMasked, payload)
}

//
// newControlFrame create new control frame with specific operation code and
// optional payload.
//
func newControlFrame(opcode opcode, isMasked bool, payload []byte) []byte {
	if len(payload) > frameSmallPayload {
		// All control frames MUST have a payload length of 125 bytes
		// or less and MUST NOT be fragmented.
		payload = payload[:frameSmallPayload]
	}
	return newFrame(opcode, isMasked, payload)
}

//
// newFrame create a single frame with specific operation code and optional
// payload.
//
func newFrame(opcode opcode, isMasked bool, payload []byte) []byte {
	f := &Frame{
		fin:     frameIsFinished,
		opcode:  opcode,
		payload: payload,
	}
	if isMasked {
		f.masked = frameIsMasked
	}
	return f.pack(isMasked)
}

//
// frameUnpack unpack the websocket data protocol from raw bytes into single
// frame.
//
// On success it will return non nil frame, and the index to the rest of
// unprocessed packet.
// On fail, it will return nil frame.
//
func frameUnpack(in []byte) (f *Frame, x uint64) {
	if len(in) == 0 {
		return nil, 0
	}

	f = new(Frame)

	f.fin = in[x] & frameIsFinished
	f.opcode = opcode(in[x] & 0x0F)
	x++

	if len(in) >= 2 {
		f.masked = in[x] & frameIsMasked
		f.len = uint64(in[x] & 0x7F)
		x++
	}

	if f.opcode == opcodeClose || f.opcode == opcodePing || f.opcode == opcodePong {
		// (5.4-P33)
		if f.fin != frameIsFinished {
			return nil, x
		}
		// (5.5-P36)
		if f.len > frameSmallPayload {
			return nil, x
		}
	}

	if f.len == frameLargePayload {
		f.len = binary.BigEndian.Uint64(in[x : x+8])
		x += 8
	} else if f.len == frameMediumPayload {
		f.len = uint64(binary.BigEndian.Uint16(in[x : x+2]))
		x += 2
	}

	if f.masked == frameIsMasked {
		f.maskKey[0] = in[x]
		x++
		f.maskKey[1] = in[x]
		x++
		f.maskKey[2] = in[x]
		x++
		f.maskKey[3] = in[x]
		x++
	}

	if f.len > 0 {
		f.payload = make([]byte, f.len)
		copy(f.payload, in[x:])

		if f.masked == frameIsMasked {
			for y := uint64(0); y < f.len; y++ {
				f.payload[y] ^= f.maskKey[y%4]
			}
		}
	}
	x += f.len

	if f.opcode == opcodeClose {
		f.closeCode = CloseCode(binary.BigEndian.Uint16(f.payload[0:2]))
		f.payload = f.payload[2:]
	}

	return f, x
}

//
// IsData return true if frame is either text or binary data frame.
//
func (f *Frame) IsData() bool {
	return f.opcode == opcodeText || f.opcode == opcodeBin
}

//
// pack websocket Frame into packet that can be sent through network.
//
// Caller must set frame fields Fin, opcode, masked, and payload.
//
// Frame payload len will be set based on length of payload.
//
// Frame maskKey will be set randomly only if masked is set and randomMask
// parameter is true.
//
// A server MUST NOT mask any frames that it sends to the client. (
// (RFC 6455 5.1-P27).
//
func (f *Frame) pack(randomMask bool) (out []byte) {
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

	if f.len == frameLargePayload {
		binary.BigEndian.PutUint64(out[x:x+8], payloadSize)
		x += 8
	} else if f.len == frameMediumPayload {
		binary.BigEndian.PutUint16(out[x:x+2], uint16(payloadSize))
		x += 2
	}

	if randomMask {
		if _rng == nil {
			_rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		}
		binary.LittleEndian.PutUint32(f.maskKey[0:], _rng.Uint32())
	}

	if f.masked == frameIsMasked {
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
