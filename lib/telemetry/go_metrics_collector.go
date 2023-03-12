// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry

import (
	"regexp"
	"runtime/metrics"
)

// RuntimeMetricsAlias define an alias for [runtime/metrics.Name] to be
// exported by GoMetricsCollector.
var RuntimeMetricsAlias = map[string]string{
	`/cgo/go-to-c-calls:calls`: `go_cgo_calls`,

	`/cpu/classes/gc/mark/assist:cpu-seconds`:    `go_cpu_gc_mark_assist_seconds`,
	`/cpu/classes/gc/mark/dedicated:cpu-seconds`: `go_cpu_gc_mark_dedicated_seconds`,
	`/cpu/classes/gc/mark/idle:cpu-seconds`:      `go_cpu_gc_mark_idle_seconds`,
	`/cpu/classes/gc/pause:cpu-seconds`:          `go_cpu_gc_pause_seconds`,
	`/cpu/classes/gc/total:cpu-seconds`:          `go_cpu_gc_total_seconds`,

	`/cpu/classes/idle:cpu-seconds`:                `go_cpu_idle_seconds`,
	`/cpu/classes/scavenge/assist:cpu-seconds`:     `go_cpu_scavenge_assist_seconds`,
	`/cpu/classes/scavenge/background:cpu-seconds`: `go_cpu_scavenge_background_seconds`,
	`/cpu/classes/scavenge/total:cpu-seconds`:      `go_cpu_scavenge_total_seconds`,
	`/cpu/classes/total:cpu-seconds`:               `go_cpu_total_seconds`,
	`/cpu/classes/user:cpu-seconds`:                `go_cpu_user_seconds`,

	`/gc/cycles/automatic:gc-cycles`: `go_gc_cycles_automatic`,
	`/gc/cycles/forced:gc-cycles`:    `go_gc_cycles_forced`,
	`/gc/cycles/total:gc-cycles`:     `go_gc_cycles_total`,

	`/gc/heap/allocs-by-size:bytes`: `go_gc_heap_alloc_by_size_bytes`,
	`/gc/heap/allocs:bytes`:         `go_gc_heap_allocs_bytes`,
	`/gc/heap/allocs:objects`:       `go_gc_heap_allocs_objects`,
	`/gc/heap/frees-by-size:bytes`:  `go_gc_heap_frees_by_size_bytes`,
	`/gc/heap/frees:bytes`:          `go_gc_heap_frees_bytes`,
	`/gc/heap/frees:objects`:        `go_gc_heap_frees_objects`,
	`/gc/heap/goal:bytes`:           `go_gc_heap_goal_bytes`,
	`/gc/heap/objects:objects`:      `go_gc_heap_objects`,
	`/gc/heap/tiny/allocs:objects`:  `go_gc_heap_tiny_allocs_objects`,

	`/gc/limiter/last-enabled:gc-cycle`: `go_gc_limiter_last_enabled`,

	`/gc/pauses:seconds`: `go_gc_pauses_seconds`,

	`/gc/stack/starting-size:bytes`: `go_gc_stack_starting_size_bytes`,

	`/godebug/non-default-behavior/execerrdot:events`:           `go_godebug_execerrdot_events`,
	`/godebug/non-default-behavior/http2client:events`:          `go_godebug_http2client_events`,
	`/godebug/non-default-behavior/http2server:events`:          `go_godebug_http2server_events`,
	`/godebug/non-default-behavior/installgoroot:events`:        `go_godebug_installgoroot_events`,
	`/godebug/non-default-behavior/panicnil:events`:             `go_godebug_panicnil_events`,
	`/godebug/non-default-behavior/randautoseed:events`:         `go_godebug_randautoseed_events`,
	`/godebug/non-default-behavior/tarinsecurepath:events`:      `go_godebug_trainsecurepath_events`,
	`/godebug/non-default-behavior/x509sha1:events`:             `go_godebug_x509sha1_events`,
	`/godebug/non-default-behavior/x509usefallbackroots:events`: `go_godebug_x509usefallbackroots_events`,
	`/godebug/non-default-behavior/zipinsecurepath:events`:      `go_godebug_zipinsecurepath_events`,

	`/memory/classes/heap/free:bytes`:     `go_memory_heap_free_bytes`,
	`/memory/classes/heap/objects:bytes`:  `go_memory_heap_objects_bytes`,
	`/memory/classes/heap/released:bytes`: `go_memory_heap_released_bytes`,
	`/memory/classes/heap/stacks:bytes`:   `go_memory_heap_stacks_bytes`,
	`/memory/classes/heap/unused:bytes`:   `go_memory_heap_unused_bytes`,

	`/memory/classes/metadata/mcache/free:bytes`:  `go_memory_metadata_mcache_free_bytes`,
	`/memory/classes/metadata/mcache/inuse:bytes`: `go_memory_metadata_mcache_inuse_bytes`,
	`/memory/classes/metadata/mspan/free:bytes`:   `go_memory_metadata_mspan_free_bytes`,
	`/memory/classes/metadata/mspan/inuse:bytes`:  `go_memory_metadata_mspan_inuse_bytes`,
	`/memory/classes/metadata/other:bytes`:        `go_memory_metadata_other_bytes`,

	`/memory/classes/os-stacks:bytes`:         `go_memory_os_stacks_bytes`,
	`/memory/classes/other:bytes`:             `go_memory_other_bytes`,
	`/memory/classes/profiling/buckets:bytes`: `go_memory_profiling_buckets_bytes`,
	`/memory/classes/total:bytes`:             `go_memory_total_bytes`,

	`/sched/gomaxprocs:threads`:    `go_sched_gomaxprocs`,
	`/sched/goroutines:goroutines`: `go_sched_goroutines`,
	`/sched/latencies:seconds`:     `go_sched_latencies_seconds`,

	`/sync/mutex/wait/total:seconds`: `go_sync_mutex_wait_total_seconds`,
}

