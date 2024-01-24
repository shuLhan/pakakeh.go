// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	"strings"
)

// Route represent a parsed path.
// A path can have a key, or binding, that can be replaced with string
// value.
// For example, "/org/:user/:repo" have two keys "user" and "repo".
//
// Route handle the path in case-insensitive manner.
type Route struct {
	// path that has been cleaned up.
	path string

	// nodes contains parsed sub-path.
	nodes []*routeNode

	// nkey contains the number of keys in path.
	nkey int
}

// NewRoute create new Route from path.
// It will store the key(s) in path if available.
//
// The key is sub-path that start with colon ":".
// For example, the following path "/:user/:repo" contains two sub-paths
// with both are keys.
// If path is invalid, for example, "/:user/:" or "/:user/:user" (key with
// duplicate names), it will return nil with an error.
func NewRoute(rpath string) (rute *Route, err error) {
	rpath = strings.Trim(rpath, `/`)
	rpath = strings.ToLower(rpath)

	var (
		paths   = strings.Split(rpath, `/`)
		subpath string
	)

	rute = &Route{}

	for _, subpath = range paths {
		subpath = strings.TrimSpace(subpath)
		if len(subpath) == 0 {
			continue
		}

		var node = &routeNode{}

		if subpath[0] == ':' {
			node.key = strings.TrimSpace(subpath[1:])
			if len(node.key) == 0 {
				return nil, ErrPathKeyEmpty
			}

			if rute.isKeyExist(node.key) {
				return nil, ErrPathKeyDuplicate
			}

			node.isKey = true
			rute.nkey++
		} else {
			node.name = subpath
		}

		rute.nodes = append(rute.nodes, node)
	}
	if len(rute.nodes) == 0 {
		rute.nodes = append(rute.nodes, &routeNode{})
	}

	rute.path = rute.String()

	return rute, nil
}

// isKeyExist will return true if the key already exist in nodes; otherwise
// it will return false.
func (rute *Route) isKeyExist(key string) bool {
	var node *routeNode
	for _, node = range rute.nodes {
		if !node.isKey {
			continue
		}
		if node.key == key {
			return true
		}
	}
	return false
}

// Parse the path and return the key-value association and true if path is
// matched with current [Route]; otherwise it will return nil and false.
func (rute *Route) Parse(rpath string) (vals map[string]string, ok bool) {
	if rute.nkey == 0 {
		if rpath == rute.path {
			return nil, true
		}
	}

	rpath = strings.Trim(rpath, `/`)
	rpath = strings.ToLower(rpath)

	var paths = strings.Split(rpath, `/`)

	if len(paths) != len(rute.nodes) {
		return nil, false
	}

	var (
		x    int
		node *routeNode
	)

	vals = make(map[string]string, rute.nkey)

	for x, node = range rute.nodes {
		if node.isKey {
			vals[node.key] = paths[x]
		} else if paths[x] != node.name {
			return nil, false
		}
	}

	return vals, true
}

// String generate a clean path without any white spaces and single "/"
// between sub-path.
func (rute *Route) String() (path string) {
	var node *routeNode
	for _, node = range rute.nodes {
		path += `/`
		if node.isKey {
			path += `:` + node.key
		} else {
			path += node.name
		}
	}
	return path
}
