// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import (
	"fmt"
	"time"
)

const (
	dateTimeLayout = "2006-01-02T15:04:05-07:00"
)

type JSONToken struct {
	Issuer    string     `json:"iss,omitempty"`
	Subject   string     `json:"sub,omitempty"`
	Audience  string     `json:"aud,omitempty"`
	ExpiredAt *time.Time `json:"exp,omitempty"`
	NotBefore *time.Time `json:"nbf,omitempty"`
	IssuedAt  *time.Time `json:"iat,omitempty"`
	TokenID   string     `json:"jti,omitempty"`
	Data      string     `json:"data"`
}

//
// Validate the ExpiredAt and NotBefore time fields.
//
func (jtoken *JSONToken) Validate() (err error) {
	now := time.Now()
	if jtoken.ExpiredAt != nil {
		if now.After(*jtoken.ExpiredAt) {
			return fmt.Errorf("token is expired")
		}
	}
	if jtoken.NotBefore != nil {
		if now.Before(*jtoken.NotBefore) {
			return fmt.Errorf("token is too early")
		}
	}
	return nil
}
