#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

: ${access_key_id:?}
: ${secret_access_key:?}
: ${region_name:?}
: ${stack_name:?}
: ${focus_regex:?}
: ${s3_endpoint_host:=unset}

# Just need these to get the stack info
export AWS_ACCESS_KEY_ID=${access_key_id}
export AWS_SECRET_ACCESS_KEY=${secret_access_key}
export AWS_DEFAULT_REGION=${region_name}
stack_info=$(get_stack_info ${stack_name})

# Some of these are optional
export ACCESS_KEY_ID=${access_key_id}
export SECRET_ACCESS_KEY=${secret_access_key}
export REGION=${region_name}
export BUCKET_NAME=$(get_stack_info_of "${stack_info}" "BucketName")
export S3_HOST=${s3_endpoint_host}

pushd s3cli-src > /dev/null
  . .envrc
  go install s3cli/s3cli

  export S3_CLI_PATH=$(which s3cli)

  ginkgo -r -focus="${focus_regex}" src/s3cli/integration/
popd > /dev/null
