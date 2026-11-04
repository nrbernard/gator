#!/bin/sh
# Fails if any place that names the Go version disagrees with go.mod's `go` directive.
# Run by `make check`.
set -eu

want=$(awk '$1 == "go" { print $2 }' go.mod)
status=0

check() { # file, version found there, version expected
	if [ "$2" != "$3" ]; then
		echo "$1 names Go '$2'; go.mod says $3" >&2
		status=1
	fi
}

check fly.toml "$(sed -n "s/^ *GO_VERSION *= *'\(.*\)'.*/\1/p" fly.toml)" "$want"
check docker-compose.yml "$(sed -n 's/^ *GO_VERSION: *\([^ ]*\).*/\1/p' docker-compose.yml)" "$want"
check Dockerfile "$(sed -n 's/^ARG GO_VERSION=\([^ ]*\).*/\1/p' Dockerfile)" "$want"
# The README names major.minor only.
check README.md "$(sed -n 's/^- Go \([0-9][0-9.]*\) .*/\1/p' README.md)" "${want%.*}"

exit $status
