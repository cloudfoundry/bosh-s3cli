#!/usr/bin/env bash

project_dir="$( cd "$(dirname "${0}")/.." && pwd )"

fly -t bosh sp -p bosh-s3cli -c "${project_dir}/ci/pipeline.yml"
