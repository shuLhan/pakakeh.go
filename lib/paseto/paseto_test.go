// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package paseto

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestPae(t *testing.T) {
	cases := []struct {
		pieces [][]byte
		exp    []byte
	}{{
		exp: []byte("\x00\x00\x00\x00\x00\x00\x00\x00"),
	}, {
		pieces: [][]byte{[]byte{}},
		exp:    []byte("\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
	}, {
		pieces: [][]byte{{}, {}},
		exp:    []byte("\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
	}, {
		pieces: [][]byte{[]byte("test")},
		exp:    []byte("\x01\x00\x00\x00\x00\x00\x00\x00\x04\x00\x00\x00\x00\x00\x00\x00test"),
	}, {
		pieces: [][]byte{[]byte("Paragon")},
		exp:    []byte("\x01\x00\x00\x00\x00\x00\x00\x00\x07\x00\x00\x00\x00\x00\x00\x00\x50\x61\x72\x61\x67\x6f\x6e"),
	}}

	for _, c := range cases {
		got, err := pae(c.pieces)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "pae", c.exp, got)
	}
}

func TestEncrypt(t *testing.T) {
	hexKey := "70717273" + "74757677" + "78797a7b" + "7c7d7e7f" +
		"80818283" + "84858687" + "88898a8b" + "8c8d8e8f"

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		t.Fatal(err)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc  string
		nonce string
		exp   string

		msg    []byte
		footer []byte
	}{{
		desc: "Encrypt with zero nonce, without footer",
		msg:  []byte(`{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`),
		nonce: "00000000" + "00000000" + "00000000" + "00000000" +
			"00000000" + "00000000",
		exp: "v2.local.97TTOvgwIxNGvV80XKiGZg_kD3tsXM_-qB4dZGHOeN1cTkgQ4Pn" +
			"W8888l802W8d9AvEGnoNBY3BnqHORy8a5cC8aKpbA0En8XELw2yDk2f1sVOD" +
			"yfnDbi6rEGMY3pSfCbLWMM2oHJxvlEl2XbQ",
	}, {
		desc: "Encrypt with zero nonce, without footer (2)",
		msg:  []byte(`{"data":"this is a secret message","exp":"2019-01-01T00:00:00+00:00"}`),
		nonce: "00000000" + "00000000" + "00000000" + "00000000" +
			"00000000" + "00000000",
		exp: "v2.local.CH50H-HM5tzdK4kOmQ8KbIvrzJfjYUGuu5Vy9ARSFHy9owVDMYg" +
			"3-8rwtJZQjN9ABHb2njzFkvpr5cOYuRyt7CRXnHt42L5yZ7siD-4l-FoNsC7" +
			"J2OlvLlIwlG06mzQVunrFNb7Z3_CHM0PK5w",
	}, {
		desc: "Encrypt with nonce, without footer",
		msg:  []byte(`{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`),
		nonce: "45742c97" + "6d684ff8" + "4ebdc0de" + "59809a97" +
			"cda2f64c" + "84fda19b",
		exp: "v2.local.5K4SCXNhItIhyNuVIZcwrdtaDKiyF81-eWHScuE0idiVqCo72bb" +
			"jo07W05mqQkhLZdVbxEa5I_u5sgVk1QLkcWEcOSlLHwNpCkvmGGlbCdNExn6" +
			"Qclw3qTKIIl5-O5xRBN076fSDPo5xUCPpBA",
	}, {
		desc: "Encrypt with nonce, with footer",
		msg:  []byte(`{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`),
		nonce: "45742c97" + "6d684ff8" + "4ebdc0de" + "59809a97" +
			"cda2f64c" + "84fda19b",
		footer: []byte(`{"kid":"zVhMiPBP9fRf2snEcT7gFTioeA9COcNy9DfgL1W60haN"}`),
		exp: "v2.local.5K4SCXNhItIhyNuVIZcwrdtaDKiyF81-eWHScuE0idiVqCo72bb" +
			"jo07W05mqQkhLZdVbxEa5I_u5sgVk1QLkcWEcOSlLHwNpCkvmGGlbCdNExn6" +
			"Qclw3qTKIIl5-zSLIrxZqOLwcFLYbVK1SrQ.eyJraWQiOiJ6VmhNaVBCUDlm" +
			"UmYyc25FY1Q3Z0ZUaW9lQTlDT2NOeTlEZmdMMVc2MGhhTiJ9",
	}, {
		desc: "Encrypt with nonce, with footer (2)",
		msg:  []byte(`{"data":"this is a secret message","exp":"2019-01-01T00:00:00+00:00"}`),
		nonce: "45742c97" + "6d684ff8" + "4ebdc0de" + "59809a97" +
			"cda2f64c" + "84fda19b",
		footer: []byte(`{"kid":"zVhMiPBP9fRf2snEcT7gFTioeA9COcNy9DfgL1W60haN"}`),
		exp: "v2.local.pvFdDeNtXxknVPsbBCZF6MGedVhPm40SneExdClOxa9HNR8wFv7" +
			"cu1cB0B4WxDdT6oUc2toyLR6jA6sc-EUM5ll1EkeY47yYk6q8m1RCpqTIzUr" +
			"Iu3B6h232h62DnMXKdHn_Smp6L_NfaEnZ-A.eyJraWQiOiJ6VmhNaVBCUDlm" +
			"UmYyc25FY1Q3Z0ZUaW9lQTlDT2NOeTlEZmdMMVc2MGhhTiJ9",
	}}

	for _, c := range cases {
		nonce, err := hex.DecodeString(c.nonce)
		if err != nil {
			t.Fatal(err)
		}

		got, err := encrypt(aead, nonce, c.msg, c.footer)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.desc, c.exp, got)
	}
}

