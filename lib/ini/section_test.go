// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewSection(t *testing.T) {
	cases := []struct {
		desc   string
		name   string
		sub    string
		expSec *section
	}{{
		desc: "With empty name",
	}, {
		desc: "With empty name but not subsection",
		sub:  "subsection",
	}, {
		desc: "With name only",
		name: "Section",
		expSec: &section{
			mode:      lineModeSection,
			name:      "Section",
			nameLower: "section",
		},
	}, {
		desc: "With name and subname",
		name: "Section",
		sub:  "Subsection",
		expSec: &section{
			mode:      lineModeSection | lineModeSubsection,
			name:      "Section",
			nameLower: "section",
			sub:       "Subsection",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := newSection(c.name, c.sub)

		test.Assert(t, "section", c.expSec, got, true)
	}
}

func TestSectionSet(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		v      string
		expOK  bool
		expSec *section
	}{{
		desc: "With empty key",
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
		},
	}, {
		desc:  "With empty value (Key-1) (will be added)",
		k:     "Key-1",
		expOK: true,
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "true",
			}},
		},
	}, {
		desc:  "With new value (Key-1)",
		k:     "Key-1",
		v:     "false",
		expOK: true,
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "false",
			}},
		},
	}, {
		desc:  "With key not found (Key-2) (added)",
		k:     "Key-2",
		v:     "2",
		expOK: true,
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "false",
			}, {
				mode:     lineModeValue,
				key:      "Key-2",
				keyLower: "key-2",
				value:    "2",
			}},
		},
	}, {
		desc:  "With empty value on Key-2 (true)",
		k:     "Key-2",
		expOK: true,
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "false",
			}, {
				mode:     lineModeValue,
				key:      "Key-2",
				keyLower: "key-2",
				value:    "true",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ok := sec.set(c.k, c.v)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func TestSectionAdd(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		v      string
		expSec *section
	}{{
		desc:   "Empty key (no change)",
		expSec: lastSec,
	}, {
		desc: "Duplicate key-1 (no value)",
		k:    "Key-1",
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "false",
			}, {
				mode:     lineModeValue,
				key:      "Key-2",
				keyLower: "key-2",
				value:    "true",
			}, {
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "true",
			}},
		},
	}, {
		desc: "Duplicate key-1 (1)",
		k:    "Key-1",
		v:    "1",
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "false",
			}, {
				mode:     lineModeValue,
				key:      "Key-2",
				keyLower: "key-2",
				value:    "true",
			}, {
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "true",
			}, {
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "1",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.add(c.k, c.v)

		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func TestSectionSet2(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		v      string
		expOK  bool
		expSec *section
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

		ok := sec.set(c.k, c.v)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func TestSectionUnset(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		expOK  bool
		expSec *section
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
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "false",
			}, {
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "true",
			}, {
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "1",
			}},
		},
	}, {
		desc:  "With valid key-2 (again)",
		k:     "key-2",
		expOK: true,
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "false",
			}, {
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "true",
			}, {
				mode:     lineModeValue,
				key:      "Key-1",
				keyLower: "key-1",
				value:    "1",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ok := sec.unset(c.k)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func TestSectionUnsetAll(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		expSec *section
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
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
		},
	}, {
		desc: "With valid key-1 (again)",
		k:    "KEY-1",
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.unsetAll(c.k)

		test.Assert(t, "section", c.expSec, sec, true)

		lastSec = c.expSec
	}
}

func TestSectionReplaceAll(t *testing.T) {
	sec.addVariable(nil)

	sec.add("key-3", "3")
	sec.add("key-3", "33")
	sec.add("key-3", "333")
	sec.add("key-3", "3333")

	cases := []struct {
		desc   string
		k      string
		v      string
		expSec *section
	}{{
		desc: "With empty key",
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3",
			}, {
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "33",
			}, {
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "333",
			}, {
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3333",
			}},
		},
	}, {
		desc: "With invalid key-4 (will be added)",
		k:    "KEY-4",
		v:    "4",
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3",
			}, {
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "33",
			}, {
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "333",
			}, {
				mode:     lineModeValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3333",
			}, {
				mode:     lineModeValue,
				key:      "KEY-4",
				keyLower: "key-4",
				value:    "4",
			}},
		},
	}, {
		desc: "With valid key-3",
		k:    "KEY-3",
		v:    "replaced",
		expSec: &section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeValue,
				key:      "KEY-4",
				keyLower: "key-4",
				value:    "4",
			}, {
				mode:     lineModeValue,
				key:      "KEY-3",
				keyLower: "key-3",
				value:    "replaced",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.replaceAll(c.k, c.v)

		test.Assert(t, "section", c.expSec, sec, true)
	}
}

func TestSectionGet(t *testing.T) {
	cases := []struct {
		desc   string
		k      string
		def    string
		expOK  bool
		expVal string
	}{{
		desc: "On empty vars",
		k:    "key-1",
	}, {
		desc:   "On empty vars with default",
		k:      "key-1",
		def:    "default value",
		expVal: "default value",
	}, {
		desc:   "Valid key",
		k:      "key-3",
		def:    "default value",
		expOK:  true,
		expVal: "replaced",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, ok := sec.get(c.k, c.def)

		test.Assert(t, "ok", c.expOK, ok, true)
		test.Assert(t, "value", c.expVal, got, true)
	}
}

func TestSectionGets(t *testing.T) {
	sec.add("dup", "value 1")
	sec.add("dup", "value 2")

	cases := []struct {
		desc  string
		key   string
		defs  []string
		exps  []string
		expOK bool
	}{{
		desc: "With empty key",
	}, {
		desc: "With no key found",
		key:  "noop",
		defs: []string{"default"},
		exps: []string{"default"},
	}, {
		desc:  "With key found",
		key:   "dup",
		defs:  []string{"default"},
		exps:  []string{"value 1", "value 2"},
		expOK: true,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, ok := sec.gets(c.key, c.defs)

		test.Assert(t, "Gets value", c.exps, got, true)
		test.Assert(t, "Gets ok", c.expOK, ok, true)
	}
}
