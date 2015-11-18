#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param bucket_name
check_param s3cmd_host
check_param s3cmd_region
check_param port

export BATS_LOG="${PWD}/bats_log"
echo "" > $BATS_LOG
trap "echo 'Output log:'; cat ${BATS_LOG}" ERR

# Hack to get bats to work with concourse
export TERM=xterm

pushd ${PWD}/s3cli > /dev/null
  . .envrc
  go install s3cli/s3cli

  export S3_CLI_CONFIG="$(mktemp -d /tmp/bats.XXXXXX)/s3cli.config"
  cat > "${S3_CLI_CONFIG}"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "credentials_source": "static",
  "port": ${port},
  "region": "${region_name}",
  "host": "${host}",
  "ssl_verify_peer": true,
  "use_ssl": true
}
EOF


  bats integration/test.bats
popd > /dev/null
