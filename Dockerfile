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
    LD_LIBRARY_PATH=/usr/lib

RUN apt-get -y update && apt-get -y upgrade && apt-get clean && \
    apt-get -y install --allow-downgrades --no-install-recommends \
    \
      lsb-release=9.20161125 \
      curl=7.52.1-5+deb9u5 \
      git=1:2.11.0-3+deb9u2 \
      cron=3.0pl1-128+deb9u1 \
      at=3.1.20-3 \
      logrotate=3.11.0-0.1 \
      rsyslog=8.24.0-1 \
      unattended-upgrades=0.93.1+nmu1 \
      ssmtp=2.64-8+b2  \
      lsof=4.89+dfsg-0.1 \
      procps=2:3.3.12-3 \
      initscripts=2.88dsf-59.9 \
      libsystemd0=232-25+deb9u2 \
      libudev1=232-25+deb9u2 \
      systemd=232-25+deb9u2 \
      systemd-sysv=232-25+deb9u2 \
      sysvinit-utils=2.88dsf-59.9 \
      udev=232-25+deb9u2 \
      util-linux=2.29.2-1+deb9u1 \
    \
      libzmq5=4.2.1-4 \
    && \
    apt-get clean && \
    sed -i '/imklog/{s/^/#/}' /etc/rsyslog.conf

COPY pkg /tmp

RUN find /tmp -type f -name 'lake_*_amd64.deb' -exec dpkg --install \{\} \; -exec rm -f \{\} \; && \
    systemctl unmask lake && \
    systemctl enable lake

STOPSIGNAL SIGTERM

ENTRYPOINT ["/lib/systemd/systemd"]
