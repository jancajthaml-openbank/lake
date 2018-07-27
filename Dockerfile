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

FROM debian:stretch AS base

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends apt-utils

# ---------------------------------------------------------------------------- #

FROM base

MAINTAINER Jan Cajthaml <jan.cajthaml@gmail.com>

RUN apt-get -y install --allow-downgrades --no-install-recommends \
    \
      rsyslog=8.24.0-1 \
      libsystemd0>=232-25 \
      systemd>=232-25 \
    && \
    apt-get clean && \
    sed -i '/imklog/{s/^/#/}' /etc/rsyslog.conf

COPY packaging/params.conf /etc/init/lake.conf

COPY packaging/bin /tmp/packages

RUN find /tmp/packages -type f -name 'lake_*_amd64.deb' -exec apt-get -y install --no-install-recommends -f \{\} \; && \
    rm -rf /tmp/packages

STOPSIGNAL SIGTERM

ENTRYPOINT ["/lib/systemd/systemd"]
