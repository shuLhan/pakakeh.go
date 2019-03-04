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
// List of valid operation code in frame.
//
const (
	OpCodeCont  = 0x0
	OpCodeText  = 0x1
	OpCodeBin   = 0x2
	OpCodeClose = 0x8
	OpCodePing  = 0x9
	OpCodePong  = 0xA
)

// List of frame length.
const (
	FrameSmallPayload  = 125
	FrameMediumPayload = 126
	FrameLargePayload  = 127
)

// List of frame FIN and MASK values.
const (
	FrameIsFinished = 0x80
	FrameIsMasked   = 0x80
)

//
// List of close code in network byte order.  The name of status is
// mimicking the "net/http" status code.
//
// Endpoints MAY use the following pre-defined status codes when sending
// a Close frame.
//
// Status code 1004-1006, and 1015 is reserved and MUST NOT be used on Close
// payload.
//
// See RFC6455 7.4.1-P45 for more information.
//
var (
	// StatusNormal (1000) indicates a normal closure, meaning that the
	// purpose for which the connection was established has been
	// fulfilled.
	StatusNormal = []byte{0x03, 0xE8} //nolint: gochecknoglobals

	// StatusGone (1001) indicates that an endpoint is "going away", such
	// as a server going down or a browser having navigated away from a
	// page.
	StatusGone = []byte{0x03, 0xE9} //nolint: gochecknoglobals

	// StatusBadRequest (1002) indicates that an endpoint is terminating
	// the connection due to a protocol error.
	StatusBadRequest = []byte{0x03, 0xEA} //nolint: gochecknoglobals

	// StatusUnsupportedType (1003) indicates that an endpoint is
	// terminating the connection because it has received a type of data
	// it cannot accept (e.g., an endpoint that understands only text data
	// MAY send this if it receives a binary message).
	StatusUnsupportedType = []byte{0x03, 0xEB} //nolint: gochecknoglobals

	// StatusInvalidData (1007) indicates that an endpoint is terminating
	// the connection because it has received data within a message that
	// was not consistent with the type of the message (e.g., non-UTF-8
	// [RFC3629] data within a text message).
	StatusInvalidData = []byte{0x03, 0xEF} //nolint: gochecknoglobals

	// StatusForbidden (1008) indicates that an endpoint is terminating
	// the connection because it has received a message that violates its
	// policy.  This is a generic status code that can be returned when
	// there is no other more suitable status code (e.g., 1003 or 1009) or
	// if there is a need to hide specific details about the policy.
	StatusForbidden = []byte{0x03, 0xF0} //nolint: gochecknoglobals

	// StatusRequestEntityTooLarge (1009) indicates that an endpoint is
	// terminating the connection because it has received a message that
	// is too big for it to process.
	StatusRequestEntityTooLarge = []byte{0x03, 0xF1} //nolint: gochecknoglobals

	// StatusBadGateway (1010) indicates that an endpoint (client) is
	// terminating the connection because it has expected the server to
	// negotiate one or more extension, but the server didn't return them
	// in the response message of the WebSocket handshake.  The list of
	// extensions that are needed SHOULD appear in the /reason/ part of
	// the Close frame.  Note that this status code is not used by the
	// server, because it can fail the WebSocket handshake instead.
	StatusBadGateway = []byte{0x03, 0xF2} //nolint: gochecknoglobals

	// StatusInternalError or 1011 indicates that a server is terminating
	// the connection because it encountered an unexpected condition that
	// prevented it from fulfilling the request.
	StatusInternalError = []byte{0x03, 0xF3} //nolint: gochecknoglobals
)

// List of unmasked control frames, MUST used only by server.
var (
	ControlFrameClose         = []byte{FrameIsFinished | OpCodeClose, 0x00} //nolint: gochecknoglobals
	ControlFrameCloseWithCode = []byte{FrameIsFinished | OpCodeClose, 0x02} //nolint: gochecknoglobals
	ControlFramePing          = []byte{FrameIsFinished | OpCodePing, 0x00}  //nolint: gochecknoglobals
	ControlFramePong          = []byte{FrameIsFinished | OpCodePong, 0x00}  //nolint: gochecknoglobals
)

