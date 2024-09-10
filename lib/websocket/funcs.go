// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

import (
	"crypto/rand"
	"crypto/sha1" //nolint:gosec
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// maxBuffer define maximum payload that we read/write from socket at one
// time.
// This number should be lower than MTU for better handling larger payload.
const maxBuffer = 1024

// Recv read packet from socket fd.
// The timeout parameter is optional, define the timeout when reading from
// socket.
// If timeout is zero the Recv operation will block until a data arrived.
// If timeout is greater than zero, the Recv operation will return
// os.ErrDeadlineExceeded when no data received after timeout duration.
func Recv(fd int, timeout time.Duration) (packet []byte, err error) {
	var (
		logp    = `Recv`
		buf     = make([]byte, maxBuffer)
		timeval = unix.Timeval{}
	)

	err = unix.SetNonblock(fd, false)
	if err != nil {
		return nil, fmt.Errorf(`%s: SetNonblock: %w`, logp, err)
	}

	if timeout > 0 {
		timeval.Sec = int64(timeout.Seconds())
		err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &timeval)
		if err != nil {
			return nil, fmt.Errorf(`%s: SetsockoptTimeval: %w`, logp, err)
		}
	}

	var n int
	for {
		n, err = unix.Read(fd, buf)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				return nil, fmt.Errorf(`%s: %w`, logp, os.ErrDeadlineExceeded)
			}
			return nil, fmt.Errorf(`%s: Read: %w`, logp, err)
		}
		if n > 0 {
			packet = append(packet, buf[:n]...)
		}
		if n < maxBuffer {
			break
		}
	}

	return packet, nil
}

// Send the packet through socket file descriptor fd.
// The timeout parameter is optional, its define the maximum duration when
// socket write should wait before considered fail.
// If timeout is zero, Send will block until buffer is available.
// If timeout is greater than zero, and Send has wait for this duration for
// buffer available then it will return os.ErrDeadlineExceeded.
func Send(fd int, packet []byte, timeout time.Duration) (err error) {
	var (
		logp    = `Send`
		timeval = unix.Timeval{}

		max int
		n   int
	)

	err = unix.SetNonblock(fd, false)
	if err != nil {
		return fmt.Errorf(`%s: SetNonblock: %w`, logp, err)
	}

	if timeout > 0 {
		timeval.Sec = int64(timeout.Seconds())
		err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, &timeval)
		if err != nil {
			return fmt.Errorf(`%s: SetsockoptTimeval: %w`, logp, err)
		}
	}

	for len(packet) > 0 {
		if len(packet) < maxBuffer {
			max = len(packet)
		} else {
			max = maxBuffer
		}

		n, err = unix.Write(fd, packet[:max])
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				return fmt.Errorf(`%s: %w`, logp, os.ErrDeadlineExceeded)
			}
			return fmt.Errorf(`%s: Write: %w`, logp, err)
		}

		if n > 0 {
			packet = packet[n:]
		}
	}

	return nil
}

// generateHandshakeAccept generate server accept key by concatenating key,
// defined in step 4 in Section 4.2.2, with the string
// "258EAFA5-E914-47DA-95CA-C5AB0DC85B11", taking the SHA-1 hash of this
// concatenated value to obtain a 20-byte value and base64-encoding (see
// Section 4 of [RFC4648]) this 20-byte hash.
func generateHandshakeAccept(key []byte) string {
	key = append(key, "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"...)
	var sum = sha1.Sum(key) //nolint:gosec
	return base64.StdEncoding.EncodeToString(sum[:])
}

// generateHandshakeKey randomly selected 16-byte value that has been
// base64-encoded (see Section 4 of [RFC4648]).
func generateHandshakeKey() (key []byte) {
	var (
		bkey = make([]byte, 16)

		err error
	)

	_, err = rand.Read(bkey)
	if err != nil {
		log.Panicf(`generateHandshakeKey: %s`, err)
	}

	key = make([]byte, base64.StdEncoding.EncodedLen(len(bkey)))
	base64.StdEncoding.Encode(key, bkey)

	return key
}
