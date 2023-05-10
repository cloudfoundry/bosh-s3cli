## S3 CLI

A CLI for uploading, fetching and deleting content to/from an S3-compatible
blobstore.

Continuous integration: <https://bosh-cpi.ci.cf-app.com/pipelines/s3cli>

Releases can be found in https://s3cli-artifacts.s3.amazonaws.com. The Linux binaries follow the regex `s3cli-(\d+\.\d+\.\d+)-linux-amd64` and the windows binaries `s3cli-(\d+\.\d+\.\d+)-windows-amd64`.

## Installation

```
go get github.com/cloudfoundry/bosh-s3cli
```

## Usage

Given a JSON config file (`config.json`)...

``` json
{
  "bucket_name":            "<string> (required)",

  "credentials_source":     "<string> [static|env_or_profile|none]",
  "access_key_id":          "<string> (required if credentials_source = 'static')",
  "secret_access_key":      "<string> (required if credentials_source = 'static')",

  "region":                 "<string> (optional - default: 'us-east-1')",
  "host":                   "<string> (optional)",
  "port":                   <int> (optional),

  "ssl_verify_peer":        <bool> (optional),
  "use_ssl":                <bool> (optional),
  "signature_version":      "<string> (optional)",
  "server_side_encryption": "<string> (optional)",
  "sse_kms_key_id":         "<string> (optional)",
  "multipart_upload":       <bool> (optional - default: true)
}
```

``` bash
# Usage
s3cli --help

# Command: "put"
# Upload a blob to an S3-compatible blobstore.
s3cli -c config.json put <path/to/file> <remote-blob>

# Command: "get"
# Fetch a blob from an S3-compatible blobstore.
# Destination file will be overwritten if exists.
s3cli -c config.json get <remote-blob> <path/to/file>

# Command: "delete"
# Remove a blob from an S3-compatible blobstore.
s3cli -c config.json delete <remote-blob>

# Command: "exists"
# Checks if blob exists in an S3-compatible blobstore.
s3cli -c config.json exists <remote-blob>

# Command: "sign"
# Create a self-signed url for an object
s3cli -c config.json sign <remote-blob> <get|put> <seconds-to-expiration>
```

## Contributing

Follow these steps to make a contribution to the project:

- Fork this repository
- Create a feature branch based upon the `main` branch (*pull requests must be made against this branch*)
  ``` bash
  git checkout -b feature-name origin/main
  ```
- Run tests to check your development environment setup
  ``` bash
  scripts/ginkgo -r -race --skip-package=integration ./
  ```
- Make your changes (*be sure to add/update tests*)
- Run tests to check your changes
  ``` bash
  scripts/ginkgo -r -race --skip-package=integration ./
  ```
- Push changes to your fork
  ``` bash
  git add .
  git commit -m "Commit message"
  git push origin feature-name
  ```
- Create a GitHub pull request, selecting `main` as the target branch

## Running integration tests

To run the integration tests, export the following variables into your environment:

```
export access_key_id=YOUR_AWS_ACCESS_KEY
export focus_regex="GENERAL AWS|AWS V2 REGION|AWS V4 REGION|AWS US-EAST-1"
export region_name=us-east-1
export s3_endpoint_host=s3.amazonaws.com
export secret_access_key=YOUR_SECRET_ACCESS_KEY
export stack_name=s3cli-iam
export bucket_name=s3cli-pipeline
```

Run `ci/tasks/setup-aws-infrastructure.sh` and `ci/tasks/teardown-infrastructure.sh` before and after the `run-integration-*` tests in `ci/tasks`.
