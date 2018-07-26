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
    LD_LIBRARY_PATH=/usr/lib

RUN apt-get update && apt-get -y install --no-install-recommends \
    \
        ca-certificates>=20161130 \
        git>=1:2.11.0-3 \
        curl>=7.52.1-5

# ---------------------------------------------------------------------------- #

FROM base AS go

ENV DEBIAN_FRONTEND=noninteractive \
    GOLANG_VERSION="1.10.1" \
    LD_LIBRARY_PATH=/usr/lib \
    GOPATH=/go \
    GOARCH=amd64 \
    GOOS=linux \
    GOROOT=/usr/lib/go

RUN apt-get -y install --no-install-recommends \
    \
        tar=1.29b-1.1

RUN curl -L "https://golang.org/dl/go${GOLANG_VERSION}.${GOOS}-${GOARCH}.tar.gz" -o /tmp/go-pkg.tar.gz && \
    tar -C "${LD_LIBRARY_PATH}" -xzf /tmp/go-pkg.tar.gz && \
    mv "${LD_LIBRARY_PATH}"/go/bin/go /usr/bin/go && \
    mv "${LD_LIBRARY_PATH}"/go/bin/godoc /usr/bin/godoc && \
    mv "${LD_LIBRARY_PATH}"/go/bin/gofmt /usr/bin/gofmt

RUN curl -L https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 -o /usr/bin/dep && \
    chmod +x /usr/bin/dep

RUN curl -L https://github.com/Masterminds/glide/releases/download/v0.13.1/glide-v0.13.1-linux-amd64.tar.gz -o /tmp/glide-pkg.tar.gz && \
    tar -C "${LD_LIBRARY_PATH}" -xzf /tmp/glide-pkg.tar.gz && \
    mv "${LD_LIBRARY_PATH}"/linux-amd64/glide /usr/bin/glide

RUN go get -u \
        \
        github.com/golang/lint/golint \
        github.com/fzipp/gocyclo \
        github.com/client9/misspell/cmd/misspell

# ---------------------------------------------------------------------------- #

FROM base AS jq

ENV DEBIAN_FRONTEND=noninteractive \
    JQ_VERSION="1.5"

RUN curl -L "https://github.com/stedolan/jq/releases/download/jq-${JQ_VERSION}/jq-linux64" -o /usr/bin/jq && \
    chmod +x /usr/bin/jq

# ---------------------------------------------------------------------------- #

FROM base

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LD_LIBRARY_PATH=/usr/lib \
    CGO_ENABLED=1 \
    GOPATH=/go \
    GOARCH=amd64 \
    GOOS=linux \
    GOROOT=/usr/lib/go

RUN apt-get -y install --no-install-recommends \
    \
        libzmq3-dev=4.2.1-4 \
        cmake=3.7.2-1 \
        make=4.1-9.1 \
        gcc=4:6.3.0-4 \
        g++=4:6.3.0-4 \
        zlib1g-dev=1:1.2.8.dfsg-5 \
        libmpc-dev=1.0.3-1+b2 \
        libmpfr-dev=3.1.5-1 \
        libgmp-dev=2:6.1.2+dfsg-1 \
        patch>=2.7.5-1 \
        build-essential=12.3 \
        debhelper=10.2.5 \
        config-package-dev=5.1.2 \
        devscripts>=2.17.6 \
        fakeroot=1.21-3.1 \
        pkg-config>=0.29-4 \
    && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN curl -L https://download.docker.com/linux/static/stable/x86_64/docker-17.09.0-ce.tgz | \
    tar -xzvf - --strip-components=1 -C /usr/bin docker/docker && \
    chmod a+x /usr/bin/docker

COPY --from=jq /usr/bin/jq /usr/bin/jq

COPY --from=go /go /go
COPY --from=go /usr/lib/go /usr/lib/go
COPY --from=go /usr/bin/go /usr/bin/go
COPY --from=go /usr/bin/dep /usr/bin/dep
COPY --from=go /usr/bin/gofmt /usr/bin/gofmt
COPY --from=go /usr/bin/godoc /usr/bin/godoc
COPY --from=go /usr/bin/glide /usr/bin/glide
COPY --from=go /go/bin/gocyclo /usr/bin/gocyclo
COPY --from=go /go/bin/golint /usr/bin/golint
COPY --from=go /go/bin/misspell /usr/bin/misspell

# ---------------------------------------------------------------------------- #
