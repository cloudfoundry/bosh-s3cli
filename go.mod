module github.com/cloudfoundry/bosh-s3cli

go 1.17

require (
	github.com/aws/aws-sdk-go v1.43.7
	github.com/cloudfoundry/bosh-utils v0.0.302
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
)

require (
	code.cloudfoundry.org/tlsconfig v0.0.0-20211123175040-23cc9f05b6b3 // indirect
	github.com/cloudfoundry/go-socks5 v0.0.0-20180221174514-54f73bdb8a8e // indirect
	github.com/cloudfoundry/socks5-proxy v0.2.44 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pivotal-cf/paraphernalia v0.0.0-20180203224945-a64ae2051c20 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220224120231-95c6836cb0e7 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.0.0-20201224043029-2b0845dc783e // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.30.15
