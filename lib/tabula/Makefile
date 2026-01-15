#!/bin/make
# SPDX-License-Identifier: BSD-3-Clause
# SPDX-FileCopyrightText: 2017 Shulhan <ms@kilabit.info>

SRC_FILES	:=$(shell go list -f '{{ join .GoFiles " " }}')
TEST_FILES	:=$(shell go list -f '{{ join .TestGoFiles " " }}')
XTEST_FILES	:=$(shell go list -f '{{ join .XTestGoFiles " " }}')
COVER_OUT	:=cover.out
COVER_HTML	:=cover.html
TARGET		:=$(shell go list -f '{{ .Target }}')

.PHONY: all clean coverbrowse

all: ${TARGET}

${TARGET}: ${COVER_HTML}
	go install -a .

${COVER_HTML}: ${COVER_OUT}
	go tool cover -html=$< -o $@

${COVER_OUT}: ${SRC_FILES} ${TEST_FILES} ${XTEST_FILES}
	go test -v -coverprofile $@

coverbrowse: ${COVER_HTML}
	xdg-open $<

clean:
	rm -f ${COVER_HTML} ${COVER_OUT} *.bench *.prof *.test
