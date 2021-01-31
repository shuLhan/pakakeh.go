//
// Package clise implements circular slice.
// A circular slice is a slice that have fixed size.
// An append to slice that has reached its length will overwrite and start
// again from index 0.
//
// For example, a clise with size 5,
//
//	c := clise.New(5)
//	c.Push(1, 2, 3, 4, 5)
//	fmt.Printf("%v\n", c.Slice()) // [1 2 3 4 5]
//
// If we push another item, it will overwrite the first index,
//
//	c.Push(6)
//	fmt.Printf("%v\n", c.Slice()) // [6 2 3 4 5]
//
// See the examples for usage of the package.
//
package clise

type Clise struct {
	v    []interface{}
	size int
	last int
	over bool
}

//
// New create and initialize circular slice with fixed size.
// It will return nil if size <= 0.
//
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

//
// Push the item into the slice.
//
func (c *Clise) Push(src ...interface{}) {
	for x := 0; x < len(src); x++ {
		c.v[c.last] = src[x]
		c.last++
		if c.last == c.size {
			c.last = 0
			c.over = true
		}
	}
}

//
// RecentSlice return the slice from index zero until the recent item.
//
func (c *Clise) RecentSlice() (dst []interface{}) {
	dst = make([]interface{}, c.last)
	copy(dst, c.v[:c.last])
	return dst
}

//
// Reset the slice, start from zero.
//
func (c *Clise) Reset() {
	c.last = 0
	c.over = false
}

//
// Slice return the content of circular slice as slice in the order of the
// last item to the recent item.
//
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
	copy(dst, c.v[start:end])
	if c.over {
		copy(dst[end-start:], c.v[0:start])
	}
	return dst
}
