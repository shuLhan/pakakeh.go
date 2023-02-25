#!/bin/sh

## Script to run go test and generate HTML coverage.
##
## Argument: path to package or default to current package and its
## subdirectory.

PKGS=${1:-./...}

CGO_ENABLED=1
go test -race -p=1 -coverprofile=cover.out ${PKGS}
go tool cover -html=cover.out -o cover.html
