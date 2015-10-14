#!/usr/bin/env bats

# Required configuration:
#   - $CONFIG_FILE must be set for use with the s3cli
#   - $BATS_LOG must be set to the path of a log file to append to
#   - $BUCKET_NAME must be set to the name of the s3 bucket
#   - $HOME/.s3cfg must be set for s3cmd

setup() {
  # Preload the target s3 bucket with a known file, then delete the file
  export BATS_KNOWN_FILE=$(cat /proc/sys/kernel/random/uuid)
  export TEST_STRING=$(cat /proc/sys/kernel/random/uuid)
  local_known_file=${BATS_TMPDIR}/${BATS_KNOWN_FILE}

  echo -n ${TEST_STRING} > ${local_known_file} 
  s3cmd put ${local_known_file} s3://${BUCKET_NAME}/
  rm ${local_known_file}
}

teardown() {
  # Clean up the s3 bucket by removing the known file
  s3cmd del s3://${BUCKET_NAME}/${BATS_KNOWN_FILE}
  rm -f ${BATS_TMPDIR}/${BATS_KNOWN_FILE}
}

print_debug() {
  description=$1
  output=$2

  cat >> ${BATS_LOG} << EOF
---
Output of "${description}":
  ${output}
EOF
}

@test "Invoking s3cli get with nonexistent key should output error" {
  random_uuid=$(cat /proc/sys/kernel/random/uuid)

  run out/s3 -c ${CONFIG_FILE} get $random_uuid $random_uuid

  print_debug "${BATS_TEST_DESCRIPTION}" "${output}"

  [ "${status}" -ne 0 ]
  [ "${output}" = "Error: The specified key does not exist." -o "${output}" = "Error: 404 Not Found" ]
}

@test "Invoking s3cli get with existing key should return the correct file" {
  local_known_file=${BATS_TMPDIR}/${BATS_KNOWN_FILE}

  run out/s3 -c ${CONFIG_FILE} get ${BATS_KNOWN_FILE} ${local_known_file}

  print_debug "${BATS_TEST_DESCRIPTION}" "${output}"

  [ "${status}" -eq 0 ]
  [ "${TEST_STRING}" = "$(cat ${local_known_file})" ]
}

@test "Invoking s3cli put should correctly write the file to the bucket" {
  local_known_file=${BATS_TMPDIR}/${BATS_KNOWN_FILE}
  random_uuid=$(cat /proc/sys/kernel/random/uuid)

  echo -n ${random_uuid} > ${local_known_file}
  run out/s3 -c $CONFIG_FILE put ${local_known_file} ${BATS_KNOWN_FILE} 
  rm ${local_known_file}

  print_debug "${BATS_TEST_DESCRIPTION}" "${output}"

  s3cmd get s3://${BUCKET_NAME}/${BATS_KNOWN_FILE} ${local_known_file}
  [ "${status}" -eq 0 ]
  [ "${random_uuid}" = "$(cat ${local_known_file})" ]
}
