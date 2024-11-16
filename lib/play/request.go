// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

const cookieNameSid = `sid`

// Request for calling [Format] and [Run].
type Request struct {
	// cookieSid contains unique session ID between request.
	// A single session can only run one command at a time, otherwise
	// the previous command will be canceled first.
	//
	// In the HTTP request, the sid is read from cookie named "sid".
	cookieSid *http.Cookie

	// The Go version that will be used in go.mod.
	GoVersion string `json:"goversion"`

	// Body contains the Go code to be Format-ed or Run.
	Body string `json:"body"`

	// WithoutRace define option opt out "-race" when running Go code.
	WithoutRace bool `json:"without_race"`
}

func (req *Request) init() {
	if req.cookieSid == nil {
		req.cookieSid = &http.Cookie{
			Name:  cookieNameSid,
			Value: req.generateSid(),
		}
	}
	req.cookieSid.Path = `/`
	req.cookieSid.MaxAge = 604800 // Seven days.
	req.cookieSid.SameSite = http.SameSiteStrictMode

	if req.GoVersion == `` {
		req.GoVersion = GoVersion
	}
}

// generateSid generate session ID from the first 16 hex of SHA256 hash of
// request body plus current epoch in.
func (req *Request) generateSid() string {
	var (
		plain = []byte(req.Body)
		epoch = now()
	)
	plain = libbytes.AppendInt64(plain, epoch)
	var cipher = sha256.Sum256(plain)
	var dst = make([]byte, hex.EncodedLen(len(cipher)))
	hex.Encode(dst, cipher[:])

	return string(dst[:16])
}
