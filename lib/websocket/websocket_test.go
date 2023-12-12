// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	_testExternalJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MzA0NjU2MDYsImhhc2giOiJiYmJiYmJiYi1iYmJiLWJiYmItYmJiYi1iYmJiYmJiYmJiYmIiLCJpYXQiOjE1MzAyMDY0MDYsIm5hZiI6MTUzMjc5ODQwNn0.15quj_gkeo9cWkLN98_2rXjtjihQym16Kn_9BQjYC14"

	_testEndpointAuth string
	_testInternalJWT  = _testExternalJWT
	_testUID          = uint64(100)
	_testPort         = 9001
	_testServer       *Server
	_testAddr         string
	_testWSAddr       string
	_testHdrValWSKey  = "dGhlIHNhbXBsZSBub25jZQ=="
	_testMaskKey      = []byte{'7', 'ú', '!', '='}

	_dummyPayload256, _dummyPayload256Masked     = generateDummyPayload(256)
	_dummyPayload65536, _dummyPayload65536Masked = generateDummyPayload(65536)
)

func generateDummyPayload(size uint64) (payload []byte, masked []byte) {
	payload = make([]byte, size)
	masked = make([]byte, size)

	payload[0] = 'A'

	var x uint64
	for x = 1; x < size; x *= 2 {
		copy(payload[x:], payload[:x])
	}

	for x = 0; x < size; x++ {
		masked[x] = payload[x] ^ _testMaskKey[x%4]
	}

	return
}

// testHandleText from websocket by echo-ing back the payload.
func testHandleText(conn int, payload []byte) {
	var (
		packet = NewFrameText(false, payload)
		err    error
	)

	err = Send(conn, packet, 1*time.Second)
	if err != nil {
		log.Println("handlePayloadText: " + err.Error())
	}
}

// testHandleBin from websocket by echo-ing back the payload.
func testHandleBin(conn int, payload []byte) {
	var (
		packet = NewFrameBin(false, payload)
		err    error
	)

	err = Send(conn, packet, 1*time.Second)
	if err != nil {
		log.Println("handlePayloadBin: " + err.Error())
	}
}

// testHandleAuth with token in query parameter
func testHandleAuth(req *Handshake) (ctx context.Context, err error) {
	var (
		q = req.URL.Query()

		extJWT string
	)

	extJWT = q.Get(_qKeyTicket)
	if len(extJWT) == 0 {
		return nil, fmt.Errorf("Missing authorization")
	}

	ctx = context.WithValue(context.Background(), CtxKeyExternalJWT, extJWT)
	ctx = context.WithValue(ctx, CtxKeyInternalJWT, _testInternalJWT)
	ctx = context.WithValue(ctx, CtxKeyUID, _testUID)

	return
}

func runTestServer() {
	_testAddr = "127.0.0.1:" + strconv.Itoa(_testPort)
	_testWSAddr = "ws://" + _testAddr + "/"
	_testEndpointAuth = _testWSAddr + "?" + _qKeyTicket + "=" +
		_testExternalJWT

	var (
		opts = &ServerOptions{
			Address:    _testAddr,
			HandleAuth: testHandleAuth,
			HandleBin:  testHandleBin,
			HandleText: testHandleText,
		}

		err error
	)
	_testServer = NewServer(opts)

	go func() {
		err = _testServer.Start()
		if err != nil {
			log.Fatal("runTestServer: " + err.Error())
		}
	}()

	time.Sleep(500 * time.Millisecond)
}

func TestMain(m *testing.M) {
	runTestServer()

	os.Exit(m.Run())
}
