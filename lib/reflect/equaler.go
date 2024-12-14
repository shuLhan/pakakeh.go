// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package reflect

// Equaler is an interface that when implemented by a struct type, it will
// be used to compare the value in [DoEqual] or [IsEqual].
type Equaler interface {
	// Equal compare the struct receiver with parameter v.
	// The v value can be converted to struct type T using (*T).
	// If both struct values are equal it should return nil.
	Equal(v any) error
}
