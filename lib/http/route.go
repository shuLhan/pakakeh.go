// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"strings"
)

//
// route represent the route to endpoint.
//
type route struct {
	path     string    // path contains Endpoint's path that has been cleaned up.
	nodes    []*node   // nodes contains sub-path.
	nkey     int       // nkey contains the number of keys in nodes.
	endpoint *Endpoint // endpoint of route.
}

//
// newRoute parse the Endpoint's path, store the key(s) in path if available
// in nodes.
//
// The key is sub-path that start with colon ":".
// For example, the following path "/:user/:repo" contains two nodes with both
// are keys.
// If path is invalid, for example, "/:user/:" or "/:user/:user" (key with
// duplicate names), it will return nil.
//
func newRoute(ep *Endpoint) (rute *route, err error) {
	rute = &route{
		endpoint: ep,
	}
	if ep.ErrorHandler == nil {
		ep.ErrorHandler = DefaultErrorHandler
	}

	paths := strings.Split(strings.ToLower(strings.Trim(ep.Path, "/")), "/")

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if len(path) == 0 {
			continue
		}

		nod := &node{}

		if path[0] == ':' {
			nod.key = strings.TrimSpace(path[1:])
			if len(nod.key) == 0 {
				return nil, ErrEndpointKeyEmpty
			}

			if rute.isKeyExist(nod.key) {
				return nil, ErrEndpointKeyDuplicate
			}

			nod.isKey = true
			rute.nkey++
		} else {
			nod.name = path
		}

		rute.nodes = append(rute.nodes, nod)
	}
	if len(rute.nodes) == 0 {
		rute.nodes = append(rute.nodes, &node{})
	}

	rute.path = rute.generatePath()

	return rute, nil
}

//
// isKeyExist will return true if the key already exist in nodes; otherwise it
// will return false.
//
func (rute *route) isKeyExist(key string) bool {
	for _, node := range rute.nodes {
		if !node.isKey {
			continue
		}
		if node.key == key {
			return true
		}
	}
	return false
}

//
// parse the path and return the key-value association and true if path is
// matched with current route; otherwise it will return nil and false.
//
func (rute *route) parse(path string) (vals map[string]string, ok bool) {
	if rute.nkey == 0 {
		if path == rute.path {
			return nil, true
		}
	}

	paths := strings.Split(strings.ToLower(strings.Trim(path, "/")), "/")

	if len(paths) != len(rute.nodes) {
		return nil, false
	}

	vals = make(map[string]string, rute.nkey)
	for x, node := range rute.nodes {
		if node.isKey {
			vals[node.key] = paths[x]
		} else if paths[x] != node.name {
			return nil, false
		}
	}

	return vals, true
}

//
// generatePath generate a clean path without any white spaces and single "/"
// between sub-path.
//
func (rute *route) generatePath() (path string) {
	for _, node := range rute.nodes {
		path += "/"
		if node.isKey {
			path += ":" + node.key
		} else {
			path += node.name
		}
	}
	return path
}
