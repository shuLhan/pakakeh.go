// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/shuLhan/share/lib/debug"
)

//
// caches of DNS answers.
//
type caches struct {
	sync.Mutex

	// v contains mapping of DNS question name (a domain name) with their
	// list of answer.
	v map[string]*answers

	// lru represent list of non local answers, ordered based on answer
	// access time in ascending order.
	lru *list.List

	// pruneDelay define a delay where caches will be pruned.
	// Default to 1 hour.
	pruneDelay time.Duration

	// pruneThreshold define negative duration where answers will be
	// pruned from caches.
	// Default to -1 hour.
	pruneThreshold time.Duration
}

//
// newCaches create new in memory caches with specific prune delay and
// threshold.
// The prune delay MUST be greater than 1 minute or it will set to 1 hour.
// The prune threshold MUST be greater than -1 minute or it will be set to -1
// hour.
//
func newCaches(pruneDelay, pruneThreshold time.Duration) (ca *caches) {
	if pruneDelay.Minutes() < 1 {
		pruneDelay = time.Hour
	}
	if pruneThreshold.Minutes() > -1 {
		pruneThreshold = -1 * time.Hour
	}

	ca = &caches{
		v:              make(map[string]*answers),
		lru:            list.New(),
		pruneDelay:     pruneDelay,
		pruneThreshold: pruneThreshold,
	}

	go ca.startWorker()

	return
}

//
// get an answer from cache based on domain-name, query type, and query class.
//
// If query name exist but the query type or class does not exist,
// it will return list of answer and nil answer.
//
// If answer exist on cache, their accessed time will be updated to current
// time and moved to back of LRU to prevent being pruned later.
//
func (c *caches) get(qname string, qtype, qclass uint16) (ans *answers, an *answer) {
	c.Lock()

	var found bool

	ans, found = c.v[qname]
	if found {
		an, _ = ans.get(qtype, qclass)
		if an != nil {
			// Move the answer to the back of LRU if its not
			// local and update its accessed time.
			if an.receivedAt > 0 {
				c.lru.MoveToBack(an.el)
				an.accessedAt = time.Now().Unix()
			}
		}
	}

	c.Unlock()
	return
}

//
// list return all answers in LRU.
//
func (c *caches) list() (list []*answer) {
	c.Lock()
	for e := c.lru.Front(); e != nil; e = e.Next() {
		list = append(list, e.Value.(*answer))
	}
	c.Unlock()
	return
}

//
// prune will remove old answers on caches based on accessed time.
//
func (c *caches) prune() {
	c.Lock()

	exp := time.Now().Add(c.pruneThreshold).Unix()

	e := c.lru.Front()
	for e != nil {
		an := e.Value.(*answer)
		if an.accessedAt > exp {
			break
		}

		if debug.Value >= 1 {
			fmt.Printf("dns: - 0:%s\n", an.msg.Question.String())
		}

		next := e.Next()
		_ = c.lru.Remove(e)
		c.remove(an)

		e = next
	}

	c.Unlock()
}

//
// remove an answer from caches.
//
func (c *caches) remove(an *answer) {
	answers, found := c.v[an.qname]
	if found {
		answers.remove(an.qtype, an.qclass)
	}
	an.clear()
}

//
// upsert update or insert answer to caches.  If the answer is inserted to
// caches it will return true, otherwise when its updated it will return
// false.
//
func (c *caches) upsert(nu *answer) (inserted bool) {
	if nu == nil || nu.msg == nil {
		return
	}

	c.Lock()

	answers, found := c.v[nu.qname]
	if !found {
		inserted = true
		c.v[nu.qname] = newAnswers(nu)
		if nu.receivedAt > 0 {
			nu.el = c.lru.PushBack(nu)
		}
	} else {
		an := answers.upsert(nu)
		if an == nil {
			inserted = true
			if nu.receivedAt > 0 {
				// Push the new answer to LRU if new answer is
				// not local and its inserted to list.
				nu.el = c.lru.PushBack(nu)
			}
		}
	}

	c.Unlock()

	return inserted
}

//
// startWorker start the worker pruning process.
//
// The worker prune process will run based on prune delay and it will remove
// any cached answer that has not been accessed less than prune threshold
// value.
//
func (c *caches) startWorker() {
	ticker := time.NewTicker(c.pruneDelay)

	for t := range ticker.C {
		fmt.Printf("dns: pruning at %v\n", t)

		c.prune()
	}
}
