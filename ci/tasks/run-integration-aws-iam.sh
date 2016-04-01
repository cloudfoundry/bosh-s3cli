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
bucket_name="bosh-s3cli-bogus" # $(get_stack_info_of "${stack_info}" "BucketName")
iam_role_arn=$(get_stack_info_of "${stack_info}" "IamRoleArn")
lambda_payload="{\"region\": \"${region_name}\", \"bucket_name\": \"${bucket_name}\", \"s3_host\": \"s3.amazonaws.com\"}"
lambda_log=$(realpath $(mktemp "XXXXXX-lambda.log"))

trap "cat ${lambda_log}" EXIT

pushd s3cli-src > /dev/null
  . .envrc
  GOOS=linux GOARCH=amd64 go build s3cli/s3cli
  GOOS=linux GOARCH=amd64 ginkgo build src/s3cli/integration

  zip -j payload.zip src/s3cli/integration/integration.test s3cli ci/assets/lambda_function.py

  lambda_function_name=s3cli-integration-$(date +%s)

  aws lambda create-function \
  --region ${region_name} \
  --function-name ${lambda_function_name} \
  --zip-file fileb://payload.zip \
  --role ${iam_role_arn} \
  --timeout 300 \
  --handler lambda_function.test_runner_handler \
  --runtime python2.7

  aws lambda invoke \
  --invocation-type RequestResponse \
  --function-name ${lambda_function_name} \
  --region ${region_name} \
  --log-type Tail \
  --payload "${lambda_payload}" \
  ${lambda_log} | tee lambda_output.json

  set +e
    log_group_name="/aws/lambda/${lambda_function_name}"

    logs_command="aws logs describe-log-streams --log-group-name=${log_group_name}"
    log_streams_json=$(${logs_command})

    tries=0
    while [[ ( $? -ne 0 ) && ( $tries -ne 5 ) ]] ; do
      sleep 2
      echo "Retrieving CloudWatch logs; attempt: $tries"
      tries=$((tries + 1))
      log_streams_json=$(${logs_command})
    done
  set -e

  log_stream_name=$(echo "${log_streams_json}" | jq -r ".logStreams[0].logStreamName")

  echo "Lambda execution log output"
  aws logs get-log-events --log-group-name=${log_group_name} --log-stream-name=${log_stream_name} | jq -r ".events | map(.message) | .[]"

  aws lambda delete-function \
  --function-name ${lambda_function_name}

  cat lambda_output.json | jq -r ".FunctionError" | grep -v -e "Handled" -e "Unhandled"
popd > /dev/null