// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sseclient

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

func (ev *Event) reset(id string) {
	ev.Type = EventTypeMessage
	ev.Data = ``
	ev.ID = id
}
