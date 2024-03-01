// Copyright 2018 Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"git.sr.ht/~shulhan/pakakeh.go/lib/text"
)

// LineChanges represents a set of change in text.
type LineChanges []LineChange

// GetAllDels return all deleted chunks.
func (changes *LineChanges) GetAllDels() (allDels text.Chunks) {
	for _, change := range *changes {
		allDels = append(allDels, change.Dels...)
	}
	return
}

// GetAllAdds return all addition chunks.
func (changes *LineChanges) GetAllAdds() (allAdds text.Chunks) {
	for _, change := range *changes {
		allAdds = append(allAdds, change.Adds...)
	}
	return
}
