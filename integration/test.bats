#!/usr/bin/env bats

setup() {
  export BATS_TMPDIR=$(mktemp -d /tmp/bats.XXXXXX)

  export CONFIG_FILE="${BATS_TMPDIR}/blobstore-s3.json"
  cat > "${CONFIG_FILE}"<< EOF
{
  "access_key_id": "${access_key_id}",
  "secret_access_key": "${secret_access_key}",
  "bucket_name": "${bucket_name}",
  "credentials_source": "static",
  "region": "${region_name}",
  "host": "${host}",
  "port": ${port},
  "use_ssl": true,
  "ssl_verify_peer": true
}
EOF

  export S3CMD_CONFIG_FILE="${BATS_TMPDIR}/s3cmd.s3cfg"
  cat > "${S3CMD_CONFIG_FILE}" << EOF
[default]
access_key = ${access_key_id}
secret_key = ${secret_access_key}
bucket_location = ${region_name}
host_base = ${host}
host_bucket = %(bucket)s.${host}
enable_multipart = True
multipart_chunk_size_mb = 15
use_https = True
EOF

  export BATS_RANDOM_ID=$(uuidgen)
}

teardown() {
  rm -rf ${BATS_TMPDIR}
}

print_debug() {
  description=$1
  output=$2

  echo "BATS_DEBUG-[$(date)] '${description}': ${output}" >> ${BATS_LOG}
}

@test "Invoking s3cli get with nonexistent key should output error" {
  local non_existant_file=${BATS_RANDOM_ID}

  run s3 -c ${CONFIG_FILE} get ${non_existant_file} ${BATS_TMPDIR}/empty_file
  print_debug "${BATS_TEST_DESCRIPTION}" "status:${status}, output:${output}"

  [ "${status}" -ne 0 ]
  [[ "${output}" =~ "NoSuchKey" ]]
}

@test "Invoking s3cli get with existing key should return the correct file" {
  local expected_string=${BATS_RANDOM_ID}
  local s3_filename="existing_file_in_s3"

  echo -n ${expected_string} > ${BATS_TMPDIR}/${s3_filename}
  s3cmd --config ${S3CMD_CONFIG_FILE} put ${BATS_TMPDIR}/${s3_filename} s3://${bucket_name}/

  run s3 -c ${CONFIG_FILE} get ${s3_filename} ${BATS_TMPDIR}/gotten_file
  print_debug "${BATS_TEST_DESCRIPTION}" "status:${status}, output:${output}"
  local actual_string=$(cat ${BATS_TMPDIR}/gotten_file)
  print_debug "actual_string" "${actual_string}"

  # Clean up the s3 bucket by removing the known file
  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/${s3_filename}

  [ "${status}" -eq 0 ]
  [ "${expected_string}" = "${actual_string}" ]
}

@test "Invoking s3cli put should correctly write the file to the bucket" {
  local expected_string=${BATS_RANDOM_ID}

  echo -n ${expected_string} > ${BATS_TMPDIR}/file_to_upload
  run s3 -c ${CONFIG_FILE} put ${BATS_TMPDIR}/file_to_upload uploaded_by_s3
  print_debug "${BATS_TEST_DESCRIPTION}" "status:${status}, output:${output}"

  s3cmd --config ${S3CMD_CONFIG_FILE} get s3://${bucket_name}/uploaded_by_s3 ${BATS_TMPDIR}/uploaded_by_s3
  s3cmd --config ${S3CMD_CONFIG_FILE} del s3://${bucket_name}/uploaded_by_s3
  local actual_string=$(cat ${BATS_TMPDIR}/uploaded_by_s3)

  [ "${status}" -eq 0 ]
  [ "${expected_string}" = "${actual_string}" ]
}
