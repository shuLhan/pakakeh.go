// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/ints"
)

var (
	ErrScheduleUnknown error = errors.New(`unknown schedule`)
)

const (
	ScheduleKindMinutely = `minutely`
	ScheduleKindHourly   = `hourly`
	ScheduleKindDaily    = `daily`
	ScheduleKindWeekly   = `weekly`
	ScheduleKindMonthly  = `monthly`
)

// Scheduler is a timer that run periodically based on calendar or day time.
//
// A schedule is divided into monthly, weekly, daily, hourly, and minutely.
// An empty schedule is equal to minutely, a schedule that run every minute.
type Scheduler struct {
	next time.Time // The next schedule.

	c     chan time.Time
	cstop chan struct{}    // The channel to stop the scheduler.
	C     <-chan time.Time // The channel on which the schedule are delivered.

	kind string

	minutes []int   // Partial minutes in hourly schedule.
	tod     []Clock // List of time in daily schedule.
	dow     []int   // List of day in weekly schedule.
	dom     []int   // List of day in monthly schedule.

	nextSeconds int64
}

// NewScheduler create new Scheduler from string schedule.
// A schedule is divided into monthly, weekly, daily, hourly, and minutely.
// An empty schedule is equal to minutely.
//
// A monthly schedule can be divided into calendar day and a time, with the
// following format,
//
//	MONTHLY      = "monthly" ["@" DAY_OF_MONTH] ["@" TIME_OF_DAY]
//	DAY_OF_MONTH = [ 1-31 *("," 1-31) ]
//	TIME_OF_DAY  = [ TIME *("," TIME) ]
//	TIME         = "00:00"-"23:59"
//
// An empty DAY_OF_MONTH is equal to 1 or the first day.
// An empty TIME_OF_DAY is equal to midnight or 00:00.
// If registered DAY_OF_MONTH is not available in the current month, it will
// be skipped, for example "monthly@31" will not run in February.
// For example,
//
//   - monthly = monthly@1@00:00 = the first day of each month at 00:00.
//   - monthly@1,15@18:00 = on day 1 and 15 every month at 6 PM.
//
// A weekly schedule can be divided into day of week and a time, with the
// following format,
//
//	WEEKLY      = "weekly" ["@" LIST_DOW] ["@" TIME_OF_DAY]
//	LIST_DOW    = [ DAY_OF_WEEK *("," DAY_OF_WEEK) ]
//	DAY_OF_WEEK = "Sunday" / "Monday" / "Tuesday" / "Wednesday"
//	            / "Thursday" / "Friday" / "Saturday"
//
// The first day of the week or empty LIST_DOW is equal to Sunday.
//
// For example,
//   - weekly = weekly@Sunday@00:00 = every Sunday at 00:00.
//   - weekly@Sunday,Tuesday,Friday@15:00 = every Sunday, Tuesday, and Friday
//     on each week at 3 PM.
//
// A daily schedule can be divided only into time.
//
//	DAILY = "daily" ["@" TIME_OF_DAY]
//
// For example,
//   - daily = daily@00:00 = every day at 00:00.
//   - daily@00:00,06:00,12:00,18:00 = every day at midnight, 6 AM, and 12 PM.
//
// A hourly schedule can be divided into minutes, with the following format,
//
//	HOURLY  = "hourly" ["@" MINUTES]
//	MINUTES = [ 0-59 *("," 0-59) ]
//
// An empty MINUTES is equal to 0.
// For example,
//   - hourly = hourly@0 = every hour at minute 0.
//   - hourly@0,15,30,45 = on minutes 0, 15, 30, 45 every hour.
//
// A minutely schedule run every minute, with the following format,
//
//	MINUTELY = "minutely"
//
// For example,
//   - minutely = every minute
func NewScheduler(schedule string) (sch *Scheduler, err error) {
	schedule = strings.ToLower(schedule)

	var (
		logp = `NewScheduler`
		list = strings.Split(schedule, `@`)
		c    = make(chan time.Time, 1)

		v string
	)

	sch = &Scheduler{
		C:     c,
		c:     c,
		cstop: make(chan struct{}, 1),
		kind:  list[0],
	}

	switch sch.kind {
	case ``:
		sch.kind = ScheduleKindMinutely
	case ScheduleKindMinutely:
		// Minutes is the lowest schedule.

	case ScheduleKindHourly:
		if len(list) >= 2 {
			sch.parseListMinutes(list[1])
		}

	case ScheduleKindDaily:
		v = ``
		if len(list) >= 2 {
			v = list[1]
		}
		sch.parseListTimeOfDay(v)

	case ScheduleKindWeekly:
		v = ``
		if len(list) >= 2 {
			v = list[1]
		}
		sch.parseListDayOfWeek(v)

		v = ``
		if len(list) >= 3 {
			v = list[2]
		}
		sch.parseListTimeOfDay(v)

	case ScheduleKindMonthly:
		v = ``
		if len(list) >= 2 {
			v = list[1]
		}
		sch.parseListDayOfMonth(v)

		v = ``
		if len(list) >= 3 {
			v = list[2]
		}
		sch.parseListTimeOfDay(v)

	default:
		return nil, fmt.Errorf(`%s: %w`, logp, ErrScheduleUnknown)
	}

	sch.calcNext(Now().UTC())

	go sch.run()

	return sch, nil
}

