// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package time

import (
	"testing"
	"testing/synctest"
	"time"

	"kilabit.info/pakakeh.go/lib/test"
)

func TestWorker_reg(t *testing.T) {
	NowFunc = func() time.Time {
		return time.Now()
	}

	listCase := []struct {
		name          string
		exp           string
		elapsedBefore time.Duration
		waitAfter     time.Duration
	}{{
		name:          `startAt1s`,
		elapsedBefore: 1 * time.Second,
		exp:           `2000-01-01 00:01:00 +0000 UTC`,
	}, {
		name:          `startAt30s`,
		elapsedBefore: 30 * time.Second,
		waitAfter:     31 * time.Second,
		exp:           `2000-01-01 00:01:00 +0000 UTC`,
	}, {
		name:          `startAt59s`,
		elapsedBefore: 59 * time.Second,
		waitAfter:     2 * time.Second,
		exp:           `2000-01-01 00:01:00 +0000 UTC`,
	}}
	for _, tcase := range listCase {
		synctest.Test(t, func(t *testing.T) {
			time.Sleep(tcase.elapsedBefore) // Set now to 2000-01-01 00:00:00 + Xs.
			synctest.Wait()

			wrk := newWorker()
			q := make(chan time.Time, 1)
			wrk.reg(q)

			got := <-q
			wrk.stop()
			synctest.Wait()

			got = got.Round(time.Minute)
			test.Assert(t, tcase.name, tcase.exp, got.String())
		})
	}
}

func TestWorker_multipleReg(t *testing.T) {
	NowFunc = func() time.Time {
		return time.Now()
	}

	listCase := []struct {
		name          string
		elapsedBefore time.Duration
	}{{
		name:          `startAt1s`,
		elapsedBefore: 1 * time.Second,
	}, {
		name:          `startAt30s`,
		elapsedBefore: 29 * time.Second,
	}, {
		name:          `startAt59s`,
		elapsedBefore: 29 * time.Second,
	}}

	listq := []chan time.Time{}
	listid := []int64{}
	synctest.Test(t, func(t *testing.T) {

		for _, tcase := range listCase {
			time.Sleep(tcase.elapsedBefore)
			synctest.Wait()
			wrk := newWorkerOnce()
			q := make(chan time.Time, 1)
			id := wrk.reg(q)
			listq = append(listq, q)
			listid = append(listid, id)
		}

		var got []string
		for _, q := range listq {
			gotTime := <-q
			got = append(got, gotTime.Round(time.Minute).String())
		}
		exp := []string{
			`2000-01-01 00:01:00 +0000 UTC`,
			`2000-01-01 00:01:00 +0000 UTC`,
			`2000-01-01 00:01:00 +0000 UTC`,
		}
		test.Assert(t, `multipleReg #1`, exp, got)

		got = nil
		for _, q := range listq {
			gotTime := <-q
			got = append(got, gotTime.String())
		}
		exp = []string{
			`2000-01-01 00:02:00 +0000 UTC`,
			`2000-01-01 00:02:00 +0000 UTC`,
			`2000-01-01 00:02:00 +0000 UTC`,
		}
		test.Assert(t, `multipleReg #2`, exp, got)

		// Unregistering all ids should stop the worker.

		wrk := newWorkerOnce()
		for _, id := range listid {
			wrk.unreg(id)
		}
		synctest.Wait() // For worker start goroutine stopped.

		var expqstop chan struct{}
		test.Assert(t, `qstop should be nil`, expqstop, wrk.qstop)
	})
}
