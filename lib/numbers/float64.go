// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package numbers

import (
	"math"
)

// Float64Round will round `v` to `nprec` digit in fraction.
func Float64Round(v float64, nprec int) float64 {
	pow := math.Pow(10, float64(nprec))
	tmp := v * pow
	_, frac := math.Modf(tmp)
	x := .5
	if frac < 0.0 {
		x = -.5
	}
	if frac >= x {
		v = math.Ceil(tmp)
	} else {
		v = math.Floor(tmp)
	}

	return v / pow
}
