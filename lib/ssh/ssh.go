// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

// List of signature algorithms.
const (
	SignAlg_ECDSA_SHA2_NISTP256 = "ecdsa-sha2-nistp256"
	SignAlg_ECDSA_SHA2_NISTP384 = "ecdsa-sha2-nistp384"
	SignAlg_ECDSA_SHA2_NISTP521 = "ecdsa-sha2-nistp521"
	SignAlg_RSA_SHA2_256        = "rsa-sha2-256"
	SignAlg_RSA_SHA2_512        = "rsa-sha2-512"
	SignAlg_SSH_ED22519         = "ssh-ed22519"
	SignAlg_SSH_RSA             = "ssh-rsa"
)

//
// patternToRegex convert the Host and Match pattern string into regex.
//
func patternToRegex(in string) (out string) {
	sr := make([]rune, 0, len(in))
	for _, r := range in {
		switch r {
		case '*', '?':
			sr = append(sr, '.')
		case '.':
			sr = append(sr, '\\')
		}
		sr = append(sr, r)
	}
	return string(sr)
}