// calcNext calculate the next schedule based on time now.
func (sch *Scheduler) calcNext(now time.Time) {
	switch sch.kind {
	case ScheduleKindMinutely:
		sch.nextMinutely(now)

	case ScheduleKindHourly:
		sch.nextHourly(now)

	case ScheduleKindDaily:
		sch.nextDaily(now)

	case ScheduleKindWeekly:
		sch.nextWeekly(now)

	case ScheduleKindMonthly:
		sch.nextMonthly(now)
	}

	sch.nextSeconds = int64(sch.next.Sub(now).Seconds())
	if sch.nextSeconds < 0 {
		sch.nextSeconds = 0
	}
}

// parseListDayOfWeek parse comma separated day (Sunday,...) into field dow.
func (sch *Scheduler) parseListDayOfWeek(days string) {
	days = strings.TrimSpace(days)
	days = strings.ToLower(days)

	var (
		listDay = strings.Split(days, `,`)

		day    string
		dayInt int
	)

	for _, day = range listDay {
		day = strings.TrimSpace(day)

		switch day {
		case `sunday`, `sun`:
			dayInt = int(time.Sunday)
		case `monday`, `mon`:
			dayInt = int(time.Monday)
		case `tuesday`, `tue`:
			dayInt = int(time.Tuesday)
		case `wednesday`, `wed`:
			dayInt = int(time.Wednesday)
		case `thursday`, `thu`:
			dayInt = int(time.Thursday)
		case `friday`, `fri`:
			dayInt = int(time.Friday)
		case `saturday`, `sat`:
			dayInt = int(time.Saturday)
		default:
			dayInt = -1
		}
		if dayInt == -1 {
			continue
		}
		if !ints.IsExist(sch.dow, dayInt) {
			sch.dow = append(sch.dow, dayInt)
		}
	}
	if len(sch.dow) == 0 {
		sch.dow = append(sch.dow, int(time.Sunday))
	} else {
		sort.Ints(sch.dow)
	}
}

func (sch *Scheduler) parseListDayOfMonth(v string) {
	v = strings.TrimSpace(v)

	var (
		list = strings.Split(v, `,`)

		day int
		err error
	)

	for _, v = range list {
		day, err = strconv.Atoi(v)
		if err != nil {
			continue
		}
		if day < 0 || day > 31 {
			continue
		}
		sch.dom = append(sch.dom, day)
	}

	if len(sch.dom) == 0 {
		sch.dom = append(sch.dom, 1)
	} else {
		sort.Ints(sch.dom)
	}
}

