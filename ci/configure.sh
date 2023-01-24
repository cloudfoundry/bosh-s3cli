#!/usr/bin/env bash

project_dir="$( cd "$(dirname "${0}")/.." && pwd )"

if [[ $(lpass status -q; echo $?) != 0 ]]; then
  echo "Login with lpass first"
  exit 1
fi

fly -t bosh-ecosystem sp -p bosh-s3cli -c "${project_dir}/ci/pipeline.yml"
