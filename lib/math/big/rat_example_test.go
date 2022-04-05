// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
)

func ExampleAddRat() {
	fmt.Println(AddRat())
	fmt.Println(AddRat(nil))
	fmt.Println(AddRat("a"))
	fmt.Println(AddRat(0, 0.0001))
	fmt.Println(AddRat("1.007", "a", "2.003")) // Invalid parameter "a" is ignored.
	//Output:
	//0
	//0
	//0
	//0.0001
	//3.01
}

func ExampleMulRat() {
	fmt.Println(MulRat())
	fmt.Println(MulRat(nil))
	fmt.Println(MulRat("a"))
	fmt.Println(MulRat(0, 1))
	fmt.Println(MulRat(6, "a", "0.3")) // Invalid parameter "a" is ignored.
	//Output:
	//0
	//0
	//0
	//0
	//1.8
}

func ExampleNewRat() {
	var (
		stdNilInt *big.Int
		stdNilRat *big.Rat
		stdRat    big.Rat
		nilRat    *Rat
	)
	inputs := []interface{}{
		nil,
		[]byte{},
		"",
		[]byte("a"),
		"0",
		"0.0000_0001",
		[]byte("14687233442.06916608"),
		"14_687_233_442.069_166_08",
		nilRat,
		NewRat("14687233442.06916608"),
		*NewRat("14687233442.06916608"),
		stdNilRat,
		stdRat,
		big.NewRat(14687233442, 100_000_000),
		*big.NewRat(14687233442, 100_000_000),
		uint16(math.MaxUint16),
		uint32(math.MaxUint32),
		uint64(math.MaxUint64),
		stdNilInt,
		big.NewInt(100_000_000),
	}

	for _, v := range inputs {
		fmt.Println(NewRat(v))
	}
	//Output:
	//0
	//0
	//0
	//0
	//0
	//0.00000001
	//14687233442.06916608
	//14687233442.06916608
	//0
	//14687233442.06916608
	//14687233442.06916608
	//0
	//0
	//146.87233442
	//146.87233442
	//65535
	//4294967295
	//18446744073709551615
	//0
	//100000000
}

func ExampleQuoRat() {
	fmt.Println(QuoRat())
	fmt.Println(QuoRat(nil))
	fmt.Println(QuoRat("a"))
	fmt.Println(QuoRat("0"))
	fmt.Println(QuoRat(2, 0, 2))
	fmt.Println(QuoRat(6, "a", "0.3"))
	fmt.Println(QuoRat(0, 1))
	fmt.Println(QuoRat(4651, 272))
	fmt.Println(QuoRat(int64(1815507979407), NewRat(100000000)))
	fmt.Println(QuoRat("25494300", "25394000000"))

	//Output:
	//0
	//0
	//0
	//0
	//0
	//0
	//0
	//17.0992647
	//18155.07979407
	//0.00100395
}

func ExampleSubRat() {
	fmt.Println(SubRat())
	fmt.Println(SubRat(nil))
	fmt.Println(SubRat("a"))
	fmt.Println(SubRat(0, 1))
	fmt.Println(SubRat(6, "a", "0.3"))
	//Output:
	//0
	//0
	//0
	//-1
	//5.7
}

func ExampleRat_Abs() {
	fmt.Println(NewRat(nil).Abs())
	fmt.Println(NewRat("-1").Abs())
	fmt.Println(NewRat("-0.00001").Abs())
	fmt.Println(NewRat("1").Abs())
	//Output:
	//0
	//1
	//0.00001
	//1
}

func ExampleRat_Add() {
	fmt.Println(NewRat(nil).Add(nil))
	fmt.Println(NewRat(1).Add(nil))
	fmt.Println(NewRat(1).Add(1))
	//Output:
	//0
	//1
	//2
}

func ExampleRat_Humanize() {
	var (
		thousandSep = "."
		decimalSep  = ","
	)
	fmt.Printf("%s\n", NewRat(nil).Humanize(thousandSep, decimalSep))
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
	//0
	//0,1234
	//100
	//100,1234
	//1.000
	//1.000,2
	//10.000,23
	//100.000,234
}

