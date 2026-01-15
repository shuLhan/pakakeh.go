// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package http

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	liberrors "git.sr.ht/~shulhan/pakakeh.go/lib/errors"
)

const defKeepAliveInterval = 5 * time.Second

// SSEEndpoint endpoint to create Server-Sent Events (SSE) on server.
//
// For creating the SSE client see subpackage [sseclient].
type SSEEndpoint struct {
	// Call handler that will called when request to Path accepted.
	Call SSECallback

	// Path where server accept the request for SSE.
	Path string

	// KeepAliveInterval define the interval where server will send an
	// empty message to active connection periodically.
	// This field is optional, default and minimum value is 5 seconds.
	KeepAliveInterval time.Duration
}

func (ep *SSEEndpoint) call(
	res http.ResponseWriter,
	req *http.Request,
	evaluators []Evaluator,
	vals map[string]string,
) {
	var err error

	err = req.ParseForm()
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// Fill the form with path binding.
	if len(vals) > 0 {
		if req.Form == nil {
			req.Form = make(url.Values, len(vals))
		}
		var k, v string
		for k, v = range vals {
			if len(k) > 0 && len(v) > 0 {
				req.Form.Set(k, v)
			}
		}
	}

	err = ep.doEvals(res, req, evaluators)
	if err != nil {
		return
	}

	var sseconn *SSEConn

	sseconn, err = ep.hijack(res, req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	sseconn.handshake()
	if ep.KeepAliveInterval < defKeepAliveInterval {
		ep.KeepAliveInterval = defKeepAliveInterval
	}
	go sseconn.workerKeepAlive(ep.KeepAliveInterval)
	ep.Call(sseconn)
	sseconn.conn.Close()
}

func (ep *SSEEndpoint) doEvals(
	res http.ResponseWriter,
	req *http.Request,
	evaluators []Evaluator,
) (err error) {
	var eval Evaluator

	for _, eval = range evaluators {
		err = eval(req, nil)
		if err != nil {
			var errInternal = &liberrors.E{}
			if !errors.As(err, &errInternal) {
				errInternal.Code = http.StatusUnprocessableEntity
			}
			http.Error(res, err.Error(), errInternal.Code)
			return err
		}
	}
	return nil
}

func (ep *SSEEndpoint) hijack(res http.ResponseWriter, req *http.Request) (sseconn *SSEConn, err error) {
	var (
		hijack http.Hijacker
		ok     bool
	)

	hijack, ok = res.(http.Hijacker)
	if !ok {
		return nil, errors.New(`http.ResponseWriter is not http.Hijacker`)
	}

	sseconn = &SSEConn{
		HTTPRequest: req,
	}

	sseconn.conn, sseconn.bufrw, err = hijack.Hijack()
	if err != nil {
		return nil, err
	}

	return sseconn, nil
}
