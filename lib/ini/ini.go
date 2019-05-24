// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shuLhan/share/lib/debug"
	libstrings "github.com/shuLhan/share/lib/strings"
)

//
// Ini contains the parsed file.
//
type Ini struct {
	secs []*section
}

//
// Open and parse INI formatted file.
//
// On fail it will return incomplete instance of Ini with error.
//
func Open(filename string) (in *Ini, err error) {
	reader := newReader()

	in, err = reader.parseFile(filename)

	if debug.Value >= 1 && err == nil {
		err = in.Write(os.Stdout)
	}

	return
}

//
// Parse INI format from text.
//
func Parse(text []byte) (in *Ini, err error) {
	reader := newReader()

	return reader.Parse(text)
}

//
// addSection append the new section to the list.
//
func (in *Ini) addSection(sec *section) {
	if sec == nil {
		return
	}
	if sec.mode != lineModeEmpty && len(sec.name) == 0 {
		return
	}

	sec.nameLower = strings.ToLower(sec.name)

	in.secs = append(in.secs, sec)
}

//
// AsMap return the INI contents as mapping of
// (section-name ":" subsection-name ":" variable-name) as key
// and the variable's values as slice of string.
// For example, given the following INI file,
//
//	[section1]
//	key = value
//
//	[section2 "sub"]
//	key2 = value2
//	key2 = value3
//
// it will be mapped as,
//
//	map["section1::key"] = []string{"value"}
//	map["section2:sub:key2"] = []string{"value2", "value3"}
//
func (in *Ini) AsMap() (out map[string][]string) {
	sep := ":"
	out = make(map[string][]string)

	for x := 0; x < len(in.secs); x++ {
		sec := in.secs[x]

		for y := 0; y < len(sec.vars); y++ {
			v := sec.vars[y]

			if v.mode == lineModeEmpty {
				continue
			}
			if v.mode&lineModeSection > 0 || v.mode&lineModeSubsection > 0 {
				continue
			}

			key := sec.nameLower + sep + sec.sub + sep + v.keyLower

			vals, ok := out[key]
			if !ok {
				out[key] = []string{v.value}
			} else {
				out[key] = libstrings.AppendUniq(vals, v.value)
			}
		}
	}

	return
}

//
// Get the last key on section and/or subsection.
//
// If key found it will return its value and true; otherwise it will return
// default value in def and false.
//
func (in *Ini) Get(section, subsection, key, def string) (val string, ok bool) {
	if len(in.secs) == 0 || len(section) == 0 || len(key) == 0 {
		return def, false
	}

	x := len(in.secs) - 1
	sec := strings.ToLower(section)

	for ; x >= 0; x-- {
		if in.secs[x].nameLower != sec {
			continue
		}

		if in.secs[x].sub != subsection {
			continue
		}

		val, ok = in.secs[x].get(key, def)
		if ok {
			return
		}
	}

	return def, false
}

//
// GetBool return key's value as boolean.  If no key found it will return
// default value.
//
func (in *Ini) GetBool(section, subsection, key string, def bool) bool {
	out, ok := in.Get(section, subsection, key, "false")
	if !ok {
		return def
	}

	return IsValueBoolTrue(out)
}

//
// Gets key's values as slice of string in the same section and subsection.
//
func (in *Ini) Gets(section, subsection, key string) (out []string) {
	section = strings.ToLower(section)

	for _, sec := range in.secs {
		if sec.mode&lineModeSection == 0 {
			continue
		}
		if sec.nameLower != section {
			continue
		}
		if sec.sub != subsection {
			continue
		}
		vals, ok := sec.gets(key, nil)
		if !ok {
			continue
		}

		out = libstrings.AppendUniq(out, vals...)
	}
	return
}

//
// Prune remove all empty lines, comments, and merge all section and
// subsection with the same name into one group.
//
func (in *Ini) Prune() {
	newSecs := make([]*section, 0, len(in.secs))

	for _, sec := range in.secs {
		if sec.mode == lineModeEmpty {
			continue
		}
		newSec := &section{
			mode:      lineModeSection,
			name:      sec.name,
			nameLower: sec.nameLower,
		}
		if len(sec.sub) > 0 {
			newSec.mode |= lineModeSubsection
			newSec.sub = sec.sub
		}
		for _, v := range sec.vars {
			if v.mode == lineModeEmpty || v.mode == lineModeComment {
				continue
			}

			newValue := v.value
			if len(v.value) == 0 {
				newValue = "true"
			}

			newSec.addUniqValue(v.key, newValue)
		}
		newSecs = mergeSection(newSecs, newSec)
	}

	in.secs = newSecs
}

//
// mergeSection merge a section (and subsection) into slice.
//
func mergeSection(secs []*section, newSec *section) []*section {
	for x := 0; x < len(secs); x++ {
		if secs[x].nameLower != newSec.nameLower {
			continue
		}
		if secs[x].sub != newSec.sub {
			continue
		}
		for _, v := range newSec.vars {
			if v.mode == lineModeEmpty || v.mode == lineModeComment {
				continue
			}
			secs[x].addUniqValue(v.keyLower, v.value)
		}
		return secs
	}
	secs = append(secs, newSec)
	return secs
}

//
// Rebase merge the other INI sections into this INI sections.
//
func (in *Ini) Rebase(other *Ini) {
	for _, otherSec := range other.secs {
		in.secs = mergeSection(in.secs, otherSec)
	}
}

//
// Save the current parsed Ini into file `filename`. It will overwrite the
// destination file if it's exist.
//
func (in *Ini) Save(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}

	err = in.Write(f)

	errClose := f.Close()
	if errClose != nil {
		println("ini.Save:", errClose)
	}

	return
}

//
// Write the current parsed Ini into writer `w`.
//
func (in *Ini) Write(w io.Writer) (err error) {
	for x := 0; x < len(in.secs); x++ {
		fmt.Fprint(w, in.secs[x])

		for _, v := range in.secs[x].vars {
			fmt.Fprint(w, v)
		}
	}

	return
}
