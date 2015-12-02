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
export s3cmd_host=$(cat ${PWD}/configs/s3_endpoint_host)

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

# Note there is a subtle use of bash environment variables where some are
# evaluated by the here-doc and some are injected into the script by the
# here-doc to be evaluated by the bats runtime.

bats_file=dynamic.bats
cat s3cli-src/integration/test.bats > ${bats_file}
export S3CLI_CONFIGS_DIR=${configs_dir}
for S3CLI_CONFIG_FILE in ${configs_dir}/*-s3cli_config.json; do
  cat >> "${bats_file}" << EOF

@test "Invoking s3cli get with nonexistent key should output error using config: ${S3CLI_CONFIG_FILE}" {
  local non_existant_file=\${BATS_RANDOM_ID}

  run_local_or_remote "\${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} get \${non_existant_file} empty_file"

  [ "\${status}" -ne 0 ]
  [[ "\${output}" =~ "NoSuchKey" ]]
}

@test "Invoking s3cli get with existing key should return the correct file using config: ${S3CLI_CONFIG_FILE}" {
  local expected_string=\${BATS_RANDOM_ID}
  local s3_filename="existing_file_in_s3"

  echo -n \${expected_string} > \${s3_filename}
  s3cmd --config ${S3CMD_CONFIG_FILE} put \${s3_filename} s3://${bucket_name}/

  run_local_or_remote "\${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} get \${s3_filename} gotten_file"

  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/\${s3_filename}

  if [ ! -z \${test_host} ]; then
    scp ec2-user@\${test_host}:./gotten_file ./
  fi
  local actual_string=\$(cat gotten_file)
  print_debug "actual_string" "\${actual_string}"

  [ "\${status}" -eq 0 ]
  [ "\${expected_string}" = "\${actual_string}" ]
}

@test "Invoking s3cli put should correctly write the file to the bucket using config: ${S3CLI_CONFIG_FILE}" {
  local expected_string=\${BATS_RANDOM_ID}

  echo -n \${expected_string} > file_to_upload
  if [ ! -z \${test_host} ]; then
    scp file_to_upload ec2-user@\${test_host}:~/file_to_upload
  fi

  run_local_or_remote "\${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} put file_to_upload uploaded_by_s3"

  s3cmd --config ${S3CMD_CONFIG_FILE} get s3://${bucket_name}/uploaded_by_s3 uploaded_by_s3 --force
  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/uploaded_by_s3
  local actual_string=\$(cat uploaded_by_s3)

  [ "\${status}" -eq 0 ]
  [ "\${expected_string}" = "\${actual_string}" ]
}
EOF
done

bats ${bats_file}
