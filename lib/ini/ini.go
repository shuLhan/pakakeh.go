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
	secs []*Section
}

//
// Open and parse INI formatted file and return it as instance of Ini struct.
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
// AddSection append the new section to the list.
//
func (in *Ini) AddSection(sec *Section) {
	if sec == nil {
		return
	}
	if sec.mode != varModeEmpty && len(sec.Name) == 0 {
		return
	}

	sec.NameLower = strings.ToLower(sec.Name)

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

		for y := 0; y < len(sec.Vars); y++ {
			v := sec.Vars[y]

			if v.mode == varModeEmpty {
				continue
			}
			if v.mode&varModeSection > 0 || v.mode&varModeSubsection > 0 {
				continue
			}

			key := sec.NameLower + sep + sec.Sub + sep + v.KeyLower

			vals, ok := out[key]
			if !ok {
				out[key] = []string{v.Value}
			} else {
				out[key] = libstrings.AppendUniq(vals, v.Value)
			}
		}
	}

	return
}

//
// Get the last key on section and/or subsection (if not empty).
//
// If section, subsection, and key found it will return key's value and true;
// otherwise it will return nil and false.
//
func (in *Ini) Get(section, subsection, key string) (val string, ok bool) {
	if len(in.secs) == 0 || len(section) == 0 || len(key) == 0 {
		return
	}

	x := len(in.secs) - 1
	sec := strings.ToLower(section)

	for ; x >= 0; x-- {
		if in.secs[x].NameLower != sec {
			continue
		}

		if in.secs[x].Sub != subsection {
			continue
		}

		val, ok = in.secs[x].Get(key, "")
		if ok {
			return
		}
	}

	return
}

//
// GetBool return key's value as boolean.  If no key found it will return
// default value.
//
func (in *Ini) GetBool(section, subsection, key string, def bool) bool {
	out, ok := in.Get(section, subsection, key)
	if !ok {
		return def
	}

	return IsValueBoolTrue(out)
}

//
// GetSection return the last section that match with section name and/or
// subsection name.
// If section name is empty or no match found it will return nil.
//
func (in *Ini) GetSection(section, subsection string) *Section {
	if len(section) == 0 {
		return nil
	}

	section = strings.ToLower(section)

	for x := len(in.secs) - 1; x >= 0; x-- {
		if in.secs[x].NameLower != section {
			continue
		}
		if in.secs[x].Sub != subsection {
			continue
		}
		return in.secs[x]
	}

	return nil
}

//
// GetSections return all section that match with "name" as slice.
//
func (in *Ini) GetSections(name string) (secs []*Section) {
	if len(name) == 0 {
		return
	}

	name = strings.ToLower(name)

	for x := 0; x < len(in.secs); x++ {
		if in.secs[x].NameLower != name {
			continue
		}
		secs = append(secs, in.secs[x])
	}

	return
}

//
// GetString return key's value as string. if no key found it will return
// default value.
//
func (in *Ini) GetString(section, subsection, key, def string) (out string) {
	out, ok := in.Get(section, subsection, key)
	if !ok {
		out = def
	}

	return
}

//
// Gets key's values as slice of string in the same section and subsection.
//
func (in *Ini) Gets(section, subsection, key string) (out []string) {
	section = strings.ToLower(section)

	for _, sec := range in.secs {
		if sec.mode&varModeSection == 0 {
			continue
		}
		if sec.NameLower != section {
			continue
		}
		if sec.Sub != subsection {
			continue
		}
		vals, ok := sec.Gets(key, nil)
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
	newSecs := make([]*Section, 0, len(in.secs))

	for _, sec := range in.secs {
		if sec.mode == varModeEmpty {
			continue
		}
		newSec := &Section{
			mode:      varModeSection,
			Name:      sec.Name,
			NameLower: sec.NameLower,
		}
		if len(sec.Sub) > 0 {
			newSec.mode |= varModeSubsection
			newSec.Sub = sec.Sub
		}
		for _, v := range sec.Vars {
			if v.mode == varModeEmpty || v.mode == varModeComment {
				continue
			}

			newValue := v.Value
			if len(v.Value) == 0 {
				newValue = "true"
			}

			newSec.AddUniqValue(v.Key, newValue)
		}
		newSecs = mergeSection(newSecs, newSec)
	}

	in.secs = newSecs
}

//
// mergeSection merge a section (and subsection) into slice.
//
func mergeSection(secs []*Section, newSec *Section) []*Section {
	for x := 0; x < len(secs); x++ {
		if secs[x].NameLower != newSec.NameLower {
			continue
		}
		if secs[x].Sub != newSec.Sub {
			continue
		}
		for _, v := range newSec.Vars {
			if v.mode == varModeEmpty || v.mode == varModeComment {
				continue
			}
			secs[x].AddUniqValue(v.KeyLower, v.Value)
		}
		return secs
	}
	secs = append(secs, newSec)
	return secs
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

		for _, v := range in.secs[x].Vars {
			fmt.Fprint(w, v)
		}
	}

	return
}
