#!/usr/bin/env bash

set -e

# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

VERSION=$1
if [[ "${VERSION}x" == "x" ]]
then
    echo Missing version parameter!
    echo Usage:
    echo $0 1.0.0
    exit 1
fi


vendor/go-bindata -pkg constants -o constants/jsonschema.go ./schema/0.1.0/
echo Building cross platform Seed CLI.
echo Building for Linux...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-X main.version=$VERSION -extldflags=\"-static\"" -o output/seed-linux-amd64
echo Building for OSX...
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags "-X main.version=$VERSION -extldflags=\"-static\"" -o output/seed-darwin-amd64
echo Building for Windows...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -ldflags "-X main.version=$VERSION -extldflags=\"-static\"" -o output/seed-windows-amd64
echo CLI build complete

echo Building example images.................................................................
SEED=""
UNAME=$(uname -s)
case "${UNAME}" in
    Linux*)     SEED=output/seed-linux-amd64; SUDO=sudo;;
    Darwin*)    SEED=output/seed-darwin-amd64;;
    CYGWIN*)    SEED=output/seed-windows-amd64;;
    *)          SEED="UNKNOWN:${UNAME}"
esac
${SEED} build -d examples/addition-algorithm/
${SEED} build -d examples/extractor/

echo Finished building example images.................................................................

echo Running example images...........................................................................
${SUDO} ${SEED} run -in addition-algorithm-0.0.1-seed:1.0.0 -i INPUT_FILE=examples/addition-algorithm/inputs.txt -rm -m MOUNT_BIN=testdata/complete/ -m MOUNT_TMP=testdata/ -e SETTING_ONE=one -e SETTING_TWO=two -o temp
echo ...
echo ...
${SUDO} ${SEED} run -in extractor-0.1.0-seed:0.1.0 -i ZIP=testdata/seed-scale.zip -i MULTIPLE=examples/extractor/seed.manifest.json -i MULTIPLE=examples/extractor/seed.outputs.json -rm -m MOUNTAIN=testdata/complete/ -e HELLO=Hello -o temp
echo Finished running example images..................................................................

popd >/dev/null