func ExampleRat_Int64() {
	fmt.Printf("MaxInt64: %d\n", NewRat("9223372036854775807").Int64())
	fmt.Printf("MaxInt64+1: %d\n", NewRat("9223372036854775808").Int64())
	fmt.Printf("MinInt64: %d\n", NewRat("-9223372036854775808").Int64())
	fmt.Printf("MinInt64-1: %d\n", NewRat("-9223372036854775809").Int64())

	fmt.Println(NewRat(nil).Int64())
	fmt.Println(NewRat("0.000_000_001").Int64())
	fmt.Println(NewRat("0.5").Int64())
	fmt.Println(NewRat("0.6").Int64())
	fmt.Println(NewRat("4011144900.02438879").Mul(100000000).Int64())
	fmt.Println(QuoRat("128_900", "0.000_0322").Int64())
	fmt.Println(QuoRat(128900, 3220).Mul(100000000).Int64())
	fmt.Println(QuoRat(25494300, QuoRat(25394000000, 100000000)).Int64())

	//Output:
	//MaxInt64: 9223372036854775807
	//MaxInt64+1: 9223372036854775807
	//MinInt64: -9223372036854775808
	//MinInt64-1: -9223372036854775808
	//0
	//0
	//0
	//0
	//401114490002438879
	//4003105590
	//4003105590
	//100394
}

func ExampleRat_IsEqual() {
	f := NewRat(1)

	fmt.Println(NewRat(nil).IsEqual(0))
	fmt.Println(f.IsEqual("a"))
	fmt.Println(f.IsEqual("1"))
	fmt.Println(f.IsEqual(1.1))
	fmt.Println(f.IsEqual(byte(1)))
	fmt.Println(f.IsEqual(int(1)))
	fmt.Println(f.IsEqual(int32(1)))
	fmt.Println(f.IsEqual(float32(1)))
	fmt.Println(f.IsEqual(NewRat(1)))
	// Equal due to String() truncation to DefaultDigitPrecision (8) digits.
	fmt.Println(NewRat("0.1234567890123").IsEqual("0.12345678"))

	//Output:
	//false
	//false
	//true
	//false
	//true
	//true
	//true
	//true
	//true
	//true
}

func ExampleRat_IsGreater() {
	r := NewRat("0.000_000_5")

	fmt.Println(NewRat(nil).IsGreater(0))
	fmt.Println(r.IsGreater(nil))
	fmt.Println(r.IsGreater("0.000_000_5"))
	fmt.Println(r.IsGreater("0.000_000_49999"))

	//Output:
	//false
	//false
	//false
	//true
}

func ExampleRat_IsGreaterOrEqual() {
	r := NewRat("0.000_000_5")

	fmt.Println(NewRat(nil).IsGreaterOrEqual(0))
	fmt.Println(r.IsGreaterOrEqual(nil))
	fmt.Println(r.IsGreaterOrEqual("0.000_000_500_000_000_001"))
	fmt.Println(r.IsGreaterOrEqual("0.000_000_5"))
	fmt.Println(r.IsGreaterOrEqual("0.000_000_49999"))

	//Output:
	//false
	//false
	//false
	//true
	//true
}

func ExampleRat_IsGreaterThanZero() {
	fmt.Println(NewRat(nil).IsGreaterThanZero())
	fmt.Println(NewRat(0).IsGreaterThanZero())
	fmt.Println(NewRat("-0.000_000_000_000_000_001").IsGreaterThanZero())
	fmt.Println(NewRat("0.000_000_000_000_000_001").IsGreaterThanZero())

	//Output:
	//false
	//false
	//false
	//true
}

func ExampleRat_IsLess() {
	r := NewRat("0.000_000_5")

	fmt.Println(NewRat(nil).IsLess(0))
	fmt.Println(r.IsLess(nil))
	fmt.Println(r.IsLess("0.000_000_5"))
	fmt.Println(r.IsLess("0.000_000_49"))
	fmt.Println(r.IsLess("0.000_000_500_000_000_001"))

	//Output:
	//false
	//false
	//false
	//false
	//true
}

func ExampleRat_IsLessOrEqual() {
	r := NewRat("0.000_000_5")

	fmt.Println(NewRat(nil).IsLessOrEqual(r))
	fmt.Println(r.IsLessOrEqual(nil))
	fmt.Println(r.IsLessOrEqual("0.000_000_5"))
	fmt.Println(r.IsLessOrEqual("0.000_000_49"))
	fmt.Println(r.IsLessOrEqual("0.000_000_500_000_000_001"))

	//Output:
	//false
	//false
	//true
	//false
	//true
}

