FROM ubuntu:20.04
ARG ARCH=amd64
# FROM arm=armhf/ubuntu:20.04 arm64=arm64v8/ubuntu:20.04

RUN apt-get update && \
    apt-get install -y gcc ca-certificates git wget curl vim less file kmod iptables xz-utils zip && \
    rm -f /bin/sh && ln -s /bin/bash /bin/sh

ENV GOLANG_ARCH_amd64=amd64 GOLANG_ARCH_arm=armv6l GOLANG_ARCH_arm64=arm64 GOLANG_ARCH=GOLANG_ARCH_${ARCH} \
    GOPATH=/go PATH=/go/bin:/usr/local/go/bin:${PATH} SHELL=/bin/bash

RUN wget -O - https://storage.googleapis.com/golang/go1.18.3.linux-${!GOLANG_ARCH}.tar.gz | tar -xzf - -C /usr/local

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.46.2

RUN curl -LO https://github.com/magefile/mage/releases/download/v1.14.0/mage_1.14.0_Linux-64bit.tar.gz && \
    tar -xf mage_1.14.0_Linux-64bit.tar.gz && \
    mv mage /usr/local/bin/ && \
    rm mage_1.14.0_Linux-64bit.tar.gz

WORKDIR /go/src/github.com/rancher/rke/
COPY . /go/src/github.com/rancher/rke/

RUN mage build

