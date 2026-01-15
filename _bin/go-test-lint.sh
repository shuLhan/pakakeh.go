#!/bin/sh
# SPDX-License-Identifier: BSD-3-Clause
# SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

## Script go-test-lint.sh run Go test and if its success it will run
## predefined linter, in the current directory.
##
## Arg 1: the method or function to test, default to ".".
##
## The linter program and its argument is derived from environment variable
## GO_LINT.
## If its empty, it will try the following linter in order: revive, or
## golangci-lint.
##
## To add additional arguments to go test set the environment variable
## GO_TEST_ARGS.
##
## == Examples
##
## Run all tests with -race condition,
##
##   $ GO_TEST_ARGS=-race go-test-lint.sh
##
## Run test named "Hello" using mylint as linter,
##
##   $ GO_LINT="mylint ./..." go-test-lint.sh "Hello"
##

FN=${1:-.}
LINTER="$GO_LINT"

go test "$GO_TEST_ARGS" -run="$FN" . || exit $?

if [[ -z $LINTER ]]; then
	LINTER=$(command -v revive)
	if [[ -z $LINTER ]]; then
		LINTER=$(command -v golangci-lint)
		if [[ -z $LINTER ]]; then
			echo "No linter found."
			exit 0
		fi
		LINTER="$LINTER run ./..."
	else
		LINTER="$LINTER ./..."
	fi
fi

echo "Running linter: $LINTER"
$LINTER
