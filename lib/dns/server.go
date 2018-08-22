// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"io"
	"log"
	"net"
	"strconv"
)

//
// Server defines DNS server.
//
type Server struct {
	Handler Handler
	udp     *net.UDPConn
	tcp     *net.TCPListener
}

func parseAddress(address string) (*net.UDPAddr, *net.TCPAddr, error) {
	host, sport, err := net.SplitHostPort(address)
	if err != nil {
		return nil, nil, err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, nil, ErrInvalidAddress
	}

	port := DefaultPort

	if len(sport) >= 0 {
		port, err = strconv.Atoi(sport)
		if err != nil {
			return nil, nil, err
		}
	}

	udpAddr := &net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	tcpAddr := &net.TCPAddr{
		IP:   ip,
		Port: port,
	}

	return udpAddr, tcpAddr, nil
}

//
// ListenAndServe run DNS server, listening on UDP and TCP connection.
//
func (srv *Server) ListenAndServe(address string) error {
	var err error

	udpAddr, tcpAddr, err := parseAddress(address)
	if err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		err = srv.ListenAndServeTCP(tcpAddr)
		if err != nil {
			cherr <- err
		}
	}()
	go func() {
		err = srv.ListenAndServeUDP(udpAddr)
		if err != nil {
			cherr <- err
		}
	}()

	err = <-cherr

	return err
}

//
// ListenAndServeTCP listen for request with TCP socket.
//
func (srv *Server) ListenAndServeTCP(tcpAddr *net.TCPAddr) error {
	var err error

	srv.tcp, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	for {
		conn, err := srv.tcp.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		cl := &TCPClient{
			conn: conn,
		}

		go srv.serveTCPClient(cl)
	}
}

//
// ListenAndServeUDP listen for request with UDP socket.
//
func (srv *Server) ListenAndServeUDP(udpAddr *net.UDPAddr) error {
	var (
		n   int
		err error
	)

	srv.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	req := _requestPool.Get().(*Request)

	for {
		n, req.UDPAddr, err = srv.udp.ReadFromUDP(req.Message.Packet)
		if err != nil {
			log.Println(err)
			req.Reset()
			continue
		}

		req.Message.Packet = req.Message.Packet[:n]

		req.Message.UnpackHeaderQuestion()

		res := srv.Handler.ServeDNS(req)
		if res == nil {
			req.Reset()
			continue
		}

		_, err = srv.udp.WriteToUDP(res.Message.Packet, req.UDPAddr)
		if err != nil {
			log.Println("ListenAndServeUDP: WriteToUDP:", err)
		}

		req.Reset()
	}
}

func (srv *Server) serveTCPClient(cl *TCPClient) {
	var err error
	req := _requestPool.Get().(*Request)

	for {
		err = cl.Recv(req.Message)
		if err != nil {
			if err == io.EOF {
				_requestPool.Put(req)
				break
			}
			log.Println("serveTCPClient:", err)
			continue
		}

		req.Message.UnpackHeaderQuestion()

		res := srv.Handler.ServeDNS(req)
		if res == nil {
			req.Reset()
			continue
		}

		err = cl.Send(res.Message)
		if err != nil {
			log.Println("serveTCPClient: Send:", err)
		}

		req.Reset()
	}
}
