#!/usr/bin/env python
# -*- coding: utf-8 -*-

import docker

from utils import progress, info, print_daemon

from systemd.lake import Lake

import platform
import tarfile
import tempfile
import errno
import os
import subprocess

class ApplianceManager(object):

  def get_arch(self):
    return {
      'x86_64': 'amd64',
      'armv7l': 'armhf',
      'armv8': 'arm64'
    }.get(platform.uname().machine, 'amd64')

  def __init__(self):
    self.arch = self.get_arch()

    self.store = {}
    self.image_version = None
    self.debian_version = None
    self.units = {}
    self.services = []
    self.docker = docker.APIClient(base_url='unix://var/run/docker.sock')

    DEVNULL = open(os.devnull, 'w')

    try:
      os.mkdir("/opt/artifacts")
    except OSError as exc:
      if exc.errno != errno.EEXIST:
        raise
      pass

    image_version = os.environ.get('UNIT_VERSION', '')
    if image_version.startswith('v'):
      image_version = image_version[1:]

    self.image_version = image_version
    self.debian_version = image_version.replace('-', '+', 1)

    scratch_docker_cmd = ['FROM alpine']

    image = 'openbank/lake:v{}'.format(self.image_version)
    package = 'lake_{}_{}'.format(self.debian_version, self.arch)
    scratch_docker_cmd.append('COPY --from={} /opt/artifacts/{}.deb /opt/artifacts/lake.deb'.format(image, package))

    temp = tempfile.NamedTemporaryFile(delete=True)
    try:
      with open(temp.name, 'w') as f:
        for item in scratch_docker_cmd:
          f.write("%s\n" % item)

      for chunk in self.docker.build(fileobj=temp, rm=True, decode=True, tag='perf_artifacts-scratch'):
        if 'stream' in chunk:
          for line in chunk['stream'].splitlines():
            if len(line):
              print_daemon(line.strip('\r\n'))

      scratch = self.docker.create_container('perf_artifacts-scratch', '/bin/true')

      if scratch['Warnings']:
        raise Exception(scratch['Warnings'])

      tar_name = '/opt/artifacts/lake.tar'
      tar_stream, stat = self.docker.get_archive(scratch['Id'], '/opt/artifacts/lake.deb')
      with open(tar_name, 'wb') as destination:
        total_bytes = 0
        for chunk in tar_stream:
          total_bytes += len(chunk)
          progress('extracting {} {:.2f}%'.format(stat['name'], min(100, 100 * (total_bytes/stat['size']))))
          destination.write(chunk)
      archive = tarfile.TarFile(tar_name)
      archive.extract('lake.deb', '/opt/artifacts')
      os.remove(tar_name)

      try:
        contents = subprocess.check_output(["dpkg", "-c", "/opt/artifacts/{}.deb".format('lake')], stderr=subprocess.STDOUT).decode("utf-8").strip()
      except subprocess.CalledProcessError as e:
        raise Exception(e.output.decode("utf-8").strip())

      self.docker.remove_container(scratch['Id'])
    finally:
      temp.close()
      self.docker.remove_image('perf_artifacts-scratch', force=True)

    progress('installing lake {}'.format(self.image_version))
    subprocess.check_call(["apt-get", "-y", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", '/opt/artifacts/lake.deb'], stdout=DEVNULL, stderr=subprocess.STDOUT)
    info('installed lake {}'.format(self.image_version))

    DEVNULL.close()

    installed = subprocess.check_output(["systemctl", "-t", "service", "--no-legend"], stderr=subprocess.STDOUT).decode("utf-8").strip()
    self.services = set([x.split(' ')[0].split('@')[0].split('.service')[0] for x in installed.splitlines()])

  def __len__(self):
    return sum([len(x) for x in self.units.values()])

  def __getitem__(self, key):
    return self.units.get(str(key), [])

  def __setitem__(self, key, value):
    self.units.setdefault(str(key), []).append(value)

  def __delitem__(self, key):
    # fixme add lock here
    if not str(key) in self.units:
      return

    for node in self.units[str(key)]:
      node.teardown()

    del self.units[str(key)]

  # fixme __iter__
  def items(self) -> list:
    return self.units.items()

  def values(self) -> list:
    return self.units.values()

  def reconfigure(self, params, key=None) -> None:
    if not key:
      for name in list(self.units):
        for node in self[name]:
          node.reconfigure(params)
      return

    for node in self[key]:
      node.reconfigure(params)

  def start(self, key=None) -> None:
    if not key:
      for name in list(self.units):
        for node in self[name]:
          node.start()
      return

    for node in self[key]:
      node.start()

  def stop(self, key=None) -> None:
    if not key:
      for name in list(self.units):
        for node in self[name]:
          node.stop()
      return

    for node in self[key]:
      node.stop()

  def restart(self, key=None) -> None:
    if not key:
      for name in list(self.units):
        for node in self[name]:
          node.restart()
      return

    for node in self[key]:
      node.restart()

  def bootstrap(self) -> None:
    if 'lake' in self.services and not self['lake']:
      self['lake'] = Lake()

  def teardown(self, key=None) -> None:
    if key:
      del self[key]
    else:
      for name in list(self.units):
        del self[name]