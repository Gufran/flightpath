#!/bin/bash

set -ex

project="github.com/Gufran/flightpath"

commit_hash="$(git rev-parse --short HEAD)"
build_time="$(date --utc +'%d-%m-%YT%H:%M:%S+00')"

ld_vars="-X ${project}/version.Commit=${commit_hash} -X ${project}/version.BuildTime=${build_time}"
ldflags="-s -w -extldflags \"-static\" ${ld_vars}"

go build -a -o /flightpath -ldflags "${ldflags}"
