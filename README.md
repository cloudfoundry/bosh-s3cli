# S3 CLI written in pure Go [![Build Status](https://travis-ci.org/pivotal-golang/s3cli.png?branch=master)](https://travis-ci.org/pivotal-golang/s3cli)

## Using Go AMZ

We are using Go AMZ library under the LGPL v3 License for this CLI.
To change the version being used by the CLI follow these steps:

 * Install go
 * Change the source code used for GoAMZ (whichever version you want), it's under src/launchpad.net/goamz
 * Compile the new version with go/build
 * Get the updated binary from out/s3

See GoAMZ home page for more details on the library: https://wiki.ubuntu.com/goamz
