#!/bin/bash

set -eou pipefail
export CGO_ENABLED=0

project="github.com/Gufran/flightpath"

commit_hash="$(git rev-parse --short HEAD)"
build_time="$(date --utc +'%d-%m-%YT%H:%M:%S+00')"

ld_vars="-X ${project}/version.Commit=${commit_hash} -X ${project}/version.BuildTime=${build_time}"
ldflags="-s -w -extldflags \"-static\" ${ld_vars}"

function allarch {
  for os in darwin linux windows; do
    GOOS="${os}" GOARCH=amd64 go build -a -o "_build/flightpath-${os}-amd64" -ldflags "${ldflags}"
  done
}

function whatever {
  go build -a -o /flightpath -ldflags "${ldflags}"
}

[ $# -gt 0 ] && { set -x; "${@}"; exit 0; }

whatever
