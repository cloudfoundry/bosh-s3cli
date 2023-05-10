#!/usr/bin/env bash
set -euo pipefail

my_dir="$( cd "$(dirname "${0}")" && pwd )"
release_dir="$( cd "${my_dir}" && cd ../.. && pwd )"

semver='1.2.3.4'
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

pushd "${release_dir}" > /dev/null
  git_rev=$(git rev-parse --short HEAD)
  version="${semver}-${git_rev}-${timestamp}"

  echo -e "\n Vetting packages for potential issues..."
  if ! command -v golangci-lint &> /dev/null; then
    go_bin="$(go env GOPATH)/bin"
    export PATH=${go_bin}:${PATH}
    go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  fi
  golangci-lint run --enable goimports ./...

  echo -e "\n Unit testing packages..."
  scripts/ginkgo -r -race --skip-package=integration ./

  echo -e "\n Running build script to confirm everything compiles..."
  go build -ldflags "-X main.version=${version}" -o out/s3cli .

  echo -e "\n Testing version information"
  app_version=$(out/s3cli -v)
  test "${app_version}" = "version ${version}"
popd > /dev/null
