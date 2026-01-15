#!/bin/sh
# SPDX-License-Identifier: BSD-3-Clause
# SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>

## Script to run Go benchmark in the current package.
##
## Arg 1: the method or function to benchmark, default to ".".
## Arg 2: the benchmark number, default to "YYYYmmDD-HHMM".
##
## This script output three files:
##  - bench_$1_$2.cpu.prof: the CPU profile
##  - bench_$1_$2.mem.prof: the memory profile.
##  - bench_$1_$2.txt: the benchmark output.

TIMESTAMP=$(date +%Y%m%d-%H%M)

FN=${1:-.}
NO=${2:-$TIMESTAMP}

BENCH_OUT="bench_${FN}_${NO}.txt"
CPU_PROF="bench_${FN}_${NO}_cpu.prof"
MEM_PROF="bench_${FN}_${NO}_mem.prof"

export GORACE=history_size=7
export CGO_ENABLED=1
go test -race -run=noop -benchmem -bench=${FN} \
	-cpuprofile=${CPU_PROF} \
	-memprofile=${MEM_PROF} \
	-count=10 \
	. \
	|& tee $BENCH_OUT
