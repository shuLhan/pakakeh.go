<!--
SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

SPDX-License-Identifier: BSD-3-Clause
-->

This note document a benchmark between gobwas vs our websocket library.

# github.com/gobwas/ws@v0.1.0

## Go v1.10.3

```
goos: linux
goarch: amd64
pkg: github.com/gobwas/ws
BenchmarkUpgrader/base-8                         5000000               377 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/lowercase-8                    5000000               380 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/uppercase-8                    5000000               394 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/subproto-8                     3000000               529 ns/op               1 B/op          1 allocs/op
BenchmarkUpgrader/subproto_comma-8               3000000               453 ns/op               1 B/op          1 allocs/op
BenchmarkUpgrader/#00-8                          1000000              1824 ns/op            1354 B/op          4 allocs/op
BenchmarkUpgrader/bad_http_method-8             10000000               143 ns/op               3 B/op          1 allocs/op
BenchmarkUpgrader/bad_http_proto-8              10000000               135 ns/op               3 B/op          1 allocs/op
BenchmarkUpgrader/bad_host-8                    10000000               224 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade-8                 10000000               215 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade#01-8              10000000               235 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade#02-8              10000000               229 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection-8              10000000               216 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection#01-8           10000000               175 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version_x-8           10000000               216 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version-8              5000000               237 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_key-8                  5000000               379 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_key#01-8               5000000               382 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/gobwas/ws    58.827s
```

## Go version devel +d6a27e8edc

```
goos: linux
goarch: amd64
pkg: github.com/gobwas/ws
BenchmarkUpgrader/base-8                         5000000               378 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/lowercase-8                    5000000               374 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/uppercase-8                    5000000               398 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/subproto-8                     3000000               533 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/subproto_comma-8               3000000               449 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/#00-8                          1000000              1653 ns/op            1354 B/op          4 allocs/op
BenchmarkUpgrader/bad_http_method-8             10000000               142 ns/op               3 B/op          1 allocs/op
BenchmarkUpgrader/bad_http_proto-8              10000000               138 ns/op               3 B/op          1 allocs/op
BenchmarkUpgrader/bad_host-8                    10000000               219 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade-8                 10000000               217 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade#01-8               5000000               233 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade#02-8               5000000               227 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection-8              10000000               215 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection#01-8           10000000               176 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version_x-8           10000000               217 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version-8              5000000               266 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_key-8                  5000000               398 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_key#01-8               5000000               391 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/gobwas/ws    57.334s
```

# git.sr.ht/~shulhan/pakakeh.go/lib/websocket

## Go v1.10.3

```
goos: linux
goarch: amd64
pkg: git.sr.ht/~shulhan/pakakeh.go/lib/websocket
BenchmarkUpgrader/base-8                         5000000              339 ns/op             176 B/op          1 allocs/op
BenchmarkUpgrader/lowercase-8                    5000000              358 ns/op             176 B/op          1 allocs/op
BenchmarkUpgrader/uppercase-8                    5000000              352 ns/op             176 B/op          1 allocs/op
BenchmarkUpgrader/subproto-8                    10000000              172 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/subproto_comma-8               3000000              389 ns/op             176 B/op          1 allocs/op
BenchmarkUpgrader/#00-8                         10000000              174 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_http_method-8            100000000               23.8 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_http_proto-8              50000000               29.3 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_host-8                   100000000               23.1 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade-8                100000000               23.2 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade#01-8               3000000              440 ns/op             453 B/op          6 allocs/op
BenchmarkUpgrader/bad_upgrade#02-8              10000000              167 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection-8              50000000               24.4 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection#01-8           20000000              108 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version_x-8           50000000               24.2 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version-8             10000000              138 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_key-8                  5000000              343 ns/op             176 B/op          1 allocs/op
BenchmarkUpgrader/bad_sec_key#01-8               5000000              369 ns/op             176 B/op          1 allocs/op
PASS
ok      git.sr.ht/~shulhan/pakakeh.go/lib/websocket  50.192s
```

## Go v1.12

```
websocket version: 8dec8c9
Benchmark date   : Thu  7 Mar 22:09:17 WIB 2019

goos: linux
goarch: amd64
pkg: git.sr.ht/~shulhan/pakakeh.go/lib/websocket
BenchmarkUpgrader/base-8                        10000000               165 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/lowercase-8                   10000000               165 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/uppercase-8                   10000000               163 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/subproto-8                    10000000               133 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/subproto_comma-8              10000000               192 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/#00-8                         10000000               144 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_http_method-8             50000000                25.4 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_http_proto-8              50000000                30.9 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_host-8                    50000000                24.2 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade-8                 50000000                24.3 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade#01-8               5000000               381 ns/op             453 B/op          6 allocs/op
BenchmarkUpgrader/bad_upgrade#02-8              10000000               133 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection-8              50000000                24.4 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection#01-8           20000000                91.8 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version_x-8           50000000                24.4 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version-8             20000000               112 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_key-8                 10000000               165 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/bad_sec_key#01-8              10000000               166 ns/op              32 B/op          1 allocs/op
PASS
ok      git.sr.ht/~shulhan/pakakeh.go/lib/websocket  49.379s
```

## Go version devel +05b3db24 (>1.12)

```
websocket version: 8dec8c9
Benchmark date   : Thu  7 Mar 22:09:17 WIB 2019

goos: linux
goarch: amd64
pkg: git.sr.ht/~shulhan/pakakeh.go/lib/websocket
BenchmarkUpgrader/base-8                        10000000               156 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/lowercase-8                   10000000               160 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/uppercase-8                   10000000               153 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/subproto-8                    10000000               137 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/subproto_comma-8              10000000               181 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/#00-8                         10000000               143 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_http_method-8             50000000                25.0 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_http_proto-8              50000000                31.6 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_host-8                    50000000                24.6 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade-8                 50000000                24.5 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_upgrade#01-8               5000000               372 ns/op             453 B/op          6 allocs/op
BenchmarkUpgrader/bad_upgrade#02-8              10000000               133 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection-8             100000000                23.4 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_connection#01-8           20000000                92.7 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version_x-8          100000000                23.3 ns/op             0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_version-8             20000000               113 ns/op               0 B/op          0 allocs/op
BenchmarkUpgrader/bad_sec_key-8                 10000000               154 ns/op              32 B/op          1 allocs/op
BenchmarkUpgrader/bad_sec_key#01-8              10000000               157 ns/op              32 B/op          1 allocs/op
PASS
ok      git.sr.ht/~shulhan/pakakeh.go/lib/websocket  52.285s
```
