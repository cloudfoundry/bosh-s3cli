FROM golang:1

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