func TestDecrypt(t *testing.T) {
	hexKey := "70717273" + "74757677" + "78797a7b" + "7c7d7e7f" +
		"80818283" + "84858687" + "88898a8b" + "8c8d8e8f"

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		t.Fatal(err)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc      string
		token     string
		exp       []byte
		expFooter []byte
	}{{
		desc: "Decrypt without nonce and footer",
		token: "v2.local.97TTOvgwIxNGvV80XKiGZg_kD3tsXM_-qB4dZGHOeN1cTkgQ4Pn" +
			"W8888l802W8d9AvEGnoNBY3BnqHORy8a5cC8aKpbA0En8XELw2yDk2f1sVOD" +
			"yfnDbi6rEGMY3pSfCbLWMM2oHJxvlEl2XbQ",
		exp: []byte(`{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`),
	}, {
		desc: "Decrypt without nonce and footer (2)",
		token: "v2.local.CH50H-HM5tzdK4kOmQ8KbIvrzJfjYUGuu5Vy9ARSFHy9owVDMYg" +
			"3-8rwtJZQjN9ABHb2njzFkvpr5cOYuRyt7CRXnHt42L5yZ7siD-4l-FoNsC7" +
			"J2OlvLlIwlG06mzQVunrFNb7Z3_CHM0PK5w",
		exp: []byte(`{"data":"this is a secret message","exp":"2019-01-01T00:00:00+00:00"}`),
	}, {
		desc: "Decrypt with nonce, without footer",
		token: "v2.local.5K4SCXNhItIhyNuVIZcwrdtaDKiyF81-eWHScuE0idiVqCo72bb" +
			"jo07W05mqQkhLZdVbxEa5I_u5sgVk1QLkcWEcOSlLHwNpCkvmGGlbCdNExn6" +
			"Qclw3qTKIIl5-O5xRBN076fSDPo5xUCPpBA",
		exp: []byte(`{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`),
	}, {
		desc: "Decrypt with nonce, with footer",
		token: "v2.local.5K4SCXNhItIhyNuVIZcwrdtaDKiyF81-eWHScuE0idiVqCo72bb" +
			"jo07W05mqQkhLZdVbxEa5I_u5sgVk1QLkcWEcOSlLHwNpCkvmGGlbCdNExn6" +
			"Qclw3qTKIIl5-zSLIrxZqOLwcFLYbVK1SrQ.eyJraWQiOiJ6VmhNaVBCUDlm" +
			"UmYyc25FY1Q3Z0ZUaW9lQTlDT2NOeTlEZmdMMVc2MGhhTiJ9",
		exp:       []byte(`{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`),
		expFooter: []byte(`{"kid":"zVhMiPBP9fRf2snEcT7gFTioeA9COcNy9DfgL1W60haN"}`),
	}, {
		desc: "Decrypt with nonce, with footer (2)",
		token: "v2.local.pvFdDeNtXxknVPsbBCZF6MGedVhPm40SneExdClOxa9HNR8wFv7" +
			"cu1cB0B4WxDdT6oUc2toyLR6jA6sc-EUM5ll1EkeY47yYk6q8m1RCpqTIzUr" +
			"Iu3B6h232h62DnMXKdHn_Smp6L_NfaEnZ-A.eyJraWQiOiJ6VmhNaVBCUDlm" +
			"UmYyc25FY1Q3Z0ZUaW9lQTlDT2NOeTlEZmdMMVc2MGhhTiJ9",
		exp:       []byte(`{"data":"this is a secret message","exp":"2019-01-01T00:00:00+00:00"}`),
		expFooter: []byte(`{"kid":"zVhMiPBP9fRf2snEcT7gFTioeA9COcNy9DfgL1W60haN"}`),
	}}

	for _, c := range cases {
		got, gotFooter, err := Decrypt(aead, c.token)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.desc, c.exp, got)
		test.Assert(t, c.desc, c.expFooter, gotFooter)
	}
}

