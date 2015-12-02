#!/usr/bin/env bash

set -e

source s3cli-src/ci/tasks/utils.sh

check_param access_key_id
check_param secret_access_key

COLOR_RED='\033[0;31m'
COLOR_GREEN='\033[0;32m'
COLOR_RESET='\033[0m'
export BATS_LOG="${PWD}/bats_output.log"
export BATS_ERRORS="${PWD}/bats_errors.log"
echo "" > ${BATS_LOG}
echo "" > ${BATS_ERRORS}

configs_dir=${PWD}/configs
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

set +e
combined_status=0
configurations_run=0
failed_runs=0
for file in configs/*-s3cli_config.json; do
  echo "### Running with ${file} #######################################################"
  echo "### Running with ${file} #######################################################" >> ${BATS_LOG}
  echo "$(cat ${file})" >> ${BATS_LOG}
  S3CLI_CONFIG_FILE=${file} bats s3cli-src/integration/test.bats
  status=$?
  let "configurations_run++"
#  printf "\n\n${COLOR_GREEN}\n"
#  cat ${BATS_LOG}
#  printf "${COLOR_RESET}\n\n"
  combined_status=$(($combined_status + ${status}))
  if [ ${status} -ne 0 ]; then
    let "failed_runs++"
    cat ${BATS_LOG} >> ${BATS_ERRORS}
  fi
  echo "" > ${BATS_LOG}
done
case "${combined_status}" in
 0) OUTPUT_COLOR=${COLOR_GREEN} ;;
 *) OUTPUT_COLOR=${COLOR_RED} ;;
esac
if [ ${combined_status} -ne 0 ]; then
  printf "\n\n${COLOR_RED}ERRORS${COLOR_RESET}:"
  cat ${BATS_ERRORS}
fi
printf "\n\n${OUTPUT_COLOR}${configurations_run} configurations run, ${failed_runs} failed${COLOR_RESET}\n"
exit $combined_status