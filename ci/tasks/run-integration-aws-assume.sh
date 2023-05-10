#!/usr/bin/env bash
set -euo pipefail

my_dir="$( cd "$(dirname "${0}")" && pwd )"
release_dir="$( cd "${my_dir}" && cd ../.. && pwd )"
workspace_dir="$( cd "${release_dir}" && cd .. && pwd )"

source "${release_dir}/ci/tasks/utils.sh"
export GOPATH=${workspace_dir}
export PATH=${GOPATH}/bin:${PATH}

: "${access_key_id:?}"
: "${secret_access_key:?}"
: "${region_name:=unset}"
: "${focus_regex:?}"
: "${assume_role_arn:=unset}"
: "${s3_endpoint_host:=unset}"


# Just need these to get the stack info
export AWS_ACCESS_KEY_ID=${access_key_id}
export AWS_SECRET_ACCESS_KEY=${secret_access_key}
export AWS_DEFAULT_REGION=${region_name}
export ASSUME_ROLE_ARN=${assume_role_arn}

# Some of these are optional
export ACCESS_KEY_ID=${access_key_id}
export SECRET_ACCESS_KEY=${secret_access_key}
export REGION=${region_name}
export S3_HOST=${s3_endpoint_host}

pushd "${release_dir}" > /dev/null
  echo -e "\n running tests with $(go version)..."
  scripts/ginkgo -r --focus="${focus_regex}" integration/
popd > /dev/null
