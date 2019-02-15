// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

//
// KeyPool maintain cached DKIM public keys.
//
type KeyPool struct {
	sync.Mutex
	pool map[string]*Key
}

//
// Clear the contents of key pool.
//
func (kp *KeyPool) Clear() {
	kp.Lock()
	for k := range kp.pool {
		delete(kp.pool, k)
	}
	kp.Unlock()
}

//
// Get cached DKIM key from pool or lookup using DNS/TXT method if not exist.
//
func (kp *KeyPool) Get(dname string) (key *Key, err error) {
	if len(dname) == 0 {
		return nil, nil
	}

	kp.Lock()
	key, ok := kp.pool[dname]
	if ok {
		if !key.IsExpired() {
			kp.Unlock()
			return key, nil
		}
	}
	kp.Unlock()

	key, err = LookupKey(QueryMethod{}, dname)
	if err != nil {
		return nil, err
	}

	kp.Put(dname, key)

	return key, nil
}

//
// Put key to pool based on DKIM domain name ("d=" value plus "s=" value).
//
func (kp *KeyPool) Put(dname string, key *Key) {
	if len(dname) == 0 || key == nil {
		return
	}
	kp.Lock()
	kp.pool[dname] = key
	kp.Unlock()
}

//
// String return text representation of DKIM key inside pool sorted by domain
// name.  Each key is printed with the following format:
// "{DomainName:ExpiredAt}"
//
func (kp *KeyPool) String() string {
	var sb strings.Builder
	kp.Lock()

	dnames := make([]string, 0, len(kp.pool))
	for k := range kp.pool {
		dnames = append(dnames, k)
	}

	sort.Strings(dnames)

	sb.WriteByte('[')
	for _, v := range dnames {
		key := kp.pool[v]
		sb.WriteByte('{')
		sb.WriteString(v)
		sb.WriteByte(' ')
		sb.WriteString(strconv.FormatInt(key.ExpiredAt, 10))
		sb.WriteByte('}')
	}
	sb.WriteByte(']')

	kp.Unlock()
	return sb.String()
}
