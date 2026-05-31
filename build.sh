#!/usr/bin/env bash
set -euo pipefail

go test ./...
go build -o assay .
echo "built: $(pwd)/assay"
