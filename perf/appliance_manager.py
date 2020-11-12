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
    self.__download()

  def __install(self, file):
    filename = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    progress('installing lake {}'.format(filename))

    (code, result, error) = execute([
      "apt-get", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "-o=Dpkg::Options::=--force-confdef", "-o=Dpkg::Options::=--force-confnew", filename
    ])

    if code != 0:
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    (code, result, error) = execute([
      "systemctl", "-t", "service", "--no-legend"
    ])

    if code != 0:
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    self.services = set([x.split(' ')[0].split('@')[0].split('.service')[0] for x in result.splitlines()])


  def __download(self):
    self.image_version = os.environ.get('IMAGE_VERSION', '')
    self.debian_version = os.environ.get('UNIT_VERSION', '')

    if self.debian_version.startswith('v'):
      self.debian_version = self.debian_version[1:]

    assert self.image_version, 'IMAGE_VERSION not provided'
    assert self.debian_version, 'UNIT_VERSION not provided'

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    self.binary = '{}/packaging/bin/lake_{}_{}.deb'.format(cwd, self.debian_version, self.arch)

    if os.path.exists(self.binary):
      self.__install(self.binary)
      return

    os.makedirs(os.path.dirname(self.binary), exist_ok=True)

    failure = None
    image = 'openbank/lake:{}'.format(self.image_version)
    package = '/opt/artifacts/lake_{}_{}.deb'.format(self.debian_version, self.arch)
    temp = tempfile.NamedTemporaryFile(delete=True)
    try:
      with open(temp.name, 'w') as fd:
        fd.write(str(os.linesep).join([
          'FROM alpine',
          'COPY --from={} {} {}'.format(image, package, self.binary)
        ]))

      image, stream = self.docker.images.build(fileobj=temp, rm=True, pull=False, tag='perf_artifacts-scratch')
      for chunk in stream:
        if not 'stream' in chunk:
          continue
        for line in chunk['stream'].splitlines():
          l = line.strip(os.linesep)
          if not len(l):
            continue
          print(l)

      scratch = self.docker.containers.run('perf_artifacts-scratch', ['/bin/true'], detach=True)

      tar_name = tempfile.NamedTemporaryFile(delete=True)
      with open(tar_name.name, 'wb') as fd:
        bits, stat = scratch.get_archive(self.binary)
        for chunk in bits:
          fd.write(chunk)

      archive = tarfile.TarFile(tar_name.name)
      archive.extract(os.path.basename(self.binary), os.path.dirname(self.binary))
      self.install(self.binary)
      scratch.remove()
    except Exception as ex:
      failure = ex
    finally:
      temp.close()
      try:
        self.docker.images.remove('perf_artifacts-scratch', force=True)
      except:
        pass

    if failure:
      raise failure

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