func TestSign(t *testing.T) {
	hexPrivate := "b4cbfb43" + "df4ce210" + "727d953e" + "4a713307" +
		"fa19bb7d" + "9f850414" + "38d9e11b" + "942a3774" +
		"1eb9dbbb" + "bc047c03" + "fd70604e" + "0071f098" +
		"7e16b28b" + "757225c1" + "1f00415d" + "0e20b1a2"

	sk, err := hex.DecodeString(hexPrivate)
	if err != nil {
		t.Fatal()
	}

	m := []byte(`{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`)

	cases := []struct {
		desc string
		exp  string

		m []byte
		f []byte
	}{{
		desc: "Sign",
		m:    m,
		exp: "v2.public.eyJkYXRhIjoidGhpcyBpcyBhIHNpZ25lZCBtZXNzYWdlIi" +
			"wiZXhwIjoiMjAxOS0wMS0wMVQwMDowMDowMCswMDowMCJ9HQr8URrGnt" +
			"Tu7Dz9J2IF23d1M7-9lH9xiqdGyJNvzp4angPW5Esc7C5huy_M8I8_Dj" +
			"JK2ZXC2SUYuOFM-Q_5Cw",
	}, {
		desc: "Sign with footer",
		m:    m,
		f:    []byte(`{"kid":"zVhMiPBP9fRf2snEcT7gFTioeA9COcNy9DfgL1W60haN"}`),
		exp: "v2.public.eyJkYXRhIjoidGhpcyBpcyBhIHNpZ25lZCBtZXNzYWdlIi" +
			"wiZXhwIjoiMjAxOS0wMS0wMVQwMDowMDowMCswMDowMCJ9flsZsx_gYC" +
			"R0N_Ec2QxJFFpvQAs7h9HtKwbVK2n1MJ3Rz-hwe8KUqjnd8FAnIJZ601" +
			"tp7lGkguU63oGbomhoBw.eyJraWQiOiJ6VmhNaVBCUDlmUmYyc25FY1Q" +
			"3Z0ZUaW9lQTlDT2NOeTlEZmdMMVc2MGhhTiJ9",
	}}

	for _, c := range cases {
		got, err := Sign(sk, c.m, c.f)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.desc, c.exp, got)
	}
}

func TestVerify(t *testing.T) {
	hexPublic := "1eb9dbbb" + "bc047c03" + "fd70604e" + "0071f098" +
		"7e16b28b" + "757225c1" + "1f00415d" + "0e20b1a2"

	public, err := hex.DecodeString(hexPublic)
	if err != nil {
		t.Fatal()
	}

	cases := []struct {
		desc  string
		token string
		exp   string
	}{{
		desc: "Verify",
		token: "v2.public.eyJkYXRhIjoidGhpcyBpcyBhIHNpZ25lZCBtZXNzYWdlIi" +
			"wiZXhwIjoiMjAxOS0wMS0wMVQwMDowMDowMCswMDowMCJ9HQr8URrGnt" +
			"Tu7Dz9J2IF23d1M7-9lH9xiqdGyJNvzp4angPW5Esc7C5huy_M8I8_Dj" +
			"JK2ZXC2SUYuOFM-Q_5Cw",
		exp: `{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`,
	}, {
		desc: "Verify with footer",
		token: "v2.public.eyJkYXRhIjoidGhpcyBpcyBhIHNpZ25lZCBtZXNzYWdlIi" +
			"wiZXhwIjoiMjAxOS0wMS0wMVQwMDowMDowMCswMDowMCJ9flsZsx_gYC" +
			"R0N_Ec2QxJFFpvQAs7h9HtKwbVK2n1MJ3Rz-hwe8KUqjnd8FAnIJZ601" +
			"tp7lGkguU63oGbomhoBw.eyJraWQiOiJ6VmhNaVBCUDlmUmYyc25FY1Q" +
			"3Z0ZUaW9lQTlDT2NOeTlEZmdMMVc2MGhhTiJ9",
		exp: `{"data":"this is a signed message","exp":"2019-01-01T00:00:00+00:00"}`,
	}}

	for _, c := range cases {
		var footer []byte

		pieces := strings.Split(c.token, ".")

		sm, err := base64.RawURLEncoding.DecodeString(pieces[2])
		if err != nil {
			t.Fatal(err)
		}
		if len(pieces) == 4 {
			footer, err = base64.RawURLEncoding.DecodeString(pieces[3])
			if err != nil {
				t.Fatal(err)
			}
		}

		got, err := Verify(public, sm, footer)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.desc, c.exp, string(got))
	}
}
