// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

// Package telemetry is a library for collecting various [Metric], for example
// from standard [runtime/metrics], and send or write it to one or more
// [Forwarder].
// Each Forwarder has capability to format the Metric before sending or
// writing it using [Formatter].
package telemetry
