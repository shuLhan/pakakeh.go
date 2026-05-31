// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package time

import (
	"log"
	"sync"
	"time"
)

var newWorkerOnce = sync.OnceValue(func() (wrk *worker) {
	wrk = newWorker()
	return wrk
})

// worker define internal worker for [time.Scheduler].
// Its run a ticker, rounded by minute, that publish the current [time.Time]
// to all subscribers.
//
// The original idea is to fix the Scheduler not running as scheduled when the
// host system is suspended or sleep.
type worker struct {
	qstop chan struct{}
	subs  map[int64]chan<- time.Time
	mtx   sync.Mutex
}

func newWorker() (wrk *worker) {
	wrk = &worker{
		subs: make(map[int64]chan<- time.Time),
	}
	return wrk
}

func (wrk *worker) start() {
	now := NowFunc().Round(time.Second).UTC()
	wait := 60 - now.Second()
	tick := time.NewTicker(time.Duration(wait) * time.Second)
	for {
		select {
		case now = <-tick.C:
			now = now.Round(time.Second).UTC()
			wrk.mtx.Lock()
			for id, q := range wrk.subs {
				select {
				case q <- now:
				default:
					log.Printf(`worker: send failed on id %d`, id)
				}
			}
			wrk.mtx.Unlock()
			wait = 60 - now.Second()
			tick.Reset(time.Duration(wait) * time.Second)

		case <-wrk.qstop:
			tick.Stop()
			wrk.qstop = nil
			return
		}
	}
}

func (wrk *worker) stop() {
	wrk.mtx.Lock()
	close(wrk.qstop)
	wrk.mtx.Unlock()
}

// reg registers a channel that will receive [time.Time] every minute.
// Reg returns a unique ID that can be used for [worker.unreg] or panic if
// its cannot allocated unique ID.
func (wrk *worker) reg(q chan<- time.Time) (id int64) {
	wrk.mtx.Lock()

	ntry := 0
	for {
		id = NowFunc().UTC().UnixNano()
		_, ok := wrk.subs[id]
		if !ok {
			break
		}
		ntry++
		if ntry > 5 {
			wrk.mtx.Unlock()
			panic(`worker: failed to register after 5 attempts`)
		}
		time.Sleep(10 * time.Millisecond)
	}

	wrk.subs[id] = q
	if wrk.qstop == nil {
		wrk.qstop = make(chan struct{}, 1)
		go wrk.start()
	}
	wrk.mtx.Unlock()

	return id
}

// unreg unregisters the previous registered channel.
func (wrk *worker) unreg(id int64) {
	wrk.mtx.Lock()
	delete(wrk.subs, id)
	if len(wrk.subs) == 0 {
		close(wrk.qstop)
	}
	wrk.mtx.Unlock()
}
