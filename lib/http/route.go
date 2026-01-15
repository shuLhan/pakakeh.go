// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package http

import (
	libpath "git.sr.ht/~shulhan/pakakeh.go/lib/path"
)

// List of kind for route.
const (
	routeKindHTTP int = iota // Normal routing.
	routeKindSSE             // Routing for Server-Sent Events (SSE).
)

// route represent the route to endpoint.
type route struct {
	*libpath.Route

	endpoint    *Endpoint    // endpoint of route.
	endpointSSE *SSEEndpoint // Endpoint for SSE.

	kind int
}

// newRoute parse the Endpoint's path, store the key(s) in path if available
// in nodes.
//
// The key is sub-path that start with colon ":".
// For example, the following path "/:user/:repo" contains two nodes with both
// are keys.
// If path is invalid, for example, "/:user/:" or "/:user/:user" (key with
// duplicate names), it will return nil.
func newRoute(ep *Endpoint) (rute *route, err error) {
	rute = &route{
		endpoint: ep,
	}
	if ep.ErrorHandler == nil {
		ep.ErrorHandler = DefaultErrorHandler
	}
	rute.Route, err = libpath.NewRoute(ep.Path)
	if err != nil {
		return nil, err
	}
	return rute, nil
}

// newRouteSSE create and initialize new route for SSE.
func newRouteSSE(ep *SSEEndpoint) (rute *route, err error) {
	rute = &route{
		endpointSSE: ep,
		kind:        routeKindSSE,
	}
	rute.Route, err = libpath.NewRoute(ep.Path)
	if err != nil {
		return nil, err
	}
	return rute, nil
}