// parseListMinutes parse comma separated minutes (x,y,...) into field
// minutes.
func (sch *Scheduler) parseListMinutes(minutes string) {
	var (
		list = strings.Split(minutes, `,`)

		err error
		str string
		m   int
	)
	for _, str = range list {
		m, err = strconv.Atoi(str)
		if err != nil {
			continue
		}
		if m < 0 || m > 59 {
			continue
		}
		sch.minutes = append(sch.minutes, m)
	}
	if len(sch.minutes) == 0 {
		sch.minutes = append(sch.minutes, 0)
	} else {
		sort.Ints(sch.minutes)
	}
}

// parseListTimeOfDay parse comma separated time of day (HOUR:MINUTE) into
// field tod under Scheduler.
// If the v is empty, it will add default "00:00" as the only tod.
func (sch *Scheduler) parseListTimeOfDay(v string) {
	v = strings.TrimSpace(v)

	var listTod = strings.Split(v, `,`)
	for _, v = range listTod {
		v = strings.TrimSpace(v)

		var tod = ParseClock(v)
		tod.sec = 0
		sch.tod = append(sch.tod, tod)
	}
	if len(sch.tod) == 0 {
		sch.tod = append(sch.tod, Clock{})
	} else {
		SortClock(sch.tod)
	}
}

// nextMinutely calculate the next event for minutely schedule.
func (sch *Scheduler) nextMinutely(now time.Time) {
	var diffSecond = 60 - now.Second()
	sch.next = now.Add(time.Duration(diffSecond) * time.Second)
}

// nextHourly calculate the next event for hourly schedule.
func (sch *Scheduler) nextHourly(now time.Time) {
	var (
		nowMinute = now.Minute()

		m int
	)
	if len(sch.minutes) == 0 {
		m = 60 - nowMinute
		sch.next = now.Add(time.Duration(m) * time.Minute).Round(time.Hour)
		return
	}

	for _, m = range sch.minutes {
		m = m - nowMinute
		if m < 0 {
			continue
		}
		sch.next = now.Add(time.Duration(m) * time.Minute).Round(time.Minute)
		return
	}

	m = (60 - nowMinute) + sch.minutes[0]
	sch.next = now.Add(time.Duration(m) * time.Minute).Round(time.Minute)
}

// nextDaily calculate the next event for daily schedule.
func (sch *Scheduler) nextDaily(now time.Time) {
	var (
		clockNow = Clock{hour: now.Hour(), min: now.Minute()}

		nextClock Clock
		found     bool
	)

	nextClock, found = sch.nextClock(clockNow)

	sch.next = time.Date(now.Year(), now.Month(), now.Day(),
		nextClock.hour, nextClock.min, 0, 0, time.UTC)

	if !found {
		// No schedule for today, apply the first clock on the next
		// day.
		sch.next = sch.next.AddDate(0, 0, 1)
	}
}

func (sch *Scheduler) nextWeekly(now time.Time) {
	var (
		today    = int(now.Weekday())
		clockNow = Clock{hour: now.Hour(), min: now.Minute()}

		nextClock Clock
		found     bool
	)

	if sch.isDayOfWeek(today) {
		nextClock, found = sch.nextClock(clockNow)
		if found {
			// Today is registered in day-of-week, and we have
			// another clock in queue.
			sch.next = time.Date(now.Year(), now.Month(), now.Day(),
				nextClock.hour, nextClock.min, 0, 0, time.UTC)
			return
		}
	}

	var (
		nextDay int
		dayInc  int
	)

	nextDay, found = sch.nextDayOfWeek(today)
	if found {
		dayInc = nextDay - today
	} else {
		dayInc = (7 - today) + nextDay
	}

	nextClock = sch.tod[0]

	sch.next = time.Date(now.Year(), now.Month(), now.Day(),
		nextClock.hour, nextClock.min, 0, 0, time.UTC)

	sch.next = sch.next.AddDate(0, 0, dayInc)
}

// isDayOfWeek return true if the dayNow is one of the day registered in dow.
func (sch *Scheduler) isDayOfWeek(dayNow int) bool {
	var dow int
	for _, dow = range sch.dow {
		if dayNow == dow {
			return true
		}
	}
	return false
}

