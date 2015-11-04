#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh


semver=`cat version-semver/number`
timestamp=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

pushd s3cli > /dev/null
  git_rev=`git rev-parse --short HEAD`
  version="${semver}-${git_rev}-${timestamp}"

  . .envrc

  echo -e "\n building artifact..."
  go build -ldflags "-X main.version ${version}" -o "out/s3cli-${semver}-linux-amd64" s3cli/s3
popd > /dev/null

