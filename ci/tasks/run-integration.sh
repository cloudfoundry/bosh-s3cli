#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key

configs_dir=./configs
if [ -e ${PWD}/configs/test_host_ip ]; then
  export test_host=$(cat ${PWD}/configs/test_host_ip)
  private_key=${PWD}/private_key.pem
  echo "${private_key_data}" > ${private_key}
  chmod 600 ${private_key}
  eval $(ssh-agent)
  ssh-add ${private_key}

  mkdir -p ~/.ssh
  cat > ~/.ssh/config << EOF
  Host *
    StrictHostKeyChecking no
EOF
fi

export bucket_name=$(cat ${PWD}/configs/bucket_name)
s3cmd_host=$(cat ${PWD}/configs/s3_endpoint_host)

pushd s3cli-src > /dev/null
  # Hack to get bats to work with concourse
  export TERM=xterm
  . .envrc
  go install s3cli/s3cli

  export S3CLI_EXE=$(which s3cli)
popd > /dev/null

export S3CMD_CONFIG_FILE="s3cmd.s3cfg"
cat > "${S3CMD_CONFIG_FILE}" << EOF
[default]
access_key = ${access_key_id}
secret_key = ${secret_access_key}
bucket_location = ${region_name}
host_base = ${s3cmd_host}
host_bucket = %(bucket)s.${s3cmd_host}
enable_multipart = True
multipart_chunk_size_mb = 15
use_https = True
EOF

bats_file=test.bats
cat s3cli-src/ci/assets/setup-template.bats > ${bats_file}
export S3CLI_CONFIGS_DIR=${configs_dir}
for file in ${configs_dir}/*-s3cli_config.json; do
  export S3CLI_CONFIG_FILE=${file}

  # We pass the `shell-format` parameter to envsubst to limit which environment
  # variables should be interpolated, leaving the rest to be interpolated at bats runtime.
  # See: https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html

  cat s3cli-src/ci/assets/examples-template.bats | \
  envsubst "\$S3CLI_CONFIG_FILE \$bucket_name \$S3CMD_CONFIG_FILE" >> "${bats_file}"
done

bats ${bats_file}
