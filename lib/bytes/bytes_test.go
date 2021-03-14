package bytes

import (
	"testing"

	"github.com/shuLhan/share/lib/ascii"
	"github.com/shuLhan/share/lib/test"
)

func TestConcat(t *testing.T) {
	var exp []byte
	t.Log("With one parameter")
	got := Concat([]byte{})
	test.Assert(t, "Concat", exp, got)

	t.Log("With first parameter is empty")
	got = Concat([]byte{}, []byte("B"))
	exp = []byte("B")
	test.Assert(t, "Concat", exp, got)

	t.Log("With two parameters")
	got = Concat([]byte("A"), []byte("B"))
	exp = []byte("AB")
	test.Assert(t, "Concat", exp, got)

	t.Log("With three parameters")
	got = Concat([]byte("A"), []byte("B"), []byte("C"))
	exp = []byte("ABC")
	test.Assert(t, "Concat", exp, got)

	t.Log("With one parameter is string")
	got = Concat([]byte("A"), "B", []byte("C"))
	exp = []byte("ABC")
	test.Assert(t, "Concat", exp, got)

	t.Log("With some parameter is not []byte or string")
	got = Concat([]byte("A"), 1, []int{2}, []byte{}, []byte("C"))
	exp = []byte("AC")
	test.Assert(t, "Concat", exp, got)
}

