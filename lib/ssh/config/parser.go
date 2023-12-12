// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parser for the SSH config file that merge all "Include" files and convert
// them all into lines of string.
type parser struct {
	files   map[string]struct{}
	workDir string
	homeDir string
}

func newParser() (p *parser, err error) {
	p = &parser{
		files: make(map[string]struct{}),
	}

	p.workDir, err = os.Getwd()
	if err != nil {
		return nil, err
	}
	p.homeDir, err = os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// load the config file(s) using glob(7) pattern and convert them into lines.
func (p *parser) load(dir, pattern string) (lines []string, err error) {
	switch pattern[0] {
	case '~':
		// File is absolute path to user's home directory.
		pattern = filepath.Join(p.homeDir, pattern[1:])

	case '/':
		// File is absolute path, do nothing.

	case '.':
		// File is relative to current working directory.
		pattern = filepath.Join(p.workDir, pattern)

	default:
		if len(dir) != 0 {
			// File is relative to previous directory.
			pattern = filepath.Join(dir, pattern)
		} else {
			// File is relative to user's .ssh directory.
			pattern = filepath.Join(p.homeDir, ".ssh", pattern)
		}
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", pattern, err)
	}

	// File did not exist.
	if len(matches) == 0 {
		return nil, nil
	}

	rawLines := make([]string, 0)
	for _, file := range matches {
		_, ok := p.files[file]
		if ok {
			// File already loaded previously.
			continue
		}

		newLines, err := readLines(file)
		if err != nil {
			return nil, err
		}
		if len(newLines) > 0 {
			rawLines = append(rawLines, newLines...)
		}
		p.files[file] = struct{}{}
	}

	lines = make([]string, 0, len(rawLines))

	dir = filepath.Dir(pattern)

	// Check and parse the Include directive in each lines.
	for x := 0; x < len(rawLines); x++ {
		if !isIncludeDirective(rawLines[x]) {
			lines = append(lines, rawLines[x])
			continue
		}

		patterns := parseInclude(rawLines[x])

		for _, pattern := range patterns {
			includeContents, err := p.load(dir, pattern)
			if err != nil {
				return nil, err
			}
			if len(includeContents) == 0 {
				continue
			}
			lines = append(lines, includeContents...)
		}
	}

	return lines, nil
}

// isIncludeDirective will return true if line started with "include", in case
// insensitive.
func isIncludeDirective(line string) bool {
	keyLen := len(keyInclude)
	if len(line) <= keyLen {
		return false
	}
	if strings.ToLower(line[:keyLen]) == keyInclude {
		if line[keyLen] == ' ' || line[keyLen] == '=' {
			return true
		}
	}
	return false
}

func parseInclude(line string) (patterns []string) {
	var (
		x = len(keyInclude)

		start    int
		end      int
		useQuote bool
	)

	for ; x < len(line); x++ {
		if line[x] != ' ' {
			break
		}
	}
	if line[x] == '=' {
		x++
	}
	start = x

	for x < len(line) {
		if line[x] == '"' {
			useQuote = true
			x++
			start = x
		}

		for ; x < len(line); x++ {
			if line[x] == ' ' {
				if useQuote {
					continue
				}
				end = x
				break
			}
			if line[x] == '"' {
				if useQuote {
					useQuote = false
					end = x
					x++
					break
				}
			}
		}
		if end == 0 {
			end = len(line)
		}
		if end > start {
			patterns = append(patterns, line[start:end])
		}
		end = 0

		for ; x < len(line); x++ {
			if line[x] != ' ' {
				break
			}
		}
		start = x
	}

	return patterns
}

// readLines convert the contents of file into lines as slice of string.
// Any empty lines or line start with comment '#' will be removed.
func readLines(file string) (lines []string, err error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	rawLines := bytes.Split(contents, []byte{'\n'})
	for x := 0; x < len(rawLines); x++ {
		rawLines[x] = bytes.TrimSpace(rawLines[x])
		if len(rawLines[x]) == 0 {
			continue
		}
		if rawLines[x][0] == '#' {
			continue
		}
		lines = append(lines, string(rawLines[x]))
	}

	return lines, nil
}

// parseArgs split single line arguments into list of string, separated by
// `sep` (default to space), grouped by double quote.
//
// For example, given raw argument `a "b c" d` it would return "a", "b c", and
// "d".
func parseArgs(raw string, sep byte) (args []string) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}
	if sep == 0 {
		sep = ' '
	}
	var (
		x        int
		begin    int
		useQuote bool
	)
	args = make([]string, 0)
	for ; x < len(raw); x++ {
		c := raw[x]
		if useQuote {
			if c != '"' {
				continue
			}
			args = append(args, raw[begin:x])
			begin = len(raw)
			useQuote = false
			continue
		}
		if c == sep {
			if begin < x {
				args = append(args, raw[begin:x])
				begin = len(raw)
			}
			continue
		}
		if c == '"' {
			useQuote = true
			begin = x + 1
		} else if begin == len(raw) {
			begin = x
		}
	}
	if begin < x {
		args = append(args, raw[begin:x])
	}
	return args
}