// nextClock return the next clock that is after now.
// If not found, return the first item and false.
func (sch *Scheduler) nextClock(now Clock) (tod Clock, found bool) {
	for _, tod = range sch.tod {
		if now.Before(tod) {
			return tod, true
		}
	}
	return sch.tod[0], false
}

// nextDayOfWeek return the next day of week that is greater than today;
// or the first item in dow if not found.
func (sch *Scheduler) nextDayOfWeek(today int) (day int, found bool) {
	for _, day = range sch.dow {
		if day > today {
			return day, true
		}
	}
	// If no weekday found, use the first item.
	return sch.dow[0], false
}

// nextMonthly calculate the next event for monthly schedule.
func (sch *Scheduler) nextMonthly(now time.Time) {
	var (
		nowMonth = now.Month()
		today    = now.Day()
		nowClock = Clock{hour: now.Hour(), min: now.Minute()}

		nextClock Clock
		nextDay   int
		found     bool
	)

	if sch.isDayOfMonth(today) {
		nextClock, found = sch.nextClock(nowClock)
		if found {
			// Today is registered in day-of-week, and we have
			// another clock in queue.
			sch.next = time.Date(now.Year(), now.Month(), today,
				nextClock.hour, nextClock.min, 0, 0, time.UTC)
			return
		}
	}

	nextClock = sch.tod[0]

	nextDay, found = sch.nextDayOfMonth(today)
	if found {
		sch.next = time.Date(now.Year(), now.Month(), nextDay,
			nextClock.hour, nextClock.min, 0, 0, time.UTC)

		if sch.next.Month() != nowMonth {
			// The next day is out of range for the current month.
			// Set the next day to the first day of next month.
			sch.next = time.Date(now.Year(), now.Month()+1, sch.dom[0],
				nextClock.hour, nextClock.min, 0, 0, time.UTC)
		}
	} else {
		sch.next = time.Date(now.Year(), now.Month()+1, nextDay,
			nextClock.hour, nextClock.min, 0, 0, time.UTC)
	}
}

// isDayOfWeek return true if today is one of registered day of month.
func (sch *Scheduler) isDayOfMonth(today int) bool {
	var dom int
	for _, dom = range sch.dom {
		if today == dom {
			return true
		}
	}
	return false
}

// nextDayOfMonth return the next day of month that is greater than today.
// If no day found it will return the first day registered in dom.
func (sch *Scheduler) nextDayOfMonth(today int) (nextDay int, found bool) {
	for _, nextDay = range sch.dom {
		if nextDay > today {
			return nextDay, true
		}
	}
	return sch.dom[0], false
}

// run the ticker for scheduler.
func (sch *Scheduler) run() {
	var (
		dur int64 = 60

		ticker  *time.Ticker
		nextDur int64
	)

	if sch.nextSeconds < 60 {
		dur = sch.nextSeconds
	}
	ticker = time.NewTicker(time.Duration(dur) * time.Second)

	for {
		select {
		case <-ticker.C:
			sch.nextSeconds -= dur

			if sch.nextSeconds <= 0 {
				// Notify the user and calculate the next
				// event.
				select {
				case sch.c <- sch.next:
				default:
				}

				sch.calcNext(Now().UTC())
			}

			if sch.nextSeconds < 60 {
				nextDur = sch.nextSeconds
			} else {
				nextDur = 60
			}
			if dur != nextDur {
				ticker.Reset(time.Duration(nextDur) * time.Second)
			}
			dur = nextDur

		case <-sch.cstop:
			ticker.Stop()
			// Send ACK back to the stopper.
			sch.cstop <- struct{}{}
			return
		}
	}
}

// Stop the scheduler.
func (sch *Scheduler) Stop() {
	select {
	case sch.cstop <- struct{}{}:
		// Wait for ACK.
		<-sch.cstop
	default:
		// noop.
	}
}
