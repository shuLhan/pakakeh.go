// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"strconv"
	"testing"
)

var (
	_testExternalJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MzA0NjU2MDYsImhhc2giOiJiYmJiYmJiYi1iYmJiLWJiYmItYmJiYi1iYmJiYmJiYmJiYmIiLCJpYXQiOjE1MzAyMDY0MDYsIm5hZiI6MTUzMjc5ODQwNn0.15quj_gkeo9cWkLN98_2rXjtjihQym16Kn_9BQjYC14" //nolint: lll, gochecknoglobals

	_testInternalJWT    = _testExternalJWT               //nolint: gochecknoglobals
	_testUID            = 100                            //nolint: gochecknoglobals
	_testPort           = 9001                           //nolint: gochecknoglobals
	_testServer         *Server                          //nolint: gochecknoglobals
	_testWSAddr         string                           //nolint: gochecknoglobals
	_testHdrValWSAccept = "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" //nolint: gochecknoglobals
	_testHdrValWSKey    = "dGhlIHNhbXBsZSBub25jZQ=="     //nolint: gochecknoglobals
	_testMaskKey        = [4]byte{'7', 'ú', '!', '='}    //nolint: gochecknoglobals
)

var (
	_dummyPayload256, _dummyPayload256Masked     = generateDummyPayload(256)   //nolint: gochecknoglobals
	_dummyPayload65536, _dummyPayload65536Masked = generateDummyPayload(65536) //nolint: gochecknoglobals
)

func generateDummyPayload(size uint64) (payload []byte, masked []byte) {
	payload = make([]byte, size)
	masked = make([]byte, size)

	payload[0] = 'A'

	for x := uint64(1); x < size; x *= 2 {
		copy(payload[x:], payload[:x])
	}

	for x := uint64(0); x < size; x++ {
		masked[x] = payload[x] ^ _testMaskKey[x%4]
	}

	return
}

//
// testHandleText from websocket by echo-ing back the payload.
//
func testHandleText(conn int, payload []byte) {
	packet := NewFrameText(false, payload)

	err := Send(conn, packet)
	if err != nil {
		log.Println("handlePayloadText: " + err.Error())
	}
}

//
// testHandleBin from websocket by echo-ing back the payload.
//
func testHandleBin(conn int, payload []byte) {
	packet := NewFrameBin(false, payload)

	err := Send(conn, packet)
	if err != nil {
		log.Println("handlePayloadBin: " + err.Error())
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
	ctx = context.WithValue(ctx, CtxKeyUID, uint64(_testUID))

	return
}

func runTestServer() {
	var err error

	_testWSAddr = "ws://127.0.0.1:" + strconv.Itoa(_testPort) + "/"

	_testServer, err = NewServer(_testPort)
	if err != nil {
		log.Println("runTestServer: " + err.Error())
		os.Exit(2)
	}

	_testServer.HandleAuth = testHandleAuth
	_testServer.HandleBin = testHandleBin
	_testServer.HandleText = testHandleText

	go _testServer.Start()
}

func TestMain(m *testing.M) {
	runTestServer()

	s := m.Run()

	os.Exit(s)
}
