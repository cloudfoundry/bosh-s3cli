#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh


semver='1.2.3.4'
timestamp=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

pushd s3cli > /dev/null
  git_rev=`git rev-parse --short HEAD`
  version="${semver}-${git_rev}-${timestamp}"

  . .envrc

  go install github.com/onsi/ginkgo/ginkgo
  go install github.com/golang/lint/golint

  echo -e "\n Vetting packages for potential issues..."
  go vet s3cli/...

  echo -e "\n Checking with golint..."
  golint s3cli/...

  echo -e "\n Testing packages..."
  ginkgo -r -race src/s3cli

  echo -e "\n Running build script to confirm everything compiles..."
  go run -ldflags "-X main.version ${version}" -o out/s3 s3cli/s3

  echo -e "\n Testing version information"
  app_version=$(out/s3 -v)
  test "${app_version}" = "version ${version}"

  echo -e "\n suite success"
popd > /dev/null

