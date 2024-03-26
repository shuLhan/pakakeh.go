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
	"log"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
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

	// zone contains internal zones, with its origin as the key.
	zone map[string]*Zone

	// lru contains list of external answers, ordered by access time in
	// ascending order (the least recently used, LRU, record will be on
	// the top).
	lru *list.List

	debug int

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
func (c *Caches) init(pruneDelay, pruneThreshold time.Duration, debug int) {
	if pruneDelay.Minutes() < 1 {
		pruneDelay = time.Hour
	}
	if pruneThreshold.Minutes() > -1 {
		pruneThreshold = -1 * time.Hour
	}

	c.internal = make(map[string]*answers)
	c.external = make(map[string]*answers)
	c.zone = make(map[string]*Zone)
	c.lru = list.New()
	c.debug = debug

	go c.worker(pruneDelay, pruneThreshold)
}

// ExternalClear remove all external answers.
func (c *Caches) ExternalClear() (listAnswer []*Answer) {
	listAnswer = c.prune(math.MaxInt64)
	return listAnswer
}

// ExternalLoad the gob encoded external answers from r.
func (c *Caches) ExternalLoad(r io.Reader) (answers []*Answer, err error) {
	var (
		logp = "Caches.ExternalLoad"

		answer *Answer
	)

	answers, err = c.read(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	for _, answer = range answers {
		_ = c.upsert(answer)
	}
	return answers, nil
}

// ExternalLRU return list of external caches ordered by the least recently
// used.
func (c *Caches) ExternalLRU() (answers []*Answer) {
	var (
		e *list.Element
	)

	c.Lock()
	defer c.Unlock()

	for e = c.lru.Front(); e != nil; e = e.Next() {
		answers = append(answers, e.Value.(*Answer))
	}
	return answers
}

// externalRemoveName remove an external answers by domain name.
// It will return nil if qname is not exist in the caches.
func (c *Caches) externalRemoveName(qname string) (listAnswer []*Answer) {
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

// ExternalRemoveNames remove external caches by domain names.
func (c *Caches) ExternalRemoveNames(names []string) (listAnswer []*Answer) {
	var (
		answers []*Answer
		name    string
	)
	for _, name = range names {
		answers = c.externalRemoveName(name)
		if len(answers) > 0 {
			listAnswer = append(listAnswer, answers...)
			if c.debug&DebugLevelCache != 0 {
				log.Println(`dns: - `, name)
			}
		}
	}
	return listAnswer
}

// ExternalSave write all external answers into w, encoded with gob.
// On success, it returns the number of answers written to w.
func (c *Caches) ExternalSave(w io.Writer) (n int, err error) {
	var (
		logp    = "ExternalSave"
		answers = c.ExternalLRU()
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

// ExternalSearch search external answers where domain name match with regular
// expression.
func (c *Caches) ExternalSearch(re *regexp.Regexp) (listMsg []*Message) {
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

func (c *Caches) query(msg *Message) (an *Answer) {
	var ans *answers

	c.Lock()

	ans = c.internal[msg.Question.Name]
	if ans == nil {
		ans = c.external[msg.Question.Name]
		if ans == nil {
			goto out
		}
	}

	an, _ = ans.get(msg.Question.Type, msg.Question.Class)
	if an == nil {
		goto out
	}

	// Move the answer to the back of LRU if its external answer and
	// update its accessed time.
	if an.ReceivedAt > 0 {
		c.lru.MoveToBack(an.el)
		an.AccessedAt = timeNow().Unix()
	}

out:
	c.Unlock()

	if an == nil {
		// No answers found in internal and external caches.
		// If the requested domain is subset of our internal
		// zone, return answer with error and Authority.
		var zone = c.internalZone(msg.Question.Name)
		if zone == nil {
			return nil
		}
		an = &Answer{
			msg: msg,
		}
		_ = an.msg.AddAuthority(zone.soaRecord())
		an.msg.SetResponseCode(RCodeErrName)
	}
	return an
}

// internalZone will return the zone if the query name is suffix of one of
// the Zone Origin.
func (c *Caches) internalZone(qname string) (zone *Zone) {
	qname = toDomainAbsolute(qname)
	for _, zone = range c.zone {
		if strings.HasSuffix(qname, `.`+zone.Origin) {
			return zone
		}
	}
	return nil
}

// InternalPopulate add list of message to internal caches.
func (c *Caches) InternalPopulate(msgs []*Message, from string) {
	var (
		isLocal = true

		msg      *Message
		an       *Answer
		n        int
		inserted bool
	)

	for _, msg = range msgs {
		an = newAnswer(msg, isLocal)
		inserted = c.upsert(an)
		if inserted {
			n++
		}
	}

	if c.debug&DebugLevelCache != 0 {
		log.Printf(`dns: %d out of %d records cached from %q`, n, len(msgs), from)
	}
}

// InternalPopulateRecords update or insert new ResourceRecord into internal
// caches.
func (c *Caches) InternalPopulateRecords(listRR []*ResourceRecord, from string) (err error) {
	var (
		rr *ResourceRecord
		n  int
	)

	for _, rr = range listRR {
		err = c.internalUpsertRecord(rr)
		if err != nil {
			return err
		}
		n++
	}
	if c.debug&DebugLevelCache != 0 {
		log.Printf(`dns: %d out of %d records cached from %q`, n, len(listRR), from)
	}
	return nil
}

// InternalPopulateZone populate the internal caches from Zone.
func (c *Caches) InternalPopulateZone(zone *Zone) {
	if zone == nil {
		return
	}
	if len(zone.Origin) == 0 {
		return
	}
	c.zone[zone.Origin] = zone
	c.InternalPopulate(zone.Messages(), zone.Path)
}

// InternalRemoveNames remove internal caches by domain names.
func (c *Caches) InternalRemoveNames(names []string) {
	var (
		x int
	)

	c.Lock()
	defer c.Unlock()

	for ; x < len(names); x++ {
		delete(c.internal, names[x])
		if c.debug&DebugLevelCache != 0 {
			log.Println(`dns: - `, names[x])
		}
	}
}

// InternalRemoveRecord remove the answer from caches by ResourceRecord name, type,
// class, and value.
func (c *Caches) InternalRemoveRecord(rr *ResourceRecord) (rrOut *ResourceRecord, err error) {
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

// internalUpsertRecord update or insert new answer by RR.
//
// First, it will check if the answer already exist in cache.
// If it not exist, the new message and answer will created and inserted to
// cached.
// If its exist, it will add or replace the existing RR in the message
// (dependes on RR type).
func (c *Caches) internalUpsertRecord(rr *ResourceRecord) (err error) {
	err = rr.initAndValidate()
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	var (
		rrName = strings.TrimSuffix(rr.Name, `.`)
		ans    = c.internal[rrName]

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
		c.internal[rrName] = ans
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

		if c.debug&DebugLevelCache != 0 {
			log.Printf(`dns: - 0:%s`, answer.msg.Question.String())
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

		msg, err = UnpackMessage(item.Packet)
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

// upsert update or insert answer in caches.
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

// worker for pruning unused caches.
//
// The worker prune process will run based on prune delay and it will remove
// any cached answer that has not been accessed less than prune threshold
// value.
func (c *Caches) worker(pruneDelay, pruneThreshold time.Duration) {
	var (
		pruneTimer = time.NewTimer(pruneDelay)

		now        time.Time
		listAnswer []*Answer
		exp        int64
	)

	for now = range pruneTimer.C {
		exp = now.Add(pruneThreshold).Unix()
		listAnswer = c.prune(exp)
		if c.debug&DebugLevelCache != 0 {
			log.Printf(`dns: pruning %d records from cache`, len(listAnswer))
		}
	}
}
