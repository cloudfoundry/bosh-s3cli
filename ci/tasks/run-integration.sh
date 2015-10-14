#!/usr/bin/env bash

set -e

source s3cli/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param bucket_name
check_param region_name
check_param signature_version
check_param host
check_param port

config_file="${PWD}/blobstore-s3.json"
cat > "${config_file}"<<EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "credentials_source": "static",
  "region": "${region_name}",
  "signature_version": "${signature_version}",

  "host": "${host}",
  "port": ${port}",

  "use_ssl": true,
  "ssl_verify_peer": true
}
EOF

pushd ${PWD}/s3cli
  echo "Running tests on s3cli:"
  bin/test
popd
