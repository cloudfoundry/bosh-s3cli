#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh


semver=`cat version-semver/number`
timestamp=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

output_dir=${PWD}/out

pushd s3cli-src > /dev/null
  git_rev=`git rev-parse --short HEAD`
  version="${semver}-${git_rev}-${timestamp}"

  . .envrc

  echo -e "\n building artifact..."
  go build -ldflags "-X main.version=${version}" -o "out/s3cli-${semver}-linux-amd64" s3cli/s3cli

  echo -e "\n sha1 of artifact..."
  sha1sum out/s3cli-${semver}-linux-amd64

  mv out/s3cli-${semver}-linux-amd64 ${output_dir}/
popd > /dev/null
