// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package ini implement reading and writing INI text format as defined by
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
// Git "include" and "includeIf" directives.
//
// In Git specification, an empty variable is equal to boolean true.  This
// cause inconsistency between empty string and boolean true.
//
// Syntax
//
// The '#' and ';' characters begin comments to the end of line.
//
// Blank lines are ignored.
//
// Section
//
// A section begins with the name of the section in square brackets.
//
// A section continues until the next section begins.
//
// Section name are case-insensitive.
//
// Variable name must start with an alphabetic character, no
// whitespace before name or after '['.
//
// Section name only allow alphanumeric characters, `-` and `.`.
//
// Section can be further divided into subsections.
//
// Section headers cannot span multiple lines.
//
// You can have `[section]` if you have `[section "subsection"]`, but
// you don’t need to.
//
// All the other lines (and the remainder of the line after the
// section header) are recognized as setting variables, in the form
// `name = value`.
//
// Subsection
//
// To begin a subsection put its name in double quotes, separated by
// space from the section name, in the section header, for example
//
//	[section "subsection"]
//
// Subsection name are case sensitive and can contain any characters
// except newline and the null byte.
//
// Subsection name can include doublequote `"` and backslash by
// escaping them as `\"` and `\\`, respectively.
//
// Other backslashes preceding other characters are dropped when
// reading subsection name; for example, `\t` is read as `t` and `\0` is read
// as `0`.
//
// Variable
//
// Variable name must start with an alphabetic character.
//
// Variable must belong to some section, which means that there
// must be a section header before the first setting of a variable.
//
// Variable name are case-insensitive.
//
// Variable name allow only alphanumeric characters and `-`.
// This ini library add extension to allow dot ('.') and underscore ('_')
// characters on variable name.
//
// Value
//
// Value can be empty or not set.
// (EXT) Variable name without value is a short-hand to set the value to the
// empty string value, for example
//
//	[section]
//		thisisempty # equal to thisisempty=
//
//
// Internal whitespaces within the value are retained verbatim.
// Leading and trailing whitespaces on value without double quote will
// be discarded.
//
// 	key = multiple strings     # equal to "multiple strings"
// 	key = " multiple strings " # equal to " multiple strings "
//
// Value can be continued to the next line by ending it with a backslash '\'
// character, the backquote and the end-of-line are stripped.
//
//	key = multiple \           # equal to "multiple string"
//	strings
//
// Value can contain inline comment, for example
//
//	key = value # this is inline comment
//
// Comment characters, '#' and ';', inside double quoted value will be
// read as content of value, not as comment,
//
//	key = "value # with hash"
//
// Inside value enclosed double quotes, the following escape sequences
// are recognized: `\"` for doublequote, `\\` for backslash, `\n` for newline
// character (NL), `\t` for horizontal tabulation (HT, TAB) and `\b` for
// backspace (BS).
//
// Other char escape sequences (including octal escape sequences) are
// invalid.
//
// Marshaling
//
// The container to be passed when marshaling must be struct type.
// Each exported field in the struct with "ini" tags will be marshaled based
// on the section, subsection, and key in the tag.
//
// If the field type is slice of primitive, for example "[]int", it will be
// marshaled into multiple key with the same name.
//
// If the field type is struct, it will marshaled as new section and/or
// subsection based on tag on the struct field
//
// If the field type is slice of struct, it will marshaled as multiple
// section-subsection with the same tags.
//
// Map type is supported as long as the key is string, otherwise it will be
// ignored.
// The map key will be marshaled as key.
//
// Other standard type that supported is time.Time, which will be rendered
// with the time format defined in "layout" tag.
//
// Example,
//
//	type U struct {
//		Int `ini:"::int"`
//	}
//
//	type T struct {
//		String      string            `ini:"single::string"
//		Time        time.Time         `ini:"single::time" layout:"2006-01-02"`
//		SliceString []string          `ini:"slice::string"
//		Struct      U                 `ini:"single:struct"
//		SliceStruct []U               `ini:"slice:struct"
//		Map         map[string]int    `ini:"amap:"
//		MapSub      map[string]string `ini:"amap:sub"
//	}
//
// will be marshaled into
//
//	[single]
//	string = <value of T.String>
//	time = <value of T.Time with layout "YYYY-MM-DD">
//
//	[slice]
//	string = <value of T.SliceStruct[0]>
//	...
//	string = <value of T.SliceStruct[n]>
//
//	[single "struct"]
//	int = <value of T.U.Int>
//
//	[slice "struct"]
//	int = <value of T.SliceStruct[0].Int
//
//	[slice "struct"]
//	int = <value of T.SliceStruct[n].Int
//
//	[amap]
//	<T.Map.Key[0]> = <T.Map.Value[0]>
//	...
//	<T.Map.Key[n]> = <T.Map.Value[n]>
//
//	[amap "sub"]
//	<T.MapSub.Key[0]> = <T.MapSub.Value[0]>
//	...
//	<T.MapSub.Key[n]> = <T.MapSub.Value[n]>
//
// Unmarshaling
//
// The syntax and rules for unmarshaling is equal to the marshaling.
//
// References
//
// https://git-scm.com/docs/git-config#_configuration_file
//
package ini
