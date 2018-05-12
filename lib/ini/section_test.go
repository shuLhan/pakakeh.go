// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	sec     *Section
	lastSec *Section
)

func testNewSection(t *testing.T) {
	cases := []struct {
		desc   string
		name   string
		sub    string
		expSec *Section
	}{{
		desc: "With empty name",
	}, {
		desc: "With empty name but not subsection",
		sub:  "subsection",
	}, {
		desc: "With name only",
		name: "Section",
		expSec: &Section{
			mode:  varModeSection,
			name:  []byte("Section"),
			_name: []byte("section"),
		},
	}, {
		desc: "With name and subname",
		name: "Section",
		sub:  "Subsection",
		expSec: &Section{
			mode:  varModeSection | varModeSubsection,
			name:  []byte("Section"),
			_name: []byte("section"),
			Sub:   []byte("Subsection"),
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := NewSection(c.name, c.sub)

		test.Assert(t, "section", c.expSec, got, true)
	}
}

func testSectionGet(t *testing.T) {
	cases := []struct {
		desc   string
		k      []byte
		expOK  bool
		expVal []byte
	}{{
		desc: "On empty vars",
		k:    []byte("key-1"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, ok := sec.Get(c.k)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "value", c.expVal, got, true)
	}
}

func testSectionSet(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		v      string
		expOK  bool
		expSec *Section
	}{{
		desc: "With empty key",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
		},
	}, {
		desc:  "With empty value (Key-1) (will be added)",
		k:     "Key-1",
		expOK: true,
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("true"),
			}},
		},
	}, {
		desc:  "With new value (Key-1)",
		k:     "Key-1",
		v:     "false",
		expOK: true,
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("false"),
			}},
		},
	}, {
		desc:  "With key not found (Key-2) (added)",
		k:     "Key-2",
		v:     "2",
		expOK: true,
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("false"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-2"),
				_key:  []byte("key-2"),
				value: []byte("2"),
			}},
		},
	}, {
		desc:  "With empty value on Key-2 (true)",
		k:     "Key-2",
		expOK: true,
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("false"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-2"),
				_key:  []byte("key-2"),
				value: []byte("true"),
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ok := sec.Set(c.k, c.v)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func testSectionAdd(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		v      string
		expSec *Section
	}{{
		desc:   "Empty key (no change)",
		expSec: lastSec,
	}, {
		desc: "Duplicate key-1 (no value)",
		k:    "Key-1",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("false"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-2"),
				_key:  []byte("key-2"),
				value: []byte("true"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("true"),
			}},
		},
	}, {
		desc: "Duplicate key-1 (1)",
		k:    "Key-1",
		v:    "1",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("false"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-2"),
				_key:  []byte("key-2"),
				value: []byte("true"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("true"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("1"),
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.Add(c.k, c.v)

		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func testSectionSet2(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		v      string
		expOK  bool
		expSec *Section
	}{{
		desc:   "Set duplicate Key-1",
		k:      "Key-1",
		v:      "new value",
		expSec: lastSec,
	}, {
		desc:   "Set duplicate key-1",
		k:      "key-1",
		v:      "new value",
		expSec: lastSec,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ok := sec.Set(c.k, c.v)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func testSectionUnset(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		expOK  bool
		expSec *Section
	}{{
		desc:   "With empty key",
		expOK:  true,
		expSec: lastSec,
	}, {
		desc:   "With duplicate key-1",
		k:      "key-1",
		expSec: lastSec,
	}, {
		desc:  "With valid key-2",
		k:     "key-2",
		expOK: true,
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("false"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("true"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("1"),
			}},
		},
	}, {
		desc:  "With valid key-2 (again)",
		k:     "key-2",
		expOK: true,
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("false"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("true"),
			}, {
				mode:  varModeValue,
				key:   []byte("Key-1"),
				_key:  []byte("key-1"),
				value: []byte("1"),
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ok := sec.Unset(c.k)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func testSectionUnsetAll(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		expSec *Section
	}{{
		desc:   "With empty key",
		expSec: lastSec,
	}, {
		desc:   "With invalid key-3",
		k:      "key-3",
		expSec: lastSec,
	}, {
		desc: "With valid key-1",
		k:    "KEY-1",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
		},
	}, {
		desc: "With valid key-1 (again)",
		k:    "KEY-1",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.UnsetAll(c.k)

		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func testSectionReplaceAll(t *testing.T) {
	sec.add(nil)

	sec.Add("key-3", "3")
	sec.Add("key-3", "33")
	sec.Add("key-3", "333")
	sec.Add("key-3", "3333")

	cases := []struct {
		desc   string
		k      string
		v      string
		expSec *Section
	}{{
		desc: "With empty key",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("3"),
			}, {
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("33"),
			}, {
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("333"),
			}, {
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("3333"),
			}},
		},
	}, {
		desc: "With invalid key-4 (will be added)",
		k:    "KEY-4",
		v:    "4",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("3"),
			}, {
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("33"),
			}, {
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("333"),
			}, {
				mode:  varModeValue,
				key:   []byte("key-3"),
				_key:  []byte("key-3"),
				value: []byte("3333"),
			}, {
				mode:  varModeValue,
				key:   []byte("KEY-4"),
				_key:  []byte("key-4"),
				value: []byte("4"),
			}},
		},
	}, {
		desc: "With valid key-3",
		k:    "KEY-3",
		v:    "replaced",
		expSec: &Section{
			mode:  sec.mode,
			name:  sec.name,
			_name: sec._name,
			Vars: []*Variable{{
				mode:  varModeValue,
				key:   []byte("KEY-4"),
				_key:  []byte("key-4"),
				value: []byte("4"),
			}, {
				mode:  varModeValue,
				key:   []byte("KEY-3"),
				_key:  []byte("key-3"),
				value: []byte("replaced"),
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.ReplaceAll(c.k, c.v)

		test.Assert(t, "section", c.expSec, sec, true)
	}
}

func TestSection(t *testing.T) {
	sec = NewSection("test", "")

	t.Run("New", testNewSection)
	t.Run("Get", testSectionGet)
	t.Run("Set", testSectionSet)
	t.Run("Add", testSectionAdd)
	t.Run("Set2", testSectionSet2)
	t.Run("Unset", testSectionUnset)
	t.Run("UnsetAll", testSectionUnsetAll)
	t.Run("ReplaceAll", testSectionReplaceAll)
}
