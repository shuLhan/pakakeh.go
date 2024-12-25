// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

type apoVersion byte

const apoVersionOne apoVersion = 1

// apoHeader define the header for Apo file.
type apoHeader struct {
	// Version define the version of the Apo file.
	Version apoVersion

	// TotalData number of data in the file.
	TotalData int64

	// OffFoot define the offset of Apo footer in the file.
	OffFoot int64
}

func (head *apoHeader) init() {
	head.Version = apoVersionOne
}
