#!/usr/bin/env bash

set -e

my_dir="$( cd $(dirname $0) && pwd )"
release_dir="$( cd ${my_dir} && cd ../.. && pwd )"

source ${release_dir}/ci/tasks/utils.sh

semver='1.2.3.4'
timestamp=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

pushd ${release_dir} > /dev/null
  git_rev=`git rev-parse --short HEAD`
  version="${semver}-${git_rev}-${timestamp}"

  . .envrc
  S3CLI_FILES=$(find . -type f -name '*.go' -not -path "*/vendor/*")

  echo -e "\n Vetting packages for potential issues..."
  for f in $S3CLI_FILES ; do
    go vet $f
  done

  echo -e "\n Checking with golint..."
  for f in $S3CLI_FILES ; do
    golint $f
  done

  echo -e "\n Unit testing packages..."
  ginkgo -r -race -skipPackage=integration src/s3cli/

  echo -e "\n Running build script to confirm everything compiles..."
  go build -ldflags "-X main.version=${version}" -o out/s3cli s3cli

  echo -e "\n Testing version information"
  app_version=$(out/s3cli -v)
  test "${app_version}" = "version ${version}"

  echo -e "\n suite success"
popd > /dev/null
