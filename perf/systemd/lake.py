#!/usr/bin/env python3


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
    self.start()

  def __repr__(self):
    return 'Lake()'

  def teardown(self):
    for unit in ['lake-relay', 'lake']:
      execute_shell(['systemctl', 'stop', unit])
      (code, result, error) = execute_shell([
        'journalctl', '-o', 'short-precise', '-u', '{}.service'.format(unit), '--no-pager'
      ], True)
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

  @property
  def is_healthy(self) -> bool:
    (code, result, error) = execute_shell([
      'bash',
      '-c',
      'while [ "$(systemctl show -p SubState lake-relay)" != "SubState=running" ]; do sleep 0.5; done;',
    ])
    return code == 0
