#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param bucket_name
check_param s3_endpoint_host
check_param s3_endpoint_port

cd ${PWD}/configs

echo ${s3_endpoint_host} > s3_endpoint_host
echo ${bucket_name} > bucket_name

cat > "s3_compatible_w_port-s3cli_config.json"<< EOF
{
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}",
  "ssl_verify_peer": true,
  "use_ssl": true
}
EOF

cat > "static_w_host_wout_region-s3cli_config.json"<< EOF
{
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}",
  "port": "${s3_endpoint_port}",
  "ssl_verify_peer": true,
  "use_ssl": true
}
EOF
