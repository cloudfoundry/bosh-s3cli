#!/usr/bin/env bats

set -e

run_local_or_remote() {
  local cmd=$1
  if [ ! -z ${test_host} ]; then
    cmd="ssh ec2-user@${test_host} ${cmd}"
  fi
  run ${cmd}
  echo "-------------"
  echo   "Test:        ${BATS_TEST_DESCRIPTION}"
  printf "Config file: ${current_config_file}\n$(cat ${current_config_file})\n"
  echo   "Command:     ${cmd}"
  echo   "Status:      ${status}"
  echo   "Output:      ${output}"
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
