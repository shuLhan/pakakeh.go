// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"container/list"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/shuLhan/share/lib/debug"
)

const (
	cachesFileFormatV1 = 1
)

// Caches of DNS answers.
//
// There are two type of answer: internal and external.
// Internal answer is a DNS record that is loaded from hosts or zone files.
// Internal answer never get pruned.
// External answer is a DNS record that is received from parent name
// servers.
type Caches struct {
	// internal contains list of internal answers loaded from hosts or
	// zone files, indexed by its domain name.
	internal map[string]*answers

	// external contains list of answers from parent name servers, indexed
	// by is domain name.
	external map[string]*answers

	// lru contains list of external answers, ordered by access time in
	// ascending order (the least recently used, LRU, record will be on
	// the top).
	lru *list.List

	sync.Mutex
}

// cachesFileHeader define the file header when storing caches on storage.
type cachesFileHeader struct {
	Version int
}

// cachesFileV1 contains the format of DNS message to be stored on file.
type cachesFileV1 struct {
	// Packet contains the raw DNS message.
	Packet []byte

	// ReceivedAt contains time when message is received.
	ReceivedAt int64

	// AccessedAt contains time when message last accessed.
	AccessedAt int64
}

// init create new in memory caches with specific prune delay and
// threshold.
// The prune delay MUST be greater than 1 minute or it will set to 1 hour.
// The prune threshold MUST be greater than -1 minute or it will be set to -1
// hour.
func (c *Caches) init(pruneDelay, pruneThreshold time.Duration) {
	if pruneDelay.Minutes() < 1 {
		pruneDelay = time.Hour
	}
	if pruneThreshold.Minutes() > -1 {
		pruneThreshold = -1 * time.Hour
	}

	c.internal = make(map[string]*answers)
	c.external = make(map[string]*answers)
	c.lru = list.New()

	go c.worker(pruneDelay, pruneThreshold)
}

// get an answer based on domain-name, query type, and query class.
//
// If query name exist but the query type or class does not exist,
// it will return list of answer and nil answer.
//
// If answer exist on cache and its from external, their accessed time will be
// updated to current time and moved to back of LRU to prevent being pruned
// later.
func (c *Caches) get(qname string, rtype RecordType, rclass RecordClass) (ans *answers, an *Answer) {
	c.Lock()
	defer c.Unlock()

	ans, _ = c.internal[qname]
	if ans == nil {
		ans, _ = c.external[qname]
		if ans == nil {
			return nil, nil
		}
	}

	an, _ = ans.get(rtype, rclass)
	if an == nil {
		return ans, nil
	}

	// Move the answer to the back of LRU if its external
	// answer and update its accessed time.
	if an.ReceivedAt > 0 {
		c.lru.MoveToBack(an.el)
		an.AccessedAt = time.Now().Unix()
	}

	return ans, an
}

// list return all external answers in least-recently-used order.
func (c *Caches) list() (answers []*Answer) {
	var (
		e *list.Element
	)

	c.Lock()
	defer c.Unlock()

	for e = c.lru.Front(); e != nil; e = e.Next() {
		answers = append(answers, e.Value.(*Answer))
	}
	return
}

// prune old, external answers that have access time less or equal than
// expired "exp" time.
func (c *Caches) prune(exp int64) (listAnswer []*Answer) {
	var (
		el, next *list.Element
		answer   *Answer
		answers  *answers
	)

	c.Lock()
	defer c.Unlock()

	el = c.lru.Front()
	for el != nil {
		answer = el.Value.(*Answer)
		if answer.AccessedAt > exp {
			break
		}

		if debug.Value >= 1 {
			fmt.Printf("dns: - 0:%s\n", answer.msg.Question.String())
		}

		next = el.Next()
		_ = c.lru.Remove(el)
		answers = c.external[answer.QName]
		if answers != nil {
			answers.remove(answer.RType, answer.RClass)
			if len(answers.v) == 0 {
				delete(c.external, answer.QName)
			}
		}
		answer.clear()

		listAnswer = append(listAnswer, answer)

		el = next
	}
	return listAnswer
}

// read external caches stored on storage r.
func (c *Caches) read(r io.Reader) (answers []*Answer, err error) {
	var (
		logp   = "Caches.read"
		header = &cachesFileHeader{}
		dec    = gob.NewDecoder(r)

		item   *cachesFileV1
		msg    *Message
		answer *Answer
	)

	err = dec.Decode(header)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if header.Version != cachesFileFormatV1 {
		return nil, fmt.Errorf("%s: unknown version %d", logp, header.Version)
	}

	for {
		item = &cachesFileV1{}
		err = dec.Decode(item)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		msg = NewMessage()
		msg.packet = item.Packet
		err = msg.Unpack()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		answer = newAnswer(msg, false)
		answer.ReceivedAt = item.ReceivedAt
		answer.AccessedAt = item.AccessedAt

		answers = append(answers, answer)
	}

	return answers, nil
}

