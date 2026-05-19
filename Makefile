# SPDX-License-Identifier: BSD-3-Clause
# SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

COVER_OUT:=cover.out
COVER_HTML:=cover.html
CPU_PROF:=cpu.prof
MEM_PROF:=mem.prof
TEST_OUT:=test.txt

CIIGO := ${GOBIN}/ciigo
VERSION := $(shell git describe --tags)

.PHONY: all install build docs docs-serve clean distclean
.PHONY: lint test test.prof

all: lint build test

install:
	go install ./cmd/...

build: BUILD_FLAGS=-ldflags "-s -w -X 'kilabit.info/pakakeh.go.Version=$(VERSION)'"
build:
	mkdir -p _bin/
	go build \
		-trimpath \
		-buildmode=pie \
		-mod=readonly \
		-modcacherw \
		$(BUILD_FLAGS) -o _bin/ ./cmd/...

## Prevent git CLI to print errors in user's locale language.
test: export LC_ALL=C.UTF-8
test: export CGO_ENABLED=1
## TODO: remove after minimum go version in go.mod set to 1.27 or later.
## See https://go-review.googlesource.com/c/go/+/773962.
test: export GOEXPERIMENT=nojsonv2
test:
	go test -failfast -timeout=2m -race \
		-coverprofile=$(COVER_OUT) ./... \
		2>&1 | tee $(TEST_OUT)
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)

test.prof:
	go test -race -timeout=1m -cpuprofile $(CPU_PROF) -memprofile $(MEM_PROF) ./...

.PHONY: test.top10
test.top10: $(TEST_OUT)
	@echo ""
	@echo "=== Top 10 longest tests"
	@cat $(TEST_OUT) | rg '\d+\.\d+s' | awk '{ print $$3" "$$2 }' \
		| sort -r -h | head -10
	@echo ""
	@echo "=== Top 10 lowest coverages"
	@cat $(TEST_OUT) | rg '\d+\.\d+s' | rg '\d+\.\d+%' \
		| awk '{ print $$5" "$$2 }' \
		| sort -h | head -10

lint:
	go run ./cmd/gocheck ./...
	go vet ./...

$(CIIGO):
	go install kilabit.info/ciigo/cmd/ciigo

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