func ExampleRat_IsLessThanZero() {
	fmt.Println(NewRat(nil).IsLessThanZero())
	fmt.Println(NewRat(byte(0)).IsLessThanZero())
	fmt.Println(NewRat("-0.000_000_000_000_000_001").IsLessThanZero())
	fmt.Println(NewRat("0.000_000_000_000_000_001").IsLessThanZero())

	//Output:
	//false
	//false
	//true
	//false
}

func ExampleRat_IsZero() {
	fmt.Println(NewRat(nil).IsZero())
	fmt.Println(NewRat(byte(0)).IsZero())
	fmt.Println(NewRat(byte(-0)).IsZero())
	fmt.Println(NewRat("-0.000_000_000_000_000_001").IsZero())
	fmt.Println(NewRat("0.000_000_000_000_000_001").IsZero())

	//Output:
	//false
	//true
	//true
	//false
	//false
}

func ExampleRat_MarshalJSON() {
	inputs := []string{
		"",
		"a",
		"0.0000_0000",
		"0.1",
		"0.0000_0001",
		"0.0000_0000_1", // Truncated by DefaultDigitPrecision.
		"1234567890.0",
		"64.23738872403", // Truncated by DefaultDigitPrecision.
		"0.1234567890",
		"142660378.65368736",
		"9193394308.85771370",
		"14687233442.06916608",
	}

	MarshalJSONAsString = true
	for _, in := range inputs {
		out, _ := NewRat(in).MarshalJSON()
		fmt.Printf("%s\n", out)
	}

	// Setting this to false will make the JSON output become number.
	MarshalJSONAsString = false

	for _, in := range inputs {
		out, _ := NewRat(in).MarshalJSON()
		fmt.Printf("%s\n", out)
	}

	//Output:
	//"0"
	//"0"
	//"0"
	//"0.1"
	//"0.00000001"
	//"0"
	//"1234567890"
	//"64.23738872"
	//"0.12345678"
	//"142660378.65368736"
	//"9193394308.8577137"
	//"14687233442.06916608"
	//0
	//0
	//0
	//0.1
	//0.00000001
	//0
	//1234567890
	//64.23738872
	//0.12345678
	//142660378.65368736
	//9193394308.8577137
	//14687233442.06916608
}

func ExampleRat_MarshalJSON_withStruct() {
	type T struct {
		V *Rat
	}

	inputs := []T{
		{V: nil},
		{V: NewRat(0)},
		{V: NewRat("0.1234567890")},
	}

	MarshalJSONAsString = true
	for _, in := range inputs {
		out, _ := json.Marshal(&in)
		fmt.Printf("%s\n", out)
	}

	MarshalJSONAsString = false
	for _, in := range inputs {
		out, _ := json.Marshal(&in)
		fmt.Printf("%s\n", out)
	}

	//Output:
	//{"V":null}
	//{"V":"0"}
	//{"V":"0.12345678"}
	//{"V":null}
	//{"V":0}
	//{"V":0.12345678}
}

func ExampleRat_Mul() {
	const (
		defValue = "14687233442.06916608"
	)

	fmt.Println(NewRat(defValue).Mul("a"))
	fmt.Println(NewRat(defValue).Mul("0"))
	fmt.Println(NewRat(defValue).Mul(defValue))
	fmt.Println(NewRat("1.06916608").Mul("1.06916608"))
	//Output:
	//0
	//0
	//215714826181834884090.46087866
	//1.1431161
}

func ExampleRat_Quo() {
	const (
		defValue = "14687233442.06916608"
	)

	fmt.Println(NewRat(nil).Quo(1))
	fmt.Println(NewRat(defValue).Quo(nil))
	fmt.Println(NewRat(defValue).Quo("a"))
	fmt.Println(NewRat(defValue).Quo("100_000_000"))
	//Output:
	//0
	//0
	//0
	//146.87233442
}

