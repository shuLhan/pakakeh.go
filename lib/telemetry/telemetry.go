// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package telemetry is a library for collecting various [Metric], for example
// from standard [runtime/metrics], and send or write it to one or more
// [Forwarder].
// Each Forwarder has capability to format the Metric before sending or
// writing it using [Formatter].
package telemetry
