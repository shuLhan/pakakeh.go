// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import (
	"log"
	"regexp"
	"runtime"
	"sort"
)

// GoMemStatsCollector collect Go statistics about memory allocator, as in
// calling [runtime.ReadMemStats].
//
// This collector export the following metric names, with value from the field
// of [runtime.MemStats]:
//
//   - go_memstats_alloc_bytes: [runtime.MemStats.Alloc]
//   - go_memstats_total_alloc_bytes: [runtime.MemStats.TotalAlloc]
//   - go_memstats_sys_bytes: [runtime.MemStats.Sys]
//   - go_memstats_lookups: [runtime.MemStats.Lookups]
//   - go_memstats_mallocs_objects: [runtime.MemStats.Mallocs]
//   - go_memstats_frees_objects: [runtime.MemStats.Frees]
//   - go_memstats_heap_alloc_bytes: [runtime.MemStats.HeapAlloc]
//   - go_memstats_heap_sys_bytes: [runtime.MemStats.HeapSys]
//   - go_memstats_heap_idle_bytes: [runtime.MemStats.HeapIdle]
//   - go_memstats_heap_inuse_bytes: [runtime.MemStats.HeapInuse]
//   - go_memstats_heap_released_bytes: [runtime.MemStats.HeapReleased]
//   - go_memstats_heap_objects: [runtime.MemStats.HeapObjects]
//   - go_memstats_stack_inuse_bytes: [runtime.MemStats.StackInuse]
//   - go_memstats_stack_sys_bytes: [runtime.MemStats.StackSys]
//   - go_memstats_mspan_inuse_bytes: [runtime.MemStats.MSpanInuse]
//   - go_memstats_mspan_sys_bytes: [runtime.MemStats.MSpanSys]
//   - go_memstats_mcache_inuse_bytes: [runtime.MemStats.MCacheInuse]
//   - go_memstats_mcache_sys_bytes: [runtime.MemStats.MCacheSys]
//   - go_memstats_buck_hash_sys_bytes: [runtime.MemStats.BuckHashSys]
//   - go_memstats_gc_sys_bytes: [runtime.MemStats.GCSys]
//   - go_memstats_other_sys_bytes: [runtime.MemStats.OtherSys]
//   - go_memstats_gc_next_bytes: [runtime.MemStats.NextGC]
//   - go_memstats_gc_last: [runtime.MemStats.LastGC]
//   - go_memstats_pause_total_ns: [runtime.MemStats.PauseTotalNs]
//   - go_memstats_pause_ns: [runtime.MemStats.PauseNs]
//   - go_memstats_pause_end_ns: [runtime.MemStats.PauseEnd]
//   - go_memstats_gc_num: [runtime.MemStats.NumGC]
//   - go_memstats_gc_forced_num: [runtime.MemStats.NumForcedGC]
//   - go_memstats_gc_cpu_fraction: [runtime.MemStats.GCCPUFraction]
type GoMemStatsCollector struct {
	// map of metric name with its pointer to its value.
	nameValue map[string]any

	// names contains the filtered metric names.
	names []string

	memstats runtime.MemStats
}

// NewGoMemStatsCollector create new MemStats collector with options to filter
// the metrics by its name using regular expression.
//
// If filter is nil, none of the metrics will be collected.
func NewGoMemStatsCollector(filter *regexp.Regexp) (col *GoMemStatsCollector) {
	col = &GoMemStatsCollector{}
	if filter == nil {
		return col
	}

	col.init(filter)

	return col
}

func (col *GoMemStatsCollector) init(filter *regexp.Regexp) {
	col.nameValue = map[string]any{
		`go_memstats_alloc_bytes`:         &col.memstats.Alloc,
		`go_memstats_total_alloc_bytes`:   &col.memstats.TotalAlloc,
		`go_memstats_sys_bytes`:           &col.memstats.Sys,
		`go_memstats_lookups`:             &col.memstats.Lookups,
		`go_memstats_mallocs_objects`:     &col.memstats.Mallocs,
		`go_memstats_frees_objects`:       &col.memstats.Frees,
		`go_memstats_heap_alloc_bytes`:    &col.memstats.HeapAlloc,
		`go_memstats_heap_sys_bytes`:      &col.memstats.HeapSys,
		`go_memstats_heap_idle_bytes`:     &col.memstats.HeapIdle,
		`go_memstats_heap_inuse_bytes`:    &col.memstats.HeapInuse,
		`go_memstats_heap_released_bytes`: &col.memstats.HeapReleased,
		`go_memstats_heap_objects`:        &col.memstats.HeapObjects,
		`go_memstats_stack_inuse_bytes`:   &col.memstats.StackInuse,
		`go_memstats_stack_sys_bytes`:     &col.memstats.StackSys,
		`go_memstats_mspan_inuse_bytes`:   &col.memstats.MSpanInuse,
		`go_memstats_mspan_sys_bytes`:     &col.memstats.MSpanSys,
		`go_memstats_mcache_inuse_bytes`:  &col.memstats.MCacheInuse,
		`go_memstats_mcache_sys_bytes`:    &col.memstats.MCacheSys,
		`go_memstats_buck_hash_sys_bytes`: &col.memstats.BuckHashSys,
		`go_memstats_gc_sys_bytes`:        &col.memstats.GCSys,
		`go_memstats_other_sys_bytes`:     &col.memstats.OtherSys,
		`go_memstats_gc_next_bytes`:       &col.memstats.NextGC,
		`go_memstats_gc_last`:             &col.memstats.LastGC,
		`go_memstats_pause_total_ns`:      &col.memstats.PauseTotalNs,
		`go_memstats_pause_ns`:            &col.memstats.PauseNs,
		`go_memstats_pause_end_ns`:        &col.memstats.PauseEnd,
		`go_memstats_gc_num`:              &col.memstats.NumGC,
		`go_memstats_gc_forced_num`:       &col.memstats.NumForcedGC,
		`go_memstats_gc_cpu_fraction`:     &col.memstats.GCCPUFraction,
	}

	var key string
	for key = range col.nameValue {
		if filter.MatchString(key) {
			col.names = append(col.names, key)
		}
	}
	sort.Strings(col.names)
}

// Collect the Go MemStats.
func (col *GoMemStatsCollector) Collect(ts int64) (list []Metric) {
	if len(col.names) == 0 {
		return nil
	}

	runtime.ReadMemStats(&col.memstats)

	var name string
	for _, name = range col.names {
		var m = Metric{
			Timestamp: ts,
			Name:      name,
		}

		var val = col.nameValue[name]

		switch v := val.(type) {
		case *uint64:
			m.Value = float64(*v)
		case *uint32:
			m.Value = float64(*v)
		case *float64:
			m.Value = *v
		case *[256]uint64:
			var last = v[(col.memstats.NumGC+255)%256]
			m.Value = float64(last)
		default:
			log.Printf(`GoMemStatsCollector.Collect: unknown type: %T %v`, v, v)
		}
		list = append(list, m)
	}
	return list
}
