// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import (
	"sort"
	"strings"
)

// Metadata provides versioned Map with stable key order.
type Metadata struct {
	vals    map[string]string
	version int
}

// NewMetadata create and initialize new metadata.
func NewMetadata() (md *Metadata) {
	md = &Metadata{
		vals: map[string]string{},
	}
	return md
}

// Delete Metadata by its key.
// The versioning will be increased only if the key exist.
func (md *Metadata) Delete(key string) {
	var ok bool
	_, ok = md.vals[key]
	if ok {
		delete(md.vals, key)
		md.version++
	}
}

// Get the Metadata value by its key.
func (md *Metadata) Get(key string) string {
	return md.vals[key]
}

// Keys return the Metadata keys sorted lexicographically.
func (md *Metadata) Keys() (keys []string) {
	var key string
	for key = range md.vals {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// KeysMap return the Metadata keys sorted lexicographically and its map.
func (md *Metadata) KeysMap() (keys []string, vals map[string]string) {
	keys = md.Keys()
	return keys, md.vals
}

// Set store the key with value into Metadata.
// This method always increase the version.
func (md *Metadata) Set(key, value string) {
	md.version++
	md.vals[key] = value
}

// String return the Metadata where each item separated by comma and the
// key-value separated by equal character.
func (md *Metadata) String() string {
	var keys = md.Keys()

	var (
		sb  strings.Builder
		key string
		val string
		x   int
	)
	for x, key = range keys {
		if x > 0 {
			sb.WriteByte(',')
		}

		val = md.vals[key]

		sb.WriteString(key)
		sb.WriteByte('=')
		sb.WriteString(val)
	}
	return sb.String()
}

// Version return the current version of Metadata.
func (md *Metadata) Version() int {
	return md.version
}
