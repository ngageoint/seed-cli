#!/usr/bin/env bash
set -e

# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

UNAME=$(uname -s)

vendor/go-bindata-${UNAME} -pkg assets -o assets/assets.go ./schema/* ./images/wordmark.png

go run main.go "$@"

popd >/dev/null