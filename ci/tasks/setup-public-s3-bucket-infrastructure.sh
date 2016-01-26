#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key
check_param region_name
check_param stack_name
check_param region_optional
check_param s3_endpoint_host

export AWS_ACCESS_KEY_ID=${access_key_id}
export AWS_SECRET_ACCESS_KEY=${secret_access_key}
export AWS_DEFAULT_REGION=${region_name}

cmd="aws cloudformation create-stack \
    --stack-name    ${stack_name} \
    --template-body file://${PWD}/s3cli-src/ci/assets/cloudformation-s3cli-public-bucket.template.json"
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

cd ${PWD}/configs
echo ${s3_endpoint_host} > s3_endpoint_host

test_types=( public_read )
for test_type in "${test_types[@]}"; do
  mkdir -p ${test_type}
done

echo ${bucket_name} > bucket_name

cat > "public_read/region_minimal-s3cli_config.json"<< EOF
{
  "bucket_name": "${bucket_name}",
  "region": "${region_name}"
}
EOF
