#!/bin/bash

set -eou pipefail
export CGO_ENABLED=0

project="github.com/Gufran/flightpath"

commit_hash="$(git rev-parse --short HEAD)"
build_time="$(date --utc +'%d-%m-%YT%H:%M:%S+00')"

ld_vars="-X ${project}/version.Commit=${commit_hash} -X ${project}/version.BuildTime=${build_time}"
ldflags="-s -w -extldflags \"-static\" ${ld_vars}"

function allarch { # Build binaries for all supported OS
  set -x
  for os in darwin linux windows; do
    GOOS="${os}" GOARCH=amd64 go build -a -o "_build/flightpath-${os}-amd64" -ldflags "${ldflags}"
  done
}

function whatever { # Build binary for host OS and architecture
  go build -a -o flightpath -ldflags "${ldflags}"
}

function optsmd { # Generate a markdown table of available options
  go run -tags docs doc.go flags.go
}

function docs { # Generate documentation or serve local site
  if [ $# -eq 0 ]; then
    docker run --rm -it -v "${PWD}":/docs squidfunk/mkdocs-material build --clean --site-dir docs
    echo "docs.flightpath.xyz" > docs/CNAME
  else
    docker run --rm -it -p 8000:8000 -v "${PWD}":/docs squidfunk/mkdocs-material
  fi
}

function help { # Print help text
  echo "Available commands:"
  grep '^function' "${BASH_SOURCE[0]}" | sed -e 's/function /  /' -e 's/{ //' | column -t -s\#
}

[ $# -gt 0 ] && { "${@}"; exit 0; }

whatever