//
// Frame represent a websocket data protocol.
//
//	5.2 Base Framing Protocol
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
//	Mask:  1 bit
//
//	   Defines whether the "Payload data" is masked.  If set to 1, a
//	   masking key is present in masking-key, and this is used to unmask
//	   the "Payload data" as per Section 5.3.  All frames sent from
//	   client to server have this bit set to 1.
//
//	Payload length:  7 bits, 7+16 bits, or 7+64 bits
//
//	   The length of the "Payload data", in bytes: if 0-125, that is the
//	   payload length.  If 126, the following 2 bytes interpreted as a
//	   16-bit unsigned integer are the payload length.  If 127, the
//	   following 8 bytes interpreted as a 64-bit unsigned integer (the
//	   most significant bit MUST be 0) are the payload length.  Multibyte
//	   length quantities are expressed in network byte order.  Note that
//	   in all cases, the minimal number of bytes MUST be used to encode
//	   the length, for example, the length of a 124-byte-long string
//	   can't be encoded as the sequence 126, 0, 124.  The payload length
//	   is the length of the "Extension data" + the length of the
//	   "Application data".  The length of the "Extension data" may be
//	   zero, in which case the payload length is the length of the
//	   "Application data".
//
//	Masking-key:  0 or 4 bytes
//
//	   All frames sent from the client to the server are masked by a
//	   32-bit value that is contained within the frame.  This field is
//	   present if the mask bit is set to 1 and is absent if the mask bit
//	   is set to 0.  See Section 5.3 for further information on client-
//	   to-server masking.
//
//	Payload data:  (x+y) bytes
//
//	   The "Payload data" is defined as "Extension data" concatenated
//	   with "Application data".
//
//	Extension data:  x bytes
//
//	   The "Extension data" is 0 bytes unless an extension has been
//	   negotiated.  Any extension MUST specify the length of the
//	   "Extension data", or how that length may be calculated, and how
//	   the extension use MUST be negotiated during the opening handshake.
//	   If present, the "Extension data" is included in the total payload
//	   length.
//
//	Application data:  y bytes
//
//	   Arbitrary "Application data", taking up the remainder of the frame
//	   after any "Extension data".  The length of the "Application data"
//	   is equal to the payload length minus the length of the "Extension
//	   data".
//
type Frame struct {
	Fin    byte
	Opcode byte
	Masked byte
	// closeCode represent the status of control frame close request.
	closeCode uint16
	len       uint64
	maskKey   [4]byte
	Payload   []byte
}

//
// NewFrameBin create a single binary data frame with optional payload.
// Client frame must be masked.
//
func NewFrameBin(isMasked bool, payload []byte) []byte {
	return newFrame(OpCodeBin, isMasked, payload)
}

//
// NewFrameClose create a masked CLOSE control frame.
// Server must use predefined, unmasked, packet ControlFrameClose, while
// client frame must be masked.
//
func NewFrameClose(payload []byte) []byte {
	return newControlFrame(OpCodeClose, payload)
}

//
// NewFramePing create a masked PING control frame.
// Server must use predefined unmasked packet ControlFramePing, while client
// frame must be masked.
//
func NewFramePing(payload []byte) (packet []byte) {
	return newControlFrame(OpCodePing, payload)
}

//
// NewFramePong create a masked PONG control frame to be used by client.
// Server must use predefined unmasked packet ControlFramePong.
// Client frame must be masked.
//
func NewFramePong(payload []byte) (packet []byte) {
	return newControlFrame(OpCodePong, payload)
}

//
// NewFrameText create a single text data frame with optional payload.
// Client frame must be masked.
//
func NewFrameText(isMasked bool, payload []byte) []byte {
	return newFrame(OpCodeText, isMasked, payload)
}

//
// newControlFrame create new control frame with specific operation code and
// optional payload.
//
func newControlFrame(opcode byte, payload []byte) []byte {
	if len(payload) > FrameSmallPayload {
		// All control frames MUST have a payload length of 125 bytes
		// or less and MUST NOT be fragmented.
		payload = payload[:FrameSmallPayload]
	}
	return newFrame(opcode, true, payload)
}

