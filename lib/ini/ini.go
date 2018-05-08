//
// Package ini implement reading and writing INI configuration as defined by
// Git configuration file syntax [1].
//
// ## Feature Promises
//
// * Reading and writing on the same file should not change the content of
// file (including comment), as long as no variable has been added or updated.
//
// ## Unsupported Features
//
// * Git `include` and `includeIf`
//
// ## Syntax
//
// (S.1.0) The `#` and `;` characters begin comments to the end of line.
// (S.1.1) Blank lines are ignored.
//
// (S.2.0) A section begins with the name of the section in square brackets.
// (S.2.1) A section continues until the next section begins.
// (S.2.2) Section name are case-insensitive.
// (S.2.3) Section name only allow alphanumeric characters, `-` and `.`.
// (S.2.4) Section can be further divided into subsections.
// (S.2.5) Section headers cannot span multiple lines.
// (S.2.6) You can have `[section]` if you have `[section "subsection"]`, but
// you don’t need to.
// (S.2.7) All the other lines (and the remainder of the line after the
// section header) are recognized as setting variables, in the form
// `name = value`.
//
// (S.3.0) To begin a subsection put its name in double quotes, separated by
// space from the section name, in the section header, for example
//
//	[section "subsection"]
//
// (S.3.1) Subsection name are case sensitive and can contain any characters
// except newline and the null byte.
// (S.3.2) Subsection name can include doublequote `"` and backslash by
// escaping them as `\"` and `\\`, respectively.
// (S.3.3) Other backslashes preceding other characters are dropped when
// reading subsection name; for example, `\t` is read as `t` and `\0` is read
// as `0`
//
// (S.4.0) Variable must belong to some section, which means that there
// must be a section header before the first setting of a variable.
//
// (S.5.1) Variable name are case-insensitive.
// (S.5.2) Variable name allow only alphanumeric characters and `-`.
// (S.5.3) Variable name must start with an alphabetic character.
// (S.5.4) Variable name without value is a short-hand to set the value to the
// boolean "true" value, e.g.
//
//	[section]
//		thisistrue # equal to thisistrue=true
//
// (S.6.0) Value can be empty or not set, see S.5.4.
// (S.6.1) Internal whitespaces within the value are retained verbatim.
// (S.6.2) Value can be continued to the next line by ending it with a `\`;
// the backquote and the end-of-line are stripped.
// (S.6.3) Leading and trailing.whitespaces on value without double quote will
// be discarded.
// (S.6.4) Value can contain inline comment, e.g.
//
//		key = value # this is inline comment
//
// (S.6.5) Comment characters, '#' and ';', inside double quoted value will be
// read as content of value, not as comment,
//
//		key = "value # with hash"
//
// (S.6.6) Inside value enclosed double quotes, the following escape sequences
// are recognized: `\"` for doublequote, `\\` for backslash, `\n` for newline
// character (NL), `\t` for horizontal tabulation (HT, TAB) and `\b` for
// backspace (BS).
// (S.6.8) Other char escape sequences (including octal escape sequences) are
// invalid.
//
// [1] https://git-scm.com/docs/git-config#_configuration_file
//
package ini

import (
	"bytes"
	"fmt"
	"strings"
)

//
// Ini contains the parsed file.
//
type Ini struct {
	secs []*section
}

//
// Open will open and parse INI formatted `file` and return it as instance of
// ini struct.
//
// On success it will return instance of ini without error.
// On fail it may return incomplete instance of ini with error.
//
func Open(filename string) (in *Ini, err error) {
	in = &Ini{}
	reader := NewReader()

	err = reader.ParseFile(in, filename)

	if debug >= debugL1 {
		for x := 0; x < len(in.secs); x++ {
			switch in.secs[x].m {
			case sectionModeNormal:
				fmt.Printf("[%s]\n", in.secs[x].name)

			case sectionModeSub:
				fmt.Printf("[%s \"%s\"]\n", in.secs[x].name,
					in.secs[x].subName)
			}

			for _, v := range in.secs[x].vars {
				switch v.m {
				case varModeNewline:
					fmt.Println()
				case varModeComment:
					fmt.Printf("%s\n", v.c)
				case varModeNormal:
					fmt.Printf("\t%s=%s%s\n", v.k, v.v, v.c)
				}
			}
		}
	}

	return
}

//
// Reset will clear all parsed data. This function can be used if you want to
// reuse the same Ini instance for parsing different Ini content.
//
func (in *Ini) Reset() {
	in.secs = nil
}

//
// Get will return the last key on section or subsection (if not empty).
// It will return nil and false,
// (1) If Ini file contains no sections,
// (2) section or key parameter is empty, or
// (3) no key found.
//
// Otherwise it will return key's value and true.
//
func (in *Ini) Get(section, subsection, key string) (val []byte, ok bool) {
	// (1) (2)
	if len(in.secs) == 0 || len(section) == 0 || len(key) == 0 {
		return
	}

	x := len(in.secs) - 1
	bsec := []byte(strings.ToLower(section))
	bsub := []byte(subsection)
	bkey := []byte(key)

	for ; x >= 0; x-- {
		if !bytes.Equal(in.secs[x].name, bsec) {
			continue
		}

		if !bytes.Equal(in.secs[x].subName, bsub) {
			continue
		}

		val, ok = in.secs[x].Get(bkey)
		if ok {
			return
		}
	}

	// (3)
	val = nil
	return
}
