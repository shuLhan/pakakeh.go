// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"strings"
	"unicode"
)

// IsValidVarName check if "v" is valid variable name, where the
// first character must be a letter and the rest should contains only letter,
// digit, period, hyphen, or underscore.
// If "v" is valid it will return true.
func IsValidVarName(v string) bool {
	if len(v) == 0 {
		return false
	}
	for x, r := range v {
		if x == 0 && !unicode.IsLetter(r) {
			return false
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) ||
			r == tokHyphen || r == tokDot || r == tokUnderscore {
			continue
		}
		return false
	}
	return true
}

// IsValueBoolTrue will return true if variable contains boolean value for
// true. The following conditions is boolean true for value: "" (empty
// string), "true", "yes", "ya", "t", "1" (all of string is case insensitive).
func IsValueBoolTrue(v string) bool {
	if len(v) == 0 {
		return false
	}
	v = strings.ToLower(v)
	if v == "true" || v == "t" || v == "ya" || v == "yes" || v == "1" {
		return true
	}
	return false
}

// parseTag parse the ini field tag as used in the struct's field.
// This returned slice always have 4 string element: section, subsection, key,
// and default value.
func parseTag(in string) (tags []string) {
	var (
		sb       strings.Builder
		r        rune
		x        int
		foundSep bool
		isEsc    bool
	)

	tags = append(tags, ``, ``, ``, ``)

	in = strings.TrimSpace(in)
	if len(in) == 0 {
		return tags
	}

	// Parse the section.
	for x, r = range in {
		if r == ':' {
			foundSep = true
			break
		}
	}
	if !foundSep {
		// If no ":" found, the tag is the section.
		tags[0] = in
		return tags
	}

	tags[0] = in[:x]
	in = in[x+1:]

	// Parse the subsection.
	foundSep = false
	sb.Reset()
	for x, r = range in {
		if r == '\\' {
			if isEsc {
				sb.WriteRune('\\')
				isEsc = false
			} else {
				isEsc = true
			}
			continue
		}
		if r == '"' {
			if isEsc {
				sb.WriteRune(r)
				isEsc = false
				continue
			}
			break
		}
		if r == ':' {
			if isEsc {
				sb.WriteRune(r)
				isEsc = false
				continue
			}
			foundSep = true
			break
		}
		sb.WriteRune(r)
	}
	tags[1] = sb.String()

	if !foundSep {
		return tags
	}
	in = in[x+1:]

	// Parse variable name.
	foundSep = false
	for x, r = range in {
		if r == ':' {
			foundSep = true
			break
		}
	}
	if !foundSep {
		tags[2] = in
		return tags
	}

	tags[2] = in[:x]
	in = in[x+1:]

	// The rest is the default value.
	tags[3] = in

	return tags
}
