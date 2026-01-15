// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package numbers

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestFloat64Round(t *testing.T) {
	data := []float64{0.553, -0.553, 0.49997, -0.49997, 0.4446, -0.4446}
	exps := [][]float64{
		{1, 0.6, 0.55},
		{-1, -0.6, -0.55},
		{0.0, 0.5, 0.5},
		{0.0, -0.5, -0.5},
		{0, 0.4, 0.44},
		{0, -0.4, -0.44},
	}

	for x := range data {
		for y, exp := range exps[x] {
			got := Float64Round(data[x], y)
			test.Assert(t, "", exp, got)
		}
	}
}
