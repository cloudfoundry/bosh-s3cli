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

echo ${s3_endpoint_host} > s3_endpoint_host
echo ${test_host_ip} > test_host_ip
echo ${bucket_name} > bucket_name

cat > "static_wout_host_w_region-s3cli_config.json"<< EOF
{
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}",
  "ssl_verify_peer": true,
  "use_ssl": true
}
EOF

if [ "${region_optional}" = true ]; then
  cat > "static_w_host_wout_region-s3cli_config.json"<< EOF
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

  cat > "static_wout_host_wout_region-s3cli_config.json"<< EOF
{
  "credentials_source": "static",
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "ssl_verify_peer": true,
  "use_ssl": true
}
EOF
fi

cat > "profile_wout_host_w_region-s3cli_config.json"<< EOF
{
  "credentials_source": "env_or_profile",
  "bucket_name": "${bucket_name}",
  "region": "${region_name}",
  "ssl_verify_peer": true,
  "use_ssl": true
}
EOF

if [ "${region_optional}" = true ]; then
  cat > "profile_wout_host_wout_region-s3cli_config.json"<< EOF
{
  "credentials_source": "env_or_profile",
  "bucket_name": "${bucket_name}",
  "ssl_verify_peer": true,
  "use_ssl": true
}
EOF
fi
