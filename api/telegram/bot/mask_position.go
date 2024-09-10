// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// MaskPosition describes the position on faces where a mask should be placed
// by default.
type MaskPosition struct {
	// The part of the face relative to which the mask should be placed.
	// One of “forehead”, “eyes”, “mouth”, or “chin”.
	Point string `json:"point"`

	// Shift by X-axis measured in widths of the mask scaled to the face
	// size, from left to right.
	// For example, choosing -1.0 will place mask just to the left of the
	// default mask position.
	XShift float64 `json:"x_shift"`

	// Shift by Y-axis measured in heights of the mask scaled to the face
	// size, from top to bottom.
	// For example, 1.0 will place the mask just below the default mask
	// position.
	YShift float64 `json:"y_shift"`

	// Mask scaling coefficient. For example, 2.0 means double size.
	Scale float64 `json:"scale"`
}
