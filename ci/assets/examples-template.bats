@test "Invoking s3cli get with nonexistent key should output error using config: ${S3CLI_CONFIG_FILE}" {
  local non_existant_file=${BATS_RANDOM_ID}

  run_local_or_remote "${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} get ${non_existant_file} empty_file"

  [ "${status}" -ne 0 ]
  [[ "${output}" =~ "NoSuchKey" ]]
}

@test "Invoking s3cli get with existing key should return the correct file using config: ${S3CLI_CONFIG_FILE}" {
  local expected_string=${BATS_RANDOM_ID}
  local s3_filename="existing_file_in_s3"

  echo -n ${expected_string} > ${s3_filename}
  s3cmd --config ${S3CMD_CONFIG_FILE} put ${s3_filename} s3://${bucket_name}/

  run_local_or_remote "${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} get ${s3_filename} gotten_file"

  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/${s3_filename}

  if [ ! -z ${test_host} ]; then
    scp ec2-user@${test_host}:./gotten_file ./
  fi
  local actual_string=$(cat gotten_file)

  [ "${status}" -eq 0 ]
  [ "${expected_string}" = "${actual_string}" ]
}

@test "Invoking s3cli put should correctly write the file to the bucket using config: ${S3CLI_CONFIG_FILE}" {
  local expected_string=${BATS_RANDOM_ID}

  echo -n ${expected_string} > file_to_upload
  if [ ! -z ${test_host} ]; then
    scp file_to_upload ec2-user@${test_host}:~/file_to_upload
  fi

  run_local_or_remote "${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} put file_to_upload uploaded_by_s3"

  s3cmd --config ${S3CMD_CONFIG_FILE} get s3://${bucket_name}/uploaded_by_s3 uploaded_by_s3 --force
  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/uploaded_by_s3
  local actual_string=$(cat uploaded_by_s3)

  [ "${status}" -eq 0 ]
  [ "${expected_string}" = "${actual_string}" ]
}
