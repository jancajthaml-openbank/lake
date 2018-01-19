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

FROM library/ruby:2

MAINTAINER Jan Cajthaml <jan.cajthaml@gmail.com>

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        netcat-openbsd \
        bsdmainutils \
        libzmq3-dev && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN gem install \
    turnip:2.1.1 \
    excon \
    bigdecimal \
    turnip_formatter:0.5.0 \
    rspec_junit_formatter \
    byebug \
    rspec-instafail \
    ffi-rzmq:2.0.4

RUN curl -L https://download.docker.com/linux/static/stable/x86_64/docker-17.09.0-ce.tgz | \
    tar -xzvf - --strip-components=1 -C /usr/bin docker/docker && \
    chmod a+x /usr/bin/docker

RUN curl -Lo /usr/local/bin/docker-compose https://github.com/docker/compose/releases/download/1.17.1/docker-compose-Linux-x86_64 && \
    chmod a+x /usr/local/bin/docker-compose

WORKDIR /opt/bbtest

ENTRYPOINT ["rspec", "--require", "/opt/bbtest/spec.rb"]
CMD ["--pattern", "/opt/bbtest/features/*.feature"]
