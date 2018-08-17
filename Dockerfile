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

FROM debian:stretch

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LAKE_VERSION=1.0.2

MAINTAINER Jan Cajthaml <jan.cajthaml@gmail.com>

RUN apt-get -y update && \
    apt-get -y upgrade && \
    apt-get clean && \
    apt-get -y install \
      apt-utils \
      lsb-release \
      curl \
      git \
      cron \
      at \
      logrotate \
      rsyslog \
      unattended-upgrades \
      ssmtp \
      lsof \
      procps \
      initscripts \
      libsystemd0 \
      libudev1 \
      systemd \
      sysvinit-utils \
      udev \
      util-linux && \
  apt-get clean && \
  sed -i '/imklog/{s/^/#/}' /etc/rsyslog.conf

RUN echo "root:Docker!" | chpasswd

RUN cd /lib/systemd/system/sysinit.target.wants/ && \
    ls | grep -v systemd-tmpfiles-setup.service | xargs rm -f && \
    rm -f /lib/systemd/system/sockets.target.wants/*udev* && \
    systemctl mask -- \
      tmp.mount \
      etc-hostname.mount \
      etc-hosts.mount \
      etc-resolv.conf.mount \
      -.mount \
      swap.target \
      getty.target \
      getty-static.service \
      dev-mqueue.mount \
      cgproxy.service \
      systemd-tmpfiles-setup-dev.service \
      systemd-remount-fs.service \
      systemd-ask-password-wall.path \
      systemd-logind.service && \
    systemctl set-default multi-user.target || :

RUN sed -ri /etc/systemd/journald.conf -e 's!^#?Storage=.*!Storage=volatile!'

COPY packaging/bin /tmp/packages

RUN find /tmp/packages -type f -name 'lake_*_amd64.deb' -exec \
    apt-get -y install --no-install-recommends \
    -f \{\} \; && rm -rf /tmp/packages

VOLUME [ "/sys/fs/cgroup", "/run", "/run/lock", "/tmp" ]

ENTRYPOINT ["/lib/systemd/systemd"]

