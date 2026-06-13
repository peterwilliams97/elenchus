#!/usr/bin/env bash
# build.sh — the build gate. Chosen over a Makefile because this is a fixed
# linear sequence with fail-loud semantics, not a dependency graph; a shell
# script expresses "run these three, stop on the first failure" directly.
#
# Runs, in order: golangci-lint run, go test ./..., go build ./...
# Any failing step aborts the whole gate with a non-zero exit and a loud banner.

set -euo pipefail

cd "$(dirname "$0")"

fail() {
	echo >&2
	echo "########################################" >&2
	echo "# GATE FAILED: $1" >&2
	echo "########################################" >&2
	exit 1
}

echo ">> golangci-lint run"
command -v golangci-lint >/dev/null 2>&1 || fail "golangci-lint not installed"
golangci-lint run || fail "golangci-lint run"

echo ">> go test ./..."
go test ./... || fail "go test ./..."

echo ">> go build ./..."
go build ./... || fail "go build ./..."

echo
echo "GATE PASSED"
