// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"reflect"
	"testing"
)

func TestNewSectionMatch(t *testing.T) {
	cases := []struct {
		raw      string
		exp      *Section
		expError string
	}{{
		raw:      "test",
		expError: `unknown criteria "test"`,
	}}

	for _, c := range cases {
		got, err := newSectionMatch(c.raw)
		if err != nil {
			if c.expError != err.Error() {
				t.Fatalf("parseCriteriaWithArg: expecting error %s, got %s",
					c.expError, err.Error())
			}
			continue
		}
		got.init(testParser.workDir, testParser.homeDir)

		if !reflect.DeepEqual(*c.exp, *got) {
			t.Fatalf("newSectionMatch: expecting %v, got %v", c.exp, got)
		}
	}
}

func TestParseCriteriaAll(t *testing.T) {
	cases := []struct {
		raw      string
		exp      func(def Section) *Section
		expError string
	}{{
		raw: "all ",
		exp: func(exp Section) *Section {
			exp.criteria = []*matchCriteria{{
				name: criteriaAll,
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {

		raw: "canonical all",
		exp: func(exp Section) *Section {
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
		exp: func(exp Section) *Section {
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
			if c.expError != err.Error() {
				t.Fatalf("parseCriteriaWithArg: expecting error %s, got %s",
					c.expError, err.Error())
			}
			continue
		}
		got.init(testParser.workDir, testParser.homeDir)

		exp := c.exp(*testDefaultSection)
		if !reflect.DeepEqual(*exp, *got) {
			t.Fatalf("parseCriteriaAll: expecting %v, got %v", exp, got)
		}
	}
}

func TestNewSectionMatch_ParseCriteriaExec(t *testing.T) {
	cases := []struct {
		raw      string
		exp      func(def Section) *Section
		expError string
	}{{
		raw: `exec "echo true"`,
		exp: func(exp Section) *Section {
			exp.criteria = []*matchCriteria{{
				name: criteriaExec,
				arg:  `echo true`,
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw: `exec "echo true`,
		exp: func(exp Section) *Section {
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
			if c.expError != err.Error() {
				t.Fatalf("parseCriteriaWithArg: expecting error %s, got %s",
					c.expError, err.Error())
			}
			continue
		}
		got.init(testParser.workDir, testParser.homeDir)

		exp := c.exp(*testDefaultSection)
		if !reflect.DeepEqual(*exp, *got) {
			t.Fatalf("parseCriteriaExec: expecting %v, got %v", exp, got)
		}
	}
}

func TestParseCriteriaWithArg(t *testing.T) {
	cases := []struct {
		raw      string
		exp      func(exp Section) *Section
		expError string
	}{{
		raw: `user name*`,
		exp: func(exp Section) *Section {
			exp.criteria = []*matchCriteria{{
				name: criteriaUser,
				arg:  `name*`,
				patterns: []*pattern{{
					value: "name*",
				}},
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw: `user "a*,b*"`,
		exp: func(exp Section) *Section {
			exp.criteria = []*matchCriteria{{
				name: criteriaUser,
				arg:  `a*,b*`,
				patterns: []*pattern{{
					value: "a*",
				}, {
					value: "b*",
				}},
			}}
			exp.useCriteria = true
			return &exp
		},
	}, {
		raw: `user "a*,b*`,
		exp: func(exp Section) *Section {
			exp.criteria = []*matchCriteria{{
				name: criteriaUser,
				arg:  `a*,b*`,
				patterns: []*pattern{{
					value: "a*",
				}, {
					value: "b*",
				}},
			}}
			exp.useCriteria = true
			return &exp
		},
	}}

	for _, c := range cases {
		got, err := newSectionMatch(c.raw)
		if err != nil {
			if c.expError != err.Error() {
				t.Fatalf("parseCriteriaWithArg: expecting error %s, got %s",
					c.expError, err.Error())
			}
			continue
		}
		got.init(testParser.workDir, testParser.homeDir)

		exp := c.exp(*testDefaultSection)
		if !reflect.DeepEqual(*exp, *got) {
			t.Fatalf("parseCriteriaWithArg: expecting %v, got %v", exp, got)
		}
	}
}
