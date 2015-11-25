FROM ubuntu:latest

RUN locale-gen en_US.UTF-8
RUN dpkg-reconfigure locales
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8

RUN apt-get update; apt-get -y upgrade; apt-get clean
RUN apt-get install -y git curl python uuid-runtime; apt-get clean

RUN cd /tmp && \
    curl -O -L https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go*.tar.gz && \
    rm go*.tar.gz

ENV PATH /usr/local/go/bin:$PATH

RUN apt-get install -y python-dateutil python-magic; apt-get clean

RUN cd /tmp && \
    git clone https://github.com/sstephenson/bats.git && \
    cd bats && \
    ./install.sh /usr/local && \
    cd / && \
    rm -rf /tmp/bats

RUN cd /tmp && \
    curl -O -L https://github.com/s3tools/s3cmd/archive/v1.6.0.tar.gz && \
    tar xzf v1.6.0.tar.gz && \
    cd s3cmd-1.6.0 && \
    cp -R s3cmd S3 /usr/local/bin && \
    cd /tmp && \
    rm -rf s3cmd-1.6.0/ v1.6.0.tar.gz

RUN apt-get install -y python python-pip groff jq; apt-get clean; pip install awscli;
