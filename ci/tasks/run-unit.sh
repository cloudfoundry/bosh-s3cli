#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh


pushd s3cli > /dev/null
  . .envrc

  echo -e "\n Vetting packages for potential issues..."
  go vet s3cli/...

  echo -e "\n Checking with golint..."
  golint s3cli/...

  echo -e "\n Testing packages..."
  ginkgo -r -race src/s3cli

  echo -e "\n Running build script to confirm everything compiles..."
  go build -o out/s3 s3cli/s3
popd > /dev/null

