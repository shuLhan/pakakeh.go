// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package ini implement reading and writing INI configuration as defined by
// Git configuration file syntax [1].
//
// Feature Promises
//
// Reading and writing on the same file should not change the content of
// file (including comment).
//
// Unsupported Features
//
// Git `include` and `includeIf`
//
// Syntax
//
// (S.1.0) The `#` and `;` characters begin comments to the end of line.
//
// (S.1.1) Blank lines are ignored.
//
// (Section)
//
// (S.2.0) A section begins with the name of the section in square brackets.
//
// (S.2.1) A section continues until the next section begins.
//
// (S.2.2) Section name are case-insensitive.
//
// (S.2.3) Variable name must start with an alphabetic character, no
// whitespace before name or after '['.
//
// (S.2.4) Section name only allow alphanumeric characters, `-` and `.`.
//
// (S.2.5) Section can be further divided into subsections.
//
// (S.2.6) Section headers cannot span multiple lines.
//
// (S.2.7) You can have `[section]` if you have `[section "subsection"]`, but
// you don’t need to.
//
// (S.2.8) All the other lines (and the remainder of the line after the
// section header) are recognized as setting variables, in the form
// `name = value`.
//
// (SubSection)
//
// (S.3.0) To begin a subsection put its name in double quotes, separated by
// space from the section name, in the section header, for example
//
//	[section "subsection"]
//
// (S.3.1) Subsection name are case sensitive and can contain any characters
// except newline and the null byte.
//
// (S.3.2) Subsection name can include doublequote `"` and backslash by
// escaping them as `\"` and `\\`, respectively.
//
// (S.3.3) Other backslashes preceding other characters are dropped when
// reading subsection name; for example, `\t` is read as `t` and `\0` is read
// as `0`
//
// (Variable)
//
// (S.4.0) Variable must belong to some section, which means that there
// must be a section header before the first setting of a variable.
//
// (S.5.1) Variable name are case-insensitive.
//
// (S.5.2) Variable name allow only alphanumeric characters and `-`.
//
// (S.5.3) Variable name must start with an alphabetic character.
//
// (S.5.4) Variable name without value is a short-hand to set the value to the
// boolean "true" value, e.g.
//
//	[section]
//		thisistrue # equal to thisistrue=true
//
// (Value)
//
// (S.6.0) Value can be empty or not set, see S.5.4.
//
// (S.6.1) Internal whitespaces within the value are retained verbatim.
//
// (S.6.2) Value can be continued to the next line by ending it with a `\`;
// the backquote and the end-of-line are stripped.
//
// (S.6.3) Leading and trailing.whitespaces on value without double quote will
// be discarded.
//
// (S.6.4) Value can contain inline comment, e.g.
//
//	key = value # this is inline comment
//
// (S.6.5) Comment characters, '#' and ';', inside double quoted value will be
// read as content of value, not as comment,
//
//	key = "value # with hash"
//
// (S.6.6) Inside value enclosed double quotes, the following escape sequences
// are recognized: `\"` for doublequote, `\\` for backslash, `\n` for newline
// character (NL), `\t` for horizontal tabulation (HT, TAB) and `\b` for
// backspace (BS).
//
// (S.6.8) Other char escape sequences (including octal escape sequences) are
// invalid.
//
// References
//
// [1] https://git-scm.com/docs/git-config#_configuration_file
//
package ini

import (
	"fmt"
	"io"
	"os"
	"strings"
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
	reader := NewReader()

	in, err = reader.ParseFile(filename)

	if debug >= debugL1 && err == nil {
		err = in.Write(os.Stdout)
	}

	return
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
// Get the last key on section and/or subsection (if not empty).
//
// It will return nil and false,
// (1) If Ini file contains no sections,
// (2) section or key parameter is empty, or
// (3) no key found.
//
// Otherwise it will return key's value and true.
//
func (in *Ini) Get(section, subsection, key string) (val string, ok bool) {
	// (1) (2)
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

	// (3)
	return
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
