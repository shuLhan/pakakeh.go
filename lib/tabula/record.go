// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

import (
	"math"
	"reflect"
	"strconv"
)

const (
	// TUndefined for undefined type
	TUndefined = -1
	// TString string type.
	TString = 0
	// TInteger integer type (64 bit).
	TInteger = 1
	// TReal float type (64 bit).
	TReal = 2
)

// Record represent the smallest building block of data-set.
type Record struct {
	v interface{}
}

// NewRecord will create and return record with nil value.
func NewRecord() *Record {
	return &Record{v: nil}
}

// NewRecordBy create new record from string with type set to `t`.
func NewRecordBy(v string, t int) (r *Record, e error) {
	r = NewRecord()
	e = r.SetValue(v, t)
	return
}

// NewRecordString will create new record from string.
func NewRecordString(v string) (r *Record) {
	return &Record{v: v}
}

// NewRecordInt create new record from integer value.
func NewRecordInt(v int64) (r *Record) {
	return &Record{v: v}
}

// NewRecordReal create new record from float value.
func NewRecordReal(v float64) (r *Record) {
	return &Record{v: v}
}

// Clone will create and return a clone of record.
func (r *Record) Clone() *Record {
	return &Record{v: r.v}
}

// IsNil return true if record has not been set with value, or nil.
func (r *Record) IsNil() bool {
	return r.v == nil
}

// Type of record.
func (r *Record) Type() int {
	switch r.v.(type) {
	case int64:
		return TInteger
	case float64:
		return TReal
	}
	return TString
}

// SetValue set the record value from string using type `t`. If value can not
// be converted to type, it will return an error.
func (r *Record) SetValue(v string, t int) error {
	switch t {
	case TString:
		r.v = v

	case TInteger:
		i64, e := strconv.ParseInt(v, 10, 64)
		if nil != e {
			return e
		}

		r.v = i64

	case TReal:
		f64, e := strconv.ParseFloat(v, 64)
		if nil != e {
			return e
		}

		r.v = f64
	}
	return nil
}

// SetString will set the record value with string value.
func (r *Record) SetString(v string) {
	r.v = v
}

// SetFloat will set the record value with float 64bit.
func (r *Record) SetFloat(v float64) {
	r.v = v
}

// SetInteger will set the record value with integer 64bit.
func (r *Record) SetInteger(v int64) {
	r.v = v
}

// IsMissingValue check wether the value is a missing attribute.
//
// If its string the missing value is indicated by character '?'.
//
// If its integer the missing value is indicated by minimum negative integer,
// or math.MinInt64.
//
// If its real the missing value is indicated by -Inf.
func (r *Record) IsMissingValue() bool {
	switch v := r.v.(type) {
	case string:
		if v == "?" {
			return true
		}

	case int64:
		if v == math.MinInt64 {
			return true
		}

	case float64:
		return math.IsInf(v, -1)
	}

	return false
}

// Interface return record value as interface.
func (r *Record) Interface() interface{} {
	return r.v
}

// Bytes convert record value to slice of byte.
func (r *Record) Bytes() []byte {
	return []byte(r.String())
}

// String convert record value to string.
func (r Record) String() (s string) {
	switch v := r.v.(type) {
	case string:
		s = v

	case int64:
		s = strconv.FormatInt(v, 10)

	case float64:
		s = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return
}

// Float convert given record to float value. If its failed it will return
// the -Infinity value.
func (r *Record) Float() (f64 float64) {
	var e error

	switch v := r.v.(type) {
	case string:
		f64, e = strconv.ParseFloat(v, 64)

		if nil != e {
			f64 = math.Inf(-1)
		}

	case int64:
		f64 = float64(v)

	case float64:
		f64 = v
	}

	return
}

// Integer convert given record to integer value. If its failed, it will return
// the minimum integer in 64bit.
func (r *Record) Integer() (i64 int64) {
	var e error

	switch v := r.v.(type) {
	case string:
		i64, e = strconv.ParseInt(v, 10, 64)

		if nil != e {
			i64 = math.MinInt64
		}

	case int64:
		i64 = v

	case float64:
		i64 = int64(v)
	}

	return
}

// IsEqual return true if record is equal with other, otherwise return false.
func (r *Record) IsEqual(o *Record) bool {
	return reflect.DeepEqual(r.v, o.Interface())
}

// IsEqualToString return true if string representation of record value is
// equal to string `v`.
func (r *Record) IsEqualToString(v string) bool {
	return r.String() == v
}

// IsEqualToInterface return true if interface type and value equal to record
// type and value.
func (r *Record) IsEqualToInterface(v interface{}) bool {
	return reflect.DeepEqual(r.v, v)
}

// Reset will reset record value to empty string or zero, depend on type.
func (r *Record) Reset() {
	switch r.v.(type) {
	case string:
		r.v = ""
	case int64:
		r.v = int64(0)
	case float64:
		r.v = float64(0)
	}
}
