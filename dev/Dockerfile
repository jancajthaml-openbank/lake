# Copyright (c) 2017-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ---------------------------------------------------------------------------- #

FROM debian:stretch AS base

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8

RUN dpkg --add-architecture armhf
RUN dpkg --add-architecture amd64

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        apt-utils \
        ca-certificates>=20161130 \
        git>=1:2.11.0-3 \
        curl>=7.52.1-5

# ---------------------------------------------------------------------------- #

FROM base AS go

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    GOLANG_VERSION=1.11.1 \
    GOPATH=/go

RUN apt-get -y install --no-install-recommends tar=1.29b-1.1

RUN curl -L "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    -# \
    -o /tmp/go-pkg.tar.gz && \
    \
    tar -C /usr/local -xzf /tmp/go-pkg.tar.gz && \
    mv /usr/local/go/bin/go /usr/bin/go && \
    mv /usr/local/go/bin/godoc /usr/bin/godoc && \
    mv /usr/local/go/bin/gofmt /usr/bin/gofmt

RUN curl -L https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 \
    -# \
    -o /usr/bin/dep && \
    \
    chmod +x /usr/bin/dep

RUN curl -L https://github.com/Masterminds/glide/releases/download/v0.13.1/glide-v0.13.1-linux-amd64.tar.gz \
    -# \
    -o /tmp/glide-pkg.tar.gz && \
    \
    tar -C /usr/lib -xzf /tmp/glide-pkg.tar.gz && \
    mv /usr/lib/linux-amd64/glide /usr/bin/glide

RUN go get -u \
        \
        golang.org/x/lint/golint \
        github.com/fzipp/gocyclo \
        github.com/client9/misspell/cmd/misspell \
        github.com/alexkohler/prealloc \
        github.com/mdempsky/maligned \
        github.com/jgautheron/goconst/cmd/goconst

RUN curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b /usr/bin latest

# ---------------------------------------------------------------------------- #

FROM base

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LIBRARY_PATH=/usr/lib \
    LD_LIBRARY_PATH=/usr/lib \
    CGO_ENABLED=1 \
    GOPATH=/go \
    GOARCH=amd64 \
    GOOS=linux \
    GOHOSTOS=linux \
    CC=gcc \
    CXX=g++

RUN apt-get -y install --no-install-recommends \
    \
        cmake=3.7.2-1 \
        make=4.1-9.1 \
        patch>=2.7.5-1 \
        debhelper=10.2.5 \
        config-package-dev=5.1.2 \
        fakeroot=1.21-3.1 \
        pkg-config>=0.29-4 \
        gcc \
        gcc-arm-linux-gnueabi \
        gcc-arm-linux-gnueabihf \
        libc6 \
        libc6-armhf-cross \
        libc6-dev \
        libc6-dev-armhf-cross \
    \
        libzmq3-dev:amd64=4.2.1-4 \
        libzmq3-dev:armhf=4.2.1-4 \
    && \
    \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=go /go /go
COPY --from=go /usr/local/go /usr/local/go
COPY --from=go /usr/bin/go /usr/bin/go
COPY --from=go /usr/bin/dep /usr/bin/dep
COPY --from=go /usr/bin/gosec /usr/bin/gosec
COPY --from=go /usr/bin/gofmt /usr/bin/gofmt
COPY --from=go /usr/bin/godoc /usr/bin/godoc
COPY --from=go /usr/bin/glide /usr/bin/glide
COPY --from=go /go/bin/gocyclo /usr/bin/gocyclo
COPY --from=go /go/bin/golint /usr/bin/golint
COPY --from=go /go/bin/misspell /usr/bin/misspell
COPY --from=go /go/bin/prealloc /usr/bin/prealloc
COPY --from=go /go/bin/maligned /usr/bin/maligned
COPY --from=go /go/bin/goconst /usr/bin/goconst
COPY --from=jancajthaml/jq /usr/local/bin/jq /usr/bin/jq
COPY --from=library/docker:18.06 /usr/local/bin/docker /usr/bin/docker

# ---------------------------------------------------------------------------- #
