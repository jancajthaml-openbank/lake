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

FROM library/debian:stretch

MAINTAINER Jan Cajthaml <jan.cajthaml@gmail.com>

ENV DEBIAN_FRONTEND=noninteractive \
    GOLANG_VERSION=1.9 \
    LD_LIBRARY_PATH=/usr/lib \
    GOPATH=/go \
    GOARCH=amd64 \
    GOOS=linux \
    GOROOT=/usr/lib/go \
    OSXCROSS_GCC_NO_STATIC_RUNTIME=1 \
    OSXCROSS_NO_INCLUDE_PATH_WARNINGS=1

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
        ca-certificates=20161130+nmu1 \
        pkg-config=0.29-4+b1 \
        libzmq3-dev=4.2.1-4 \
        cmake=3.7.2-1 \
        gcc=4:6.3.0-4 \
        g++=4:6.3.0-4 \
        zlib1g-dev=1:1.2.8.dfsg-5 \
        libmpc-dev=1.0.3-1+b2 \
        libmpfr-dev=3.1.5-1 \
        libgmp-dev=2:6.1.2+dfsg-1 \
        patch=2.7.5-1+b2 \
        git=1:2.11.0-3+deb9u2 \
        wget=1.18-5+deb9u1 \
        tar=1.29b-1.1 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.${GOOS}-${GOARCH}.tar.gz" && \
            echo 'd70eadefce8e160638a9a6db97f7192d8463069ab33138893ad3bf31b0650a79 *go.tgz' | sha256sum -c - && \
            tar -C "${LD_LIBRARY_PATH}" -xzf go.tgz && rm go.tgz

ENV PATH "${GOPATH}/bin:${LD_LIBRARY_PATH}/go/bin:${PATH}"

RUN go get -u \
        \
        github.com/golang/dep/cmd/dep \
        github.com/Masterminds/glide \
        github.com/golang/lint/golint \
        github.com/fzipp/gocyclo \
        github.com/client9/misspell/cmd/misspell
