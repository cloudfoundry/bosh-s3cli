#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param region_name
check_param stack_name
check_param region_optional
check_param ec2_ami
check_param public_key_name
check_param alt_host
check_param invalid_host
check_param valid_host
check_param alt_region

export AWS_ACCESS_KEY_ID=${access_key_id}
export AWS_SECRET_ACCESS_KEY=${secret_access_key}
export AWS_DEFAULT_REGION=${region_name}

cloudformation_parameters="ParameterKey=AmazonMachineImageID,ParameterValue=${ec2_ami} ParameterKey=KeyPairName,ParameterValue=${public_key_name}"

cmd="aws cloudformation create-stack \
    --stack-name    ${stack_name} \
    --template-body file://${PWD}/s3cli-src/ci/assets/cloudformation-s3cli-iam.template.json \
    --capabilities  CAPABILITY_IAM
    --parameters    ${cloudformation_parameters}"
echo "Running: ${cmd}"; ${cmd}

while true; do
  stack_status=$(get_stack_status $stack_name)
  echo "StackStatus ${stack_status}"
  if [ $stack_status == 'CREATE_IN_PROGRESS' ]; then
    echo "sleeping 5s"; sleep 5s
  else
    break
  fi
done

if [ $stack_status != 'CREATE_COMPLETE' ]; then
  echo "cloudformation failed stack info:\n$(get_stack_info $stack_name)"
  exit 1
fi

stack_info=$(get_stack_info ${stack_name})
bucket_name=$(get_stack_info_of "${stack_info}" "BucketName")
s3_endpoint_host=$(get_stack_info_of "${stack_info}" "S3EndpointHost")
test_host_ip=$(get_stack_info_of "${stack_info}" "TestHostIP")

cd ${PWD}/configs
test_types=( generic negative_sig_version negative_region_invalid negative_region_and_host )
for test_type in "${test_types[@]}"; do
  mkdir -p ${test_type}
done

echo ${s3_endpoint_host} > s3_endpoint_host
echo ${test_host_ip} > test_host_ip
echo ${bucket_name} > bucket_name

cat > "generic/region_minimal-s3cli_config.json"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}"
}
EOF

cat > "generic/v4_static_wout_host_w_region-s3cli_config.json"<< EOF
{
  "signature_version": "4",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}"
}
EOF

cat > "generic/v4_profile_wout_host_w_region-s3cli_config.json"<< EOF
{
  "signature_version": "4",
  "credentials_source": "env_or_profile",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}"
}
EOF

if [ "${region_optional}" = true ]; then
  cat > "generic/minimal-s3cli_config.json"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}"
}
EOF

  cat > "generic/v2_minimal-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}"
}
EOF

  cat > "generic/v2_static_w_host_wout_region-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}"
}
EOF

  cat > "generic/w_host_wout_region-s3cli_config.json"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}"
}
EOF

  cat > "generic/v2_static_wout_host_w_region-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}"
}
EOF

  cat > "generic/profile_wout_host_wout_region-s3cli_config.json"<< EOF
{
  "credentials_source": "env_or_profile",
  "bucket_name": "${bucket_name}"
}
EOF

  cat > "negative_region_invalid/v4_w_host_wout_region-s3cli_config.json"<< EOF
{
  "signature_version": "4",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${s3_endpoint_host}"
}
EOF
else
  cat > "negative_sig_version/v2_wout_host_w_region-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}"
}
EOF
fi

if [ ! -z "${invalid_host}" ]; then
cat > "negative_region_and_host/v4_static_w_wrong_host_w_region-s3cli_config.json"<< EOF
{
  "signature_version": "4",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}",
  "host": "${invalid_host}"
}
EOF
fi

if [ ! -z "${valid_host}" ]; then
  cat > "generic/v4_static_w_host_w_region-s3cli_config.json"<< EOF
{
    "signature_version": "4",
    "credentials_source": "static",
    "access_key_id": "${access_key_id}",
    "secret_access_key": "${secret_access_key}",
    "bucket_name": "${bucket_name}",
    "host": "${valid_host}",
    "region": "${region_name}"
  }
EOF
fi

if [ ! -z "${alt_host}" ]; then
  cat > "generic/v2_static_w_alt_host_wout_region-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${alt_host}"
}
EOF
  if [ ! -z "${alt_region}" ]; then
    cat > "generic/v2_static_w_alt_host_w_alt_region-s3cli_config.json"<< EOF
{
  "signature_version": "2",
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "host": "${alt_host}",
  "region": "${alt_region}"
}
EOF
  fi
fi
