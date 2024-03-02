// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewSection(t *testing.T) {
	cases := []struct {
		expSec *Section

		desc string
		name string
		sub  string
	}{{
		desc: "With empty name",
	}, {
		desc: "With empty name but not subsection",
		sub:  "subsection",
	}, {
		desc: "With name only",
		name: "Section",
		expSec: &Section{
			mode:      lineModeSection,
			name:      "Section",
			nameLower: "section",
		},
	}, {
		desc: "With name and subname",
		name: "Section",
		sub:  "Subsection",
		expSec: &Section{
			mode:      lineModeSection | lineModeSubsection,
			name:      "Section",
			nameLower: "section",
			sub:       "Subsection",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := newSection(c.name, c.sub)

		test.Assert(t, "section", c.expSec, got)
	}
}

func TestSectionSet(t *testing.T) {
	sec := &Section{
		mode:      lineModeSection,
		name:      "section",
		nameLower: "section",
		vars: []*variable{{
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v1",
		}, {
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v2",
		}},
	}

	cases := []struct {
		expSec *Section

		desc string
		k    string
		v    string

		expOK bool
	}{{
		desc:  "With empty value",
		k:     "k",
		expOK: true,
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
			}},
		},
	}, {
		desc:  "With value",
		k:     "k",
		v:     "false",
		expOK: true,
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "false",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ok := sec.set(c.k, c.v)

		test.Assert(t, "ok", c.expOK, ok)
		test.Assert(t, "section", c.expSec, sec)
	}
}

func TestSection_add(t *testing.T) {
	sec := &Section{
		mode:      lineModeSection,
		name:      "section",
		nameLower: "section",
		vars: []*variable{{
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v1",
		}, {
			mode:     lineModeKeyValue,
			key:      "k2",
			keyLower: "k2",
			value:    "v2",
		}},
	}

	cases := []struct {
		expSec *Section

		desc string
		k    string
		v    string
	}{{
		desc: "With empty key",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k2",
				keyLower: "k2",
				value:    "v2",
			}},
		},
	}, {
		desc: "With no value",
		k:    "k",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
			}, {
				mode:     lineModeKeyValue,
				key:      "k2",
				keyLower: "k2",
				value:    "v2",
			}},
		},
	}, {
		desc: "Duplicate key and value",
		k:    "k",
		v:    "v1",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
			}, {
				mode:     lineModeKeyValue,
				key:      "k2",
				keyLower: "k2",
				value:    "v2",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.add(c.k, c.v)

		test.Assert(t, "section", c.expSec, sec)
	}
}

func TestSectionUnset(t *testing.T) {
	sec := &Section{
		mode:      lineModeSection,
		name:      "section",
		nameLower: "section",
		vars: []*variable{{
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v1",
		}, {
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v2",
		}},
	}

	cases := []struct {
		expSec *Section

		desc  string
		k     string
		expOK bool
	}{{
		desc:  "With empty key",
		expOK: false,
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v2",
			}},
		},
	}, {
		desc:  "With duplicate key",
		k:     "k",
		expOK: true,
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}},
		},
	}, {
		desc: "With invalid key",
		k:    "key-2",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}},
		},
	}, {
		desc:  "With valid key (again)",
		k:     "k",
		expOK: true,
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars:      []*variable{},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ok := sec.unset(c.k)

		test.Assert(t, "ok", c.expOK, ok)
		test.Assert(t, "section", c.expSec, sec)
	}
}

func TestSectionUnsetAll(t *testing.T) {
	sec := &Section{
		mode:      lineModeSection,
		name:      "section",
		nameLower: "section",
		vars: []*variable{{
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v1",
		}, {
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v2",
		}},
	}

	cases := []struct {
		expSec *Section

		desc string
		k    string
	}{{
		desc: "With empty key",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v2",
			}},
		},
	}, {
		desc: "With unmatch key",
		k:    "unmatch",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v1",
			}, {
				mode:     lineModeKeyValue,
				key:      "k",
				keyLower: "k",
				value:    "v2",
			}},
		},
	}, {
		desc: "With valid k",
		k:    "K",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
		},
	}, {
		desc: "With valid key (again)",
		k:    "K",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.unsetAll(c.k)

		test.Assert(t, "section", c.expSec, sec)
	}
}

func TestSection_replaceAll(t *testing.T) {
	sec := &Section{
		mode:      lineModeSection,
		name:      "section",
		nameLower: "section",
	}

	sec.add("key-3", "3")
	sec.add("key-3", "33")
	sec.add("key-3", "333")
	sec.add("key-3", "3333")

	cases := []struct {
		expSec *Section

		desc string
		k    string
		v    string
	}{{
		desc: "With empty key",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3",
			}, {
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "33",
			}, {
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "333",
			}, {
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3333",
			}},
		},
	}, {
		desc: "With invalid key",
		k:    "KEY-4",
		v:    "4",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3",
			}, {
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "33",
			}, {
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "333",
			}, {
				mode:     lineModeKeyValue,
				key:      "key-3",
				keyLower: "key-3",
				value:    "3333",
			}, {
				mode:     lineModeKeyValue,
				key:      "KEY-4",
				keyLower: "key-4",
				value:    "4",
			}},
		},
	}, {
		desc: "With valid key",
		k:    "KEY-3",
		v:    "replaced",
		expSec: &Section{
			mode:      sec.mode,
			name:      sec.name,
			nameLower: sec.nameLower,
			vars: []*variable{{
				mode:     lineModeKeyValue,
				key:      "KEY-4",
				keyLower: "key-4",
				value:    "4",
			}, {
				mode:     lineModeKeyValue,
				key:      "KEY-3",
				keyLower: "key-3",
				value:    "replaced",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		sec.replaceAll(c.k, c.v)

		test.Assert(t, "section", c.expSec, sec)
	}
}

func TestSectionGet(t *testing.T) {
	sec := &Section{
		mode:      lineModeSection,
		name:      "section",
		nameLower: "section",
		vars: []*variable{{
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v1",
		}, {
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v2",
		}},
	}

	cases := []struct {
		desc   string
		k      string
		def    string
		expVal string

		expOK bool
	}{{
		desc:   "With invalid key and default",
		k:      "key-1",
		def:    "default value",
		expVal: "default value",
	}, {
		desc:   "Valid key",
		k:      "k",
		def:    "default value",
		expOK:  true,
		expVal: "v2",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, ok := sec.get(c.k, c.def)

		test.Assert(t, "ok", c.expOK, ok)
		test.Assert(t, "value", c.expVal, got)
	}
}

func TestSectionGets(t *testing.T) {
	sec := &Section{
		mode:      lineModeSection,
		name:      "section",
		nameLower: "section",
		vars: []*variable{{
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v1",
		}, {
			mode:     lineModeKeyValue,
			key:      "k",
			keyLower: "k",
			value:    "v2",
		}},
	}

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

		test.Assert(t, "Gets value", c.exps, got)
		test.Assert(t, "Gets ok", c.expOK, ok)
	}
}
