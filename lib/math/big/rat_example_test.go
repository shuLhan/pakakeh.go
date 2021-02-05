// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"fmt"
)

func ExampleRat_Humanize() {
	var (
		thousandSep = "."
		decimalSep  = ","
	)
	fmt.Printf("%s\n", NewRat("0").Humanize(thousandSep, decimalSep))
	fmt.Printf("%s\n", NewRat("0.1234").Humanize(thousandSep, decimalSep))
	fmt.Printf("%s\n", NewRat("100").Humanize(thousandSep, decimalSep))
	fmt.Printf("%s\n", NewRat("100.1234").Humanize(thousandSep, decimalSep))
	fmt.Printf("%s\n", NewRat("1000").Humanize(thousandSep, decimalSep))
	fmt.Printf("%s\n", NewRat("1000.2").Humanize(thousandSep, decimalSep))
	fmt.Printf("%s\n", NewRat("10000.23").Humanize(thousandSep, decimalSep))
	fmt.Printf("%s\n", NewRat("100000.234").Humanize(thousandSep, decimalSep))
	//Output:
	//0
	//0,1234
	//100
	//100,1234
	//1.000
	//1.000,2
	//10.000,23
	//100.000,234
}

func ExampleRat_RoundNearestFraction() {
	fmt.Printf("0.000000001: %s\n", NewRat("0").RoundNearestFraction()) // Affected by DefaultDigitPrecision (8)
	fmt.Printf("0.00545: %s\n", NewRat("0.00545").RoundNearestFraction())
	fmt.Printf("0.00555: %s\n", NewRat("0.00555").RoundNearestFraction())
	fmt.Printf("0.0545: %s\n", NewRat("0.0545").RoundNearestFraction())
	fmt.Printf("0.0555: %s\n", NewRat("0.0555").RoundNearestFraction())
	fmt.Printf("0.545: %s\n", NewRat("0.545").RoundNearestFraction())
	fmt.Printf("0.555: %s\n", NewRat("0.555").RoundNearestFraction())
	fmt.Printf("0.5: %s\n", NewRat("0.5").RoundNearestFraction())
	fmt.Printf("-0.5: %s\n", NewRat("-0.5").RoundNearestFraction())
	fmt.Printf("-0.555: %s\n", NewRat("-0.555").RoundNearestFraction())
	fmt.Printf("-0.545: %s\n", NewRat("-0.545").RoundNearestFraction())
	//Output:
	//0.000000001: 0
	//0.00545: 0.005
	//0.00555: 0.006
	//0.0545: 0.05
	//0.0555: 0.06
	//0.545: 0.5
	//0.555: 0.6
	//0.5: 0.5
	//-0.5: -0.5
	//-0.555: -0.6
	//-0.545: -0.5
}

func ExampleRat_RoundToNearestAway() {
	fmt.Printf("0.5455: %s\n", NewRat("0.5455").RoundToNearestAway(2))
	fmt.Printf("0.5555: %s\n", NewRat("0.5555").RoundToNearestAway(2))
	fmt.Printf("0.5566: %s\n", NewRat("0.5567").RoundToNearestAway(2))
	fmt.Printf("0.5566: %s\n", NewRat("0.5566").RoundToNearestAway(0))
	fmt.Printf("0.5: %s\n", NewRat("0.5").RoundToNearestAway(0))
	fmt.Printf("-0.5: %s\n", NewRat("-0.5").RoundToNearestAway(0))
	//Output:
	//0.5455: 0.55
	//0.5555: 0.56
	//0.5566: 0.56
	//0.5566: 1
	//0.5: 1
	//-0.5: -1
}
