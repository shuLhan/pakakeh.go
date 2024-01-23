// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package numbers

import (
	"crypto/rand"
	"log"
	"math/big"
)

// IntCreateSeq will create and return sequence of integer from `min` to
// `max`.
//
// E.g. if min is 0 and max is 5 then it will return `[0 1 2 3 4 5]`.
func IntCreateSeq(min, max int) (seq []int) {
	for ; min <= max; min++ {
		seq = append(seq, min)
	}
	return
}

// IntPickRandPositive return random integer value from 0 to maximum value
// `maxVal`.
//
// The random value is checked with already picked index: `pickedListID`.
//
// If `dup` is true, allow duplicate value in `pickedListID`, otherwise only
// single unique value allowed in `pickedListID`.
//
// If excluding index `exsListID` is not empty, do not pick the integer value
// listed in there.
func IntPickRandPositive(maxVal int, dup bool, pickedListID, exsListID []int) (idx int) {
	var (
		logp    = `IntPickRandPositive`
		randMax = big.NewInt(int64(maxVal))

		randv    *big.Int
		err      error
		excluded bool
		picked   bool
	)

	for {
		randv, err = rand.Int(rand.Reader, randMax)
		if err != nil {
			log.Panicf(`%s: %s`, logp, err)
		}

		idx = int(randv.Int64())

		// Check in exclude indices.
		excluded = false
		for _, v := range exsListID {
			if idx == v {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		if dup {
			// Allow duplicate idx.
			return
		}

		// Check if its already picked.
		picked = false
		for _, v := range pickedListID {
			if idx == v {
				picked = true
				break
			}
		}

		if picked {
			// Get another random idx again.
			continue
		}

		// Bingo, we found unique idx that has not been picked.
		return
	}
}
