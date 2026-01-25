# SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
#
# SPDX-License-Identifier: BSD-3-Clause

COVER_OUT:=cover.out
COVER_HTML:=cover.html
CPU_PROF:=cpu.prof
MEM_PROF:=mem.prof

CIIGO := ${GOBIN}/ciigo
VERSION := $(shell git describe --tags)

.PHONY: all install build docs docs-serve clean distclean
.PHONY: lint test test.prof

all: lint build test

install:
	go install ./cmd/...

build: BUILD_FLAGS=-ldflags "-s -w -X 'git.sr.ht/~shulhan/pakakeh.go.Version=$(VERSION)'"
build:
	mkdir -p _bin/
	go build \
		-trimpath \
		-buildmode=pie \
		-mod=readonly \
		-modcacherw \
		$(BUILD_FLAGS) -o _bin/ ./cmd/...

test:
	CGO_ENABLED=1 go test -failfast -timeout=2m -race -coverprofile=$(COVER_OUT) ./...
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)

test.prof:
	go test -race -timeout=1m -cpuprofile $(CPU_PROF) -memprofile $(MEM_PROF) ./...

lint:
	go run ./cmd/gocheck ./...
	go vet ./...

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
