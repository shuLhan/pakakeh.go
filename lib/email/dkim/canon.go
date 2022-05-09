// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"bytes"
	"fmt"
)

// Canon define type of canonicalization algorithm.
type Canon byte

// List of valid and known canonicalization algorithms.
const (
	CanonSimple Canon = iota // "simple" (default)
	CanonRelaxed
)

// canonNames contains mapping between canonical type and their human
// readabale names.
var canonNames = map[Canon][]byte{
	CanonSimple:  []byte("simple"),
	CanonRelaxed: []byte("relaxed"),
}

// unpackCanons unpack Signature canonicalization algorithms.
func unpackCanons(v []byte) (canonHeader, canonBody *Canon, err error) {
	var vHeader, vBody []byte

	canons := bytes.Split(v, sepSlash)

	switch len(canons) {
	case 0:
	case 1:
		vHeader = canons[0]
	case 2:
		vHeader = canons[0]
		vBody = canons[1]
	default:
		err = fmt.Errorf("dkim: invalid canonicalization: '%s'", v)
		return nil, nil, err
	}

	canonHeader, err = parseCanonValue(vHeader)
	if err != nil {
		return nil, nil, err
	}
	if canonHeader != nil {
		canonBody, err = parseCanonValue(vBody)
		if err != nil {
			return nil, nil, err
		}
	}

	return canonHeader, canonBody, nil
}

// parseCanonValue parse canonicalization name and return their numeric type.
func parseCanonValue(v []byte) (*Canon, error) {
	if len(v) == 0 {
		return nil, nil
	}
	for k, cname := range canonNames {
		if bytes.Equal(v, cname) {
			k := k
			return &k, nil
		}
	}
	return nil, fmt.Errorf("dkim: invalid canonicalization: '%s'", v)
}
