// Copyright 2015 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gini

import (
	"fmt"
	"testing"
)

var (
	data = [][]float64{
		{1.0, 6.0, 5.0, 4.0, 7.0, 3.0, 8.0, 7.0, 5.0},
		{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
	}

	targetValues = []string{"P", "P", "N", "P", "N", "N", "N", "P", "N"}
	classes      = []string{"P", "N"}

	discreteSamples = [][]string{
		{"T", "T", "T", "F", "F", "F", "F", "T", "F"},
		{"T", "T", "F", "F", "T", "T", "F", "F", "T"},
		{"T", "T", "F", "T", "F", "F", "F", "T", "F"},
	}

	discreteValues = []string{"T", "F"}
)

func TestComputeContinu(_ *testing.T) {
	target := make([]string, len(targetValues))

	copy(target, targetValues)

	fmt.Println("target:", target)

	fmt.Println("data:", data[0])
	GINI := Gini{}
	GINI.ComputeContinu(&data[0], &target, &classes)
	fmt.Println(">>> gini:", GINI)

	fmt.Println("data:", data[1])
	GINI = Gini{}
	GINI.ComputeContinu(&data[1], &target, &classes)
	fmt.Println(">>> gini:", GINI)
}

func TestComputeDiscrete(_ *testing.T) {
	gini := Gini{}
	target := make([]string, len(targetValues))

	for _, sample := range discreteSamples {
		copy(target, targetValues)

		fmt.Println("target:", target)
		fmt.Println("data:", sample)

		sample := sample
		gini.ComputeDiscrete(&sample, &discreteValues, &target, &classes)

		fmt.Println(gini)
	}
}
