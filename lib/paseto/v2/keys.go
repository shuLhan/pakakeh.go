// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package pasetov2

import (
	"sync"

	"kilabit.info/pakakeh.go/lib/paseto"
)

// keys contains and maintains list of public Keys and its configuration.
type keys struct {
	v map[string]paseto.Peer

	sync.Mutex
}

func newKeys() *keys {
	return &keys{
		v: make(map[string]paseto.Peer),
	}
}

func (p *keys) upsert(k paseto.Peer) {
	p.Lock()
	p.v[k.ID] = k
	p.Unlock()
}

func (p *keys) get(id string) (k paseto.Peer, ok bool) {
	p.Lock()
	k, ok = p.v[id]
	p.Unlock()
	return
}

func (p *keys) delete(id string) {
	p.Lock()
	delete(p.v, id)
	p.Unlock()
}
