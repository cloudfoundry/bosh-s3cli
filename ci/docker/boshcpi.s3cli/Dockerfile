FROM ubuntu:xenial

# install base utils {
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    curl \
    git \
    jq \
    python3-pip \
    python3-setuptools \
    wget \
    zip \
  && pip3 install awscli \
  && apt-get autoremove \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*
# }

# golang {
WORKDIR /usr/local

RUN \
  GO_INFO=$(curl 'https://golang.org/dl/?mode=json' | jq '.[0].files[] | select(.os == "linux" and .arch == "amd64")') && \
  GO_TAR="$(echo "$GO_INFO" | jq -r '.filename')" && \
  GO_SHA="$(echo "$GO_INFO" | jq -r '.sha256')" && \
  curl -fSL https://storage.googleapis.com/golang/$GO_TAR -o $GO_TAR \
  && echo $GO_SHA $GO_TAR | sha256sum -c - \
  && tar -xzf $GO_TAR \
  && rm -f $GO_TAR

ENV \
  PATH=/usr/local/go/bin:$PATH \
  GOPATH=/root/go \
  GOROOT=/usr/local/go
# }

# install go packages {
RUN \
  go get github.com/onsi/ginkgo/ginkgo \
  && go get github.com/golang/lint/golint
# }
