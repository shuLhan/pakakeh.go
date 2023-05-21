#!/bin/sh

usage() {
cat << EOF
Script to run Go test and generate HTML coverage.

Arguments,
1: optional, path to package. Default "./...".
2: optional, function name to be tested. Default to all functions.
EOF
}

case "$1" in
help)
	usage
	exit 0
	;;
*)
	PKGS=${1:-./...}
	NAME=${2:-.}
	;;
esac

CGO_ENABLED=1 go test -race -p=1 -coverprofile=cover.out -run=$NAME $PKGS
go tool cover -html=cover.out -o cover.html
