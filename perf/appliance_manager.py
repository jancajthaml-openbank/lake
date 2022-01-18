#!/usr/bin/env python3

from utils import info, print_daemon
from openbank_testkit import Shell, Package, Platform
from unit.lake import Lake
import platform
import tarfile
import tempfile
import errno
import os


class ApplianceManager(object):

  def __init__(self):
    self.store = {}
    self.units = {}
    self.services = []

  def __install(self):
    version = os.environ.get('VERSION', '')
    if version.startswith('v'):
      version = version[1:]

    assert version, 'VERSION not provided'

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    filename = '{}/packaging/bin/lake_{}_{}.deb'.format(cwd, version, Platform.arch)

    (code, result, error) = Shell.run([
      "apt-get", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "-o=Dpkg::Options::=--force-confdef", "-o=Dpkg::Options::=--force-confnew", filename
    ])

    if code != 'OK':
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    (code, result, error) = Shell.run([
      "systemctl", "-t", "service", "--all", "--no-legend"
    ])

    if code != 'OK':
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    self.services = set([item.replace('*', '').strip().split(' ')[0].split('@')[0].split('.service')[0] for item in result.split(os.linesep)])

  def __download(self):
    version = os.environ.get('VERSION', '')
    meta = os.environ.get('META', '')

    if version.startswith('v'):
      version = version[1:]

    assert version, 'VERSION not provided'
    assert meta, 'META not provided'

    package = Package('lake')

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    assert package.download(version, meta, '{}/packaging/bin'.format(cwd)), 'unable to download package lake'

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

  def __configure(self) -> None:
    options = {
      'LOG_LEVEL': 'DEBUG',
      'PORT_PULL': '5562',
      'PORT_PUB': '5561',
      'STATSD_ENDPOINT': '127.0.0.1:8125',
    }

    os.makedirs("/etc/lake/conf.d", exist_ok=True)
    with open('/etc/lake/conf.d/init.conf', 'w') as fd:
      for k, v in sorted(options.items()):
        fd.write('LAKE_{}={}\n'.format(k, v))

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
    self.__configure()
    self.__download()
    self.__install()

    if 'lake' in self.services and not self['lake']:
      self['lake'] = Lake()

  def teardown(self, key=None) -> None:
    if key:
      del self[key]
    else:
      for name in list(self.units):
        del self[name]
