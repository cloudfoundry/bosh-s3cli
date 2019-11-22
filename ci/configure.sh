#!/usr/bin/env bash

until lpass status;do
  LPASS_DISABLE_PINENTRY=1 lpass ls a
  sleep 1
done

until fly -t director status;do
  fly -t director login
  sleep 1
done

fly -t director sp -p s3cli -c ${PROJECT_DIR}/ci/pipeline.yml \
  -l <(lpass show --notes "s3cli concourse secrets") \
  -l <(lpass show --notes "pivotal-tracker-resource-keys") \
  -l <(lpass show --note "bosh:docker-images concourse secrets")
