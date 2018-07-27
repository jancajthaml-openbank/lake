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

FROM ruby:2.5.0-stretch

MAINTAINER Jan Cajthaml <jan.cajthaml@gmail.com>

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    \
        libzmq3-dev=4.2.1-4 \
    \
    && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN gem install \
    \
        turnip:2.1.1 \
        turnip_formatter:0.5.0 \
        rspec_junit_formatter:0.3.0 \
        rspec-instafail:1.0.0 \
        excon:0.61.0 \
        bigdecimal:1.3.4 \
        byebug:10.0.1 \
        ffi-rzmq:2.0.4

COPY --from=library/docker:18.06 /usr/local/bin/docker /usr/bin/docker

WORKDIR /opt/bbtest

ENTRYPOINT ["rspec", "--require", "/opt/bbtest/spec.rb"]
CMD ["--format", "documentation", "--format", "RspecJunitFormatter", "--out", "junit.xml", "--pattern", "/opt/bbtest/features/*.feature"]
