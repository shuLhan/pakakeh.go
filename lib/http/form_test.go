package http

import (
	"net/url"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

type X struct {
	Int int `form:"int"`
}

type Y struct {
	String string `form:"string"`
	X
}

type Z struct {
	Y
	Bool bool `form:"bool"`
}

func TestMarshalForm(t *testing.T) {
	var (
		astruct = Z{
			Y: Y{
				X: X{
					Int: 1000,
				},
				String: `string in Y`,
			},
			Bool: true,
		}
		exp = url.Values{
			`int`:    []string{`1000`},
			`string`: []string{`string in Y`},
			`bool`:   []string{`true`},
		}

		got url.Values
		err error
	)

	got, err = MarshalForm(astruct)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, `Embedded`, exp, got)
}

func TestUnmarshalForm(t *testing.T) {
	var (
		form = url.Values{}
		exp  = Z{
			Y: Y{
				X: X{
					Int: 1000,
				},
				String: `string in Y`,
			},
			Bool: true,
		}

		got Z
		err error
	)

	form.Set(`int`, `1000`)
	form.Set(`string`, `string in Y`)
	form.Set(`bool`, `1`)

	err = UnmarshalForm(form, &got)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, `Embedded`, exp, got)
}
