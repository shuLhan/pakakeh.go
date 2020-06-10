// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// Location represents a point on the map.
type Location struct {
	Longitude float64 `json:"longitude"` // Longitude as defined by sender.
	Latitude  float64 `json:"latitude"`  // Latitude as defined by sender.
}
