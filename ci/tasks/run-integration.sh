#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param bucket_name
check_param region_name
check_param host
check_param port

export BUCKET_NAME=${bucket_name}

export CONFIG_FILE="${PWD}/blobstore-s3.json"
cat > "${CONFIG_FILE}"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "credentials_source": "static",
  "region": "${region_name}",
  "host": "${host}",
  "port": ${port},
  "use_ssl": true,
  "ssl_verify_peer": true
}
EOF

cat > "${HOME}/.s3cfg" << EOF
[default]
access_key = ${access_key_id}
secret_key = ${secret_access_key}
host_base = ${host}
host_bucket = %(bucket)s.${host}
enable_multipart = True
multipart_chunk_size_mb = 15
use_https = True
EOF

export BATS_LOG="${PWD}/bats_log"
# Hack to get bats to work with concourse
export TERM=xterm

trap "echo 'Output log:'; cat ${BATS_LOG}" EXIT

pushd ${PWD}/s3cli
  echo "Running tests on s3cli:"
  bin/test
  echo "Running integration suite on s3cli:"
  bats integration/test.bats
popd
