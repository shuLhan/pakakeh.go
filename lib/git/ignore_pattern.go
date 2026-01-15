// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2025 M. Shulhan <ms@kilabit.info>

package git

import (
	"bytes"
	"path/filepath"
	"regexp"
)

// IgnorePattern is a type that store the parsed ignore pattern line from
// gitignore file.
type IgnorePattern struct {
	pattern  *regexp.Regexp
	isDir    bool // True if pattern end with '/'.
	isNegate bool // True if pattern start with '!'.
}

// ParseIgnorePattern parse the line from gitignore.
// At this point, the line must be not empty and not a comment.
// If the pattern is invalid it will be ignored and [IsMatch] will always
// return false.
func ParseIgnorePattern(line []byte) (ign IgnorePattern) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		// Skip empty line.
		return ign
	}
	if line[0] == '#' {
		// Skip comment.
		return ign
	}
	if line[0] == '!' {
		ign.isNegate = true
		line = line[1:]
	}
	line = removeComment(line)

	if line[len(line)-1] == '/' {
		ign.isDir = true
		line = line[:len(line)-1]
	}

	var sepIdx int

	// The "**/foo" pattern is equal to "foo", so we can remove the "**/".
	for bytes.HasPrefix(line, []byte("**/")) {
		sepIdx = -1 // Flag it as zero or more directory before.
		line = line[3:]
	}
	if len(line) == 0 {
		return ign
	}
	for bytes.HasPrefix(line, []byte("**")) {
		line = line[1:]
	}
	if len(line) == 0 || len(line) == 1 && line[0] == '*' {
		ign.pattern, _ = regexp.Compile(`^/?.*$`)
		return ign
	}

	if sepIdx == 0 {
		sepIdx = bytes.LastIndexByte(line, '/')
	}

	var RE_EVERYTHING_INSIDE = []byte(`/(.*)`)
	var RE_FILE_OR_DIR = []byte(`/?$`)
	var RE_NO_DIR_BEFORE = []byte(`^/?`)
	var RE_ONE_CHAR_EXCEPT_SEP = []byte(`[^/]`)
	var RE_ZERO_OR_MORE_CHAR_EXCEPT_SEP = []byte(`[^/]*`)
	var RE_ZERO_OR_MORE_DIR = []byte(`(/.*)?/`)
	var RE_ZERO_OR_MORE_DIR_BEFORE = []byte(`^(.*/|/)?`)

	// First replacement,
	// - Replace single '*' with RE_ZERO_OR_MORE_CHAR_EXCEPT_SEP
	// - Replace '?' with RE_ONE_CHAR_EXCEPT_SEP
	// - Escape regex metacharacters '.', '+', '|', '(', and ')'
	var newline = make([]byte, 0, len(line))
	var lastidx = len(line) - 1
	var x = 0
	var c byte
	for x < len(line) {
		c = line[x]
		switch c {
		case '*':
			if x != lastidx && line[x+1] == '*' {
				// The '**' is for regex.
				newline = append(newline, c)
				newline = append(newline, c)
				x++
			} else {
				newline = append(newline, RE_ZERO_OR_MORE_CHAR_EXCEPT_SEP...)
			}
		case '?':
			newline = append(newline, RE_ONE_CHAR_EXCEPT_SEP...)
		case '.', '+', '|', '(', ')':
			newline = append(newline, '\\', c)
		default:
			newline = append(newline, c)
		}
		x++
	}
	line = newline

	line = bytes.ReplaceAll(line, []byte("/**/"), RE_ZERO_OR_MORE_DIR)
	line = bytes.ReplaceAll(line, []byte("/**"), RE_EVERYTHING_INSIDE)
	line = bytes.ReplaceAll(line, []byte("**"), RE_ZERO_OR_MORE_CHAR_EXCEPT_SEP)

	switch sepIdx {
	case -1:
		// "foo" single string without directory separator should match only
		// if its end with it.
		// "foo" match with "/foo" or "a/foo" but not "afoo" or
		// "a/foo/bar".
		line = append(RE_ZERO_OR_MORE_DIR_BEFORE, line...)
	case 0:
		// "/foo" match with "foo" or "/foo" but not "a/foo" nor
		// "a/foo/bar".
		line = append(RE_NO_DIR_BEFORE, line[1:]...)
	default:
		// "foo/bar" should match with "/foo/bar" but not "a/foo/bar".
		if line[0] == '/' {
			line = line[1:]
		}
		line = append(RE_NO_DIR_BEFORE, line...)
	}
	if ign.isDir {
		line = append(line, '/', '$')
	} else {
		line = append(line, RE_FILE_OR_DIR...)
	}
	ign.pattern, _ = regexp.Compile(string(line))
	return ign
}

func removeComment(line []byte) []byte {
	var x = bytes.LastIndexByte(line, '#')
	if x == -1 {
		return line
	}
	for line[x-1] == '\\' {
		x = bytes.LastIndexByte(line[:x-1], '#')
		if x == -1 {
			return line
		}
	}
	return bytes.TrimSpace(line[:x])
}

// IsMatch return true if the `path` match with the pattern.
func (pat *IgnorePattern) IsMatch(path string) bool {
	if pat.pattern.MatchString(path) {
		return true
	}
	if !pat.isDir {
		return false
	}
	path = filepath.Dir(path)
	for path != `.` {
		if pat.pattern.MatchString(path + "/") {
			return true
		}
		path = filepath.Dir(path)
	}
	return false
}
