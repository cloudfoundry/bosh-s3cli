FROM bosh/golang-release

ENV DEBIAN_FRONTEND="noninteractive"

RUN apt-get update -y \
    && apt-get install -y --no-install-recommends \
        awscli \
        build-essential \
        curl \
        git \
        jq \
        wget \
        zip \
    &&  apt-get clean
