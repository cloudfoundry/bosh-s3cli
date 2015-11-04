#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param bucket_name
check_param region_name
check_param host
check_param port

export BATS_LOG="${PWD}/bats_log"
echo "" > $BATS_LOG
trap "echo 'Output log:'; cat ${BATS_LOG}" ERR

# Hack to get bats to work with concourse
export TERM=xterm

pushd ${PWD}/s3cli > /dev/null
  . .envrc
  go install s3cli/s3cli
  bats integration/test.bats
popd > /dev/null
