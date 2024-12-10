// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package test

func sum(listNumber ...int) (total int) {
	for _, num := range listNumber {
		total += num
	}
	return total
}
