#!/usr/bin/env python

#from systemd.common import Unit
from shell.process import execute_shell

import subprocess
import multiprocessing
import string
import threading
import signal
import time
import os

class Lake(object):

  def __init__(self):
    self.reconfigure({
      'METRICS_REFRESHRATE': '1000ms',
      'METRICS_CONTINUOUS': 'false',
    })
    self.stop()

  def __repr__(self):
    return 'Lake()'

  def teardown(self):
    for unit in ['lake-relay', 'lake']:
      execute_shell(['systemctl', 'stop', unit])
      (code, result, error) = execute_shell([
        'journalctl', '-o', 'short-precise', '-u', '{}.service'.format(unit), '--no-pager'
      ])
      if code == 0:
        with open('/tmp/reports/perf-tests/logs/{}.log'.format(unit), 'w') as f:
          f.write(result)

  def restart(self) -> bool:
    (code, result, error) = execute_shell(['systemctl', 'restart', 'lake-relay'])
    if code != 0:
      raise RuntimeError("Failed to restart lake-relay, stdout: {}, stderr: {}".format(result, error))
    if not self.is_healthy:
      raise RuntimeError("Failed to restart lake-relay, stdout: {}, stderr: {}".format(result, error))

  def stop(self) -> bool:
    (code, result, error) = execute_shell(['systemctl', 'stop', 'lake-relay'])
    if code != 0:
      raise RuntimeError("Failed to stop lake-relay, stdout: {}, stderr: {}".format(result, error))

  def start(self) -> bool:
    (code, result, error) = execute_shell(['systemctl', 'start', 'lake-relay'])
    if code != 0:
      raise RuntimeError("Failed to start lake-relay, stdout: {}, stderr: {}".format(result, error))
    if not self.is_healthy:
      raise RuntimeError("Failed to start lake-relay, stdout: {}, stderr: {}".format(result, error))

  def reconfigure(self, params) -> None:
    d = {}

    with open('/etc/init/lake.conf', 'r') as f:
      for line in f:
        (key, val) = line.rstrip().split('=')
        d[key] = val

    for k, v in params.items():
      key = 'LAKE_{}'.format(k)
      if key in d:
        d[key] = v

    with open('/etc/init/lake.conf', 'w') as f:
      f.write('\n'.join("{}={}".format(key, val) for (key,val) in d.items()))

    self.restart()

  @property
  def is_healthy(self) -> bool:
    (code, result, error) = execute_shell([
      'bash',
      '-c',
      'while [ "$(systemctl show -p SubState lake-relay)" != "SubState=running" ]; do sleep 0.5; done;',
    ])

    return code == 0
