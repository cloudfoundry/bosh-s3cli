#!/usr/bin/env bash

project_dir="$( cd "$(dirname "${0}")/.." && pwd )"

fly -t "${CONCOURSE_TARGET:-storage-cli}" sp -p bosh-s3cli -c "${project_dir}/ci/pipeline.yml"
