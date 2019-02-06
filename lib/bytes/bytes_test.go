package bytes

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

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

		test.Assert(t, "cut", c.exp, string(got), true)
		test.Assert(t, "idx", c.expIdx, idx, true)
		test.Assert(t, "found", c.expFound, found, true)
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

		test.Assert(t, "", c.exp, string(got), true)
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

		test.Assert(t, "newline", c.exp, string(got), true)
		test.Assert(t, "changed", c.changed, changed, true)
	}
}

func TestIsTokenAt(t *testing.T) {
	line := []byte("Hello, world")

	cases := []struct {
		token []byte
		p     int
		exp   bool
	}{{
		// empty
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
		test.Assert(t, "IsTokenAt", c.exp, got, true)
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

		test.Assert(t, "b", c.exp, got, true)
		test.Assert(t, "ok", c.expOK, ok, true)
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
		test.Assert(t, "Index", c.exp, got, true)
		test.Assert(t, "Found", c.expFound, found, true)
	}
}

func TestToLower(t *testing.T) {
	cases := []struct {
		in  []byte
		exp []byte
	}{{
		in:  []byte("@ABCDEFG"),
		exp: []byte("@abcdefg"),
	}, {
		in:  []byte("@ABCDEFG12345678"),
		exp: []byte("@abcdefg12345678"),
	}, {
		in:  []byte("@ABCDEFGhijklmno12345678"),
		exp: []byte("@abcdefghijklmno12345678"),
	}, {
		in:  []byte("@ABCDEFGhijklmnoPQRSTUVW12345678"),
		exp: []byte("@abcdefghijklmnopqrstuvw12345678"),
	}, {
		in:  []byte("@ABCDEFGhijklmnoPQRSTUVWxyz{12345678"),
		exp: []byte("@abcdefghijklmnopqrstuvwxyz{12345678"),
	}}

	for _, c := range cases {
		ToLower(&c.in)
		test.Assert(t, "ToLower", c.exp, c.in, true)
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

	test.Assert(t, "TokenFind", exp, got, true)
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

func BenchmarkToLowerStd(b *testing.B) {
	randomInput256 := Random([]byte(HexaLetters), 256)

	in := make([]byte, len(randomInput256))
	copy(in, randomInput256)

	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		bytes.ToLower(in)
	}
}

func BenchmarkToLower(b *testing.B) {
	randomInput256 := Random([]byte(HexaLetters), 256)

	in := make([]byte, len(randomInput256))
	copy(in, randomInput256)

	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		ToLower(&in)
		copy(in, randomInput256)
	}
}
