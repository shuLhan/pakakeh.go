// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import "strings"

func newSectionHost(rawPattern string) (host *ConfigSection) {
	patterns := strings.Fields(rawPattern)

	host = newConfigSection()
	host.patterns = make([]*configPattern, 0, len(patterns))

	for _, pattern := range patterns {
		pat := newConfigPattern(pattern)
		host.patterns = append(host.patterns, pat)
	}
	return host
}