func ExampleRat_RoundNearestFraction() {
	fmt.Printf("nil: %s\n", NewRat(nil).RoundNearestFraction())
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
	//nil: 0
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
	fmt.Printf("nil: %s\n", NewRat(nil).RoundToNearestAway(2))
	fmt.Printf("0.0054: %s\n", NewRat("0.0054").RoundToNearestAway(2))
	fmt.Printf("0.0054: %s\n", NewRat("0.0054").RoundToNearestAway(1))
	fmt.Printf("0.5455: %s\n", NewRat("0.5455").RoundToNearestAway(2))
	fmt.Printf("0.5555: %s\n", NewRat("0.5555").RoundToNearestAway(2))
	fmt.Printf("0.5566: %s\n", NewRat("0.5567").RoundToNearestAway(2))
	fmt.Printf("0.5566: %s\n", NewRat("0.5566").RoundToNearestAway(0))

	fmt.Printf("0.02514135: %s\n", NewRat("0.02514135").RoundToNearestAway(6))
	fmt.Printf("0.02514145: %s\n", NewRat("0.02514145").RoundToNearestAway(6))
	fmt.Printf("0.02514155: %s\n", NewRat("0.02514155").RoundToNearestAway(6))
	fmt.Printf("0.02514165: %s\n", NewRat("0.02514165").RoundToNearestAway(6))

	fmt.Printf("0.5: %s\n", NewRat("0.5").RoundToNearestAway(0))
	fmt.Printf("-0.5: %s\n", NewRat("-0.5").RoundToNearestAway(0))
	//Output:
	//nil: 0
	//0.0054: 0.01
	//0.0054: 0
	//0.5455: 0.55
	//0.5555: 0.56
	//0.5566: 0.56
	//0.5566: 1
	//0.02514135: 0.025141
	//0.02514145: 0.025141
	//0.02514155: 0.025142
	//0.02514165: 0.025142
	//0.5: 1
	//-0.5: -1
}

func ExampleRat_RoundToZero() {
	fmt.Println(NewRat(nil).RoundToZero(2))
	fmt.Println(NewRat("0.5455").RoundToZero(2))
	fmt.Println(NewRat("0.5555").RoundToZero(2))
	fmt.Println(NewRat("0.5566").RoundToZero(2))
	fmt.Println(NewRat("0.5566").RoundToZero(0))
	// In Go <= 1.18, this will print "-0", but on Go tip "0".
	// So to make test success on all versions, we multiple it to 1.
	fmt.Println(NewRat("-0.5").RoundToZero(0).Mul(1))
	//Output:
	//0
	//0.54
	//0.55
	//0.55
	//0
	//0
}

func ExampleRat_Scan() {
	var (
		r   = &Rat{}
		err error
	)

	inputs := []interface{}{
		1234,
		nil,
		"0.0001",
		float64(0.0001),
	}
	for _, in := range inputs {
		err = r.Scan(in)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
		fmt.Println(r)
	}

	//Output:
	//1234
	//error: Rat.Scan: unknown type <nil>
	//1234
	//0.0001
	//0.0001
}

func ExampleRat_String() {
	inputs := []interface{}{
		nil,
		"12345",
		"0.00000000",
		float64(0.00000000),
		"0.1",
		float64(0.1),
		"0.0000001",
		float64(0.0000001),
		"0.000000001", // Truncated due to rounding.
		"64.23738872403",
		float64(64.23738872403),
	}
	//Output:
	//0
	//12345
	//0
	//0
	//0.1
	//0.1
	//0.0000001
	//0.0000001
	//0
	//64.23738872
	//64.23738872

	for _, in := range inputs {
		fmt.Println(NewRat(in).String())
	}
}

func ExampleRat_Sub() {
	fmt.Println(NewRat(nil).Sub(1))
	fmt.Println(NewRat(1).Sub(nil))
	fmt.Println(NewRat(1).Sub(1))
	//Output:
	//0
	//1
	//0
}

func ExampleRat_UnmarshalJSON_withStruct() {
	type T struct {
		V *Rat `json:"V"`
		W *Rat `json:"W,omitempty"`
	}

	inputs := []string{
		`{"V":"ab"}`,
		`{}`,
		`{"V":}`,
		`{"V":0,"W":0}`,
		`{"V":"1"}`,
		`{"V":0.123456789}`,
		`{"V":"0.1234", "W":0.5678}`,
	}
	for _, in := range inputs {
		t := T{}
		err := json.Unmarshal([]byte(in), &t)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s %s\n", t.V, t.W)
	}
	//Output:
	//Rat.UnmarshalJSON: cannot convert []uint8([97 98]) to Rat
	//0 0
	//invalid character '}' looking for beginning of value
	//0 0
	//1 0
	//0.12345678 0
	//0.1234 0.5678
}

func ExampleRat_Value() {
	inputs := []interface{}{
		nil,
		0,
		1.2345,
		"12345.6789_1234_5678_9",
	}
	for _, in := range inputs {
		out, _ := NewRat(in).Value()
		fmt.Printf("%s\n", out)
	}
	//Output:
	//0
	//0
	//1.2345
	//12345.67891234
}
