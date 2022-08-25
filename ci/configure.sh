#!/usr/bin/env bash

project_dir="$( cd "$(dirname "${0}")/.." && pwd )"

until lpass status;do
  LPASS_DISABLE_PINENTRY=1 lpass ls a
  sleep 1
done

fly -t bosh-ecosystem sp -p bosh-s3cli -c "${project_dir}/ci/pipeline.yml" \
  -l <(lpass show --notes "s3cli concourse secrets") \
  -l <(lpass show --notes "pivotal-tracker-resource-keys") \
  -l <(lpass show --note "bosh:docker-images concourse secrets")
