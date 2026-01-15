// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package ascii

import (
	"testing"
)

func TestRandom(t *testing.T) {
	var (
		n = 5

		got  []byte
		gotn int
	)

	got = Random([]byte(Letters), n)
	gotn = len(got)
	if gotn != n {
		t.Fatalf(`Random 5 Letters: expecting %d characters, got %d`, n, gotn)
	}

	got = Random([]byte(LettersNumber), 5)
	gotn = len(got)
	if gotn != n {
		t.Fatalf(`Random 5 LettersNumber: expecting %d characters, got %d`, n, gotn)
	}

	got = Random([]byte(HexaLETTERS), 5)
	gotn = len(got)
	if gotn != n {
		t.Fatalf(`Random 5 HexaLETTERS: expecting %d characters, got %d`, n, gotn)
	}

	got = Random([]byte(HexaLetters), 5)
	gotn = len(got)
	if gotn != n {
		t.Fatalf(`Random 5 HexaLetters: expecting %d characters, got %d`, n, gotn)
	}

	got = Random([]byte(Hexaletters), 5)
	gotn = len(got)
	if gotn != n {
		t.Fatalf(`Random 5 Hexaletters: expecting %d characters, got %d`, n, gotn)
	}

	got = Random([]byte(`01`), 5)
	gotn = len(got)
	if gotn != n {
		t.Fatalf(`Random 5 binary: expecting %d characters, got %d`, n, gotn)
	}
}
