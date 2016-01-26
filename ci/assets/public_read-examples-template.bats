@test "Invoking s3cli get with existing key should return the correct file using config: ${S3CLI_CONFIG_FILE}" {
  local expected_string=${BATS_RANDOM_ID}
  local s3_filename="existing_file_in_s3"

  echo -n ${expected_string} > ${s3_filename}
  s3cmd --config ${S3CMD_CONFIG_FILE} --acl-public put ${s3_filename} s3://${bucket_name}/

  current_config_file=${S3CLI_CONFIG_FILE}
  run_local_or_remote "${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} get ${s3_filename} gotten_file"

  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/${s3_filename}

  if [ ! -z ${test_host} ]; then
    scp ec2-user@${test_host}:./gotten_file ./
  fi
  local actual_string=$(cat gotten_file)

  [ "${status}" -eq 0 ]
  [ "${expected_string}" = "${actual_string}" ]
}
