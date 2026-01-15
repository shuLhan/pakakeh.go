// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package test

// Writer contains common methods between testing.T and testing.B, a subset
// of testing.TB that cannot be used due to private methods.
type Writer interface {
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Helper()
	Log(args ...any)
	Logf(format string, args ...any)
}
