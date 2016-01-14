#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key

# Hack to get bats to work with concourse
export TERM=xterm

pushd s3cli-src > /dev/null
  . .envrc
  go install s3cli/s3cli

  export S3CLI_EXE=$(which s3cli)
popd > /dev/null

export S3CLI_CONFIGS_DIR=./configs
if [ -e ${S3CLI_CONFIGS_DIR}/test_host_ip ]; then
  export test_host=$(cat ${S3CLI_CONFIGS_DIR}/test_host_ip)
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
  scp -r ${S3CLI_CONFIGS_DIR} ec2-user@${test_host}:/home/ec2-user/
  scp ${S3CLI_EXE} ec2-user@${test_host}:/home/ec2-user/s3cli
  S3CLI_EXE="/home/ec2-user/s3cli"
fi

export bucket_name=$(cat ${S3CLI_CONFIGS_DIR}/bucket_name)
s3cmd_host=$(cat ${S3CLI_CONFIGS_DIR}/s3_endpoint_host)

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
test_types=( generic negative_sig_version negative_region_invalid negative_region_and_host )
for test_type in "${test_types[@]}"; do
  for file in ${S3CLI_CONFIGS_DIR}/${test_type}/*-s3cli_config.json; do
    if [ -e "${file}" ]; then
      export S3CLI_CONFIG_FILE=${file}

      # We pass the `shell-format` parameter to envsubst to limit which environment
      # variables should be interpolated, leaving the rest to be interpolated at bats runtime.
      # See: https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html

      cat s3cli-src/ci/assets/${test_type}-examples-template.bats | \
      envsubst "\$S3CLI_CONFIG_FILE \$bucket_name \$S3CMD_CONFIG_FILE" >> "${bats_file}"
    fi
  done
done

bats ${bats_file}
