#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh

result=0

pushd s3cli > /dev/null
  . .envrc

  echo -e "\n Vetting packages for potential issues..."
  go vet s3cli/...
  let "result+=$?"

  echo -e "\n Checking with golint..."
  golint s3cli/...
  let "result+=$?"

  echo -e "\n Testing packages..."
  ginkgo -r -race src/s3cli
  let "result+=$?"

  echo -e "\n Running build script to confirm everything compiles..."
  go build -o out/s3 s3cli/s3
  let "result+=$?"

  if [ $result -eq 0 ]; then
  	echo -e "\nSUITE SUCCESS"
  else
  	echo -e "\nSUITE FAILURE"
  fi
popd > /dev/null

exit $result
