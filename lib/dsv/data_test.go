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
	`[{name [] [A-B] 0 0} {value [] [AB] 0 0} {integer [] [1] 1 0} {real [] [0.1] 2 0}]`,
	"[{name [] [A-B-C] 0 0} {value [] [BCD] 0 0} {integer [] [2] 1 0} {real [] [0.02] 2 0}]",
	"[{name [] [A;B-C,D] 0 0} {value [] [A;B C,D] 0 0} {integer [] [3] 1 0} {real [] [0.003] 2 0}]",
	"[{name [] [A;B-C,D] 0 0} {value [] [A;B C D] 0 0} {integer [] [4] 1 0} {real [] [0.0004] 2 0}]",
	"[{name [] [] 0 0} {value [] [] 0 0} {integer [] [6] 1 0} {real [] [0.000006] 2 0}]",
	"[{name [] [ok] 0 0} {value [] [missing right-quote];8;0.00000008\n9;\"ok\"-[[ok] 0 0} {integer [] [9] 1 0} {real [] [0.000000009] 2 0}]",
	"[{name [] [test] 0 0} {value [] [integer] 0 0} {integer [] [10] 1 0} {real [] [0.101] 2 0}]",
	"[{name [] [test] 0 0} {value [] [real] 0 0} {integer [] [123456789] 1 0} {real [] [0.123456789] 2 0}]",
	"[{name [] [string with] 0 0} {value [] [string with] 0 0} {integer [] [13] 1 0} {real [] [13] 2 0}]",
	"[{name [] [string with\" quote] 0 0} {value [] [string with]] escape] 0 0} {integer [] [14] 1 0} {real [] [14] 2 0}]",
}

var expSkipColumnsAll = []string{
	"{name [] [A-B A-B-C A;B-C,D A;B-C,D  ok test test string with string with\" quote] 0 0}",
	"{value [] [AB BCD A;B C,D A;B C D  missing right-quote];8;0.00000008\n9;\"ok\"-[[ok integer real string with string with]] escape] 0 0}",
	"{integer [] [1 2 3 4 6 9 10 123456789 13 14] 1 0}",
	"{real [] [0.1 0.02 0.003 0.0004 0.000006 0.000000009 0.101 0.123456789 13 14] 2 0}",
}

var expSkipColumnsAllRev = []string{
	"{name [] [string with\" quote string with test test ok  A;B-C,D A;B-C,D A-B-C A-B] 0 0}",
	"{value [] [string with]] escape string with real integer missing right-quote];8;0.00000008\n9;\"ok\"-[[ok  A;B C D A;B C,D BCD AB] 0 0}",
	"{integer [] [14 13 123456789 10 9 6 4 3 2 1] 1 0}",
	"{real [] [14 13 0.123456789 0.101 0.000000009 0.000006 0.0004 0.003 0.02 0.1] 2 0}",
}
