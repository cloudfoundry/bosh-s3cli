#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

: ${access_key_id:?}
: ${secret_access_key:?}
: ${bucket_name:?}
: ${s3_endpoint_host:?}
: ${s3_endpoint_port:?}

export ACCESS_KEY_ID=${access_key_id}
export SECRET_ACCESS_KEY=${secret_access_key}
export BUCKET_NAME=${bucket_name}
export S3_HOST=${s3_endpoint_host}
export S3_PORT=${s3_endpoint_port}

pushd s3cli-src > /dev/null
  . .envrc
  go install s3cli/s3cli

  export S3_CLI_PATH=$(which s3cli)

  ginkgo -r -focus="S3 COMPATIBLE" src/s3cli/integration/
popd > /dev/null
