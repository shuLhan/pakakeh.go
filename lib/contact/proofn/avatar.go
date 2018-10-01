// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proofn

//
// Avatar define an URI to Proofn contact avatar image.
//
type Avatar struct {
	AvatarPathSmall  string `json:"avatarPathSmall,omitempty"`
	AvatarPathMedium string `json:"avatarPathMedium,omitempty"`
	AvatarPathLarge  string `json:"avatarPathLarge,omitempty"`
}
