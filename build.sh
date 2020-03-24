#!/bin/bash

set -eou pipefail
export CGO_ENABLED=0

version="$(awk -F\" '/Version/ { print $2; exit }' version/version.go)"
go_version="1.14"

project="github.com/Gufran/flightpath"

commit_hash="$(git rev-parse --short HEAD)"
build_time="$(date --utc +'%d-%m-%YT%H:%M:%S+00')"

ld_vars="-X ${project}/version.Commit=${commit_hash} -X ${project}/version.BuildTime=${build_time}"
ldflags="-s -w -extldflags \"-static\" ${ld_vars}"

function native() { # Build binary for host OS and architecture
  go build -a -o flightpath -ldflags "${ldflags}"
}

function gen-usage-doc() { # Update the usage documentation page
  go run -tags docs doc.go flags.go >site/usage.md
}

function docs() { # Generate documentation or serve local site
  if [ $# -eq 0 ]; then
    docker run --rm -it -v "${PWD}":/docs squidfunk/mkdocs-material build --clean --site-dir docs
    echo "docs.flightpath.xyz" >docs/CNAME
  else
    docker run --rm -it -p 8000:8000 -v "${PWD}":/docs squidfunk/mkdocs-material
  fi
}

function allarch() { # Build binaries for all supported OS
  for os in darwin linux windows; do
    echo " Building for ${os}"
    docker run \
      --interactive \
      --rm \
      --dns="1.1.1.1" \
      --volume="${PWD}:/go/src/flightpath" \
      --workdir="/go/src/flightpath" \
      "golang:${go_version}" \
      env \
      CGO_ENABLED=0 GOOS="${os}" GOARCH=amd64 \
      go build -a -o "_build/flightpath-${os}-amd64" -ldflags "${ldflags}"
  done
}

function test() { # Run tests
    docker run \
      --interactive \
      --rm \
      --dns="1.1.1.1" \
      --volume="${PWD}:/go/src/flightpath" \
      --workdir="/go/src/flightpath" \
      "golang:${go_version}" \
      go test -v -race -mod vendor ./...
}

function release() { # Build binaries and sign off for release
  if [ -z "${GPG_KEY}" ]; then
    echo "GPG_KEY environment variable is not set"
    exit 1
  fi

  gen-usage-doc
  docs
  allarch

  shasum_file="flightpath_${version}_SHA256SUMS"
  shasum --algorithm 256 _build/flightpath-*-amd64 >"_build/${shasum_file}"
  gpg --default-key "${GPG_KEY}" --detach-sig "_build/${shasum_file}"
  git commit --allow-empty --gpg-sign="${GPG_KEY}" --message "Release ${version}" --quiet --signoff
  git tag --annotate --create-reflog --local-user "${GPG_KEY}" --message "Version ${version}" --sign "v${version}"

  echo ""
  echo "Release is ready. Do not forget to run:"
  echo ""
  echo "    git push && git push --tags"
  echo ""
  echo "And then upload the binaries in _build"
}

function help() { # Print help text
  echo "Available commands:"
  grep '^function' "${BASH_SOURCE[0]}" | sed -e 's/function /  /' -e 's/()//' -e 's/{ //' | column -t -s\#
  exit 1
}

 function _has_subcommand() {
  grep '^function' "${BASH_SOURCE[0]}" | sed -e 's/function //' -e 's/().*//' | grep -w "${1}" >/dev/null 2>&1
}

[ $# -gt 0 ] && {
  _has_subcommand "${1}" && "${@}" && exit 0
}

help
