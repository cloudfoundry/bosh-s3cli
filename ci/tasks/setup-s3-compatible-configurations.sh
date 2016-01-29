#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param bucket_name
check_param s3_endpoint_host
check_param s3_endpoint_port
check_param region_name

cd ${PWD}/configs
test_types=( generic negative_sig_version negative_region_and_host )
for test_type in "${test_types[@]}"; do
  mkdir -p ${test_type}
done

echo ${s3_endpoint_host} > s3_endpoint_host
echo ${bucket_name} > bucket_name

cat > "generic/s3_compatible_minimal-s3cli_config.json"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}"
}
EOF

cat > "generic/s3_compatible_maximal-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}",
  "port": ${s3_endpoint_port},
  "use_ssl": true,
  "ssl_verify_peer": true
}
EOF

cat > "generic/s3_compatible_maximal_w_region-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}",
  "port": ${s3_endpoint_port},
  "use_ssl": true,
  "ssl_verify_peer": true,
  "region": "${region_name}"
}
EOF

cat > "generic/s3_compatible_wout_ssl-s3cli_config.json"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}",
  "use_ssl": false
}
EOF

cat > "negative_sig_version/v4_static-s3cli_config.json"<< EOF
{
  "signature_version": "4",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}"
}
EOF
