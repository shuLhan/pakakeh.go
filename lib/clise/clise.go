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
	"fmt"
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

// Close implement io.Closer, equal to Reset().
func (c *Clise) Close() error {
	c.Reset()
	return nil
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

	c.Lock()
	defer c.Unlock()

	if c.over {
		dst = make([]interface{}, c.size)
		start = c.last
	} else {
		dst = make([]interface{}, c.last)
		end = c.last
	}

	copy(dst, c.v[start:end])
	if c.over {
		copy(dst[end-start:], c.v[0:start])
	}

	return dst
}

// MarshalJSON call Slice on c and convert it into JSON.
func (c *Clise) MarshalJSON() (out []byte, err error) {
	var slice = c.Slice()
	out, err = json.Marshal(slice)
	return out, err
}

// Write implement io.Writer, equal to Push(b).
func (c *Clise) Write(b []byte) (n int, err error) {
	c.Push(b)
	return len(b), nil
}

// WriteByte implement io.ByteWriter, equal to Push(b).
func (c *Clise) WriteByte(b byte) error {
	c.Push(b)
	return nil
}

// WriteString implement io.StringWriter, equal to Push(s).
func (c *Clise) WriteString(s string) (n int, err error) {
	c.Push(s)
	return len(s), nil
}

// UnmarshalJSON unmarshal JSON array into Clise.
// If the size is zero, it will be set to the length of JSON array.
func (c *Clise) UnmarshalJSON(jsonb []byte) (err error) {
	var (
		logp  = `UnmarshalJSON`
		array = make([]interface{}, 0)
	)

	err = json.Unmarshal(jsonb, &array)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	if c.size == 0 {
		c.size = len(array)
		c.v = make([]interface{}, c.size)
	}

	c.Push(array...)

	return nil
}
