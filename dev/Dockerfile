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

FROM library/debian:stretch AS base

ENV DEBIAN_FRONTEND=noninteractive \
    LD_LIBRARY_PATH=/usr/lib

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
        ca-certificates=20161130+nmu1 \
        git=1:2.11.0-3+deb9u2 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------------------- #

FROM base AS go

ENV DEBIAN_FRONTEND=noninteractive \
    GOLANG_VERSION="1.10" \
    LD_LIBRARY_PATH=/usr/lib \
    GOPATH=/go \
    GOARCH=amd64 \
    GOOS=linux \
    GOROOT=/usr/lib/go

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
        pkg-config=0.29-4+b1 \
        unzip \
        tar=1.29b-1.1 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

ADD "https://golang.org/dl/go${GOLANG_VERSION}.${GOOS}-${GOARCH}.tar.gz" /tmp/go-pkg.tar.gz
RUN tar -C "${LD_LIBRARY_PATH}" -xzf /tmp/go-pkg.tar.gz && \
    mv "${LD_LIBRARY_PATH}"/go/bin/go /usr/bin/go && \
    mv "${LD_LIBRARY_PATH}"/go/bin/godoc /usr/bin/godoc && \
    mv "${LD_LIBRARY_PATH}"/go/bin/gofmt /usr/bin/gofmt

ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

ADD https://github.com/Masterminds/glide/releases/download/v0.13.1/glide-v0.13.1-linux-amd64.zip /tmp/glide-pkg.zip
RUN unzip -p /tmp/glide-pkg.zip linux-amd64/glide > /usr/bin/glide && \
    chmod +x /usr/bin/glide

RUN go get -u \
        \
        github.com/golang/lint/golint \
        github.com/fzipp/gocyclo \
        github.com/client9/misspell/cmd/misspell

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

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
        pkg-config=0.29-4+b1 \
        libzmq3-dev=4.2.1-4 \
        cmake=3.7.2-1 \
        gcc=4:6.3.0-4 \
        g++=4:6.3.0-4 \
        zlib1g-dev=1:1.2.8.dfsg-5 \
        libmpc-dev=1.0.3-1+b2 \
        libmpfr-dev=3.1.5-1 \
        libgmp-dev=2:6.1.2+dfsg-1 \
        patch=2.7.5-1+b2 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=go /go /go
COPY --from=go /usr/lib/go /usr/lib/go

COPY --from=go /usr/bin/go /usr/bin/go
COPY --from=go /usr/bin/dep /usr/bin/dep
COPY --from=go /usr/bin/gofmt /usr/bin/gofmt
COPY --from=go /usr/bin/godoc /usr/bin/godoc
COPY --from=go /usr/bin/glide /usr/bin/glide

# ---------------------------------------------------------------------------- #
