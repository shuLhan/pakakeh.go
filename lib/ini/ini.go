//
// Package ini implement reading and writing INI configuration as defined by
// Git configuration file syntax [1].
//
// Syntax,
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
// --
// [1] https://git-scm.com/docs/git-config#_configuration_file
//
package ini

import (
	"fmt"
)

//
// Ini contains the parsed file.
//
type Ini struct {
	Filename string
	secs     []*section
}

//
// New will create, initialize default section, and return a new instance
// of ini.
//
func New(file string) (in *Ini) {
	in = &Ini{
		Filename: file,
	}

	return
}

//
// Open will open and parse INI formatted `file` and return it as instance of
// ini struct.
//
// On success it will return instance of ini without error.
// On fail it may return incomplete instance of ini with error.
//
func Open(filename string) (in *Ini, err error) {
	reader, err := NewReader(filename)
	if err != nil {
		return
	}

	in = New(filename)

	err = reader.parse(in)

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
// Save will save the parsed INI values back into the same file as input.
//
func (in *Ini) Save() (err error) {
	return
}
