// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"
)

var (
	_testExternalJWT    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MzA0NjU2MDYsImhhc2giOiJiYmJiYmJiYi1iYmJiLWJiYmItYmJiYi1iYmJiYmJiYmJiYmIiLCJpYXQiOjE1MzAyMDY0MDYsIm5hZiI6MTUzMjc5ODQwNn0.15quj_gkeo9cWkLN98_2rXjtjihQym16Kn_9BQjYC14"
	_testInternalJWT    = _testExternalJWT
	_testUID            = 100
	_testPort           = 9001
	_testServer         *Server
	_testWSAddr         string
	_testHdrValWSAccept = "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	_testHdrValWSKey    = "dGhlIHNhbXBsZSBub25jZQ=="
	_testMaskKey        = [4]byte{'7', 'ú', '!', '='}
)

var (
	_dummyPayload256, _dummyPayload256Masked     = generateDummyPayload(256)
	_dummyPayload65536, _dummyPayload65536Masked = generateDummyPayload(65536)
)

func generateDummyPayload(size uint64) (payload []byte, masked []byte) {
	payload = make([]byte, size)
	masked = make([]byte, size)

	payload[0] = 'A'

	for x := uint64(1); x < size; x = x * 2 {
		copy(payload[x:], payload[:x])
	}

	for x := uint64(0); x < size; x++ {
		masked[x] = payload[x] ^ _testMaskKey[x%4]
	}

	return
}

//
// handleRequest from websocket by echo-ing back the payload.
//
func handleRequest(conn int, req *Frame) {
	req.Fin = FrameIsFinished
	req.Masked = 0

	err := SendFrame(conn, req, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "handleRequest: error:", err.Error())
	}
}

//
// testHandleAuth with token in query parameter
//
func testHandleAuth(req *Handshake) (ctx context.Context, err error) {
	URL, err := url.ParseRequestURI(string(req.URI))
	if err != nil {
		return
	}

	q := URL.Query()

	extJWT := q.Get(_qKeyTicket)
	if len(extJWT) == 0 {
		err = errors.New("Missing authorization")
		return
	}

	ctx = context.WithValue(context.Background(), CtxKeyExternalJWT, extJWT)
	ctx = context.WithValue(ctx, CtxKeyInternalJWT, _testInternalJWT)
	ctx = context.WithValue(ctx, CtxKeyUID, _testUID)

	return
}

func TestMain(m *testing.M) {
	var err error

	_testWSAddr = "ws://127.0.0.1:" + strconv.Itoa(_testPort) + "/"

	_testServer, err = NewServer(_testPort)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	_testServer.HandleText = handleRequest
	_testServer.HandleBin = handleRequest
	_testServer.HandleAuth = testHandleAuth

	go _testServer.Start()

	s := m.Run()

	os.Exit(s)
}
