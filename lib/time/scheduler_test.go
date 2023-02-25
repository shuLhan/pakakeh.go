// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestScheduler_minutely(t *testing.T) {
	type testCase struct {
		schedule string
		now      time.Time
		exp      Scheduler
	}

	var cases = []testCase{{
		now: time.Date(2013, time.January, 30, 14, 26, 59, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindMinutely,
			next:        time.Date(2013, time.January, 30, 14, 27, 0, 0, time.UTC),
			nextSeconds: 1,
		},
	}, {
		now: time.Date(2013, time.January, 30, 14, 27, 1, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindMinutely,
			next:        time.Date(2013, time.January, 30, 14, 28, 0, 0, time.UTC),
			nextSeconds: 59,
		},
	}}

	var (
		c   testCase
		got *Scheduler
		err error
	)

	for _, c = range cases {
		got, err = newScheduler(c.schedule, c.now)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, `NewScheduler`, c.exp, *got)
	}
}

func TestNewScheduler_hourly(t *testing.T) {
	type testCase struct {
		schedule string
		now      time.Time
		exp      Scheduler
	}

	var cases = []testCase{{
		schedule: `hourly`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindHourly,
			next:        time.Date(2013, time.January, 20, 15, 0, 0, 0, time.UTC),
			nextSeconds: 2040,
		},
	}, {
		schedule: `hourly@5,11,-1,55`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindHourly,
			next:        time.Date(2013, time.January, 20, 14, 55, 0, 0, time.UTC),
			minutes:     []int{5, 11, 55},
			nextSeconds: 1740,
		},
	}}

	var (
		c   testCase
		got *Scheduler
		err error
	)

	for _, c = range cases {
		got, err = newScheduler(c.schedule, c.now)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.schedule, c.exp, *got)
	}
}

func TestNewScheduler_daily(t *testing.T) {
	type testCase struct {
		schedule string
		now      time.Time
		exp      Scheduler
	}

	var cases = []testCase{{
		schedule: `daily`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindDaily,
			next:        time.Date(2013, time.January, 21, 0, 0, 0, 0, time.UTC),
			nextSeconds: 34440,
			tod: []Clock{
				Clock{},
			},
		},
	}, {
		schedule: `daily@00:15,06:16,12:99,24:15`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindDaily,
			next:        time.Date(2013, time.January, 21, 0, 15, 0, 0, time.UTC),
			nextSeconds: 35340,
			tod: []Clock{
				{hour: 0, min: 15},
				{hour: 0, min: 15},
				{hour: 6, min: 16},
				{hour: 12, min: 0},
			},
		},
	}}

	var (
		c   testCase
		got *Scheduler
		err error
	)

	for _, c = range cases {
		got, err = newScheduler(c.schedule, c.now)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.schedule, c.exp, *got)
	}
}

func TestNewScheduler_weekly(t *testing.T) {
	type testCase struct {
		schedule string
		now      time.Time
		exp      Scheduler
	}

	var cases = []testCase{{
		schedule: `weekly`,
		// The Weekday is Sunday.
		now: time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindWeekly,
			next:        time.Date(2013, time.January, 27, 0, 0, 0, 0, time.UTC),
			nextSeconds: 552840,
			tod: []Clock{
				Clock{},
			},
			dow: []int{0},
		},
	}, {
		schedule: `weekly@@00:15,06:16,12:99,24:15`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindWeekly,
			next:        time.Date(2013, time.January, 27, 0, 15, 0, 0, time.UTC),
			nextSeconds: 553740,
			dow:         []int{0},
			tod: []Clock{
				{hour: 0, min: 15},
				{hour: 0, min: 15},
				{hour: 6, min: 16},
				{hour: 12, min: 0},
			},
		},
	}, {
		schedule: `weekly@Sunday,Mon@00:15,06:16`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindWeekly,
			next:        time.Date(2013, time.January, 21, 0, 15, 0, 0, time.UTC),
			nextSeconds: 35340,
			dow:         []int{0, 1},
			tod: []Clock{
				{hour: 0, min: 15},
				{hour: 6, min: 16},
			},
		},
	}, {
		schedule: `weekly@Sunday,Mon@00:15,06:16`,
		// Sunday 21:00
		now: time.Date(2013, time.January, 20, 21, 0, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindWeekly,
			next:        time.Date(2013, time.January, 21, 0, 15, 0, 0, time.UTC), // Monday 00:15
			nextSeconds: 11700,
			dow:         []int{0, 1},
			tod: []Clock{
				{hour: 0, min: 15},
				{hour: 6, min: 16},
			},
		},
	}, {
		schedule: `weekly@Sunday,Mon@00:15,06:16`,
		// Monday 21:00
		now: time.Date(2013, time.January, 21, 21, 0, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindWeekly,
			next:        time.Date(2013, time.January, 27, 0, 15, 0, 0, time.UTC), // Sunday 00:15
			nextSeconds: 443700,
			dow:         []int{0, 1},
			tod: []Clock{
				{hour: 0, min: 15},
				{hour: 6, min: 16},
			},
		},
	}, {
		schedule: `weekly@Fri,Sat@00:15,06:16`,
		// Saturday 21:00
		now: time.Date(2013, time.January, 26, 21, 0, 0, 0, time.UTC),
		exp: Scheduler{
			kind:        ScheduleKindWeekly,
			next:        time.Date(2013, time.February, 1, 0, 15, 0, 0, time.UTC), // Friday 00:15
			nextSeconds: 443700,
			dow:         []int{5, 6},
			tod: []Clock{
				{hour: 0, min: 15},
				{hour: 6, min: 16},
			},
		},
	}}

	var (
		c   testCase
		got *Scheduler
		err error
	)

	for _, c = range cases {
		got, err = newScheduler(c.schedule, c.now)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.schedule, c.exp, *got)
	}
}

func TestNewScheduler_monthly(t *testing.T) {
	type testCase struct {
		schedule string
		now      time.Time
		exp      Scheduler
	}

	var cases = []testCase{{
		schedule: `monthly`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC), // The Weekday is Sunday.
		exp: Scheduler{
			kind:        ScheduleKindMonthly,
			next:        time.Date(2013, time.February, 1, 0, 0, 0, 0, time.UTC),
			nextSeconds: 984840,
			tod: []Clock{
				Clock{},
			},
			dom: []int{1},
		},
	}, {
		schedule: `monthly@15,31@00:15`,
		now:      time.Date(2013, time.January, 20, 14, 26, 0, 0, time.UTC), // The Weekday is Sunday.
		exp: Scheduler{
			kind:        ScheduleKindMonthly,
			next:        time.Date(2013, time.January, 31, 0, 15, 0, 0, time.UTC),
			nextSeconds: 899340,
			tod: []Clock{
				Clock{min: 15},
			},
			dom: []int{15, 31},
		},
	}, {
		schedule: `monthly@15,31@00:15`,
		// 2013-02-15 01:00
		now: time.Date(2013, time.February, 15, 1, 0, 0, 0, time.UTC),
		exp: Scheduler{
			kind: ScheduleKindMonthly,
			// 2013-03-15 01:00
			next:        time.Date(2013, time.March, 15, 0, 15, 0, 0, time.UTC),
			nextSeconds: 2416500,
			tod: []Clock{
				Clock{min: 15},
			},
			dom: []int{15, 31},
		},
	}}

	var (
		c   testCase
		got *Scheduler
		err error
	)
	for _, c = range cases {
		got, err = newScheduler(c.schedule, c.now)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, c.schedule, c.exp, *got)
	}
}
