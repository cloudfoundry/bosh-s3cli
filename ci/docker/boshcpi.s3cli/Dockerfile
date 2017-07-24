FROM ubuntu:latest

ENV LANG C
ENV LC_ALL C

RUN apt-get update; apt-get -y upgrade; apt-get clean
RUN apt-get install -y realpath git curl python python-pip python-dateutil python-magic zip jq; apt-get clean; pip install awscli

RUN cd /tmp && \
    curl -O -L https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go*.tar.gz && \
    rm go*.tar.gz

RUN mkdir -p /opt/go
ENV GOPATH /opt/go
ENV PATH /usr/local/go/bin:/opt/go/bin:$PATH

RUN go get github.com/onsi/ginkgo/ginkgo; go get github.com/golang/lint/golint
