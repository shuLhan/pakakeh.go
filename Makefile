## Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
## Use of this source code is governed by a BSD-style
## license that can be found in the LICENSE file.

SRC:=$(shell go list -f '{{$$d:=.Dir}} {{ range .GoFiles }}{{$$d}}/{{.}} {{end}}' ./...)
SRC_TEST:=$(shell go list -f '{{$$d:=.Dir}} {{ range .TestGoFiles }}{{$$d}}/{{.}} {{end}}' ./...)

COVER_OUT:=cover.out
COVER_HTML:=cover.html
CPU_PROF:=cpu.prof
MEM_PROF:=mem.prof

.PHONY: all install lint
.PHONY: test test.prof coverbrowse

all: install

install: test lint
	go install ./...

test: $(COVER_HTML)

test.prof:
	go test -race -cpuprofile $(CPU_PROF) -memprofile $(MEM_PROF) ./...

bench.lib.websocket:
	export GORACE=history_size=7 && \
		export CGO_ENABLED=1 && \
		go test -race -run=none -bench -benchmem \
			-cpuprofile=$(CPU_PROF) \
			-memprofile=$(MEM_PROF) \
			. ./lib/websocket

$(COVER_HTML): $(COVER_OUT)
	go tool cover -html=$< -o $@

$(COVER_OUT): $(SRC) $(SRC_TEST)
	export GORACE=history_size=7 && \
		export CGO_ENABLED=1 && \
		go test -race -count=1 -coverprofile=$@ ./...

coverbrowse: $(COVER_HTML)
	xdg-open $<

lint:
	golangci-lint run --enable-all \
		--disable=dupl \
		--disable=funlen \
		--disable=godox \
		--disable=gomnd \
		--disable=wsl \
		--disable=gocognit \
		./...

genhtml:
	ciigo -template=html.tmpl convert doc/

clean:
	rm -f $(COVER_OUT) $(COVER_HTML)
	rm -f ./**/${CPU_PROF}
	rm -f ./**/${MEM_PROF}

distclean:
	go clean -i ./...
