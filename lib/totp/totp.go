// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package totp implement Time-Based One-Time Password Algorithm based on RFC
// 6238 [1].
//
// [1] https://tools.ietf.org/html/rfc6238
package totp

import (
	"crypto"
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"log"
	"time"
)

// CryptoHash define an alias to limit the type of hash algorithm to be
// used in TOTP.
type CryptoHash crypto.Hash

// List of available hash function that can be used in TOTP.
//
// See RFC 6238 Section 1.2.
const (
	CryptoHashSHA1   CryptoHash = CryptoHash(crypto.SHA1) // Default hash algorithm.
	CryptoHashSHA256            = CryptoHash(crypto.SHA256)
	CryptoHashSHA512            = CryptoHash(crypto.SHA512)
)

// Default value for hash, digits, time-step, and maximum step backs.
const (
	DefHash = CryptoHashSHA1

	// DefCodeDigits default digits generated when verifying or generating
	// OTP.
	DefCodeDigits = 6
	DefTimeStep   = 30

	// DefStepsBack maximum value for stepsBack parameter on Verify.
	// The value 20 means the Verify() method will check maximum 20 TOTP
	// tokens or 10 minutes to the past.
	DefStepsBack = 20
)

var _digitsPower = []int{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000}

// Protocol contain methods to work with TOTP using the number of digits and
// time steps defined from New().
type Protocol struct {
	fnHash     func() hash.Hash
	codeDigits int
	timeStep   int
}

// New create TOTP protocol for prover or verifier using "cryptoHash" as the
// hmac-sha hash function, "codeDigits" as the number of digits to be
// generated and/or verified, and "timeStep" as the time divisor.
//
// There are only three hash functions that can be used: SHA1, SHA256, and
// SHA512.
// Passing hash value other than that, will revert the value default to SHA1.
//
// The maximum value for codeDigits parameter is 8.
func New(cryptoHash CryptoHash, codeDigits, timeStep int) Protocol {
	var fnHash func() hash.Hash

	switch cryptoHash {
	case CryptoHashSHA256:
		fnHash = sha256.New
	case CryptoHashSHA512:
		fnHash = sha512.New
	default:
		fnHash = sha1.New
	}
	if codeDigits <= 0 || codeDigits > 8 {
		codeDigits = DefCodeDigits
	}
	if timeStep <= 0 {
		timeStep = DefTimeStep
	}

	return Protocol{
		fnHash:     fnHash,
		codeDigits: codeDigits,
		timeStep:   timeStep,
	}
}

// Generate one time password using the secret and current timestamp.
func (p *Protocol) Generate(secret []byte) (otp string, err error) {
	var (
		mac = hmac.New(p.fnHash, secret)
		now = time.Now().Unix()
	)

	return p.generateWithTimestamp(mac, now)
}

// GenerateWithTime generate one time password using ts as time and secret.
func (p *Protocol) GenerateWithTime(ts time.Time, secret []byte) (otp string, err error) {
	var mac = hmac.New(p.fnHash, secret)
	return p.generateWithTimestamp(mac, ts.Unix())
}

// GenerateN generate n number of passwords from (ts - 0*timeStep)
// until (ts - N*timeStep).
func (p *Protocol) GenerateN(secret []byte, n int) (listOTP []string, err error) {
	var (
		mac = hmac.New(p.fnHash, secret)
		ts  = time.Now().Unix()
	)
	return p.generateN(mac, ts, n)
}

// GenerateNWithTime generate n number of passwords from (ts - 0*timeStep)
// until (ts - N*timeStep).
func (p *Protocol) GenerateNWithTime(ts time.Time, secret []byte, n int) (listOTP []string, err error) {
	var mac = hmac.New(p.fnHash, secret)
	return p.generateN(mac, ts.Unix(), n)
}

func (p *Protocol) generateN(mac hash.Hash, ts int64, n int) (listOTP []string, err error) {
	var (
		otp string
		t   int64
		x   int
	)
	for x = 0; x < n; x++ {
		t = ts - int64(x*p.timeStep)
		otp, err = p.generateWithTimestamp(mac, t)
		if err != nil {
			return nil, fmt.Errorf(`GenerateN: %w`, err)
		}
		listOTP = append(listOTP, otp)
	}
	return listOTP, nil
}

// Verify the token based on the prover secret key.
// It will return true if the token matched, otherwise it will return false.
//
// The stepsBack parameter define number of steps in the pass to be checked
// for valid OTP.
// For example, if stepsBack = 2 and timeStep = 30, the time range to
// checking OTP is in between
//
//	(current_timestamp - (2*30)) ... current_timestamp
//
// For security reason, the maximum stepsBack is limited to DefStepsBack.
func (p *Protocol) Verify(secret []byte, token string, stepsBack int) bool {
	var (
		mac = hmac.New(p.fnHash, secret)
		now = time.Now().Unix()
	)

	if stepsBack <= 0 || stepsBack > DefStepsBack {
		stepsBack = DefStepsBack
	}
	return p.verifyWithTimestamp(mac, token, stepsBack, now)
}

func (p *Protocol) verifyWithTimestamp(mac hash.Hash, token string, steps int, ts int64) bool {
	var (
		otp string
		err error
		t   int64
		x   int
	)

	for x = 0; x < steps; x++ {
		t = ts - int64(x*p.timeStep)
		otp, err = p.generateWithTimestamp(mac, t)
		if err != nil {
			log.Printf(`Verify %d: %s`, t, err.Error())
			continue
		}
		if otp == token {
			return true
		}
	}
	return false
}

func (p *Protocol) generateWithTimestamp(mac hash.Hash, time int64) (otp string, err error) {
	var (
		steps = int64((float64(time) / float64(p.timeStep)))
		msg   = fmt.Sprintf(`%016X`, steps)

		fmtZeroPadding string
		binary         int
		vbytes         []byte
		offset         byte
	)

	vbytes, err = hex.DecodeString(msg)
	if err != nil {
		return ``, err
	}

	mac.Reset()
	_, _ = mac.Write(vbytes)
	vbytes = mac.Sum(nil)

	offset = vbytes[len(vbytes)-1] & 0xf

	binary = int(vbytes[offset]&0x7f) << 24
	binary |= int(vbytes[offset+1]&0xff) << 16
	binary |= int(vbytes[offset+2]&0xff) << 8
	binary |= int(vbytes[offset+3] & 0xff)

	binary %= _digitsPower[p.codeDigits]

	fmtZeroPadding = fmt.Sprintf(`%%0%dd`, p.codeDigits)

	otp = fmt.Sprintf(fmtZeroPadding, binary)

	return otp, nil
}
