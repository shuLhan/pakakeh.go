// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewSectionMatch(t *testing.T) {
	cases := []struct {
		raw      string
		exp      *ConfigSection
		expError string
	}{{
		raw:      "test",
		expError: `unknown criteria "test"`,
	}}

	for _, c := range cases {
		got, err := newSectionMatch(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}
		got.postConfig(testParser.homeDir)
		test.Assert(t, "newSectionMatch", c.exp, got, true)
	}
}

func TestParseCriteriaAll(t *testing.T) {
	cases := []struct {
		raw      string
		exp      func(def ConfigSection) *ConfigSection
		expError string
	}{{
		raw: "all ",
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaAll,
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {

		raw: "canonical all",
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaCanonical,
			}, {
				name: criteriaAll,
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw: "final all",
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaFinal,
			}, {
				name: criteriaAll,
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw:      "user name all",
		expError: errCriteriaAll.Error(),
	}, {
		raw:      "all canonical",
		expError: errCriteriaAll.Error(),
	}}

	for _, c := range cases {
		got, err := newSectionMatch(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}
		got.postConfig(testParser.homeDir)
		test.Assert(t, "parseCriteriaAll",
			c.exp(*testDefaultSection), got, true)
	}
}

func TestNewSectionMatch_ParseCriteriaExec(t *testing.T) {
	cases := []struct {
		raw      string
		exp      func(def ConfigSection) *ConfigSection
		expError string
	}{{
		raw: `exec "echo true"`,
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaExec,
				arg:  `echo true`,
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw: `exec "echo true`,
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaExec,
				arg:  `echo true`,
			}}
			exp.useCriteria = true
			return &exp
		},
	}}

	for _, c := range cases {
		got, err := newSectionMatch(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}
		got.postConfig(testParser.homeDir)
		t.Logf("got: %+v", got)
		test.Assert(t, "parseCriteriaExec",
			c.exp(*testDefaultSection), got, true)
	}
}

func TestParseCriteriaWithArg(t *testing.T) {
	cases := []struct {
		raw      string
		exp      func(exp ConfigSection) *ConfigSection
		expError string
	}{{
		raw: `user name*`,
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaUser,
				arg:  `name*`,
				patterns: []*configPattern{{
					pattern: "name*",
				}},
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw: `user "a*,b*"`,
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaUser,
				arg:  `a*,b*`,
				patterns: []*configPattern{{
					pattern: "a*",
				}, {
					pattern: "b*",
				}},
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw: `user "a*,b*`,
		exp: func(exp ConfigSection) *ConfigSection {
			exp.criteria = []*matchCriteria{{
				name: criteriaUser,
				arg:  `a*,b*`,
				patterns: []*configPattern{{
					pattern: "a*",
				}, {
					pattern: "b*",
				}},
			}}
			exp.useCriteria = true
			return &exp
		},
	}}

	for _, c := range cases {
		got, err := newSectionMatch(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}
		got.postConfig(testParser.homeDir)
		test.Assert(t, "parseCriteriaWithArg",
			c.exp(*testDefaultSection), got, true)
	}
}
