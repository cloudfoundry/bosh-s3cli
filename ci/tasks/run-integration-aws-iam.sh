#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

: ${access_key_id:?}
: ${secret_access_key:?}
: ${region_name:?}
: ${stack_name:?}

# Just need these to get the stack info and to create/invoke the Lambda function
export AWS_ACCESS_KEY_ID=${access_key_id}
export AWS_SECRET_ACCESS_KEY=${secret_access_key}
export AWS_DEFAULT_REGION=${region_name}

stack_info=$(get_stack_info ${stack_name})
bucket_name=$(get_stack_info_of "${stack_info}" "BucketName")
iam_role_arn=$(get_stack_info_of "${stack_info}" "IamRoleArn")
lambda_payload="{\"region\": \"${region_name}\", \"bucket_name\": \"${bucket_name}\", \"s3_host\": \"s3.amazonaws.com\"}"

pushd s3cli-src > /dev/null
  . .envrc
  GOOS=linux GOARCH=amd64 go build s3cli/s3cli
  GOOS=linux GOARCH=amd64 ginkgo build src/s3cli/integration

  zip -j payload.zip src/s3cli/integration.test s3cli ci/assets/lambda_function.py

  lambda_function_name=s3cli-integration-$(date +%s)

  aws lambda create-function \
  --region ${region_name} \
  --function-name ${lambda_function_name} \
  --zip-file fileb://payload.zip \
  --role ${iam_role_arn} \
  --handler payload.test_runner_handler \
  --runtime python

  aws lambda invoke \
  --invocation-type RequestResponse \
  --function-name ${lambda_function_name} \
  --region ${region_name} \
  --log-type Tail \
  --payload $lambda_payload \
  lambda.log

popd > /dev/null
