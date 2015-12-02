#!/usr/bin/env bats

set -e

print_debug() {
  local output=$1
  echo "BATS_DEBUG-[$(date)] ${output}"
}

run_local_or_remote() {
  local cmd=$1
  if [ ! -z ${test_host} ]; then
    cmd="ssh ec2-user@${test_host} ${cmd}"
  fi
  run ${cmd}
  print_debug "-------------"
  print_debug "Test:        ${BATS_TEST_DESCRIPTION}"
  print_debug "Config file: ${S3CLI_CONFIG_FILE}"
  print_debug "Command:     ${cmd}"
  print_debug "Status:      ${status}"
  print_debug "Output:      ${output}"
}

setup() {
  : "${S3CLI_CONFIGS_DIR:?Need to set S3CLI_CONFIGS_DIR non-empty}"
  : "${S3CMD_CONFIG_FILE:?Need to set S3CMD_CONFIG non-empty}"
  : "${S3CLI_EXE:?Need to set S3CLI_EXE non-empty}"
  export BATS_RANDOM_ID=$(uuidgen)

  if [ ! -z ${test_host} ]; then
    ssh ec2-user@${test_host} "mkdir -p ~/configs"
    scp -r ${S3CLI_CONFIGS_DIR}/*.json  ec2-user@${test_host}:~/configs/
    scp ${S3CLI_EXE} ec2-user@${test_host}:~/s3cli
    S3CLI_EXE="~/s3cli"
  fi
}
