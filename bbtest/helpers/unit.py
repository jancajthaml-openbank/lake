#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
from openbank_testkit import Shell, Package, Platform
from systemd import journal

class UnitHelper(object):

  @staticmethod
  def default_config():
    return {
      "LOG_LEVEL": "DEBUG",
      "PORT_PULL": "5562",
      "PORT_PUB": "5561",
      "STATSD_ENDPOINT": "127.0.0.1:8125",
    }

  def __init__(self, context):
    self.store = dict()
    self.units = list()
    self.context = context

  def download(self):
    version = os.environ.get('VERSION', '')
    meta = os.environ.get('META', '')

    if version.startswith('v'):
      version = version[1:]

    assert version, 'VERSION not provided'
    assert meta, 'META not provided'

    package = Package('lake')

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    assert package.download(version, meta, '{}/packaging/bin'.format(cwd)), 'unable to download package lake'

    self.binary = '{}/packaging/bin/lake_{}_{}.deb'.format(cwd, version, Platform.arch)

  def configure(self, params = None):
    options = dict()
    options.update(UnitHelper.default_config())
    if params:
      options.update(params)

    os.makedirs('/etc/lake/conf.d', exist_ok=True)
    with open('/etc/lake/conf.d/init.conf', 'w') as fd:
      fd.write(str(os.linesep).join("LAKE_{!s}={!s}".format(k, v) for (k, v) in options.items()))

  def __fetch_logs(self, unit=None):
    reader = journal.Reader()
    reader.this_boot()
    reader.log_level(journal.LOG_DEBUG)
    if unit:
      reader.add_match(_SYSTEMD_UNIT=unit)
    for entry in reader:
      yield entry['MESSAGE']

  def collect_logs(self):
    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    logs_dir = '{}/reports/blackbox-tests/logs'.format(cwd)
    os.makedirs(logs_dir, exist_ok=True)
    with open('{}/journal.log'.format(logs_dir), 'w') as fd:
      for line in self.__fetch_logs():
        fd.write(line)
        fd.write(os.linesep)
    for unit in set(self.__get_systemd_units() + self.units):      
      with open('{}/{}.log'.format(logs_dir, unit), 'w') as fd:
        for line in self.__fetch_logs(unit):
          fd.write(line)
          fd.write(os.linesep)

  def teardown(self):
    self.collect_logs()
    for unit in self.__get_systemd_units():
      Shell.run(['systemctl', 'stop', unit])
    self.collect_logs()

  def __get_systemd_units(self):
    (code, result, error) = Shell.run(['systemctl', 'list-units', '--all', '--no-legend'])
    result = [item.replace('*', '').strip().split(' ')[0].strip() for item in result.split(os.linesep)]
    result = [item for item in result if "lake" in item and not item.endswith('unit.slice')]
    return result
