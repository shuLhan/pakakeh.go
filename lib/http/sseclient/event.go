// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sseclient

import "strconv"

// List of system event type.
const (
	// EventTypeOpen is set when connection succesfully established.
	// The passed [Event.Data] is empty.
	EventTypeOpen = `open`

	// EventTypeMessage is set when client received message from server,
	// possibly with new ID.
	EventTypeMessage = `message`

	EventTypeError = `error`
)

// Event contains SSE message from server or client status.
type Event struct {
	Type string
	Data string
	ID   string
}

// IDInt return the ID as int64.
// If the ID cannot be converted to integer it would return 0.
func (ev *Event) IDInt() (id int64) {
	id, _ = strconv.ParseInt(ev.ID, 10, 64)
	return id
}

func (ev *Event) reset(id string) {
	ev.Type = EventTypeMessage
	ev.Data = ``
	ev.ID = id
}
