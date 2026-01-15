// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import "time"

// Timestamper a type that return a function to generate timestamp.
type Timestamper func() int64

// SecondTimestamp return the number of seconds elapsed since January 1,
// 1970 UTC
func SecondTimestamp() Timestamper {
	return func() int64 { return time.Now().Unix() }
}

// MilliTimestamp return the number of milliseconds elapsed since January 1,
// 1970 UTC.
func MilliTimestamp() Timestamper {
	return func() int64 { return time.Now().UnixMilli() }
}

// NanoTimestamp return the number of nanoseconds elapsed since January 1,
// 1970 UTC
func NanoTimestamp() Timestamper {
	return func() int64 { return time.Now().UnixNano() }
}

// DummyTimestamp return fixed epoch 1678606568, for testing only.
func DummyTimestamp() Timestamper {
	return func() int64 { return 1678606568 }
}
