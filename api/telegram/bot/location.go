// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Location represents a point on the map.
type Location struct {
	Longitude float64 `json:"longitude"` // Longitude as defined by sender.
	Latitude  float64 `json:"latitude"`  // Latitude as defined by sender.
}
