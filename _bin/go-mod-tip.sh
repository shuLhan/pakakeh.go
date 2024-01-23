#!/bin/sh

## Script to get and print the latest Go module version based on the last tag
## and the latest commit hash from current working git directory.
##
## For example, if the last tag is v0.39.1 and the latest commit hash is
## 38f9f1fb1ebc,
##
##	$ go-mod-tip.sh
##	commit timestamp: 1658864676
##	github.com/shuLhan/share v0.39.1-0.20220726194436-38f9f1fb1ebc
##
## This command usually used to fix go.mod due to force commit.

MODNAME=$(go list -m)
COMMIT_TS=$(git log -n 1 --pretty=format:'%ct')
DATE=$(date -u --date="@${COMMIT_TS}" +%Y%m%d%H%M%S)
HASH=$(git log -n 1 --pretty=format:'%h' --abbrev=12)
TAG=$(git describe --abbrev=0 --tags 2>/dev/null)
PREFIX=

echo "commit timestamp: ${COMMIT_TS}"

if [[ "${TAG}" == "" ]]; then
	TAG="v0.0.0"
else
	IFS=. read MAJOR MINOR PATCH <<<"${TAG}"
	PATCH=$((PATCH+1))
	TAG=${MAJOR}.${MINOR}.${PATCH}
	PREFIX=0.
fi

echo ${MODNAME} ${TAG}-${PREFIX}${DATE}-${HASH}