// GoMetricsCollector collect the metrics using [runtime/metrics.Read].
type GoMetricsCollector struct {
	// samples list of [runtime/metrics.Sample] that has been
	// filtered using [AgentOptions.RuntimeMetrics].
	samples []metrics.Sample
}

// NewGoMetricsCollector create new collector for [runtime/metrics] with
// options to filter specific metric by alias name using regular expression.
//
// For example, to collect all metrics pass regex `^.*$`, to collect memory
// only pass `^go_memory_.*$`.
// A nil filter means no metrics will be collected.
func NewGoMetricsCollector(filter *regexp.Regexp) (col *GoMetricsCollector) {
	col = &GoMetricsCollector{}
	if filter == nil {
		// Nothing to collect.
		return col
	}

	var (
		org   string
		alias string
	)
	for org, alias = range RuntimeMetricsAlias {
		if filter.MatchString(alias) {
			var sample = metrics.Sample{
				Name: org,
			}
			col.samples = append(col.samples, sample)
		}
	}
	return col
}

// Collect the [runtime/metrics].
func (col *GoMetricsCollector) Collect(timestamp int64) (list []Metric) {
	if len(col.samples) == 0 {
		return nil
	}

	metrics.Read(col.samples)

	var sample metrics.Sample

	list = make([]Metric, 0, len(col.samples))

	for _, sample = range col.samples {
		var m = Metric{
			Timestamp: timestamp,
			Name:      RuntimeMetricsAlias[sample.Name],
		}

		switch sample.Value.Kind() {
		case metrics.KindUint64:
			m.Value = float64(sample.Value.Uint64())
		case metrics.KindFloat64:
			m.Value = sample.Value.Float64()
		case metrics.KindFloat64Histogram:
			var hist = sample.Value.Float64Histogram()
			m.Value = medianBucket(hist)
		}
		list = append(list, m)
	}
	return list
}

// medianBucket get the median of the histogram values.
func medianBucket(hist *metrics.Float64Histogram) float64 {
	var (
		total uint64
		count uint64
	)
	for _, count = range hist.Counts {
		total += count
	}
	if total == 0 {
		return 0
	}

	var (
		thresh = total / 2
		x      int
	)

	total = 0
	for x, count = range hist.Counts {
		total += count
		if total >= thresh {
			return hist.Buckets[x]
		}
	}
	return 0
}
