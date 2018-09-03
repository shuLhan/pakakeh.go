// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"log"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const (
	testServerAddress = "127.0.0.1:5353"
)

var (
	_testServer  *Server
	_testHandler *serverHandler
)

type serverHandler struct {
	responses []*Response
}

func (h *serverHandler) generateResponses() {
	res := &Response{
		Message: &Message{
			Header: &SectionHeader{
				ID:      1,
				QDCount: 1,
				ANCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				rdlen: 4,
				Text: &RDataText{
					v: []byte("127.0.0.1"),
				},
			}},
			Authority:  []*ResourceRecord{},
			Additional: []*ResourceRecord{},
		},
	}

	_, err := res.Message.Pack()
	if err != nil {
		log.Fatal("Pack: ", err)
	}

	h.responses = append(h.responses, res)

	// kilabit.info SOA
	res = &Response{
		Message: &Message{
			Header: &SectionHeader{
				ID:      2,
				QDCount: 1,
				ANCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				SOA: &RDataSOA{
					MName:   []byte("kilabit.info"),
					RName:   []byte("admin.kilabit.info"),
					Serial:  20180832,
					Refresh: 3600,
					Retry:   60,
					Expire:  3600,
					Minimum: 3600,
				},
			}},
			Authority:  []*ResourceRecord{},
			Additional: []*ResourceRecord{},
		},
	}

	_, err = res.Message.Pack()
	if err != nil {
		log.Fatal("Pack: ", err)
	}

	h.responses = append(h.responses, res)

	// kilabit.info TXT
	res = &Response{
		Message: &Message{
			Header: &SectionHeader{
				ID:      3,
				QDCount: 1,
				ANCount: 1,
			},
			Question: &SectionQuestion{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("kilabit.info"),
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("This is a test server"),
				},
			}},
			Authority:  []*ResourceRecord{},
			Additional: []*ResourceRecord{},
		},
	}

	_, err = res.Message.Pack()
	if err != nil {
		log.Fatal("Pack: ", err)
	}

	h.responses = append(h.responses, res)
}

func (h *serverHandler) ServeDNS(req *Request) {
	var ref *Response

	qname := string(req.Message.Question.Name)
	switch qname {
	case "kilabit.info":
		switch req.Message.Question.Type {
		case QueryTypeA:
			ref = h.responses[0]
		case QueryTypeSOA:
			ref = h.responses[1]
		case QueryTypeTXT:
			ref = h.responses[2]
		}
	}
	if ref == nil {
		_testServer.FreeRequest(req)
		return
	}

	res := &Response{
		Message: &Message{
			Header:   ref.Message.Header,
			Question: ref.Message.Question,
			Answer:   ref.Message.Answer,
		},
	}

	res.Message.Header.ID = req.Message.Header.ID

	_, err := res.Message.Pack()
	if err != nil {
		_testServer.FreeRequest(req)
		return
	}

	_, err = req.Sender.Send(res.Message, req.UDPAddr)
	if err != nil {
		log.Println("ServeDNS: ", err)
	}

	_testServer.FreeRequest(req)
}

func TestMain(m *testing.M) {
	debugLevel = 2
	log.SetFlags(log.Lmicroseconds)

	_testHandler = &serverHandler{}

	_testHandler.generateResponses()

	_testServer = &Server{
		Handler: _testHandler,
	}

	go func() {
		err := _testServer.ListenAndServe(testServerAddress)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	os.Exit(m.Run())
}

func TestQueryType(t *testing.T) {
	test.Assert(t, "QueryTypeA", QueryTypeA, uint16(1), true)
	test.Assert(t, "QueryTypeTXT", QueryTypeTXT, uint16(16), true)
	test.Assert(t, "QueryTypeAXFR", QueryTypeAXFR, uint16(252), true)
	test.Assert(t, "QueryTypeALL", QueryTypeALL, uint16(255), true)
}
