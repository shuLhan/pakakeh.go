// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

var expectation = []string{
	"&[1 A-B AB 1 0.1]",
	"&[2 A-B-C BCD 2 0.02]",
	"&[3 A;B-C,D A;B C,D 3 0.003]",
	"&[4 A;B-C,D A;B C D 4 0.0004]",
	"&[6   6 0.000006]",
	"&[8 ok missing right-quote];8;0.00000008\n9;\"ok\"-[[ok 9 0.000000009]",
	"&[10 test integer 10 0.101]",
	"&[12 test real 123456789 0.123456789]",
	"&[13 string with string with 13 13]",
	"&[14 string with\" quote string with]] escape 14 14]",
}

var expSkip = []string{
	"&[A-B AB 1 0.1]",
	"&[A-B-C BCD 2 0.02]",
	"&[A;B-C,D A;B C,D 3 0.003]",
	"&[A;B-C,D A;B C D 4 0.0004]",
	"&[  6 0.000006]",
	"&[ok missing right-quote];8;0.00000008\n9;\"ok\"-[[ok 9 0.000000009]",
	"&[test integer 10 0.101]",
	"&[test real 123456789 0.123456789]",
	"&[string with string with 13 13]",
	"&[string with\" quote string with]] escape 14 14]",
}

var expSkipColumns = []string{
	"[{name 0 0 [] [A-B]} {value 0 0 [] [AB]} {integer 1 0 [] [1]} {real 2 0 [] [0.1]}]",
	"[{name 0 0 [] [A-B-C]} {value 0 0 [] [BCD]} {integer 1 0 [] [2]} {real 2 0 [] [0.02]}]",
	"[{name 0 0 [] [A;B-C,D]} {value 0 0 [] [A;B C,D]} {integer 1 0 [] [3]} {real 2 0 [] [0.003]}]",
	"[{name 0 0 [] [A;B-C,D]} {value 0 0 [] [A;B C D]} {integer 1 0 [] [4]} {real 2 0 [] [0.0004]}]",
	"[{name 0 0 [] []} {value 0 0 [] []} {integer 1 0 [] [6]} {real 2 0 [] [0.000006]}]",
	"[{name 0 0 [] [ok]} {value 0 0 [] [missing right-quote];8;0.00000008\n9;\"ok\"-[[ok]} {integer 1 0 [] [9]} {real 2 0 [] [0.000000009]}]",
	"[{name 0 0 [] [test]} {value 0 0 [] [integer]} {integer 1 0 [] [10]} {real 2 0 [] [0.101]}]",
	"[{name 0 0 [] [test]} {value 0 0 [] [real]} {integer 1 0 [] [123456789]} {real 2 0 [] [0.123456789]}]",
	"[{name 0 0 [] [string with]} {value 0 0 [] [string with]} {integer 1 0 [] [13]} {real 2 0 [] [13]}]",
	"[{name 0 0 [] [string with\" quote]} {value 0 0 [] [string with]] escape]} {integer 1 0 [] [14]} {real 2 0 [] [14]}]",
}

var expSkipColumnsAll = []string{
	"{name 0 0 [] [A-B A-B-C A;B-C,D A;B-C,D  ok test test string with string with\" quote]}",
	"{value 0 0 [] [AB BCD A;B C,D A;B C D  missing right-quote];8;0.00000008\n9;\"ok\"-[[ok integer real string with string with]] escape]}",
	"{integer 1 0 [] [1 2 3 4 6 9 10 123456789 13 14]}",
	"{real 2 0 [] [0.1 0.02 0.003 0.0004 0.000006 0.000000009 0.101 0.123456789 13 14]}",
}

var expSkipColumnsAllRev = []string{
	"{name 0 0 [] [string with\" quote string with test test ok  A;B-C,D A;B-C,D A-B-C A-B]}",
	"{value 0 0 [] [string with]] escape string with real integer missing right-quote];8;0.00000008\n9;\"ok\"-[[ok  A;B C D A;B C,D BCD AB]}",
	"{integer 1 0 [] [14 13 123456789 10 9 6 4 3 2 1]}",
	"{real 2 0 [] [14 13 0.123456789 0.101 0.000000009 0.000006 0.0004 0.003 0.02 0.1]}",
}
