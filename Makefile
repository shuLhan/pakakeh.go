## Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
## Use of this source code is governed by a BSD-style
## license that can be found in the LICENSE file.

COVER_OUT:=cover.out
COVER_HTML:=cover.html
CPU_PROF:=cpu.prof
MEM_PROF:=mem.prof

CIIGO := ${GOBIN}/ciigo
VERSION := $(shell git describe --tags)

.PHONY: all install build docs docs-serve clean distclean
.PHONY: lint test test.prof

all: test lint build

install:
	go install ./cmd/...

build: BUILD_FLAGS=-ldflags "-s -w -X 'git.sr.ht/~shulhan/pakakeh.go.Version=$(VERSION)'"
build:
	mkdir -p _bin/
	go build $(BUILD_FLAGS) -o _bin/ ./cmd/...

test:
	CGO_ENABLED=1 go test -failfast -timeout=1m -race -coverprofile=$(COVER_OUT) ./...
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)

test.prof:
	go test -race -timeout=1m -cpuprofile $(CPU_PROF) -memprofile $(MEM_PROF) ./...

lint:
	-fieldalignment ./...
	-shadow ./...
	-golangci-lint run \
		--presets bugs,metalinter,performance,unused \
		--disable exhaustive \
		--disable musttag \
		./...

$(CIIGO):
	go install git.sr.ht/~shulhan/ciigo/cmd/ciigo

docs: $(CIIGO)
	ciigo convert _doc

docs-serve: $(CIIGO)
	ciigo -address 127.0.0.1:21019 serve _doc

clean:
	rm -f $(COVER_OUT) $(COVER_HTML)
	rm -f ./**/${CPU_PROF}
	rm -f ./**/${MEM_PROF}
	rm -f ./**/$(COVER_OUT)
	rm -f ./**/$(COVER_HTML)

distclean:
	go clean -i ./...
