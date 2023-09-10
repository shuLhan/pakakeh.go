// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import "sync"

// keys contains and maintains list of public Keys and its configuration.
type keys struct {
	v map[string]Key

	sync.Mutex
}

func newKeys() *keys {
	return &keys{
		v: make(map[string]Key),
	}
}

func (p *keys) upsert(k Key) {
	p.Lock()
	p.v[k.ID] = k
	p.Unlock()
}

func (p *keys) get(id string) (k Key, ok bool) {
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