//
// newFrame create a single frame with specific operation code and optional
// payload.
//
func newFrame(opcode byte, isMasked bool, payload []byte) []byte {
	f := &Frame{
		Fin:     FrameIsFinished,
		Opcode:  opcode,
		Payload: payload,
	}
	if isMasked {
		f.Masked = FrameIsMasked
	}
	return f.Pack(isMasked)
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

	f.Fin = in[x] & FrameIsFinished
	f.Opcode = in[x] & 0x0F
	x++

	if len(in) >= 2 {
		f.Masked = in[x] & FrameIsMasked
		f.len = uint64(in[x] & 0x7F)
		x++
	}

	if f.Opcode == OpCodeClose || f.Opcode == OpCodePing || f.Opcode == OpCodePong {
		// (5.4-P33)
		if f.Fin != FrameIsFinished {
			return nil, x
		}
		// (5.5-P36)
		if f.len > FrameSmallPayload {
			return nil, x
		}
	}

	if f.len == FrameLargePayload {
		f.len = binary.BigEndian.Uint64(in[x : x+8])
		x += 8
	} else if f.len == FrameMediumPayload {
		f.len = uint64(binary.BigEndian.Uint16(in[x : x+2]))
		x += 2
	}

	if f.Masked == FrameIsMasked {
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
		f.Payload = make([]byte, f.len)
		copy(f.Payload, in[x:])

		if f.Masked == FrameIsMasked {
			for y := uint64(0); y < f.len; y++ {
				f.Payload[y] ^= f.maskKey[y%4]
			}
		}
	}
	x += f.len

	if f.Opcode == OpCodeClose {
		f.closeCode = binary.BigEndian.Uint16(f.Payload[0:2])
	}

	return f, x
}

//
// Unpack websocket data protocol from raw bytes to one or more frames.
//
// On success it will return one or more frames.
// On fail it will return zero frame.
//
func Unpack(in []byte) (fs []*Frame) {
	if len(in) == 0 {
		return
	}

	for {
		f, x := frameUnpack(in)
		if f == nil {
			break
		}

		fs = append(fs, f)

		if x >= uint64(len(in)) {
			break
		}

		in = in[x:]
	}

	return
}

//
// IsData return true if frame is either text or binary data frame.
//
func (f *Frame) IsData() bool {
	return f.Opcode == OpCodeText || f.Opcode == OpCodeBin
}

//
// Pack websocket Frame into packet that can be sent through network.
//
// Caller must set frame fields Fin, Opcode, Masked, and Payload.
//
// Frame payload len will be set based on length of payload.
//
// Frame maskKey will be set randomly only if Masked is set and randomMask
// parameter is true.
//
//	RFC6455 5.1-P27
//	A server MUST NOT mask any frames that it sends to the client.
//
func (f *Frame) Pack(randomMask bool) (out []byte) {
	headerSize := uint64(2)
	payloadSize := uint64(len(f.Payload))

	switch {
	case payloadSize > math.MaxUint16:
		f.len = FrameLargePayload
		headerSize += 8
	case payloadSize > FrameSmallPayload:
		f.len = FrameMediumPayload
		headerSize += 2
	default:
		f.len = payloadSize
	}

	if f.Masked == FrameIsMasked {
		headerSize += 4
	}

	frameSize := headerSize + payloadSize
	out = make([]byte, frameSize)

	x := 0

	out[x] = f.Fin | f.Opcode
	x++

	out[x] = f.Masked | uint8(f.len)
	x++

	if f.len == FrameLargePayload {
		binary.BigEndian.PutUint64(out[x:x+8], payloadSize)
		x += 8
	} else if f.len == FrameMediumPayload {
		binary.BigEndian.PutUint16(out[x:x+2], uint16(payloadSize))
		x += 2
	}

	if randomMask {
		if _rng == nil {
			_rng = rand.New(rand.NewSource(time.Now().UnixNano()))
		}
		binary.LittleEndian.PutUint32(f.maskKey[0:], _rng.Uint32())
	}

	if f.Masked == FrameIsMasked {
		out[x] = f.maskKey[0]
		x++
		out[x] = f.maskKey[1]
		x++
		out[x] = f.maskKey[2]
		x++
		out[x] = f.maskKey[3]
		x++

		for y := uint64(0); y < payloadSize; y++ {
			out[x] = f.Payload[y] ^ f.maskKey[y%4]
			x++
		}
	} else {
		copy(out[x:], f.Payload)
	}

	return out
}