func TestCutUntilToken(t *testing.T) {
	line := []byte(`abc \def ghi`)

	cases := []struct {
		token []byte
		exp   string

		startAt int
		expIdx  int

		expFound bool
		checkEsc bool
	}{{
		exp:      `abc \def ghi`,
		expIdx:   -1,
		expFound: false,
	}, {
		token:    []byte(`def`),
		exp:      `abc \`,
		expIdx:   8,
		expFound: true,
	}, {
		token:    []byte(`def`),
		checkEsc: true,
		exp:      `abc def ghi`,
		expIdx:   12,
		expFound: false,
	}, {
		token:    []byte(`ef`),
		checkEsc: true,
		exp:      `abc \d`,
		expIdx:   8,
		expFound: true,
	}}

	for x, c := range cases {
		t.Logf("#%d\n", x)

		got, idx, found := CutUntilToken(line, c.token, c.startAt, c.checkEsc)

		test.Assert(t, "cut", c.exp, string(got))
		test.Assert(t, "idx", c.expIdx, idx)
		test.Assert(t, "found", c.expFound, found)
	}
}

func TestEncloseRemove(t *testing.T) {
	line := []byte(`// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`)

	cases := []struct {
		line     []byte
		leftcap  []byte
		rightcap []byte
		exp      string
	}{{
		line:     line,
		leftcap:  []byte("<"),
		rightcap: []byte(">"),
		exp:      `// Copyright 2016-2018 "Shulhan ". All rights reserved.`,
	}, {
		line:     line,
		leftcap:  []byte(`"`),
		rightcap: []byte(`"`),
		exp:      `// Copyright 2016-2018 . All rights reserved.`,
	}, {
		line:     line,
		leftcap:  []byte(`/`),
		rightcap: []byte(`/`),
		exp:      ` Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`,
	}, {
		line:     []byte(`/* TEST */`),
		leftcap:  []byte(`/*`),
		rightcap: []byte(`*/`),
		exp:      "",
	}}

	for _, c := range cases {
		got, _ := EncloseRemove(c.line, c.leftcap, c.rightcap)

		test.Assert(t, "", c.exp, string(got))
	}
}

func TestEncloseToken(t *testing.T) {
	line := []byte(`// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`)

	cases := []struct {
		token, leftcap, rightcap []byte
		exp                      string
		changed                  bool
	}{{
		token:    []byte(`_`),
		leftcap:  []byte(`-`),
		rightcap: []byte(`-`),
		exp:      `// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`,
		changed:  false,
	}, {
		token:    []byte(`/`),
		leftcap:  []byte(`\`),
		rightcap: []byte{},
		exp:      `\/\/ Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`,
		changed:  true,
	}, {
		token:    []byte(`<`),
		leftcap:  []byte(`<`),
		rightcap: []byte(` `),
		exp:      `// Copyright 2016-2018 "Shulhan << ms@kilabit.info>". All rights reserved.`,
		changed:  true,
	}, {
		token:    []byte(`"`),
		leftcap:  []byte(`\`),
		rightcap: []byte(` `),
		exp:      `// Copyright 2016-2018 \" Shulhan <ms@kilabit.info>\" . All rights reserved.`,
		changed:  true,
	}}

	for _, c := range cases {
		got, changed := EncloseToken(line, c.token, c.leftcap, c.rightcap)

		test.Assert(t, "newline", c.exp, string(got))
		test.Assert(t, "changed", c.changed, changed)
	}
}

func TestIsTokenAt(t *testing.T) {
	line := []byte("Hello, world")

	cases := []struct {
		token []byte
		p     int
		exp   bool
	}{{
		token: nil,
	}, {
		token: []byte("world"),
		p:     -1,
	}, {
		token: []byte("world"),
		p:     6,
	}, {
		token: []byte("world"),
		p:     7,
		exp:   true,
	}, {
		token: []byte("world"),
		p:     8,
	}, {
		token: []byte("worlds"),
		p:     7,
	}}

	for _, c := range cases {
		got := IsTokenAt(line, c.token, c.p)
		test.Assert(t, "IsTokenAt", c.exp, got)
	}
}

func TestReadHexByte(t *testing.T) {
	cases := []struct {
		in    []byte
		exp   byte
		expOK bool
	}{{
		in: []byte{},
	}, {
		in: []byte("x0"),
	}, {
		in: []byte("0x"),
	}, {
		in:    []byte("00"),
		expOK: true,
	}, {
		in:    []byte("01"),
		exp:   1,
		expOK: true,
	}, {
		in:    []byte("10"),
		exp:   16,
		expOK: true,
	}, {
		in:    []byte("1A"),
		exp:   26,
		expOK: true,
	}, {
		in:    []byte("1a"),
		exp:   26,
		expOK: true,
	}, {
		in:    []byte("a1"),
		exp:   161,
		expOK: true,
	}}

	for _, c := range cases {
		t.Log(c.in)

		got, ok := ReadHexByte(c.in, 0)

		test.Assert(t, "b", c.exp, got)
		test.Assert(t, "ok", c.expOK, ok)
	}
}

func TestMergeSpaces(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in: "",
	}, {
		in:  " \t\v\r\n\r\n\fa \t\v\r\n\r\n\f",
		exp: " a ",
	}}
	for _, c := range cases {
		got := MergeSpaces([]byte(c.in))
		test.Assert(t, c.in, c.exp, string(got))
	}
}

func TestSkipAfterToken(t *testing.T) {
	line := []byte(`abc \def ghi`)

	cases := []struct {
		token []byte

		startAt int
		exp     int

		checkEsc bool
		expFound bool
	}{{
		token:    []byte(`def`),
		exp:      8,
		expFound: true,
	}, {
		token:    []byte(`def`),
		checkEsc: true,
		exp:      12,
	}, {
		token:    []byte(`ef`),
		checkEsc: true,
		exp:      8,
		expFound: true,
	}, {
		token:    []byte(`hi`),
		exp:      len(line),
		expFound: true,
	}}

	for x, c := range cases {
		t.Logf("#%d\n", x)
		got, found := SkipAfterToken(line, c.token, c.startAt, c.checkEsc)
		test.Assert(t, "Index", c.exp, got)
		test.Assert(t, "Found", c.expFound, found)
	}
}

func testTokenFind(t *testing.T, line, token []byte, startat int, exp []int) {
	got := []int{}
	tokenlen := len(token)

	for {
		foundat := TokenFind(line, token, startat)

		if foundat < 0 {
			break
		}

		got = append(got, foundat)
		startat = foundat + tokenlen
	}

	test.Assert(t, "TokenFind", exp, got)
}

func TestTokenFind(t *testing.T) {
	line := []byte("// Copyright 2016-2018 Shulhan <ms@kilabit.info>. All rights reserved.")

	token := []byte("//")
	exp := []int{0}

	testTokenFind(t, line, token, 0, exp)

	token = []byte(".")
	exp = []int{42, 48, 69}

	testTokenFind(t, line, token, 0, exp)

	token = []byte("d.")
	exp = []int{68}

	testTokenFind(t, line, token, 0, exp)
}

func TestInReplace(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in:  "/a/path/to/file.ext",
		exp: "_a_path_to_file_ext",
	}}

	for _, c := range cases {
		got := InReplace([]byte(c.in), []byte(ascii.LettersNumber), '_')

		test.Assert(t, "InReplace", c.exp, string(got))
	}
}

func TestIndexes(t *testing.T) {
	cases := []struct {
		desc  string
		s     []byte
		token []byte
		exp   []int
	}{{
		desc:  "With empty string",
		token: []byte("moo"),
	}, {
		desc: "With empty token",
		s:    []byte("moo moo"),
	}, {
		desc:  "With non empty string and token",
		s:     []byte("moo moomoo"),
		token: []byte("moo"),
		exp:   []int{0, 4, 7},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := Indexes(c.s, c.token)

		test.Assert(t, "Indexes", c.exp, got)
	}
}
