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

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
      libzmq5=4.2.1-4 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN groupadd -r lake && useradd --no-log-init -r -g lake lake

USER lake

COPY --chown=lake bin/lake /entrypoint

RUN chmod +x /entrypoint

ENTRYPOINT ["/entrypoint"]
