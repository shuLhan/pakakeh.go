package net

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsHostnameValid(t *testing.T) {
	cases := []struct {
		in  []byte
		exp bool
	}{{
		in: []byte(""),
	}, {
		in: []byte("-1a"),
	}, {
		in: []byte(".1a."),
	}, {
		in: []byte("1a-"),
	}, {
		in:  []byte("a"),
		exp: true,
	}, {
		in:  []byte("_a"),
		exp: true,
	}, {
		in:  []byte("11"),
		exp: true,
	}, {
		in:  []byte("a1"),
		exp: true,
	}, {
		in:  []byte("a-1"),
		exp: true,
	}, {
		in:  []byte("a.1"),
		exp: true,
	}}

	for _, c := range cases {
		t.Log("input: ", c.in)

		got := IsHostnameValid(c.in)

		test.Assert(t, "IsHostnameValid", c.exp, got, true)
	}
}
