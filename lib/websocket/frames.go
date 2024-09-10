// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

// Frames represent continuous (fragmented) frame.
//
// A fragmented message consists of a single frame with the FIN bit clear and
// an opcode other than 0, followed by zero or more frames with the FIN bit
// clear and the opcode set to 0, and terminated by a single frame with the
// FIN bit set and an opcode of 0.
type Frames struct {
	v []*Frame
}

// Unpack websocket data protocol from raw bytes to one or more frames.
//
// When receiving packet from client, the underlying protocol or operating
// system may buffered the packet.
// Client may send a single frame one at time, but server may receive one or
// more frame in one packet; and vice versa.
// That's the reason why the Unpack return multiple frame instead of
// single frame.
//
// On success it will return one or more frames.
// On fail it will return zero frame.
func Unpack(packet []byte) (frames *Frames) {
	if len(packet) == 0 {
		return
	}

	frames = &Frames{}

	var (
		f *Frame
	)

	for len(packet) > 0 {
		f = &Frame{}
		packet = f.unpack(packet)
		frames.Append(f)
	}

	return frames
}

// Append a frame as part of continuous frame.
// This function does not check if the appended frame is valid (i.e. zero
// operation code on second or later frame).
func (frames *Frames) Append(f *Frame) {
	if f != nil {
		frames.v = append(frames.v, f)
	}
}

// fin merge all continuous frame with last frame and return it as single
// frame.
func (frames *Frames) fin(last *Frame) (frame *Frame) {
	frame = frames.v[0]

	var x int
	for x = 1; x < len(frames.v); x++ {
		if frames.v[x].opcode == OpcodeClose {
			break
		}

		// Ignore control PING or PONG frame.
		if frames.v[x].opcode == OpcodePing ||
			frames.v[x].opcode == OpcodePong {
			continue
		}

		frame.payload = append(frame.payload, frames.v[x].payload...)
	}
	if last != nil {
		frame.payload = append(frame.payload, last.payload...)
	}

	return frame
}

// isClosed will return true if one of the frame is control CLOSE frame.
func (frames *Frames) isClosed() bool {
	if len(frames.v) == 0 {
		return false
	}

	var x int
	for ; x < len(frames.v); x++ {
		if frames.v[x].opcode == OpcodeClose {
			return true
		}
	}
	return false
}

// Opcode return the operation code of the first frame.
func (frames *Frames) Opcode() Opcode {
	if len(frames.v) == 0 {
		return OpcodeCont
	}
	return frames.v[0].opcode
}

// payload return the concatenation of continuous data frame's payload.
//
// The first frame must be a data frame, either text or binary, otherwise it
// will be considered empty payload, even if frames list is not empty.
//
// Any control CLOSE frame of frame with fin set will considered the last
// frame.
func (frames *Frames) payload() (payload []byte) {
	if len(frames.v) == 0 {
		return
	}
	if !frames.v[0].IsData() {
		return
	}

	var x int
	for ; x < len(frames.v); x++ {
		if frames.v[x].opcode == OpcodeClose {
			break
		}

		// Ignore control PING or PONG frame.
		if frames.v[x].opcode == OpcodePing ||
			frames.v[x].opcode == OpcodePong {
			continue
		}

		payload = append(payload, frames.v[x].payload...)

		if frames.v[x].fin == frameIsFinished {
			break
		}
	}

	return
}
