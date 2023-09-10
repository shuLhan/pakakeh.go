// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// Venue represents a venue.
type Venue struct {
	// Name of the venue.
	Title string `json:"title"`

	// Address of the venue.
	Address string `json:"address"`

	// Optional. Foursquare identifier of the venue.
	FoursquareID string `json:"foursquare_id"`

	// Optional. Foursquare type of the venue. (For example,
	// “arts_entertainment/default”, “arts_entertainment/aquarium” or
	// “food/icecream”).
	FoursquareType string `json:"foursquare_type"`

	// Venue location.
	Location Location `json:"location"`
}
