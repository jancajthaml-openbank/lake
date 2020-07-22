#!/usr/bin/env python3

import docker
from utils import progress, info, print_daemon
from helpers.shell import execute
from systemd.lake import Lake
import platform
import tarfile
import tempfile
import errno
import os


class ApplianceManager(object):

  def get_arch(self):
    return {
      'x86_64': 'amd64',
      'armv7l': 'armhf',
      'armv8': 'arm64'
    }.get(platform.uname().machine, 'amd64')

  def __init__(self):
    self.configure()
    self.arch = self.get_arch()

    self.store = {}
    self.image_version = None
    self.debian_version = None
    self.units = {}
    self.services = []
    self.docker = docker.APIClient(base_url='unix://var/run/docker.sock')

    try:
      os.mkdir("/opt/artifacts")
    except OSError as exc:
      if exc.errno != errno.EEXIST:
        raise
      pass

    self.image_version = os.environ.get('IMAGE_VERSION', '')
    self.debian_version = os.environ.get('UNIT_VERSION', '')

    if self.debian_version.startswith('v'):
      self.debian_version = self.debian_version[1:]

    scratch_docker_cmd = ['FROM alpine']

    image = 'openbank/lake:{}'.format(self.image_version)
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

      (code, result, error) = execute([
        'dpkg', '-c', '/opt/artifacts/lake.deb'
      ])

      if code != 0:
        raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

      self.docker.remove_container(scratch['Id'])
    finally:
      temp.close()
      self.docker.remove_image('perf_artifacts-scratch', force=True)

    progress('installing lake {}'.format(self.image_version))

    (code, result, error) = execute([
      "apt-get", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "-o=Dpkg::Options::=--force-confdef", "-o=Dpkg::Options::=--force-confnew", '/opt/artifacts/lake.deb'
    ])

    if code != 0:
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    (code, result, error) = execute([
      "systemctl", "-t", "service", "--no-legend"
    ])

    if code != 0:
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    self.services = set([x.split(' ')[0].split('@')[0].split('.service')[0] for x in result.splitlines()])

  def __len__(self):
    return sum([len(x) for x in self.units.values()])

  def __getitem__(self, key):
    return self.units.get(str(key), [])

  def __setitem__(self, key, value):
    self.units.setdefault(str(key), []).append(value)

  def __delitem__(self, key):
    if not str(key) in self.units:
      return
    for node in self.units[str(key)]:
      node.teardown()
    del self.units[str(key)]

  def configure(self) -> None:
    options = {
      'LOG_LEVEL': 'INFO',
      'PORT_PULL': '5562',
      'PORT_PUB': '5561',
      'METRICS_OUTPUT': '/opt/lake/metrics',
      'METRICS_REFRESHRATE': '1000ms',
      'METRICS_CONTINUOUS': 'false',
    }

    os.makedirs("/etc/lake/conf.d", exist_ok=True)
    with open('/etc/lake/conf.d/init.conf', 'w') as fd:
      for k, v in sorted(options.items()):
        fd.write('LAKE_{}={}\n'.format(k, v))

  # fixme __iter__
  def items(self) -> list:
    return self.units.items()

  def values(self) -> list:
    return self.units.values()

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
