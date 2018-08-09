#!/usr/bin/env bash

set -e
# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

# VERSION=$1
# if [[ "${VERSION}x" == "x" ]]
# then
#     echo Missing version parameter - setting to snapshot
#     VERSION=snapshot
# fi

# jq_in_place() {
#     tmp=$(mktemp)
#     jq $1 $2 > ${tmp}
#     mv ${tmp} $2
# }

# Update version placeholders
# FILES=$(grep -r SEED_VERSION spec | cut -d ':' -f 1 | sort | uniq)

# for FILE in ${FILES}
# do
#     sed -i "" -e "s/SEED_VERSION/"$1"/g" ${FILE}
# done

echo Generating manual page entry for Linux and OSX...
asciidoctor -b manpage -D output/ docs/seed-cli.adoc