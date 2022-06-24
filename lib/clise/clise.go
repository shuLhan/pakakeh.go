// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package clise implements circular slice.
// A circular slice is a slice that have fixed size.
// An append to slice that has reached its length will overwrite and start
// again from index 0.
//
// For example, a clise with size 5,
//
//	var c *clise.Clise = clise.New(5)
//	c.Push(1, 2, 3, 4, 5)
//	fmt.Println(c.Slice()) // [1 2 3 4 5]
//
// If we push another item, it will overwrite the first index,
//
//	c.Push(6)
//	fmt.Println(c.Slice()) // [6 2 3 4 5]
//
// See the examples for usage of the package.
package clise

import (
	"encoding/json"
	"sync"
)

type Clise struct {
	v    []interface{}
	size int
	last int
	sync.Mutex
	over bool
}

// New create and initialize circular slice with fixed size.
// It will return nil if size <= 0.
func New(size int) (c *Clise) {
	if size <= 0 {
		return nil
	}
	c = &Clise{
		v:    make([]interface{}, size),
		size: size,
	}
	return c
}

// Pop remove the last Push()-ed item and return it to caller.
// It will return nil if no more item inside it.
func (c *Clise) Pop() (item interface{}) {
	c.Lock()
	if c.over {
		if c.last == 0 {
			c.last = c.size - 1
		} else {
			c.last--
		}
	} else {
		if c.last == 0 {
			c.Unlock()
			return nil
		}
		c.last--
	}
	item = c.v[c.last]
	c.v[c.last] = nil
	c.Unlock()
	return item
}

// Push the item into the slice.
func (c *Clise) Push(src ...interface{}) {
	var x int
	c.Lock()
	for ; x < len(src); x++ {
		c.v[c.last] = src[x]
		c.last++
		if c.last == c.size {
			c.last = 0
			c.over = true
		}
	}
	c.Unlock()
}

// RecentSlice return the slice from index zero until the recent item.
func (c *Clise) RecentSlice() (dst []interface{}) {
	c.Lock()
	dst = make([]interface{}, c.last)
	copy(dst, c.v[:c.last])
	c.Unlock()
	return dst
}

// Reset the slice, start from zero.
func (c *Clise) Reset() {
	c.Lock()
	c.last = 0
	c.over = false
	c.Unlock()
}

// Slice return the content of circular slice as slice in the order of the
// last item to the recent item.
func (c *Clise) Slice() (dst []interface{}) {
	var (
		start int
		end   int = c.size
	)
	if c.over {
		dst = make([]interface{}, c.size)
		start = c.last
	} else {
		dst = make([]interface{}, c.last)
		end = c.last
	}
	c.Lock()
	copy(dst, c.v[start:end])
	if c.over {
		copy(dst[end-start:], c.v[0:start])
	}
	c.Unlock()
	return dst
}

// MarshalJSON call Slice on c and convert it into JSON.
func (c *Clise) MarshalJSON() (out []byte, err error) {
	var slice = c.Slice()
	out, err = json.Marshal(slice)
	return out, err
}
