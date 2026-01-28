## S3 CLI

A CLI for uploading, fetching and deleting content to/from an S3-compatible
blobstore.

Continuous integration: <https://bosh-cpi.ci.cf-app.com/pipelines/s3cli>

Releases can be found in `https://s3.amazonaws.com/bosh-s3cli-artifacts`. The Linux binaries follow the regex `s3cli-(\d+\.\d+\.\d+)-linux-amd64` and the windows binaries `s3cli-(\d+\.\d+\.\d+)-windows-amd64`.

## Installation

```
go get github.com/cloudfoundry/bosh-s3cli
```

## Usage

Given a JSON config file (`config.json`)...

``` json
{
  "bucket_name":                           "<string> (required)",

  "credentials_source":                    "<string> [static|env_or_profile|none]",
  "access_key_id":                         "<string> (required if credentials_source = 'static')",
  "secret_access_key":                     "<string> (required if credentials_source = 'static')",

  "region":                                "<string> (optional - default: 'us-east-1')",
  "host":                                  "<string> (optional)",
  "port":                                  "<int> (optional)",

  "ssl_verify_peer":                       "<bool> (optional - default: true)",
  "use_ssl":                               "<bool> (optional - default: true)",
  "signature_version":                     "<string> (optional)",
  "server_side_encryption":                "<string> (optional)",
  "sse_kms_key_id":                        "<string> (optional)",
  "multipart_upload":                      "<bool> (optional - default: true)",
  "download_concurrency":                  "<string> (optional - default: '5')",
  "download_part_size":                    "<string> (optional - default: '5242880')", # 5 MB
  "upload_concurrency":                    "<string> (optional - default: '5')",
  "upload_part_size":                      "<string> (optional - default: '5242880')" # 5 MB
  "disable_checksums":                     "<bool> (optional - default: false)"
}
```

> Note: **multipart_upload** is not supported by Google - it's automatically set to false by parsing the provided 'host'

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
### Steps to run the integration tests on AWS
1. Export the following variables into your environment
```
export access_key_id=<YOUR_AWS_ACCESS_KEY>
export focus_regex="GENERAL AWS|AWS V2 REGION|AWS V4 REGION|AWS US-EAST-1"
export region_name=us-east-1
export s3_endpoint_host=s3.amazonaws.com
export secret_access_key=<YOUR_SECRET_ACCESS_KEY>
export stack_name=s3cli-iam
export bucket_name=s3cli-pipeline
```
2. Setup infrastructure with `ci/tasks/setup-aws-infrastructure.sh`
3. Run the desired tests by executing one or more of the scripts `run-integration-*` in `ci/tasks` (to run `run-integration-s3-compat` see [Setup for GCP](#setup-for-GCP) or [Setup for AliCloud](#setup-for-alicloud))
4. Teardown infrastructure with `ci/tasks/teardown-infrastructure.sh`

### Setup for GCP
1. Create a bucket in GCP
2. Create access keys
3. Navigate to **IAM & Admin > Service Accounts**.
4. Select your service account or create a new one if needed.
5. Ensure your service account has necessary permissions (like `Storage Object Creator`, `Storage Object Viewer`, `Storage Admin`) depending on what access you want.
6. Go to **Cloud Storage** and select **Settings**.
7. In the **Interoperability** section, create an HMAC key for your service account. This generates an "access key ID" and a "secret access key".
8. Export the following variables into your environment:
```
export access_key_id=<YOUR_ACCESS_KEY>
export secret_access_key=<YOUR_SECRET_ACCESS_KEY>
export bucket_name=<YOUR_BUCKET_NAME>
export s3_endpoint_host=storage.googleapis.com
export s3_endpoint_port=443
```
4. Run `run-integration-s3-compat.sh` in `ci/tasks`

### Setup for AliCloud
1. Create bucket in AliCloud
2. Create access keys from `RAM -> User -> Create Accesskey`
3. Export the following variables into your environment:
```
export access_key_id=<YOUR_ACCESS_KEY>
export secret_access_key=<YOUR_SECRET_ACCESS_KEY>
export bucket_name=<YOUR_BUCKET_NAME>
export s3_endpoint_host="oss-<YOUR_REGION>.aliyuncs.com"
export s3_endpoint_port=443
```
4. Run `run-integration-s3-compat.sh` in `ci/tasks`
