//
// Package totp implement Time-Based One-Time Password Algorithm based on RFC
// 6238 [1].
//
// [1] https://tools.ietf.org/html/rfc6238
//
package totp

import (
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"hash"
	"log"
	"time"
)

const (
	// DefCodeDigits default digits generated when verifying or generating
	// OTP.
	DefCodeDigits = 6
	DefTimeStep   = 30
)

var _digitsPower = []int{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000}

//
// Protocol contain methods to work with TOTP using the number of digits and
// time steps defined from New().
//
type Protocol struct {
	fnHash     func() hash.Hash
	codeDigits int
	timeStep   int
}

//
// New create TOTP protocol for prover or verifier using "fnHash" as the hmac-sha
// hash function, "codeDigits" as the number of digits to be generated
// and/or verified, and "timeStep" as the time divisor.
//
func New(fnHash func() hash.Hash, codeDigits, timeStep int) Protocol {
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

//
// Generate one time password using the secret and current timestamp.
//
func (p *Protocol) Generate(secret []byte) (otp string, err error) {
	mac := hmac.New(p.fnHash, secret)
	now := time.Now().Unix()
	return p.generateWithTimestamp(mac, now)
}

//
// GenerateN generate n number of passwords from (current time - N*timeStep)
// until the curent time.
//
func (p *Protocol) GenerateN(secret []byte, n int) (listOTP []string, err error) {
	mac := hmac.New(p.fnHash, secret)
	ts := time.Now().Unix()
	for x := 0; x < n; x++ {
		t := ts - int64(x*p.timeStep)
		otp, err := p.generateWithTimestamp(mac, t)
		if err != nil {
			return nil, fmt.Errorf("GenerateN: %w", err)
		}
		listOTP = append(listOTP, otp)
	}
	return listOTP, nil
}

//
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
// For security reason, the maximum stepsBack is limited to 4.
//
func (p *Protocol) Verify(secret []byte, token string, stepsBack int) bool {
	mac := hmac.New(p.fnHash, secret)
	now := time.Now().Unix()
	if stepsBack <= 0 || stepsBack > 4 {
		stepsBack = 1
	}
	return p.verifyWithTimestamp(mac, token, stepsBack, now)
}

func (p *Protocol) verifyWithTimestamp(
	mac hash.Hash, token string, steps int, ts int64,
) bool {
	for x := 0; x < steps; x++ {
		t := ts - int64(x*p.timeStep)
		otp, err := p.generateWithTimestamp(mac, t)
		if err != nil {
			log.Printf("Verify %d: %s", t, err.Error())
			continue
		}
		if otp == token {
			return true
		}
	}
	return false
}

func (p *Protocol) generateWithTimestamp(mac hash.Hash, time int64) (
	otp string, err error,
) {
	steps := int64((float64(time) / float64(p.timeStep)))

	msg := fmt.Sprintf("%016X", steps)
	msgb, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	mac.Reset()
	_, _ = mac.Write(msgb)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0xf

	var binary int = int(hash[offset]&0x7f) << 24
	binary |= int(hash[offset+1]&0xff) << 16
	binary |= int(hash[offset+2]&0xff) << 8
	binary |= int(hash[offset+3] & 0xff)

	otpb := binary % _digitsPower[p.codeDigits]

	fmtZeroPadding := fmt.Sprintf("%%0%dd", p.codeDigits)

	otp = fmt.Sprintf(fmtZeroPadding, otpb)

	return otp, nil
}
