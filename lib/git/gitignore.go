// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2025 M. Shulhan <ms@kilabit.info>

package git

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Gitignore is a type that represent ".gitignore" file.
// The content of Gitignore can be populated from [LoadGitignore] function or
// [Gitignore.Parse] method.
type Gitignore struct {
	// dir path to directory that contains ".gitignore" file.
	dir string

	// path to the ".gitignore" file.
	path string

	// excludePatterns contains list of excluded pattern from
	// ".gitignore" file.
	excludePatterns []ignorePattern

	// includePatterns contains list of include pattern, the one that
	// start with "!".
	includePatterns []ignorePattern
}

// LoadGitignore load the gitignore file inside directory `dir`.
// It will return nil without error if the ".gitignore" file is not exists.
//
// Any invalid pattern will be ignored.
func LoadGitignore(dir string) (ign *Gitignore, err error) {
	var logp = `LoadGitignore`
	var content []byte

	ign = &Gitignore{
		path: filepath.Join(dir, `.gitignore`),
	}
	content, err = os.ReadFile(ign.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	ign.Parse(dir, content)
	return ign, nil
}

// Parse the raw content of ".gitignore" file that located inside the `dir`
// directory.
// This is an alternative to populate Gitignore content beside
// [LoadGitignore].
// Any invalid pattern inside the `content` will be ignored.
func (ign *Gitignore) Parse(dir string, content []byte) {
	ign.dir = dir
	var lines = bytes.Split(content, []byte{'\n'})
	var line []byte
	for _, line = range lines {
		var pat ignorePattern
		pat = parsePattern(line)
		if pat.pattern == nil {
			// Skip invalid pattern.
			continue
		}
		if pat.isNegate {
			ign.includePatterns = append(ign.includePatterns, pat)
		} else {
			ign.excludePatterns = append(ign.excludePatterns, pat)
		}
	}
}

// IsIgnored return true if the `path` is ignored by this Gitignore content.
// The `path` is relative to Gitignore directory.
func (ign *Gitignore) IsIgnored(path string) bool {
	path = strings.TrimSpace(path)
	if path == `` {
		return true
	}
	var fullpath = filepath.Join(ign.dir, path)
	var fi os.FileInfo
	fi, _ = os.Stat(fullpath)
	if fi != nil {
		if fi.IsDir() {
			path += "/"
		}
	}
	var pat ignorePattern
	for _, pat = range ign.includePatterns {
		if pat.isMatch(path) {
			return false
		}
	}
	for _, pat = range ign.excludePatterns {
		if pat.isMatch(path) {
			return true
		}
	}
	return false
}
