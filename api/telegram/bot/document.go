// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// Document represents a general file (as opposed to photos, voice messages
// and audio files).
type Document struct {
	// Optional. Document thumbnail as defined by sender.
	Thumb *PhotoSize `json:"thumb"`

	// Identifier for this file, which can be used to download or reuse
	// the file.
	FileID string `json:"file_id"`

	// Unique identifier for this file, which is supposed to be the same
	// over time and for different bots. Can't be used to download or
	// reuse the file.
	FileUniqueID string `json:"file_unique_id"`

	// Optional. MIME type of the file as defined by sender.
	MimeType string `json:"mime_type"`

	// Optional. Original filename as defined by sender.
	FileName string `json:"file_name"`

	// Optional. File size.
	FileSize int `json:"file_size"`
}