// remove an external answers by query name.
// It will return nil if qname is not exist in the caches.
func (c *Caches) remove(qname string) (listAnswer []*Answer) {
	var (
		an  *Answer
		ans *answers
	)

	c.Lock()
	defer c.Unlock()

	ans = c.external[qname]
	if ans == nil {
		return nil
	}

	for _, an = range ans.v {
		c.lru.Remove(an.el)
		an.clear()
	}
	listAnswer = ans.v
	ans.v = nil

	return listAnswer
}

// removeInternalByRR remove internal cache by its record name, type, class,
// and value.
func (c *Caches) removeInternalByRR(rr *ResourceRecord) (rrOut *ResourceRecord, err error) {
	var (
		ans *answers
		an  *Answer
	)

	c.Lock()
	defer c.Unlock()

	ans = c.internal[rr.Name]
	if ans == nil {
		return nil, nil
	}
	for _, an = range ans.v {
		if an.RType != rr.Type {
			continue
		}
		if an.RClass != rr.Class {
			continue
		}
		rrOut, err = an.msg.RemoveAnswer(rr)
		break
	}
	return rrOut, err
}

// search external answers that match with regular expression.
func (c *Caches) search(re *regexp.Regexp) (listMsg []*Message) {
	var (
		an    *Answer
		ans   *answers
		dname string
	)

	c.Lock()
	defer c.Unlock()

	for dname, ans = range c.external {
		if re.MatchString(dname) {
			for _, an = range ans.v {
				listMsg = append(listMsg, an.msg)
			}
		}
	}

	return listMsg
}

// upsert update or insert answer to external caches.
//
// If the answer is inserted it will return true, otherwise it will return
// false.
func (c *Caches) upsert(nu *Answer) (inserted bool) {
	if nu == nil || nu.msg == nil {
		return
	}

	var (
		answers *answers
		an      *Answer
	)

	c.Lock()
	defer c.Unlock()

	if nu.ReceivedAt == 0 {
		answers = c.internal[nu.QName]
		if answers == nil {
			answers = newAnswers(nu)
			c.internal[nu.QName] = answers
			inserted = true
		} else {
			an = answers.upsert(nu)
			if an == nil {
				inserted = true
			}
		}
	} else {
		answers = c.external[nu.QName]
		if answers == nil {
			answers = newAnswers(nu)
			c.external[nu.QName] = answers
			inserted = true
		} else {
			an = answers.upsert(nu)
			if an == nil {
				inserted = true
			}
		}
		if inserted {
			// Push the new answer to LRU if new answer is
			// external and its inserted to list.
			nu.el = c.lru.PushBack(nu)
		}
	}

	return inserted
}

// upsertInternalRR update or insert new answer by RR.
//
// First, it will check if the answer already exist in cache.
// If it not exist, the new message and answer will created and inserted to
// cached.
// If its exist, it will add or replace the existing RR in the message
// (dependes on RR type).
func (c *Caches) upsertInternalRR(rr *ResourceRecord) (err error) {
	err = rr.initAndValidate()
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	var (
		ans *answers = c.internal[rr.Name]

		an  *Answer
		msg *Message
	)

	if ans == nil {
		msg, err = NewMessageFromRR(rr)
		if err != nil {
			return err
		}
		an = newAnswer(msg, true)
		ans = newAnswers(an)
		c.internal[rr.Name] = ans
		return nil
	}

	an, _ = ans.get(rr.Type, rr.Class)
	if an == nil {
		// The domain name is already exist, but without the RR type.
		msg, err = NewMessageFromRR(rr)
		if err != nil {
			return err
		}

		an = newAnswer(msg, true)
		ans.v = append(ans.v, an)
		return nil
	}

	return an.msg.AddAnswer(rr)
}

// worker for pruning unused caches.
//
// The worker prune process will run based on prune delay and it will remove
// any cached answer that has not been accessed less than prune threshold
// value.
func (c *Caches) worker(pruneDelay, pruneThreshold time.Duration) {
	var (
		ticker = time.NewTicker(pruneDelay)

		listAnswer []*Answer
		exp        int64
	)

	for range ticker.C {
		exp = time.Now().Add(pruneThreshold).Unix()
		listAnswer = c.prune(exp)
		fmt.Printf("dns: pruning %d records from cache\n", len(listAnswer))
	}
}

// write all external answers to w.
// On success, it returns the number of answers written to w.
func (c *Caches) write(w io.Writer) (n int, err error) {
	var (
		logp    = "Caches.write"
		answers = c.list()
		header  = &cachesFileHeader{
			Version: cachesFileFormatV1,
		}
		enc = gob.NewEncoder(w)

		answer *Answer
		item   *cachesFileV1
	)

	err = enc.Encode(header)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", logp, err)
	}

	for _, answer = range answers {
		item = &cachesFileV1{
			ReceivedAt: answer.ReceivedAt,
			AccessedAt: answer.AccessedAt,
			Packet:     answer.msg.packet,
		}
		err = enc.Encode(item)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", logp, err)
		}
		n++
	}

	return n, nil
}
