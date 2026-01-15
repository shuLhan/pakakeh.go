// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2025 M. Shulhan <ms@kilabit.info>

package git

import (
	"regexp"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestParseIgnorePattern(t *testing.T) {
	type testCase struct {
		pattern string
		exp     IgnorePattern
	}
	var listCase = []testCase{{
		pattern: `#`,
		exp: IgnorePattern{
			pattern: nil,
		},
	}, {
		pattern: `a #`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?a/?$`),
		},
	}, {
		pattern: `a \#`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?a \#/?$`),
		},
	}, {
		pattern: `?`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?[^/]/?$`),
		},
	}, {
		pattern: `!a`,
		exp: IgnorePattern{
			pattern:  regexp.MustCompile(`^(.*/|/)?a/?$`),
			isNegate: true,
		},
	}, {
		pattern: `*`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?.*$`),
		},
	}, {
		pattern: `*/`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?.*$`),
			isDir:   true,
		},
	}, {
		pattern: `**`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?.*$`),
		},
	}, {
		pattern: `***`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?.*$`),
		},
	}, {
		pattern: `**/**`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?.*$`),
		},
	}, {
		pattern: `**/**/`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?.*$`),
			isDir:   true,
		},
	}, {
		pattern: `**/**foo`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?[^/]*foo/?$`),
		},
	}, {
		pattern: `**/foo/**`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?foo/(.*)/?$`),
		},
	}, {
		pattern: `foo/**`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?foo/(.*)/?$`),
		},
	}, {
		pattern: `foo`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?foo/?$`),
		},
	}, {
		pattern: `foo/`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?foo/$`),
			isDir:   true,
		},
	}, {
		pattern: `/foo`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?foo/?$`),
		},
	}, {
		pattern: `foo/**/bar`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^/?foo(/.*)?/bar/?$`),
		},
	}, {
		pattern: `a+b|c`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?a\+b\|c/?$`),
		},
	}, {
		pattern: `(a|b)`,
		exp: IgnorePattern{
			pattern: regexp.MustCompile(`^(.*/|/)?\(a\|b\)/?$`),
		},
	}}
	for _, tc := range listCase {
		var got = ParseIgnorePattern([]byte(tc.pattern))
		test.Assert(t, tc.pattern, tc.exp, got)
	}
}

func TestIgnorePattern_IsMatch(t *testing.T) {
	type testCase struct {
		listCase map[string]bool
		pattern  string
	}
	var listCase = []testCase{{
		pattern: `	foo   # comment`,
		listCase: map[string]bool{
			`foo`:       true,
			`foo/`:      true,
			`a/foo`:     true,
			`a/b/foo`:   true,
			`afoo`:      false,
			`a/foo/bar`: false,
		},
	}, {
		pattern: `foo/`,
		listCase: map[string]bool{
			`foo/bar`:   true,
			`a/foo/bar`: true,
			`foo`:       false,
			`afoo`:      false,
			`a/foo`:     false,
		},
	}, {
		pattern: `/foo`,
		listCase: map[string]bool{
			`foo`:       true,
			`foo/bar`:   false,
			`a/foo`:     false,
			`a/foo/bar`: false,
			`afoo`:      false,
		},
	}, {
		pattern: `/foo/bar`,
		listCase: map[string]bool{
			`foo/bar`:     true,
			`foo/bar/`:    true,
			`foo/bar/z`:   false,
			`afoo/bar`:    false,
			`a/foo/bar`:   false,
			`a/foo/bar/z`: false,
		},
	}, {
		pattern: `foo/bar/`,
		listCase: map[string]bool{
			`foo/bar/`:    true,
			`foo/bar/z`:   true,
			`foo/bar`:     false,
			`afoo/bar`:    false,
			`a/foo/bar`:   false,
			`a/foo/bar/z`: false,
		},
	}, {
		pattern: `/foo/bar/`,
		listCase: map[string]bool{
			`/foo/bar/`:   true,
			`foo/bar/`:    true,
			`foo/bar/z`:   true,
			`foo/bar`:     false,
			`afoo/bar`:    false,
			`a/foo/bar`:   false,
			`a/foo/bar/z`: false,
		},
	}, {
		pattern: `foo*`,
		listCase: map[string]bool{
			`foo`:       true,
			`foobar`:    true,
			`a/foo`:     true,
			`a/foobar`:  true,
			`a/foo/bar`: false,
		},
	}, {
		pattern: `foo.*`,
		listCase: map[string]bool{
			`foo.`:       true,
			`foo.bar`:    true,
			`a/foo.bar`:  true,
			`a/foo./bar`: false,
			`a/foobar`:   false,
		},
	}, {
		pattern: `*foo`,
		listCase: map[string]bool{
			`foo`:       true,
			`afoo`:      true,
			`a/foo`:     true,
			`a/bfoo`:    true,
			`foobar`:    false,
			`a/foo/bar`: false,
			`a/foobar`:  false,
		},
	}, {
		pattern: `foo?`,
		listCase: map[string]bool{
			`food`:      true,
			`a/food`:    true,
			`foo`:       false,
			`foobar`:    false,
			`afoo`:      false,
			`a/foo`:     false,
			`a/foobar`:  false,
			`a/foo/bar`: false,
		},
	}, {
		pattern: `?foo`,
		listCase: map[string]bool{
			`afoo`:     true,
			`a/afoo`:   true,
			`foo`:      false,
			`a/foo`:    false,
			`a/foobar`: false,
		},
	}, {
		pattern: `foo/*`,
		listCase: map[string]bool{
			`foo`:       false,
			`foo/bar`:   true,
			`foo/bar/z`: false,
		},
	}, {
		pattern: `**/foo`,
		listCase: map[string]bool{
			`foo`:         true,
			`/foo`:        true,
			`a/foo`:       true,
			`a/b/foo`:     true,
			`a/b/foo/bar`: false,
		},
	}, {
		pattern: `foo/**`,
		listCase: map[string]bool{
			`foo/bar`:     true,
			`foo/bar/foo`: true,
			`foo`:         false,
			`a/foo/bar`:   false,
		},
	}, {
		pattern: `foo/**/bar`,
		listCase: map[string]bool{
			`foo/bar`:     true,
			`foo/a/bar`:   true,
			`foo/a/b/bar`: true,
			`foo/bar/foo`: false,
			`bar`:         false,
			`a/foo/bar`:   false,
			`a/foo/b/bar`: false,
		},
	}, {
		pattern: `a+b|c`,
		listCase: map[string]bool{
			`a+b|c`: true,
			`aab|c`: false,
			`aab`:   false,
		},
	}, {
		pattern: `(a|b)`,
		listCase: map[string]bool{
			`(a|b)`: true,
			`a`:     false,
			`b`:     false,
		},
	}}
	for _, tc := range listCase {
		var pat = ParseIgnorePattern([]byte(tc.pattern))
		for name, exp := range tc.listCase {
			var got = pat.IsMatch(name)
			if exp != got {
				t.Fatalf("%q: on %q want %t, got %t",
					tc.pattern, name, exp, got)
			}
		}
	}
}

func TestRemoveComment(t *testing.T) {
	type testCase struct {
		pattern string
		exp     string
	}
	var listCase = []testCase{{
		pattern: `a#`,
		exp:     `a`,
	}, {
		pattern: `a\#`,
		exp:     `a\#`,
	}, {
		pattern: `a\##`,
		exp:     `a\#`,
	}}
	for _, tc := range listCase {
		got := removeComment([]byte(tc.pattern))
		test.Assert(t, tc.pattern, tc.exp, string(got))
	}
}
