// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

import (
	"bytes"
	"context"
	"fmt"
)

// RouteHandler is a function that will be called when registered method and
// target match with request.
type RouteHandler func(ctx context.Context, req *Request) (res Response)

type route struct {
	handler RouteHandler
	name    string
	childs  []*route
	isParam bool
}

// addChild to route.
func (r *route) addChild(isParam bool, name string) (c *route, err error) {
	var logp = `addChild`

	c = r.getChild(isParam, name)
	if c != nil {
		return c, nil
	}
	if isParam {
		c = r.getChildAsParam()
		if c != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, ErrRouteDupParam)
		}
	}
	c = &route{
		name:    name,
		isParam: isParam,
	}
	r.childs = append(r.childs, c)

	return c, nil
}

// getChild of current route which has the same isParam and name value.  It
// will return nil if not found.
func (r *route) getChild(isParam bool, name string) (c *route) {
	for _, c = range r.childs {
		if isParam == c.isParam && name == c.name {
			return c
		}
	}
	return nil
}

// getChildAsParam return child route which type is parameter.
func (r *route) getChildAsParam() (c *route) {
	for _, c = range r.childs {
		if c.isParam {
			return c
		}
	}
	return nil
}

// String return route representation as string.  This function is to prevent
// formatted print to print pointer to route as address.
func (r *route) String() (out string) {
	var (
		bb = _bbPool.Get().(*bytes.Buffer)

		c *route
		x int
	)
	bb.Reset()

	fmt.Fprintf(bb, "{name:%s", r.name)
	fmt.Fprintf(bb, " isParam:%v", r.isParam)
	fmt.Fprintf(bb, " handler:%v", r.handler)
	bb.WriteString(" childs:[")

	for x, c = range r.childs {
		if x > 0 {
			bb.WriteByte(' ')
		}
		fmt.Fprintf(bb, "%s", c.String())
	}

	bb.WriteString("]}")

	out = bb.String()

	_bbPool.Put(bb)

	return
}
