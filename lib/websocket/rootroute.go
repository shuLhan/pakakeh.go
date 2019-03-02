// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
)

const (
	pathQuerySep    = "?"
	pathSep         = '/'
	pathParamPrefix = ':'
)

// List of route error values.
var (
	ErrRouteInvMethod = errors.New("invalid method")
	ErrRouteInvTarget = errors.New("invalid target")
	ErrRouteDupParam  = errors.New("duplicate parameter on route")
)

type rootRoute struct {
	methodDelete *route
	methodGet    *route
	methodPatch  *route
	methodPost   *route
	methodPut    *route
}

//
// newRootRoute create and initialize each route's method with path "/" and
// nil handler.
//
func newRootRoute() (root *rootRoute) {
	root = &rootRoute{
		methodDelete: &route{
			name:    "/",
			handler: nil,
			isParam: false,
		},
		methodGet: &route{
			name:    "/",
			handler: nil,
			isParam: false,
		},
		methodPatch: &route{
			name:    "/",
			handler: nil,
			isParam: false,
		},
		methodPost: &route{
			name:    "/",
			handler: nil,
			isParam: false,
		},
		methodPut: &route{
			name:    "/",
			handler: nil,
			isParam: false,
		},
	}

	return
}

func (root *rootRoute) getParent(method string) *route {
	switch method {
	case http.MethodDelete:
		return root.methodDelete
	case http.MethodGet:
		return root.methodGet
	case http.MethodPatch:
		return root.methodPatch
	case http.MethodPost:
		return root.methodPost
	case http.MethodPut:
		return root.methodPut
	}
	return nil
}

//
// add new route handler by method and target.
//
// method is one of HTTP method that is allowed: DELETE, GET, PATCH, POST, or
// PUT.
// target is absolute path, MUST start with slash "/", and can contains
// parameter by prefixing it with colon ":".  For example, "/book/:id", will
// be parsed into,
//
//	{
//		name:"book",
//		childs: []*route{{
//			name:"id",
//			isParam:true,
//		}}
//	}`
//
func (root *rootRoute) add(method, target string, handler RouteHandler) (err error) {
	if target[0] != pathSep {
		return ErrRouteInvTarget
	}

	method = strings.ToUpper(method)

	parent := root.getParent(method)
	if parent == nil {
		return ErrRouteInvMethod
	}

	bb := _bbPool.Get().(*bytes.Buffer)
	bb.Reset()

	started := true
	isParam := false

	for x := 1; x < len(target); x++ {
		if started && target[x] == pathParamPrefix {
			isParam = true
			started = false
			continue
		}
		if target[x] != pathSep {
			_ = bb.WriteByte(target[x])
			continue
		}
		if bb.Len() == 0 {
			started = true
			isParam = false
			continue
		}

		parent, err = parent.addChild(isParam, bb.String())
		if err != nil {
			goto out
		}

		bb.Reset()
		started = true
		isParam = false
	}

	if bb.Len() > 0 {
		parent, err = parent.addChild(isParam, bb.String())
		if err != nil {
			goto out
		}
	}

	parent.handler = handler
out:
	_bbPool.Put(bb)

	return err
}

//
// get the route parameters values and their handler.
//
func (root *rootRoute) get(method, target string) (
	params targetParam, handler RouteHandler,
) {
	if len(method) == 0 || len(target) == 0 {
		return nil, nil
	}
	if target[0] != pathSep {
		return nil, nil
	}

	method = strings.ToUpper(method)

	parent := root.getParent(method)
	if parent == nil {
		return nil, nil
	}

	bb := _bbPool.Get().(*bytes.Buffer)
	bb.Reset()

	params = make(targetParam)

	for x := 1; x < len(target); x++ {
		if target[x] != pathSep {
			_ = bb.WriteByte(target[x])
			continue
		}

		child := parent.getChild(false, bb.String())
		if child == nil {
			child = parent.getChildAsParam()
			if child == nil {
				params = nil
				goto out
			}

			params[child.name] = bb.String()
		}
		parent = child
		bb.Reset()
	}
	if bb.Len() > 0 {
		child := parent.getChild(false, bb.String())
		if child == nil {
			child = parent.getChildAsParam()
			if child == nil {
				params = nil
				goto out
			}

			params[child.name] = bb.String()
		}
		parent = child
	}

	handler = parent.handler
out:
	_bbPool.Put(bb)

	return params, handler
}
