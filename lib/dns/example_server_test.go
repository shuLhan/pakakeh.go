package dns_test

import (
	"fmt"
	"log"
	"time"

	"github.com/shuLhan/share/lib/dns"
)

type serverHandler struct {
	responses []*dns.Message
}

func (h *serverHandler) generateResponses() {
	// kilabit.info A
	res := &dns.Message{
		Header: &dns.SectionHeader{
			ID:      1,
			QDCount: 1,
			ANCount: 1,
		},
		Question: &dns.SectionQuestion{
			Name:  []byte("kilabit.info"),
			Type:  dns.QueryTypeA,
			Class: dns.QueryClassIN,
		},
		Answer: []*dns.ResourceRecord{{
			Name:  []byte("kilabit.info"),
			Type:  dns.QueryTypeA,
			Class: dns.QueryClassIN,
			TTL:   3600,
			Text: &dns.RDataText{
				Value: []byte("127.0.0.1"),
			},
		}},
		Authority:  []*dns.ResourceRecord{},
		Additional: []*dns.ResourceRecord{},
	}

	_, err := res.Pack()
	if err != nil {
		log.Fatal("Pack: ", err)
	}

	h.responses = append(h.responses, res)

	// kilabit.info SOA
	res = &dns.Message{
		Header: &dns.SectionHeader{
			ID:      2,
			QDCount: 1,
			ANCount: 1,
		},
		Question: &dns.SectionQuestion{
			Name:  []byte("kilabit.info"),
			Type:  dns.QueryTypeSOA,
			Class: dns.QueryClassIN,
		},
		Answer: []*dns.ResourceRecord{{
			Name:  []byte("kilabit.info"),
			Type:  dns.QueryTypeSOA,
			Class: dns.QueryClassIN,
			TTL:   3600,
			SOA: &dns.RDataSOA{
				MName:   []byte("kilabit.info"),
				RName:   []byte("admin.kilabit.info"),
				Serial:  20180832,
				Refresh: 3600,
				Retry:   60,
				Expire:  3600,
				Minimum: 3600,
			},
		}},
		Authority:  []*dns.ResourceRecord{},
		Additional: []*dns.ResourceRecord{},
	}

	_, err = res.Pack()
	if err != nil {
		log.Fatal("Pack: ", err)
	}

	h.responses = append(h.responses, res)

	// kilabit.info TXT
	res = &dns.Message{
		Header: &dns.SectionHeader{
			ID:      3,
			QDCount: 1,
			ANCount: 1,
		},
		Question: &dns.SectionQuestion{
			Name:  []byte("kilabit.info"),
			Type:  dns.QueryTypeTXT,
			Class: dns.QueryClassIN,
		},
		Answer: []*dns.ResourceRecord{{
			Name:  []byte("kilabit.info"),
			Type:  dns.QueryTypeTXT,
			Class: dns.QueryClassIN,
			TTL:   3600,
			Text: &dns.RDataText{
				Value: []byte("This is a test server"),
			},
		}},
		Authority:  []*dns.ResourceRecord{},
		Additional: []*dns.ResourceRecord{},
	}

	_, err = res.Pack()
	if err != nil {
		log.Fatal("Pack: ", err)
	}

	h.responses = append(h.responses, res)
}

func (h *serverHandler) ServeDNS(req *dns.Request) {
	var (
		res *dns.Message
		err error
	)

	qname := string(req.Message.Question.Name)
	if qname == "kilabit.info" {
		switch req.Message.Question.Type {
		case dns.QueryTypeA:
			res = h.responses[0]
		case dns.QueryTypeSOA:
			res = h.responses[1]
		case dns.QueryTypeTXT:
			res = h.responses[2]
		}
	}

	// Return empty answer
	if res == nil {
		res := &dns.Message{
			Header: &dns.SectionHeader{
				ID:      req.Message.Header.ID,
				QDCount: 1,
			},
			Question: req.Message.Question,
		}

		_, err = res.Pack()
		if err != nil {
			return
		}
	} else {
		res.SetID(req.Message.Header.ID)
	}

	switch req.Kind {
	case dns.ConnTypeUDP:
		if req.Sender != nil {
			_, err = req.Sender.Send(res, req.UDPAddr)
			if err != nil {
				log.Println("! ServeDNS: Sender.Send: ", err)
			}
		}

	case dns.ConnTypeTCP:
		if req.Sender != nil {
			_, err = req.Sender.Send(res, nil)
			if err != nil {
				log.Println("! ServeDNS: Sender.Send: ", err)
			}
		}

	case dns.ConnTypeDoH:
		if req.ResponseWriter != nil {
			_, err = req.ResponseWriter.Write(res.Packet)
			if err != nil {
				log.Println("! ServeDNS: ResponseWriter.Write: ", err)
			}
			req.ChanResponded <- true
		}
	}
}

func clientLookup(nameserver string) {
	cl, err := dns.NewUDPClient(nameserver)
	if err != nil {
		log.Println(err)
		return
	}

	msg, err := cl.Lookup(dns.QueryTypeA, dns.QueryClassIN, []byte("kilabit.info"))
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Receiving DNS message: %s\n", msg)
	for x, answer := range msg.Answer {
		fmt.Printf("Answer %d: %s\n", x, answer.RData())
	}
	for x, auth := range msg.Authority {
		fmt.Printf("Authority %d: %s\n", x, auth.RData())
	}
	for x, add := range msg.Additional {
		fmt.Printf("Additional %d: %s\n", x, add.RData())
	}
}

func ExampleServer() {
	serverAddress := "127.0.0.1:5300"

	handler := &serverHandler{}

	handler.generateResponses()

	serverOptions := &dns.ServerOptions{
		IPAddress:        "127.0.0.1",
		TCPPort:          5300,
		UDPPort:          5300,
		DoHPort:          8443,
		CertFile:         "testdata/domain.crt",
		PrivateKeyFile:   "testdata/domain.key",
		DoHAllowInsecure: true,
	}

	server, err := dns.NewServer(serverOptions, handler)
	if err != nil {
		log.Fatal(err)
	}

	server.Start()

	// Wait for all listeners running.
	time.Sleep(500 * time.Millisecond)

	clientLookup(serverAddress)

	server.Stop()
}
