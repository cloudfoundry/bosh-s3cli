@test "Negative - Invoking s3cli get with region and invalid host should fail using config: ${S3CLI_CONFIG_FILE}" {
  local s3_filename="existing_file_in_s3"
  echo -n ${BATS_RANDOM_ID} > ${s3_filename}

  s3cmd --config ${S3CMD_CONFIG_FILE} put ${s3_filename} s3://${bucket_name}/

  current_config_file=${S3CLI_CONFIG_FILE}
  run_local_or_remote "${S3CLI_EXE} -c ${S3CLI_CONFIG_FILE} get ${s3_filename} found_file"

  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/${s3_filename}

  [ "${status}" -ne 0 ]
  non_isolated_region_regex="the region '.*' is wrong; expecting '.*'"
  isolated_region_regex="InvalidAccessKeyId"
  [[ $output =~ $non_isolated_region_regex || $output =~ $isolated_region_regex ]]
}
