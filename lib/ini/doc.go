// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package ini implement reading and writing INI configuration as defined by
// Git configuration file syntax.
//
// Features
//
// *  Reading and writing on the same file should not change the content of
// file (including comment).
//
// *  Template friendly, through Val(), Vals(), and Subs().
//
// Unsupported features
//
// Git `include` and `includeIf` directives.
//
// In Git specification, an empty variable is equal to boolean true.  This
// cause inconsistency between empty string and boolean true.
//
// Syntax
//
// S.1.0.  The `#` and `;` characters begin comments to the end of line.
//
// S.1.1.  Blank lines are ignored.
//
// ## Section
//
// S.2.0.  A section begins with the name of the section in square brackets.
//
// S.2.1.  A section continues until the next section begins.
//
// S.2.2.  Section name are case-insensitive.
//
// S.2.3.  Variable name must start with an alphabetic character, no
// whitespace before name or after '['.
//
// S.2.4.  Section name only allow alphanumeric characters, `-` and `.`.
//
// S.2.5.  Section can be further divided into subsections.
//
// S.2.6.  Section headers cannot span multiple lines.
//
// S.2.7.  You can have `[section]` if you have `[section "subsection"]`, but
// you don’t need to.
//
// S.2.8.  All the other lines (and the remainder of the line after the
// section header) are recognized as setting variables, in the form
// `name = value`.
//
// ## SubSection
//
// S.3.0.  To begin a subsection put its name in double quotes, separated by
// space from the section name, in the section header, for example
//
//	[section "subsection"]
//
// S.3.1.  Subsection name are case sensitive and can contain any characters
// except newline and the null byte.
//
// S.3.2.  Subsection name can include doublequote `"` and backslash by
// escaping them as `\"` and `\\`, respectively.
//
// S.3.3.  Other backslashes preceding other characters are dropped when
// reading subsection name; for example, `\t` is read as `t` and `\0` is read
// as `0`.
//
// ## Variable
//
// S.4.0.  Variable must belong to some section, which means that there
// must be a section header before the first setting of a variable.
//
// S.4.1.  Variable name are case-insensitive.
//
// S.4.2.  Variable name allow only alphanumeric characters and `-`.
//
// S.4.3.  Variable name must start with an alphabetic character.
//
// ## Value
//
// S.5.0.  Value can be empty or not set, see E.4.1.
//
// S.5.1.  Internal whitespaces within the value are retained verbatim.
//
// S.5.2.  Value can be continued to the next line by ending it with a `\`;
// the backquote and the end-of-line are stripped.
//
// S.5.3.  Leading and trailing.whitespaces on value without double quote will
// be discarded.
//
// S.5.4.  Value can contain inline comment, e.g.
//
//	key = value # this is inline comment
//
// S.5.5.  Comment characters, '#' and ';', inside double quoted value will be
// read as content of value, not as comment,
//
//	key = "value # with hash"
//
// S.5.6.  Inside value enclosed double quotes, the following escape sequences
// are recognized: `\"` for doublequote, `\\` for backslash, `\n` for newline
// character (NL), `\t` for horizontal tabulation (HT, TAB) and `\b` for
// backspace (BS).
//
// S.5.7.  Other char escape sequences (including octal escape sequences) are
// invalid.
//
// Extensions
//
// ## Variable
//
// E.4.0.  Allow dot ('.') and underscore ('_') characters on variable name.
//
// E.4.1.  Variable name without value is a short-hand to set the value to the
// empty string value, e.g.
//
//	[section]
//		thisisempty # equal to thisisempty=
//
// References
//
// https://git-scm.com/docs/git-config#_configuration_file
//
package ini
